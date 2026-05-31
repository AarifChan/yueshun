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

// DepartmentHandler 部门管理处理器
type DepartmentHandler struct {
	db *gorm.DB
}

func NewDepartmentHandler(db *gorm.DB) *DepartmentHandler {
	return &DepartmentHandler{db: db}
}

func (h *DepartmentHandler) RegisterRoutes(r *gin.RouterGroup) {
	dept := r.Group("/departments")
	{
		dept.GET("", h.ListDepartments)
		dept.GET("/tree", h.GetDepartmentTree)
		dept.POST("", h.CreateDepartment)
		dept.PUT("/:id", h.UpdateDepartment)
		dept.DELETE("/:id", h.DeleteDepartment)
	}
}

type DepartmentListReq struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"pageSize,default=20"`
	Keyword  string `form:"keyword"`
}

type CreateDepartmentReq struct {
	ParentID  uint   `json:"parentId"`
	Name      string `json:"name" binding:"required,max=64"`
	Code      string `json:"code" binding:"max=64"`
	ManagerID uint   `json:"managerId"`
	Sort      int    `json:"sort"`
	Status    int8   `json:"status" binding:"oneof=0 1"`
}

type UpdateDepartmentReq struct {
	ParentID  uint   `json:"parentId"`
	Name      string `json:"name" binding:"max=64"`
	Code      string `json:"code" binding:"max=64"`
	ManagerID uint   `json:"managerId"`
	Sort      int    `json:"sort"`
	Status    int8   `json:"status" binding:"oneof=0 1"`
}

func (h *DepartmentHandler) ListDepartments(c *gin.Context) {
	var req DepartmentListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.Department{}).Where("company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	var total int64
	query.Count(&total)

	var list []model.Department
	query.Order("sort ASC, created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *DepartmentHandler) GetDepartmentTree(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	var list []model.Department
	h.db.Where("company_id = ?", companyID).Order("sort ASC, created_at DESC").Find(&list)

	// 构建树形结构
	tree := buildDeptTree(list, 0)
	response.Ok(c, tree)
}

type DeptTreeNode struct {
	model.Department
	Children []DeptTreeNode `json:"children"`
}

func buildDeptTree(depts []model.Department, parentID uint) []DeptTreeNode {
	var result []DeptTreeNode
	for _, d := range depts {
		if d.ParentID == parentID {
			node := DeptTreeNode{
				Department: d,
				Children:   buildDeptTree(depts, d.ID),
			}
			result = append(result, node)
		}
	}
	return result
}

func (h *DepartmentHandler) CreateDepartment(c *gin.Context) {
	var req CreateDepartmentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	// 如果指定了 parentId，验证父部门存在
	if req.ParentID > 0 {
		var parent model.Department
		if err := h.db.Where("id = ? AND company_id = ?", req.ParentID, companyID).First(&parent).Error; err != nil {
			response.BadRequest(c, "父部门不存在")
			return
		}
	}

	dept := model.Department{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		ParentID:             req.ParentID,
		Name:                 req.Name,
		Code:                 req.Code,
		ManagerID:            req.ManagerID,
		Sort:                 req.Sort,
		Status:               req.Status,
	}
	if err := h.db.Create(&dept).Error; err != nil {
		log.Error().Err(err).Msg("create department failed")
		response.ServerError(c, "创建失败")
		return
	}
	response.Ok(c, dept)
}

func (h *DepartmentHandler) UpdateDepartment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req UpdateDepartmentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var dept model.Department
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&dept).Error; err != nil {
		response.NotFound(c, "部门不存在")
		return
	}

	// 验证新的 parentId 不会形成循环
	if req.ParentID > 0 {
		if uint(id) == req.ParentID {
			response.BadRequest(c, "不能将自身设为父部门")
			return
		}
		var parent model.Department
		if err := h.db.Where("id = ? AND company_id = ?", req.ParentID, companyID).First(&parent).Error; err != nil {
			response.BadRequest(c, "父部门不存在")
			return
		}
	}

	if req.Name != "" {
		dept.Name = req.Name
	}
	if req.Code != "" {
		dept.Code = req.Code
	}
	dept.ParentID = req.ParentID
	dept.ManagerID = req.ManagerID
	dept.Sort = req.Sort
	dept.Status = req.Status

	if err := h.db.Save(&dept).Error; err != nil {
		log.Error().Err(err).Msg("update department failed")
		response.ServerError(c, "更新失败")
		return
	}
	response.Ok(c, dept)
}

func (h *DepartmentHandler) DeleteDepartment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	// 检查是否有子部门
	var childCount int64
	h.db.Model(&model.Department{}).Where("parent_id = ? AND company_id = ?", id, companyID).Count(&childCount)
	if childCount > 0 {
		response.Fail(c, response.CodeBadRequest, "该部门下存在子部门，无法删除")
		return
	}

	// 检查是否有关联的职员
	var empCount int64
	h.db.Model(&model.Employee{}).Where("dept_id = ?", id).Count(&empCount)
	if empCount > 0 {
		response.Fail(c, response.CodeBadRequest, "该部门下存在职员，无法删除")
		return
	}

	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.Department{}).Error; err != nil {
		log.Error().Err(err).Msg("delete department failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}
