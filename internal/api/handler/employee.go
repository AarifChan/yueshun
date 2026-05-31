package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/response"
)

// EmployeeHandler 职员管理处理器
type EmployeeHandler struct {
	db *gorm.DB
}

func NewEmployeeHandler(db *gorm.DB) *EmployeeHandler {
	return &EmployeeHandler{db: db}
}

func (h *EmployeeHandler) RegisterRoutes(r *gin.RouterGroup) {
	emp := r.Group("/employees")
	{
		emp.GET("", h.ListEmployees)
		emp.POST("", h.CreateEmployee)
		emp.PUT("/:id", h.UpdateEmployee)
		emp.DELETE("/:id", h.DeleteEmployee)
		emp.PUT("/:id/status", h.UpdateEmployeeStatus)
		emp.PUT("/:id/reset-password", h.ResetPassword)
	}
}

type EmployeeListReq struct {
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"pageSize,default=20"`
	Keyword    string `form:"keyword"`
	DeptID     uint   `form:"deptId"`
	RoleID     uint   `form:"roleId"`
	Status     int8   `form:"status"`
}

type CreateEmployeeReq struct {
	DeptID   uint   `json:"deptId" binding:"required"`
	RoleID   uint   `json:"roleId" binding:"required"`
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=6,max=64"`
	Name     string `json:"name" binding:"required,max=64"`
	Phone    string `json:"phone" binding:"max=20"`
	Email    string `json:"email" binding:"max=128"`
	Status   int8   `json:"status" binding:"oneof=0 1"`
}

type UpdateEmployeeReq struct {
	DeptID   uint   `json:"deptId"`
	RoleID   uint   `json:"roleId"`
	Name     string `json:"name" binding:"max=64"`
	Phone    string `json:"phone" binding:"max=20"`
	Email    string `json:"email" binding:"max=128"`
	Status   int8   `json:"status" binding:"oneof=0 1"`
}

func (h *EmployeeHandler) ListEmployees(c *gin.Context) {
	var req EmployeeListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.Employee{}).Where("company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("username LIKE ? OR name LIKE ? OR phone LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if req.DeptID > 0 {
		query = query.Where("dept_id = ?", req.DeptID)
	}
	if req.RoleID > 0 {
		query = query.Where("role_id = ?", req.RoleID)
	}
	if req.Status == 0 || req.Status == 1 {
		query = query.Where("status = ?", req.Status)
	}

	var total int64
	query.Count(&total)

	var list []model.Employee
	query.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *EmployeeHandler) CreateEmployee(c *gin.Context) {
	var req CreateEmployeeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	// 验证部门存在
	var dept model.Department
	if err := h.db.Where("id = ? AND company_id = ?", req.DeptID, companyID).First(&dept).Error; err != nil {
		response.BadRequest(c, "部门不存在")
		return
	}
	// 验证角色存在
	var role model.Role
	if err := h.db.Where("id = ? AND company_id = ?", req.RoleID, companyID).First(&role).Error; err != nil {
		response.BadRequest(c, "角色不存在")
		return
	}
	// 检查用户名是否重复
	var count int64
	h.db.Model(&model.Employee{}).Where("company_id = ? AND username = ?", companyID, req.Username).Count(&count)
	if count > 0 {
		response.Fail(c, response.CodeBadRequest, "用户名已存在")
		return
	}

	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	emp := model.Employee{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		DeptID:               req.DeptID,
		RoleID:               req.RoleID,
		Username:             req.Username,
		Password:             string(hashedPwd),
		Name:                 req.Name,
		Phone:                req.Phone,
		Email:                req.Email,
		Status:               req.Status,
	}
	if err := h.db.Create(&emp).Error; err != nil {
		log.Error().Err(err).Msg("create employee failed")
		response.ServerError(c, "创建失败")
		return
	}
	response.Ok(c, emp)
}

func (h *EmployeeHandler) UpdateEmployee(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req UpdateEmployeeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var emp model.Employee
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&emp).Error; err != nil {
		response.NotFound(c, "职员不存在")
		return
	}

	if req.DeptID > 0 {
		var dept model.Department
		if err := h.db.Where("id = ? AND company_id = ?", req.DeptID, companyID).First(&dept).Error; err != nil {
			response.BadRequest(c, "部门不存在")
			return
		}
		emp.DeptID = req.DeptID
	}
	if req.RoleID > 0 {
		var role model.Role
		if err := h.db.Where("id = ? AND company_id = ?", req.RoleID, companyID).First(&role).Error; err != nil {
			response.BadRequest(c, "角色不存在")
			return
		}
		emp.RoleID = req.RoleID
	}
	if req.Name != "" {
		emp.Name = req.Name
	}
	if req.Phone != "" {
		emp.Phone = req.Phone
	}
	if req.Email != "" {
		emp.Email = req.Email
	}
	emp.Status = req.Status

	if err := h.db.Save(&emp).Error; err != nil {
		log.Error().Err(err).Msg("update employee failed")
		response.ServerError(c, "更新失败")
		return
	}
	response.Ok(c, emp)
}

func (h *EmployeeHandler) DeleteEmployee(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.Employee{}).Error; err != nil {
		log.Error().Err(err).Msg("delete employee failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

func (h *EmployeeHandler) UpdateEmployeeStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req struct {
		Status int8 `json:"status" binding:"oneof=0 1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	if err := h.db.Model(&model.Employee{}).Where("id = ? AND company_id = ?", id, companyID).Update("status", req.Status).Error; err != nil {
		log.Error().Err(err).Msg("update employee status failed")
		response.ServerError(c, "更新失败")
		return
	}
	response.OkWithMessage(c, "更新成功", nil)
}

func (h *EmployeeHandler) ResetPassword(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req struct {
		Password string `json:"password" binding:"required,min=6,max=64"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err := h.db.Model(&model.Employee{}).Where("id = ? AND company_id = ?", id, companyID).Updates(map[string]interface{}{
		"password":   string(hashedPwd),
		"updated_at": time.Now(),
	}).Error; err != nil {
		log.Error().Err(err).Msg("reset password failed")
		response.ServerError(c, "重置密码失败")
		return
	}
	response.OkWithMessage(c, "密码重置成功", nil)
}
