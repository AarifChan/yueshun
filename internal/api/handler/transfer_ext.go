package handler

import (
	"errors"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/database"
	"zhizhang-server/internal/pkg/response"
)

// TransferExtHandler 调拨处理器（调拨申请/出库/入库/差异处理）
type TransferExtHandler struct {
	db *gorm.DB
}

func NewTransferExtHandler(db *gorm.DB) *TransferExtHandler {
	return &TransferExtHandler{db: db}
}

func (h *TransferExtHandler) RegisterRoutes(r *gin.RouterGroup) {
	apply := r.Group("/transfer-applies")
	{
		apply.GET("", h.ListApplies)
		apply.POST("", h.CreateApply)
		apply.GET("/:id", h.GetApply)
		apply.PUT("/:id", h.UpdateApply)
		apply.DELETE("/:id", h.DeleteApply)
		apply.PUT("/:id/approve", h.ApproveApply)
	}
	out := r.Group("/transfer-outs")
	{
		out.GET("", h.ListOuts)
		out.POST("", h.CreateOut)
		out.GET("/:id", h.GetOut)
		out.PUT("/:id", h.UpdateOut)
		out.DELETE("/:id", h.DeleteOut)
		out.PUT("/:id/complete", h.CompleteOut)
	}
	in := r.Group("/transfer-ins")
	{
		in.GET("", h.ListIns)
		in.POST("", h.CreateIn)
		in.GET("/:id", h.GetIn)
		in.PUT("/:id/complete", h.CompleteIn)
	}
	r.GET("/transfer-diffs", h.ListDiffs)
	r.PUT("/transfer-diffs/:itemId/handle", h.HandleDiff)
}

// ==================== 调拨申请单 ====================

type applyListReq struct {
	Page            int    `form:"page,default=1"`
	PageSize        int    `form:"pageSize,default=20"`
	Keyword         string `form:"keyword"`
	ProductKw       string `form:"productKw"`
	FromWarehouseID uint   `form:"fromWarehouseId"`
	Status          string `form:"status"`
	StartDate       string `form:"startDate"`
	EndDate         string `form:"endDate"`
}

type applyRow struct {
	model.TransferApply
	FromWarehouseName string `json:"fromWarehouseName"`
	ToWarehouseName   string `json:"toWarehouseName"`
	HandlerName       string `json:"handlerName"`
}

