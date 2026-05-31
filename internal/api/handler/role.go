package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/response"
)

// RoleHandler 角色管理处理器
type RoleHandler struct {
	db *gorm.DB
}

func NewRoleHandler(db *gorm.DB) *RoleHandler {
	return &RoleHandler{db: db}
}

func (h *RoleHandler) RegisterRoutes(r *gin.RouterGroup) {
	role := r.Group("/roles")
	{
		role.GET("", h.ListRoles)
		role.POST("", h.CreateRole)
		role.PUT("/:id", h.UpdateRole)
		role.DELETE("/:id", h.DeleteRole)
		role.GET("/:id/permissions", h.GetRolePermissions)
		role.PUT("/:id/permissions", h.SetRolePermissions)
	}
}

type RoleListReq struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"pageSize,default=20"`
	Keyword  string `form:"keyword"`
	Status   int8   `form:"status"`
}

type CreateRoleReq struct {
	Name        string `json:"name" binding:"required,max=64"`
	Code        string `json:"code" binding:"required,max=64"`
	Description string `json:"description" binding:"max=255"`
	Status      int8   `json:"status" binding:"oneof=0 1"`
}

type UpdateRoleReq struct {
	Name        string `json:"name" binding:"max=64"`
	Code        string `json:"code" binding:"max=64"`
	Description string `json:"description" binding:"max=255"`
	Status      int8   `json:"status" binding:"oneof=0 1"`
}

func (h *RoleHandler) ListRoles(c *gin.Context) {
	var req RoleListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.Role{}).Where("company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if req.Status == 0 || req.Status == 1 {
		query = query.Where("status = ?", req.Status)
	}

	var total int64
	query.Count(&total)

	var list []model.Role
	query.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req CreateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	// 检查 code 是否重复
	var count int64
	h.db.Model(&model.Role{}).Where("company_id = ? AND code = ?", companyID, req.Code).Count(&count)
	if count > 0 {
		response.Fail(c, response.CodeBadRequest, "角色编码已存在")
		return
	}

	role := model.Role{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		Name:                 req.Name,
		Code:                 req.Code,
		Description:          req.Description,
		Status:               req.Status,
	}
	if err := h.db.Create(&role).Error; err != nil {
		log.Error().Err(err).Msg("create role failed")
		response.ServerError(c, "创建失败")
		return
	}
	response.Ok(c, role)
}

func (h *RoleHandler) UpdateRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req UpdateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var role model.Role
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&role).Error; err != nil {
		response.NotFound(c, "角色不存在")
		return
	}

	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Code != "" {
		role.Code = req.Code
	}
	role.Description = req.Description
	role.Status = req.Status

	if err := h.db.Save(&role).Error; err != nil {
		log.Error().Err(err).Msg("update role failed")
		response.ServerError(c, "更新失败")
		return
	}
	response.Ok(c, role)
}

func (h *RoleHandler) DeleteRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	// 检查是否有职员使用此角色
	var count int64
	h.db.Model(&model.Employee{}).Where("role_id = ?", id).Count(&count)
	if count > 0 {
		response.Fail(c, response.CodeBadRequest, "该角色下存在职员，无法删除")
		return
	}

	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.Role{}).Error; err != nil {
		log.Error().Err(err).Msg("delete role failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

func (h *RoleHandler) GetRolePermissions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var role model.Role
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&role).Error; err != nil {
		response.NotFound(c, "角色不存在")
		return
	}

	var permissionIDs []uint
	h.db.Model(&model.RolePermission{}).Where("role_id = ?", id).Pluck("permission_id", &permissionIDs)

	response.Ok(c, gin.H{
		"roleID":        role.ID,
		"roleName":      role.Name,
		"permissionIDs": permissionIDs,
	})
}

func (h *RoleHandler) SetRolePermissions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req struct {
		PermissionIDs []uint `json:"permissionIDs" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var role model.Role
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&role).Error; err != nil {
		response.NotFound(c, "角色不存在")
		return
	}

	// 验证所有权限 ID 存在
	if len(req.PermissionIDs) > 0 {
		var count int64
		h.db.Model(&model.Permission{}).Where("id IN ?", req.PermissionIDs).Count(&count)
		if int(count) != len(req.PermissionIDs) {
			response.BadRequest(c, "存在无效的权限ID")
			return
		}
	}

	// 事务：删除旧关联，添加新关联
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", id).Delete(&model.RolePermission{}).Error; err != nil {
			return err
		}
		for _, pid := range req.PermissionIDs {
			if err := tx.Create(&model.RolePermission{RoleID: uint(id), PermissionID: pid}).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("set role permissions failed")
		response.ServerError(c, "设置权限失败")
		return
	}

	response.OkWithMessage(c, "权限设置成功", nil)
}
