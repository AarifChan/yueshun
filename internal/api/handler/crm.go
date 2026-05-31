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

// CRMHandler CRM处理器
type CRMHandler struct {
	db *gorm.DB
}

func NewCRMHandler(db *gorm.DB) *CRMHandler {
	return &CRMHandler{db: db}
}

func (h *CRMHandler) RegisterRoutes(r *gin.RouterGroup) {
	fu := r.Group("/follow-ups")
	{
		fu.GET("", h.ListFollowUps)
		fu.POST("", h.CreateFollowUp)
		fu.GET("/:id", h.GetFollowUp)
		fu.PUT("/:id", h.UpdateFollowUp)
		fu.DELETE("/:id", h.DeleteFollowUp)
	}

	o := r.Group("/opportunities")
	{
		o.GET("", h.ListOpportunities)
		o.POST("", h.CreateOpportunity)
		o.GET("/:id", h.GetOpportunity)
		o.PUT("/:id", h.UpdateOpportunity)
		o.DELETE("/:id", h.DeleteOpportunity)
	}

	ct := r.Group("/contracts")
	{
		ct.GET("", h.ListContracts)
		ct.POST("", h.CreateContract)
		ct.GET("/:id", h.GetContract)
		ct.PUT("/:id", h.UpdateContract)
		ct.DELETE("/:id", h.DeleteContract)
	}
}

// ==================== 跟进记录 ====================

type FollowUpListReq struct {
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"pageSize,default=20"`
	CustomerID uint   `form:"customerId"`
	Type       string `form:"type"`
	StartDate  string `form:"startDate"`
	EndDate    string `form:"endDate"`
}

func (h *CRMHandler) ListFollowUps(c *gin.Context) {
	var req FollowUpListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.FollowUp{}).Where("company_id = ?", companyID)
	if req.CustomerID > 0 {
		query = query.Where("customer_id = ?", req.CustomerID)
	}
	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}
	if req.StartDate != "" {
		query = query.Where("contact_date >= ?", req.StartDate)
	}
	if req.EndDate != "" {
		query = query.Where("contact_date <= ?", req.EndDate)
	}

	var total int64
	query.Count(&total)

	var list []model.FollowUp
	query.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *CRMHandler) GetFollowUp(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var followUp model.FollowUp
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&followUp).Error; err != nil {
		response.NotFound(c, "跟进记录不存在")
		return
	}
	response.Ok(c, followUp)
}

type CreateFollowUpReq struct {
	CustomerID  uint   `json:"customerId" binding:"required"`
	ContactDate string `json:"contactDate" binding:"required"`
	Type        string `json:"type" binding:"required"`
	Content     string `json:"content"`
	NextPlan    string `json:"nextPlan"`
}

func (h *CRMHandler) CreateFollowUp(c *gin.Context) {
	var req CreateFollowUpReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	contactDate, _ := time.Parse("2006-01-02", req.ContactDate)
	followUp := model.FollowUp{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		CustomerID:         req.CustomerID,
		ContactDate:          contactDate,
		Type:                 req.Type,
		Content:              req.Content,
		NextPlan:             req.NextPlan,
		OperatorID:           middleware.GetUserID(c),
		Status:               1,
	}

	if err := h.db.Create(&followUp).Error; err != nil {
		log.Error().Err(err).Msg("create follow up failed")
		response.ServerError(c, "创建失败")
		return
	}

	response.Ok(c, followUp)
}

func (h *CRMHandler) UpdateFollowUp(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req CreateFollowUpReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var followUp model.FollowUp
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&followUp).Error; err != nil {
		response.NotFound(c, "跟进记录不存在")
		return
	}

	contactDate, _ := time.Parse("2006-01-02", req.ContactDate)
	followUp.CustomerID = req.CustomerID
	followUp.ContactDate = contactDate
	followUp.Type = req.Type
	followUp.Content = req.Content
	followUp.NextPlan = req.NextPlan

	if err := h.db.Save(&followUp).Error; err != nil {
		log.Error().Err(err).Msg("update follow up failed")
		response.ServerError(c, "更新失败")
		return
	}

	response.Ok(c, followUp)
}

func (h *CRMHandler) DeleteFollowUp(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.FollowUp{}).Error; err != nil {
		log.Error().Err(err).Msg("delete follow up failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

// ==================== 商机 ====================

type OpportunityListReq struct {
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"pageSize,default=20"`
	CustomerID uint   `form:"customerId"`
	Stage      string `form:"stage"`
	Status     int8   `form:"status"`
}

func (h *CRMHandler) ListOpportunities(c *gin.Context) {
	var req OpportunityListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.Opportunity{}).Where("company_id = ?", companyID)
	if req.CustomerID > 0 {
		query = query.Where("customer_id = ?", req.CustomerID)
	}
	if req.Stage != "" {
		query = query.Where("stage = ?", req.Stage)
	}
	if req.Status != 0 {
		query = query.Where("status = ?", req.Status)
	}

	var total int64
	query.Count(&total)

	var list []model.Opportunity
	query.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *CRMHandler) GetOpportunity(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var opp model.Opportunity
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&opp).Error; err != nil {
		response.NotFound(c, "商机不存在")
		return
	}
	response.Ok(c, opp)
}