// ListApplies 调拨申请单列表
// @Summary 调拨申请单列表
// @Tags 库存
// @Router /api/v1/transfer-applies [get]
func (h *TransferExtHandler) ListApplies(c *gin.Context) {
	var req applyListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	q := h.db.Model(&model.TransferApply{}).
		Select(`transfer_applies.*, COALESCE(fw.name,'') AS from_warehouse_name, COALESCE(tw.name,'') AS to_warehouse_name, COALESCE(hd.name,'') AS handler_name`).
		Joins("LEFT JOIN warehouses fw ON fw.id = transfer_applies.from_warehouse_id").
		Joins("LEFT JOIN warehouses tw ON tw.id = transfer_applies.to_warehouse_id").
		Joins("LEFT JOIN employees hd ON hd.id = transfer_applies.handler_id").
		Where("transfer_applies.company_id = ?", middleware.GetCompanyID(c))
	if req.Keyword != "" {
		q = q.Where("transfer_applies.bill_no LIKE ?", "%"+req.Keyword+"%")
	}
	if req.FromWarehouseID > 0 {
		q = q.Where("transfer_applies.from_warehouse_id = ?", req.FromWarehouseID)
	}
	if req.Status != "" {
		q = q.Where("transfer_applies.status = ?", req.Status)
	}
	if req.StartDate != "" {
		q = q.Where("transfer_applies.created_at >= ?", req.StartDate+" 00:00:00")
	}
	if req.EndDate != "" {
		q = q.Where("transfer_applies.created_at <= ?", req.EndDate+" 23:59:59")
	}
	if req.ProductKw != "" {
		q = q.Where(`EXISTS (SELECT 1 FROM transfer_apply_items ai JOIN products p ON p.id = ai.product_id
			WHERE ai.apply_id = transfer_applies.id AND (p.name LIKE ? OR p.code LIKE ?))`, "%"+req.ProductKw+"%", "%"+req.ProductKw+"%")
	}
	var total int64
	q.Count(&total)
	var list []applyRow
	q.Order("transfer_applies.id DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Scan(&list)
	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

// GetApply 调拨申请单详情
// @Summary 调拨申请单详情
// @Tags 库存
// @Router /api/v1/transfer-applies/:id [get]
func (h *TransferExtHandler) GetApply(c *gin.Context) {
	var bill model.TransferApply
	if err := h.db.Where("id = ? AND company_id = ?", c.Param("id"), middleware.GetCompanyID(c)).
		Preload("Items.Product").First(&bill).Error; err != nil {
		response.NotFound(c, "单据不存在")
		return
	}
	response.Ok(c, bill)
}

type applyItemReq struct {
	ProductID uint    `json:"productId" binding:"required"`
	Quantity  float64 `json:"quantity" binding:"required,gt=0"`
	Remark    string  `json:"remark"`
}

type applyReq struct {
	FromWarehouseID uint           `json:"fromWarehouseId" binding:"required"`
	ToWarehouseID   uint           `json:"toWarehouseId" binding:"required"`
	BillDate        string         `json:"billDate"`
	HandlerID       uint           `json:"handlerId"`
	Remark          string         `json:"remark"`
	Items           []applyItemReq `json:"items" binding:"required,min=1"`
}

func parseBillDate(s string) time.Time {
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t
	}
	return time.Now()
}

// CreateApply 新增调拨申请单
// @Summary 新增调拨申请单
// @Tags 库存
// @Router /api/v1/transfer-applies [post]
func (h *TransferExtHandler) CreateApply(c *gin.Context) {
	var req applyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	var bill model.TransferApply
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		bill = model.TransferApply{
			FromWarehouseID: req.FromWarehouseID, ToWarehouseID: req.ToWarehouseID,
			BillNo: genDailyBillNo(tx, "transfer_applies", "DBSQ"), BillDate: parseBillDate(req.BillDate),
			Status: "draft", HandlerID: req.HandlerID, OperatorID: middleware.GetUserID(c), Remark: req.Remark,
		}
		bill.CompanyID = companyID
		for _, it := range req.Items {
			bill.TotalQty += it.Quantity
			bill.Items = append(bill.Items, model.TransferApplyItem{ProductID: it.ProductID, Quantity: it.Quantity, Remark: it.Remark})
		}
		return tx.Create(&bill).Error
	}); err != nil {
		log.Error().Err(err).Msg("create transfer apply failed")
		response.ServerError(c, "创建失败")
		return
	}
	response.Ok(c, bill)
}

// UpdateApply 修改调拨申请单（草稿）
// @Summary 修改调拨申请单
// @Tags 库存
// @Router /api/v1/transfer-applies/:id [put]
func (h *TransferExtHandler) UpdateApply(c *gin.Context) {
	var req applyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	var bill model.TransferApply
	if err := h.db.Where("id = ? AND company_id = ?", c.Param("id"), companyID).First(&bill).Error; err != nil {
		response.NotFound(c, "单据不存在")
		return
	}
	if bill.Status != "draft" && bill.Status != "pending" {
		response.Fail(c, response.CodeBadRequest, "当前状态不可编辑")
		return
	}
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("apply_id = ?", bill.ID).Delete(&model.TransferApplyItem{}).Error; err != nil {
			return err
		}
		var totalQty float64
		var items []model.TransferApplyItem
		for _, it := range req.Items {
			totalQty += it.Quantity
			items = append(items, model.TransferApplyItem{ApplyID: bill.ID, ProductID: it.ProductID, Quantity: it.Quantity, Remark: it.Remark})
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		return tx.Model(&bill).Updates(map[string]interface{}{
			"from_warehouse_id": req.FromWarehouseID, "to_warehouse_id": req.ToWarehouseID,
			"bill_date": parseBillDate(req.BillDate), "handler_id": req.HandlerID, "remark": req.Remark, "total_qty": totalQty,
		}).Error
	}); err != nil {
		log.Error().Err(err).Msg("update transfer apply failed")
		response.ServerError(c, "更新失败")
		return
	}
	response.OkWithMessage(c, "更新成功", nil)
}

