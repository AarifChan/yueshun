package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/response"
)

// PermissionHandler 权限管理处理器
type PermissionHandler struct {
	db *gorm.DB
}

func NewPermissionHandler(db *gorm.DB) *PermissionHandler {
	return &PermissionHandler{db: db}
}

func (h *PermissionHandler) RegisterRoutes(r *gin.RouterGroup) {
	perm := r.Group("/permissions")
	{
		perm.GET("", h.ListPermissions)
		perm.GET("/tree", h.GetPermissionTree)
		perm.POST("", h.CreatePermission)
		perm.PUT("/:id", h.UpdatePermission)
		perm.DELETE("/:id", h.DeletePermission)
	}
}

type PermissionListReq struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"pageSize,default=20"`
	Keyword  string `form:"keyword"`
	Type     string `form:"type"`
	Status   int8   `form:"status"`
}

type CreatePermissionReq struct {
	Name     string `json:"name" binding:"required,max=64"`
	Code     string `json:"code" binding:"required,max=128"`
	Type     string `json:"type" binding:"required,oneof=menu button api data"`
	ParentID uint   `json:"parentId"`
	Path     string `json:"path" binding:"max=255"`
	Sort     int    `json:"sort"`
	Status   int8   `json:"status" binding:"oneof=0 1"`
}

type UpdatePermissionReq struct {
	Name     string `json:"name" binding:"max=64"`
	Code     string `json:"code" binding:"max=128"`
	Type     string `json:"type" binding:"oneof=menu button api data"`
	ParentID uint   `json:"parentId"`
	Path     string `json:"path" binding:"max=255"`
	Sort     int    `json:"sort"`
	Status   int8   `json:"status" binding:"oneof=0 1"`
}

func (h *PermissionHandler) ListPermissions(c *gin.Context) {
	var req PermissionListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	query := h.db.Model(&model.Permission{})
	if req.Keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}
	if req.Status == 0 || req.Status == 1 {
		query = query.Where("status = ?", req.Status)
	}

	var total int64
	query.Count(&total)

	var list []model.Permission
	query.Order("sort ASC, created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *PermissionHandler) GetPermissionTree(c *gin.Context) {
	var list []model.Permission
	h.db.Order("sort ASC, created_at DESC").Find(&list)

	tree := buildPermTree(list, 0)
	response.Ok(c, tree)
}

type PermTreeNode struct {
	model.Permission
	Children []PermTreeNode `json:"children"`
}

func buildPermTree(perms []model.Permission, parentID uint) []PermTreeNode {
	var result []PermTreeNode
	for _, p := range perms {
		if p.ParentID == parentID {
			node := PermTreeNode{
				Permission: p,
				Children:   buildPermTree(perms, p.ID),
			}
			result = append(result, node)
		}
	}
	return result
}

func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	var req CreatePermissionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	// 检查 code 是否重复
	var count int64
	h.db.Model(&model.Permission{}).Where("code = ?", req.Code).Count(&count)
	if count > 0 {
		response.Fail(c, response.CodeBadRequest, "权限编码已存在")
		return
	}

	// 验证父权限存在
	if req.ParentID > 0 {
		var parent model.Permission
		if err := h.db.First(&parent, req.ParentID).Error; err != nil {
			response.BadRequest(c, "父权限不存在")
			return
		}
	}

	perm := model.Permission{
		Name:     req.Name,
		Code:     req.Code,
		Type:     req.Type,
		ParentID: req.ParentID,
		Path:     req.Path,
		Sort:     req.Sort,
		Status:   req.Status,
	}
	if err := h.db.Create(&perm).Error; err != nil {
		log.Error().Err(err).Msg("create permission failed")
		response.ServerError(c, "创建失败")
		return
	}
	response.Ok(c, perm)
}

func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req UpdatePermissionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	var perm model.Permission
	if err := h.db.First(&perm, id).Error; err != nil {
		response.NotFound(c, "权限不存在")
		return
	}

	// 验证新的 parentId 不会形成循环
	if req.ParentID > 0 {
		if uint(id) == req.ParentID {
			response.BadRequest(c, "不能将自身设为父权限")
			return
		}
		var parent model.Permission
		if err := h.db.First(&parent, req.ParentID).Error; err != nil {
			response.BadRequest(c, "父权限不存在")
			return
		}
	}

	if req.Name != "" {
		perm.Name = req.Name
	}
	if req.Code != "" {
		perm.Code = req.Code
	}
	if req.Type != "" {
		perm.Type = req.Type
	}
	perm.ParentID = req.ParentID
	if req.Path != "" {
		perm.Path = req.Path
	}
	perm.Sort = req.Sort
	perm.Status = req.Status

	if err := h.db.Save(&perm).Error; err != nil {
		log.Error().Err(err).Msg("update permission failed")
		response.ServerError(c, "更新失败")
		return
	}
	response.Ok(c, perm)
}

func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}

	// 检查是否有子权限
	var childCount int64
	h.db.Model(&model.Permission{}).Where("parent_id = ?", id).Count(&childCount)
	if childCount > 0 {
		response.Fail(c, response.CodeBadRequest, "该权限下存在子权限，无法删除")
		return
	}

	// 检查是否被角色使用
	var roleCount int64
	h.db.Model(&model.RolePermission{}).Where("permission_id = ?", id).Count(&roleCount)
	if roleCount > 0 {
		response.Fail(c, response.CodeBadRequest, "该权限已被角色使用，无法删除")
		return
	}

	if err := h.db.Delete(&model.Permission{}, id).Error; err != nil {
		log.Error().Err(err).Msg("delete permission failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}
