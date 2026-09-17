package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/errors"
	"zhizhang-server/internal/pkg/response"
)

// --- Request/Response DTOs ---

// LoginReq 登录请求
type LoginReq struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=6,max=64"`
}

// LoginResp 登录响应
type LoginResp struct {
	AccessToken      string   `json:"accessToken"`
	RefreshToken     string   `json:"refreshToken"`
	AccessExpiresIn  int      `json:"accessExpiresIn"`  // 访问令牌有效期（秒）
	RefreshExpiresIn int      `json:"refreshExpiresIn"` // 刷新令牌有效期（秒）
	User             UserInfo `json:"user"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	RoleID   uint   `json:"roleId"`
	DeptID   uint   `json:"deptId"`
}

// RefreshReq 刷新令牌请求
type RefreshReq struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// issueTokenPair 为员工签发 access/refresh 令牌对（密码登录、刷新、企业微信登录共用）
func issueTokenPair(emp *model.Employee) (*LoginResp, error) {
	accessToken, err := middleware.GenerateToken(emp.ID, emp.Username, emp.RoleID, emp.DeptID, emp.CompanyID)
	if err != nil {
		return nil, err
	}
	refreshToken, err := middleware.GenerateRefreshToken(emp.ID)
	if err != nil {
		return nil, err
	}
	return &LoginResp{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		AccessExpiresIn:  middleware.GetAccessTTL(),
		RefreshExpiresIn: middleware.GetRefreshTTL(),
		User: UserInfo{
			ID:       emp.ID,
			Username: emp.Username,
			Name:     emp.Name,
			Phone:    emp.Phone,
			RoleID:   emp.RoleID,
			DeptID:   emp.DeptID,
		},
	}, nil
}

// AuthHandler 认证处理器
type AuthHandler struct {
	db *gorm.DB
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

// RegisterRoutes 注册路由
func (h *AuthHandler) RegisterRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.Refresh)
		auth.POST("/logout", middleware.JWTMiddleware(), h.Logout)
		auth.GET("/me", middleware.JWTMiddleware(), h.Me)
	}
}

// Login 登录
// @Summary 用户登录
// @Description 使用用户名和密码登录，返回 JWT 令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param body body LoginReq true "登录参数"
// @Success 200 {object} response.Response{data=LoginResp} "登录成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "用户名或密码错误"
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 查找用户
	var emp model.Employee
	if err := h.db.Where("username = ? AND status = 1", req.Username).First(&emp).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Fail(c, response.CodeUnauthorized, "用户名或密码错误")
			return
		}
		log.Error().Err(err).Str("username", req.Username).Msg("login query failed")
		response.ServerError(c, "登录失败")
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(emp.Password), []byte(req.Password)); err != nil {
		response.Fail(c, response.CodeUnauthorized, "用户名或密码错误")
		return
	}

	// 密码登录仅超级管理员可用，其余角色必须走企业微信登录
	var role model.Role
	if err := h.db.First(&role, emp.RoleID).Error; err != nil || role.Code != "super_admin" {
		response.Fail(c, response.CodeNeedWecom, "当前账号请使用企业微信登录")
		return
	}

	// 更新登录时间
	now := time.Now()
	emp.LastLoginAt = &now
	emp.LastLoginIP = c.ClientIP()
	h.db.Save(&emp)

	resp, err := issueTokenPair(&emp)
	if err != nil {
		log.Error().Err(err).Msg("issue token pair failed")
		response.ServerError(c, "令牌生成失败")
		return
	}

	response.Ok(c, resp)
}

// Refresh 刷新访问令牌
// @Summary 刷新访问令牌
// @Description 使用 RefreshToken 获取新的 AccessToken（同时轮换 RefreshToken）
// @Tags 认证
// @Accept json
// @Produce json
// @Param body body RefreshReq true "刷新参数"
// @Success 200 {object} response.Response{data=LoginResp} "刷新成功"
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	// 解析并验证 refresh token
	userID, err := middleware.ParseRefreshToken(req.RefreshToken)
	if err != nil {
		if err == errors.ErrExpiredToken {
			response.Fail(c, response.CodeUnauthorized, "刷新令牌已过期，请重新登录")
			return
		}
		response.Fail(c, response.CodeUnauthorized, "无效的刷新令牌，请重新登录")
		return
	}

	// 校验用户仍然存在且可用
	var emp model.Employee
	if err := h.db.Where("id = ? AND status = 1", userID).First(&emp).Error; err != nil {
		response.Fail(c, response.CodeUnauthorized, "用户不可用，请重新登录")
		return
	}

	// 签发新的令牌对
	resp, err := issueTokenPair(&emp)
	if err != nil {
		log.Error().Err(err).Msg("issue token pair failed")
		response.ServerError(c, "令牌生成失败")
		return
	}

	response.Ok(c, resp)
}

// Logout 登出
// @Summary 用户登出
// @Description 清除当前登录状态
// @Tags 认证
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response "登出成功"
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// TODO: 将令牌加入黑名单（Redis）
	response.OkWithMessage(c, "登出成功", nil)
}

// Me 获取当前用户信息
// @Summary 获取当前用户信息
// @Description 根据 JWT 获取登录用户信息
// @Tags 认证
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=UserInfo} "获取成功"
// @Router /api/v1/auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	var emp model.Employee
	if err := h.db.First(&emp, userID).Error; err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	response.Ok(c, UserInfo{
		ID:       emp.ID,
		Username: emp.Username,
		Name:     emp.Name,
		Phone:    emp.Phone,
		RoleID:   emp.RoleID,
		DeptID:   emp.DeptID,
	})
}

// --- Seed Data ---

// InitSeedData 初始化种子数据
func InitSeedData(db *gorm.DB) error {
	// 检查是否已有数据
	var count int64
	db.Model(&model.Company{}).Count(&count)
	if count > 0 {
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		// 创建默认公司
		company := model.Company{
			Name: "默认公司",
			Code: "DEFAULT",
		}
		if err := tx.Create(&company).Error; err != nil {
			return err
		}

		// 创建默认部门
		dept := model.Department{
			BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: company.ID},
			Name:                 "总部",
			Code:                 "HQ",
		}
		if err := tx.Create(&dept).Error; err != nil {
			return err
		}

		// 创建默认角色
		role := model.Role{
			BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: company.ID},
			Name:                 "超级管理员",
			Code:                 "super_admin",
		}
		if err := tx.Create(&role).Error; err != nil {
			return err
		}

		// 创建默认用户（密码: admin123）
		hashedPwd, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		admin := model.Employee{
			BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: company.ID},
			DeptID:               dept.ID,
			RoleID:               role.ID,
			Username:             "admin",
			Password:             string(hashedPwd),
			Name:                 "管理员",
			Phone:                "13800138000",
			Status:               1,
		}
		if err := tx.Create(&admin).Error; err != nil {
			return err
		}

		log.Info().Msg("seed data created: admin/admin123")
		return nil
	})
}