// DeleteApply 删除调拨申请单
// @Summary 删除调拨申请单
// @Tags 库存
// @Router /api/v1/transfer-applies/:id [delete]
func (h *TransferExtHandler) DeleteApply(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var bill model.TransferApply
	if err := h.db.Where("id = ? AND company_id = ?", c.Param("id"), companyID).First(&bill).Error; err != nil {
		response.NotFound(c, "单据不存在")
		return
	}
	if bill.Status != "draft" && bill.Status != "pending" {
		response.Fail(c, response.CodeBadRequest, "当前状态不可删除")
		return
	}
	h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("apply_id = ?", bill.ID).Delete(&model.TransferApplyItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&bill).Error
	})
	response.Ok(c, nil)
}

// ApproveApply 审核调拨申请单：draft/pending → approved（待出库）
// @Summary 审核调拨申请单
// @Tags 库存
// @Router /api/v1/transfer-applies/:id/approve [put]
func (h *TransferExtHandler) ApproveApply(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var bill model.TransferApply
	if err := h.db.Where("id = ? AND company_id = ?", c.Param("id"), companyID).First(&bill).Error; err != nil {
		response.NotFound(c, "单据不存在")
		return
	}
	if bill.Status != "draft" && bill.Status != "pending" {
		response.Fail(c, response.CodeBadRequest, "当前状态不可审核")
		return
	}
	h.db.Model(&bill).UpdateColumn("status", "approved")
	response.OkWithMessage(c, "审核成功", nil)
}

// ==================== 调拨出库单 ====================

type transferOutListReq struct {
	Page            int    `form:"page,default=1"`
	PageSize        int    `form:"pageSize,default=20"`
	Keyword         string `form:"keyword"`
	ProductKw       string `form:"productKw"`
	TransferType    string `form:"transferType"`
	Status          string `form:"status"`
	FromWarehouseID uint   `form:"fromWarehouseId"`
	ShowCancelled   bool   `form:"showCancelled"`
	StartDate       string `form:"startDate"`
	EndDate         string `form:"endDate"`
}

type transferOutRow struct {
	model.TransferOut
	FromWarehouseName string `json:"fromWarehouseName"`
	ToWarehouseName   string `json:"toWarehouseName"`
	HandlerName       string `json:"handlerName"`
}

