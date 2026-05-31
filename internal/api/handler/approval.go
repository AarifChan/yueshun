package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/response"
)

// ApprovalHandler 审批处理器
type ApprovalHandler struct {
	db *gorm.DB
}

func NewApprovalHandler(db *gorm.DB) *ApprovalHandler {
	return &ApprovalHandler{db: db}
}

func (h *ApprovalHandler) RegisterRoutes(r *gin.RouterGroup) {
	ap := r.Group("/approval-processes")
	{
		ap.GET("", h.ListProcesses)
		ap.POST("", h.CreateProcess)
		ap.GET("/:id", h.GetProcess)
		ap.PUT("/:id", h.UpdateProcess)
		ap.DELETE("/:id", h.DeleteProcess)
	}

	ar := r.Group("/approval-records")
	{
		ar.GET("", h.ListRecords)
		ar.POST("", h.CreateRecord)
		ar.GET("/:id", h.GetRecord)
		ar.PUT("/:id/approve", h.ApproveRecord)
		ar.PUT("/:id/reject", h.RejectRecord)
	}
}

// ==================== 审批流程 ====================

type ProcessListReq struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"pageSize,default=20"`
	Keyword  string `form:"keyword"`
	Type     string `form:"type"`
	Status   int8   `form:"status"`
}

func (h *ApprovalHandler) ListProcesses(c *gin.Context) {
	var req ProcessListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.ApprovalProcess{}).Where("company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("name LIKE ?", "%"+req.Keyword+"%")
	}
	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}
	if req.Status != 0 {
		query = query.Where("status = ?", req.Status)
	}

	var total int64
	query.Count(&total)

	var list []model.ApprovalProcess
	query.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *ApprovalHandler) GetProcess(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var process model.ApprovalProcess
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&process).Error; err != nil {
		response.NotFound(c, "流程不存在")
		return
	}
	response.Ok(c, process)
}

type CreateProcessReq struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

func (h *ApprovalHandler) CreateProcess(c *gin.Context) {
	var req CreateProcessReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	process := model.ApprovalProcess{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		Name:                 req.Name,
		Code:                 req.Code,
		Type:                 req.Type,
		Description:          req.Description,
		Status:               1,
	}

	if err := h.db.Create(&process).Error; err != nil {
		log.Error().Err(err).Msg("create approval process failed")
		response.ServerError(c, "创建失败")
		return
	}

	response.Ok(c, process)
}

func (h *ApprovalHandler) UpdateProcess(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req CreateProcessReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var process model.ApprovalProcess
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&process).Error; err != nil {
		response.NotFound(c, "流程不存在")
		return
	}

	process.Name = req.Name
	process.Code = req.Code
	process.Type = req.Type
	process.Description = req.Description

	if err := h.db.Save(&process).Error; err != nil {
		log.Error().Err(err).Msg("update approval process failed")
		response.ServerError(c, "更新失败")
		return
	}

	response.Ok(c, process)
}

func (h *ApprovalHandler) DeleteProcess(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.ApprovalProcess{}).Error; err != nil {
		log.Error().Err(err).Msg("delete approval process failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

// ==================== 审批记录 ====================

type RecordListReq struct {
	Page         int    `form:"page,default=1"`
	PageSize     int    `form:"pageSize,default=20"`
	ProcessID    uint   `form:"processId"`
	BusinessType string `form:"businessType"`
	Action       string `form:"action"`
}

func (h *ApprovalHandler) ListRecords(c *gin.Context) {
	var req RecordListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.ApprovalRecord{}).Where("company_id = ?", companyID)
	if req.ProcessID > 0 {
		query = query.Where("process_id = ?", req.ProcessID)
	}
	if req.BusinessType != "" {
		query = query.Where("business_type = ?", req.BusinessType)
	}
	if req.Action != "" {
		query = query.Where("action = ?", req.Action)
	}

	var total int64
	query.Count(&total)

	var list []model.ApprovalRecord
	query.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *ApprovalHandler) GetRecord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var record model.ApprovalRecord
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&record).Error; err != nil {
		response.NotFound(c, "记录不存在")
		return
	}
	response.Ok(c, record)
}

type CreateRecordReq struct {
	ProcessID    uint   `json:"processId" binding:"required"`
	BusinessType string `json:"businessType" binding:"required"`
	BusinessID   uint   `json:"businessId" binding:"required"`
	Comment      string `json:"comment"`
}

func (h *ApprovalHandler) CreateRecord(c *gin.Context) {
	var req CreateRecordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	record := model.ApprovalRecord{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		ProcessID:            req.ProcessID,
		BusinessType:         req.BusinessType,
		BusinessID:           req.BusinessID,
		ApplicantID:          middleware.GetUserID(c),
		Action:               "pending",
		Comment:              req.Comment,
		Status:               1,
	}

	if err := h.db.Create(&record).Error; err != nil {
		log.Error().Err(err).Msg("create approval record failed")
		response.ServerError(c, "创建失败")
		return
	}

	response.Ok(c, record)
}

func (h *ApprovalHandler) ApproveRecord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var record model.ApprovalRecord
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&record).Error; err != nil {
		response.NotFound(c, "记录不存在")
		return
	}
	if record.Action != "pending" {
		response.Fail(c, response.CodeBadRequest, "当前状态不可审批")
		return
	}

	record.Action = "approved"
	record.ApproverID = middleware.GetUserID(c)
	record.ApproveDate = time.Now()

	if err := h.db.Save(&record).Error; err != nil {
		log.Error().Err(err).Msg("approve record failed")
		response.ServerError(c, "审批失败")
		return
	}
	response.OkWithMessage(c, "审批通过", nil)
}

func (h *ApprovalHandler) RejectRecord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var record model.ApprovalRecord
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&record).Error; err != nil {
		response.NotFound(c, "记录不存在")
		return
	}
	if record.Action != "pending" {
		response.Fail(c, response.CodeBadRequest, "当前状态不可审批")
		return
	}

	record.Action = "rejected"
	record.ApproverID = middleware.GetUserID(c)
	record.ApproveDate = time.Now()

	if err := h.db.Save(&record).Error; err != nil {
		log.Error().Err(err).Msg("reject record failed")
		response.ServerError(c, "拒绝失败")
		return
	}
	response.OkWithMessage(c, "已拒绝", nil)
}