type CreateOpportunityReq struct {
	CustomerID   uint    `json:"customerId" binding:"required"`
	Name         string  `json:"name" binding:"required"`
	Amount       float64 `json:"amount"`
	Stage        string  `json:"stage"`
	Probability  float64 `json:"probability"`
	ExpectedDate string  `json:"expectedDate"`
	Remark       string  `json:"remark"`
}

func (h *CRMHandler) CreateOpportunity(c *gin.Context) {
	var req CreateOpportunityReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	expectedDate, _ := time.Parse("2006-01-02", req.ExpectedDate)
	opp := model.Opportunity{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		CustomerID:         req.CustomerID,
		Name:               req.Name,
		Amount:             req.Amount,
		Stage:              req.Stage,
		Probability:        req.Probability,
		ExpectedDate:       expectedDate,
		OperatorID:         middleware.GetUserID(c),
		Remark:             req.Remark,
		Status:             1,
	}

	if err := h.db.Create(&opp).Error; err != nil {
		log.Error().Err(err).Msg("create opportunity failed")
		response.ServerError(c, "创建失败")
		return
	}

	response.Ok(c, opp)
}

func (h *CRMHandler) UpdateOpportunity(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req CreateOpportunityReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var opp model.Opportunity
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&opp).Error; err != nil {
		response.NotFound(c, "商机不存在")
		return
	}

	expectedDate, _ := time.Parse("2006-01-02", req.ExpectedDate)
	opp.CustomerID = req.CustomerID
	opp.Name = req.Name
	opp.Amount = req.Amount
	opp.Stage = req.Stage
	opp.Probability = req.Probability
	opp.ExpectedDate = expectedDate
	opp.Remark = req.Remark

	if err := h.db.Save(&opp).Error; err != nil {
		log.Error().Err(err).Msg("update opportunity failed")
		response.ServerError(c, "更新失败")
		return
	}

	response.Ok(c, opp)
}

func (h *CRMHandler) DeleteOpportunity(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.Opportunity{}).Error; err != nil {
		log.Error().Err(err).Msg("delete opportunity failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

// ==================== 合同 ====================

type ContractListReq struct {
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"pageSize,default=20"`
	Keyword    string `form:"keyword"`
	CustomerID uint   `form:"customerId"`
	Status     string `form:"status"`
}

func (h *CRMHandler) ListContracts(c *gin.Context) {
	var req ContractListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.Contract{}).Where("company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("contract_no LIKE ?", "%"+req.Keyword+"%")
	}
	if req.CustomerID > 0 {
		query = query.Where("customer_id = ?", req.CustomerID)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	var total int64
	query.Count(&total)

	var list []model.Contract
	query.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *CRMHandler) GetContract(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var contract model.Contract
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&contract).Error; err != nil {
		response.NotFound(c, "合同不存在")
		return
	}
	response.Ok(c, contract)
}

type CreateContractReq struct {
	CustomerID    uint   `json:"customerId" binding:"required"`
	OpportunityID uint   `json:"opportunityId"`
	ContractNo    string `json:"contractNo" binding:"required"`
	ContractDate  string `json:"contractDate" binding:"required"`
	Amount        float64 `json:"amount"`
	StartDate     string `json:"startDate"`
	EndDate       string `json:"endDate"`
	Status        string `json:"status"`
	Remark        string `json:"remark"`
}

func (h *CRMHandler) CreateContract(c *gin.Context) {
	var req CreateContractReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	contractDate, _ := time.Parse("2006-01-02", req.ContractDate)
	startDate, _ := time.Parse("2006-01-02", req.StartDate)
	endDate, _ := time.Parse("2006-01-02", req.EndDate)
	contract := model.Contract{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		CustomerID:         req.CustomerID,
		OpportunityID:      req.OpportunityID,
		ContractNo:         req.ContractNo,
		ContractDate:       contractDate,
		Amount:             req.Amount,
		StartDate:          startDate,
		EndDate:            endDate,
		Status:             req.Status,
		OperatorID:         middleware.GetUserID(c),
		Remark:             req.Remark,
	}

	if err := h.db.Create(&contract).Error; err != nil {
		log.Error().Err(err).Msg("create contract failed")
		response.ServerError(c, "创建失败")
		return
	}

	response.Ok(c, contract)
}

func (h *CRMHandler) UpdateContract(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req CreateContractReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var contract model.Contract
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&contract).Error; err != nil {
		response.NotFound(c, "合同不存在")
		return
	}

	contractDate, _ := time.Parse("2006-01-02", req.ContractDate)
	startDate, _ := time.Parse("2006-01-02", req.StartDate)
	endDate, _ := time.Parse("2006-01-02", req.EndDate)
	contract.CustomerID = req.CustomerID
	contract.OpportunityID = req.OpportunityID
	contract.ContractNo = req.ContractNo
	contract.ContractDate = contractDate
	contract.Amount = req.Amount
	contract.StartDate = startDate
	contract.EndDate = endDate
	contract.Status = req.Status
	contract.Remark = req.Remark

	if err := h.db.Save(&contract).Error; err != nil {
		log.Error().Err(err).Msg("update contract failed")
		response.ServerError(c, "更新失败")
		return
	}

	response.Ok(c, contract)
}

func (h *CRMHandler) DeleteContract(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.Contract{}).Error; err != nil {
		log.Error().Err(err).Msg("delete contract failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}