// ListOuts 调拨出库单列表
// @Summary 调拨出库单列表
// @Tags 库存
// @Router /api/v1/transfer-outs [get]
func (h *TransferExtHandler) ListOuts(c *gin.Context) {
	var req transferOutListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	q := h.db.Model(&model.TransferOut{}).
		Select(`transfer_outs.*, COALESCE(fw.name,'') AS from_warehouse_name, COALESCE(tw.name,'') AS to_warehouse_name, COALESCE(hd.name,'') AS handler_name`).
		Joins("LEFT JOIN warehouses fw ON fw.id = transfer_outs.from_warehouse_id").
		Joins("LEFT JOIN warehouses tw ON tw.id = transfer_outs.to_warehouse_id").
		Joins("LEFT JOIN employees hd ON hd.id = transfer_outs.handler_id").
		Where("transfer_outs.company_id = ?", middleware.GetCompanyID(c))
	if !req.ShowCancelled {
		q = q.Where("transfer_outs.status != 'cancelled'")
	}
	if req.Keyword != "" {
		q = q.Where("transfer_outs.bill_no LIKE ?", "%"+req.Keyword+"%")
	}
	if req.TransferType != "" {
		q = q.Where("transfer_outs.transfer_type = ?", req.TransferType)
	}
	if req.Status != "" {
		q = q.Where("transfer_outs.status = ?", req.Status)
	}
	if req.FromWarehouseID > 0 {
		q = q.Where("transfer_outs.from_warehouse_id = ?", req.FromWarehouseID)
	}
	if req.StartDate != "" {
		q = q.Where("transfer_outs.created_at >= ?", req.StartDate+" 00:00:00")
	}
	if req.EndDate != "" {
		q = q.Where("transfer_outs.created_at <= ?", req.EndDate+" 23:59:59")
	}
	if req.ProductKw != "" {
		q = q.Where(`EXISTS (SELECT 1 FROM transfer_out_items oi JOIN products p ON p.id = oi.product_id
			WHERE oi.out_id = transfer_outs.id AND (p.name LIKE ? OR p.code LIKE ?))`, "%"+req.ProductKw+"%", "%"+req.ProductKw+"%")
	}
	var total int64
	q.Count(&total)
	var list []transferOutRow
	q.Order("transfer_outs.id DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Scan(&list)
	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

// GetOut 调拨出库单详情
// @Summary 调拨出库单详情
// @Tags 库存
// @Router /api/v1/transfer-outs/:id [get]
func (h *TransferExtHandler) GetOut(c *gin.Context) {
	var bill model.TransferOut
	if err := h.db.Where("id = ? AND company_id = ?", c.Param("id"), middleware.GetCompanyID(c)).
		Preload("Items.Product").First(&bill).Error; err != nil {
		response.NotFound(c, "单据不存在")
		return
	}
	response.Ok(c, bill)
}

type transferOutItemReq struct {
	ProductID uint    `json:"productId" binding:"required"`
	Quantity  float64 `json:"quantity" binding:"required,gt=0"`
	Price     float64 `json:"price"`
	Remark    string  `json:"remark"`
}

type transferOutReq struct {
	ApplyID         uint                  `json:"applyId"`
	FromWarehouseID uint                  `json:"fromWarehouseId" binding:"required"`
	ToWarehouseID   uint                  `json:"toWarehouseId" binding:"required"`
	TransferType    string                `json:"transferType" binding:"required,oneof=same diff"`
	BillDate        string                `json:"billDate"`
	HandlerID       uint                  `json:"handlerId"`
	Remark          string                `json:"remark"`
	Items           []transferOutItemReq  `json:"items" binding:"required,min=1"`
}

// CreateOut 新增调拨出库单
// @Summary 新增调拨出库单
// @Tags 库存
// @Router /api/v1/transfer-outs [post]
func (h *TransferExtHandler) CreateOut(c *gin.Context) {
	var req transferOutReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	var bill model.TransferOut
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		bill = model.TransferOut{
			ApplyID: req.ApplyID, FromWarehouseID: req.FromWarehouseID, ToWarehouseID: req.ToWarehouseID,
			BillNo: genDailyBillNo(tx, "transfer_outs", "DBCK"), BillDate: parseBillDate(req.BillDate),
			TransferType: req.TransferType, Status: "draft",
			HandlerID: req.HandlerID, OperatorID: middleware.GetUserID(c), Remark: req.Remark,
		}
		bill.CompanyID = companyID
		for _, it := range req.Items {
			amount := it.Quantity * it.Price
			bill.TotalQty += it.Quantity
			bill.Amount += amount
			bill.Items = append(bill.Items, model.TransferOutItem{
				ProductID: it.ProductID, Quantity: it.Quantity, Price: it.Price, Amount: amount, Remark: it.Remark,
			})
		}
		return tx.Create(&bill).Error
	}); err != nil {
		log.Error().Err(err).Msg("create transfer out failed")
		response.ServerError(c, "创建失败")
		return
	}
	response.Ok(c, bill)
}

// UpdateOut 修改调拨出库单（草稿）
// @Summary 修改调拨出库单
// @Tags 库存
// @Router /api/v1/transfer-outs/:id [put]
func (h *TransferExtHandler) UpdateOut(c *gin.Context) {
	var req transferOutReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	var bill model.TransferOut
	if err := h.db.Where("id = ? AND company_id = ?", c.Param("id"), companyID).First(&bill).Error; err != nil {
		response.NotFound(c, "单据不存在")
		return
	}
	if bill.Status != "draft" {
		response.Fail(c, response.CodeBadRequest, "只有草稿状态可编辑")
		return
	}
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("out_id = ?", bill.ID).Delete(&model.TransferOutItem{}).Error; err != nil {
			return err
		}
		var totalQty, amount float64
		var items []model.TransferOutItem
		for _, it := range req.Items {
			amt := it.Quantity * it.Price
			totalQty += it.Quantity
			amount += amt
			items = append(items, model.TransferOutItem{
				OutID: bill.ID, ProductID: it.ProductID, Quantity: it.Quantity, Price: it.Price, Amount: amt, Remark: it.Remark,
			})
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		return tx.Model(&bill).Updates(map[string]interface{}{
			"from_warehouse_id": req.FromWarehouseID, "to_warehouse_id": req.ToWarehouseID,
			"transfer_type": req.TransferType, "bill_date": parseBillDate(req.BillDate),
			"handler_id": req.HandlerID, "remark": req.Remark, "total_qty": totalQty, "amount": amount,
		}).Error
	}); err != nil {
		log.Error().Err(err).Msg("update transfer out failed")
		response.ServerError(c, "更新失败")
		return
	}
	response.OkWithMessage(c, "更新成功", nil)
}

// DeleteOut 删除调拨出库单（草稿）
// @Summary 删除调拨出库单
// @Tags 库存
// @Router /api/v1/transfer-outs/:id [delete]
func (h *TransferExtHandler) DeleteOut(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var bill model.TransferOut
	if err := h.db.Where("id = ? AND company_id = ?", c.Param("id"), companyID).First(&bill).Error; err != nil {
		response.NotFound(c, "单据不存在")
		return
	}
	if bill.Status != "draft" {
		response.Fail(c, response.CodeBadRequest, "只有草稿状态可删除")
		return
	}
	h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("out_id = ?", bill.ID).Delete(&model.TransferOutItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&bill).Error
	})
	response.Ok(c, nil)
}

// CompleteOut 调拨出库单过账：调出仓减库存；若关联申请单则推进其状态
// @Summary 调拨出库单过账
// @Tags 库存
// @Router /api/v1/transfer-outs/:id/complete [put]
func (h *TransferExtHandler) CompleteOut(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var bill model.TransferOut
	if err := h.db.Where("id = ? AND company_id = ?", c.Param("id"), companyID).Preload("Items").First(&bill).Error; err != nil {
		response.NotFound(c, "单据不存在")
		return
	}
	if bill.Status != "draft" {
		response.Fail(c, response.CodeBadRequest, "只有草稿状态可过账")
		return
	}
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range bill.Items {
			if err := database.ChangeStock(tx, companyID, bill.FromWarehouseID, item.ProductID, -item.Quantity); err != nil {
				return err
			}
		}
		if err := tx.Model(&bill).UpdateColumn("status", "completed").Error; err != nil {
			return err
		}
		if bill.ApplyID > 0 {
			if err := tx.Model(&model.TransferApply{}).
				Where("id = ? AND company_id = ? AND status = 'approved'", bill.ApplyID, companyID).
				UpdateColumn("status", "shipped").Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		if errors.Is(err, database.ErrStockNotEnough) {
			response.Fail(c, response.CodeBadRequest, err.Error())
			return
		}
		log.Error().Err(err).Msg("complete transfer out failed")
		response.ServerError(c, "过账失败")
		return
	}
	response.OkWithMessage(c, "过账成功", nil)
}

// ==================== 调拨入库单 ====================

// ListIns 调拨入库单列表
// @Summary 调拨入库单列表
// @Tags 库存
// @Router /api/v1/transfer-ins [get]
func (h *TransferExtHandler) ListIns(c *gin.Context) {
	var req transferOutListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	q := h.db.Model(&model.TransferIn{}).
		Select(`transfer_ins.*, COALESCE(fw.name,'') AS from_warehouse_name, COALESCE(tw.name,'') AS to_warehouse_name, COALESCE(hd.name,'') AS handler_name`).
		Joins("LEFT JOIN warehouses fw ON fw.id = transfer_ins.from_warehouse_id").
		Joins("LEFT JOIN warehouses tw ON tw.id = transfer_ins.to_warehouse_id").
		Joins("LEFT JOIN employees hd ON hd.id = transfer_ins.handler_id").
		Where("transfer_ins.company_id = ?", middleware.GetCompanyID(c))
	if !req.ShowCancelled {
		q = q.Where("transfer_ins.status != 'cancelled'")
	}
	if req.Keyword != "" {
		q = q.Where("transfer_ins.bill_no LIKE ?", "%"+req.Keyword+"%")
	}
	if req.TransferType != "" {
		q = q.Where("transfer_ins.transfer_type = ?", req.TransferType)
	}
	if req.Status != "" {
		q = q.Where("transfer_ins.status = ?", req.Status)
	}
	if req.StartDate != "" {
		q = q.Where("transfer_ins.created_at >= ?", req.StartDate+" 00:00:00")
	}
	if req.EndDate != "" {
		q = q.Where("transfer_ins.created_at <= ?", req.EndDate+" 23:59:59")
	}
	if req.ProductKw != "" {
		q = q.Where(`EXISTS (SELECT 1 FROM transfer_in_items ii JOIN products p ON p.id = ii.product_id
			WHERE ii.in_id = transfer_ins.id AND (p.name LIKE ? OR p.code LIKE ?))`, "%"+req.ProductKw+"%", "%"+req.ProductKw+"%")
	}
	var total int64
	q.Count(&total)
	var list []transferOutRow
	q.Order("transfer_ins.id DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Scan(&list)
	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

// GetIn 调拨入库单详情
// @Summary 调拨入库单详情
// @Tags 库存
// @Router /api/v1/transfer-ins/:id [get]
func (h *TransferExtHandler) GetIn(c *gin.Context) {
	var bill model.TransferIn
	if err := h.db.Where("id = ? AND company_id = ?", c.Param("id"), middleware.GetCompanyID(c)).
		Preload("Items.Product").First(&bill).Error; err != nil {
		response.NotFound(c, "单据不存在")
		return
	}
	response.Ok(c, bill)
}

type transferInItemReq struct {
	ProductID uint    `json:"productId" binding:"required"`
	Quantity  float64 `json:"quantity" binding:"required,gt=0"`
	Price     float64 `json:"price"`
	Remark    string  `json:"remark"`
}

type transferInReq struct {
	OutID  uint                `json:"outId" binding:"required"`
	Remark string              `json:"remark"`
	Items  []transferInItemReq `json:"items" binding:"required,min=1"`
}

// CreateIn 新增调拨入库单（基于已过账调拨出库单）
// @Summary 新增调拨入库单
// @Tags 库存
// @Router /api/v1/transfer-ins [post]
func (h *TransferExtHandler) CreateIn(c *gin.Context) {
	var req transferInReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	var outBill model.TransferOut
	if err := h.db.Where("id = ? AND company_id = ? AND status = 'completed'", req.OutID, companyID).First(&outBill).Error; err != nil {
		response.NotFound(c, "关联调拨出库单不存在或未过账")
		return
	}
	var bill model.TransferIn
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		bill = model.TransferIn{
			OutID: outBill.ID, OutBillNo: outBill.BillNo,
			FromWarehouseID: outBill.FromWarehouseID, ToWarehouseID: outBill.ToWarehouseID,
			BillNo: genDailyBillNo(tx, "transfer_ins", "DBRK"), BillDate: time.Now(),
			TransferType: outBill.TransferType, Status: "draft",
			HandlerID: outBill.HandlerID, OperatorID: middleware.GetUserID(c), Remark: req.Remark,
		}
		bill.CompanyID = companyID
		for _, it := range req.Items {
			amount := it.Quantity * it.Price
			bill.TotalQty += it.Quantity
			bill.Amount += amount
			bill.Items = append(bill.Items, model.TransferInItem{
				ProductID: it.ProductID, Quantity: it.Quantity, Price: it.Price, Amount: amount, Remark: it.Remark,
			})
		}
		return tx.Create(&bill).Error
	}); err != nil {
		log.Error().Err(err).Msg("create transfer in failed")
		response.ServerError(c, "创建失败")
		return
	}
	response.Ok(c, bill)
}

// CompleteIn 调拨入库单过账：调入仓加库存，回写出库行已入库数量与差异状态
// @Summary 调拨入库单过账
// @Tags 库存
// @Router /api/v1/transfer-ins/:id/complete [put]
func (h *TransferExtHandler) CompleteIn(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var bill model.TransferIn
	if err := h.db.Where("id = ? AND company_id = ?", c.Param("id"), companyID).Preload("Items").First(&bill).Error; err != nil {
		response.NotFound(c, "单据不存在")
		return
	}
	if bill.Status != "draft" {
		response.Fail(c, response.CodeBadRequest, "只有草稿状态可过账")
		return
	}
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range bill.Items {
			if err := database.ChangeStock(tx, companyID, bill.ToWarehouseID, item.ProductID, item.Quantity); err != nil {
				return err
			}
		}
		if err := tx.Model(&bill).UpdateColumn("status", "completed").Error; err != nil {
			return err
		}
		// 回写出库单行：累加已入库数量，标记差异
		if bill.OutID > 0 {
			for _, item := range bill.Items {
				if err := tx.Exec(`UPDATE transfer_out_items SET received_qty = received_qty + ?,
					diff_status = CASE WHEN received_qty + ? >= quantity THEN 'none' ELSE 'pending' END
					WHERE out_id = ? AND product_id = ?`, item.Quantity, item.Quantity, bill.OutID, item.ProductID).Error; err != nil {
					return err
				}
			}
			// 出库单关联的申请单：若全部行已收满则置 completed
			var outBill model.TransferOut
			if err := tx.Where("id = ?", bill.OutID).First(&outBill).Error; err == nil && outBill.ApplyID > 0 {
				var pending int64
				tx.Model(&model.TransferOutItem{}).Where("out_id = ? AND diff_status = 'pending'", bill.OutID).Count(&pending)
				if pending == 0 {
					tx.Model(&model.TransferApply{}).
						Where("id = ? AND company_id = ? AND status = 'shipped'", outBill.ApplyID, companyID).
						UpdateColumn("status", "completed")
				}
			}
		}
		return nil
	}); err != nil {
		if errors.Is(err, database.ErrStockNotEnough) {
			response.Fail(c, response.CodeBadRequest, err.Error())
			return
		}
		log.Error().Err(err).Msg("complete transfer in failed")
		response.ServerError(c, "过账失败")
		return
	}
	response.OkWithMessage(c, "过账成功", nil)
}

// ==================== 调拨差异处理 ====================

type diffRow struct {
	ItemID            uint    `json:"itemId"`
	OutID             uint    `json:"outId"`
	OutBillNo         string  `json:"outBillNo"`
	TransferType      string  `json:"transferType"`
	FromWarehouseName string  `json:"fromWarehouseName"`
	ToWarehouseName   string  `json:"toWarehouseName"`
	ProductID         uint    `json:"productId"`
	ProductName       string  `json:"productName"`
	Specification     string  `json:"specification"`
	Unit              string  `json:"unit"`
	OutQty            float64 `json:"outQty"`
	InQty             float64 `json:"inQty"`
	DiffQty           float64 `json:"diffQty"`
	DiffStatus        string  `json:"diffStatus"`
	DiffResult        string  `json:"diffResult"`
	CreatedAt         string  `json:"createdAt"`
}

// ListDiffs 调拨差异列表（按出库单行 出-入 差异）
// @Summary 调拨差异处理列表
// @Tags 库存
// @Param includeNoDiff query bool false "包含没有差异商品行"
// @Param includeNotReceived query bool false "包含未入库商品行"
// @Router /api/v1/transfer-diffs [get]
func (h *TransferExtHandler) ListDiffs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")
	transferType := c.Query("transferType")
	productKw := c.Query("productKw")
	fromWarehouseID, _ := strconv.Atoi(c.DefaultQuery("fromWarehouseId", "0"))
	includeNoDiff := c.Query("includeNoDiff") == "true"
	includeNotReceived := c.Query("includeNotReceived") == "true"
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	q := h.db.Model(&model.TransferOutItem{}).
		Select(`transfer_out_items.id AS item_id, transfer_out_items.out_id, transfer_outs.bill_no AS out_bill_no,
			transfer_outs.transfer_type, COALESCE(fw.name,'') AS from_warehouse_name, COALESCE(tw.name,'') AS to_warehouse_name,
			transfer_out_items.product_id, COALESCE(p.name,'') AS product_name, COALESCE(p.specification,'') AS specification, COALESCE(p.unit,'') AS unit,
			transfer_out_items.quantity AS out_qty, transfer_out_items.received_qty AS in_qty,
			(transfer_out_items.quantity - transfer_out_items.received_qty) AS diff_qty,
			transfer_out_items.diff_status, COALESCE(transfer_out_items.diff_result,'') AS diff_result,
			transfer_outs.created_at`).
		Joins("JOIN transfer_outs ON transfer_outs.id = transfer_out_items.out_id AND transfer_outs.status = 'completed'").
		Joins("LEFT JOIN warehouses fw ON fw.id = transfer_outs.from_warehouse_id").
		Joins("LEFT JOIN warehouses tw ON tw.id = transfer_outs.to_warehouse_id").
		Joins("LEFT JOIN products p ON p.id = transfer_out_items.product_id").
		Where("transfer_outs.company_id = ?", middleware.GetCompanyID(c))
	if !includeNoDiff {
		q = q.Where("transfer_out_items.quantity != transfer_out_items.received_qty OR transfer_out_items.diff_status = 'pending'")
	}
	if !includeNotReceived {
		q = q.Where("transfer_out_items.received_qty > 0")
	}
	if keyword != "" {
		q = q.Where("transfer_outs.bill_no LIKE ?", "%"+keyword+"%")
	}
	if transferType != "" {
		q = q.Where("transfer_outs.transfer_type = ?", transferType)
	}
	if fromWarehouseID > 0 {
		q = q.Where("transfer_outs.from_warehouse_id = ?", fromWarehouseID)
	}
	if productKw != "" {
		q = q.Where("(p.name LIKE ? OR p.code LIKE ?)", "%"+productKw+"%", "%"+productKw+"%")
	}
	if startDate != "" {
		q = q.Where("transfer_outs.created_at >= ?", startDate+" 00:00:00")
	}
	if endDate != "" {
		q = q.Where("transfer_outs.created_at <= ?", endDate+" 23:59:59")
	}
	var total int64
	q.Count(&total)
	var list []diffRow
	q.Order("transfer_out_items.id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&list)
	response.OkWithPage(c, list, page, pageSize, int(total))
}

type handleDiffReq struct {
	Result string `json:"result" binding:"required,max=64"` // 处理结果，如 补发/退货/其他
}

// HandleDiff 处理差异行
// @Summary 处理调拨差异
// @Tags 库存
// @Router /api/v1/transfer-diffs/:itemId/handle [put]
func (h *TransferExtHandler) HandleDiff(c *gin.Context) {
	var req handleDiffReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	itemID, err := strconv.ParseUint(c.Param("itemId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	res := h.db.Model(&model.TransferOutItem{}).
		Where("id = ? AND EXISTS (SELECT 1 FROM transfer_outs WHERE transfer_outs.id = transfer_out_items.out_id AND transfer_outs.company_id = ?)",
			itemID, middleware.GetCompanyID(c)).
		Updates(map[string]interface{}{"diff_status": "handled", "diff_result": req.Result})
	if res.RowsAffected == 0 {
		response.NotFound(c, "差异行不存在")
		return
	}
	response.OkWithMessage(c, "处理成功", nil)
}
