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

// InventoryHandler 库存模块处理器
type InventoryHandler struct {
	db *gorm.DB
}

func NewInventoryHandler(db *gorm.DB) *InventoryHandler {
	return &InventoryHandler{db: db}
}

func (h *InventoryHandler) RegisterRoutes(r *gin.RouterGroup) {
	ic := r.Group("/inventory-checks")
	{
		ic.GET("", h.ListChecks)
		ic.POST("", h.CreateCheck)
		ic.GET("/:id", h.GetCheck)
		ic.PUT("/:id", h.UpdateCheck)
		ic.DELETE("/:id", h.DeleteCheck)
		ic.PUT("/:id/complete", h.CompleteCheck)
	}

	it := r.Group("/inventory-transfers")
	{
		it.GET("", h.ListTransfers)
		it.POST("", h.CreateTransfer)
		it.GET("/:id", h.GetTransfer)
		it.PUT("/:id", h.UpdateTransfer)
		it.DELETE("/:id", h.DeleteTransfer)
		it.PUT("/:id/complete", h.CompleteTransfer)
	}

	iw := r.Group("/inventory-warnings")
	{
		iw.GET("", h.ListWarnings)
		iw.POST("", h.CreateWarning)
		iw.POST("/refresh", h.RefreshWarnings)
		iw.GET("/:id", h.GetWarning)
		iw.PUT("/:id", h.UpdateWarning)
		iw.DELETE("/:id", h.DeleteWarning)
	}
}

// ==================== 库存盘点 ====================

type CheckListReq struct {
	Page        int    `form:"page,default=1"`
	PageSize    int    `form:"pageSize,default=20"`
	Keyword     string `form:"keyword"`
	WarehouseID uint   `form:"warehouseId"`
	Status      string `form:"status"`
	StartDate   string `form:"startDate"`
	EndDate     string `form:"endDate"`
}

func (h *InventoryHandler) ListChecks(c *gin.Context) {
	var req CheckListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.InventoryCheck{}).Where("company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("bill_no LIKE ?", "%"+req.Keyword+"%")
	}
	if req.WarehouseID > 0 {
		query = query.Where("warehouse_id = ?", req.WarehouseID)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.StartDate != "" {
		query = query.Where("bill_date >= ?", req.StartDate)
	}
	if req.EndDate != "" {
		query = query.Where("bill_date <= ?", req.EndDate)
	}

	var total int64
	query.Count(&total)

	var list []model.InventoryCheck
	query.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *InventoryHandler) GetCheck(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var check model.InventoryCheck
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Preload("Items.Product").First(&check).Error; err != nil {
		response.NotFound(c, "盘点单不存在")
		return
	}
	response.Ok(c, check)
}

type CheckItemReq struct {
	ProductID  uint    `json:"productId" binding:"required"`
	PositionID uint    `json:"positionId"`
	BookQty    float64 `json:"bookQty"`
	ActualQty  float64 `json:"actualQty" binding:"required,gte=0"`
	Price      float64 `json:"price"`
	Remark     string  `json:"remark"`
}

type CreateCheckReq struct {
	WarehouseID uint         `json:"warehouseId" binding:"required"`
	BillDate    string       `json:"billDate" binding:"required"`
	Remark      string       `json:"remark"`
	Items       []CheckItemReq `json:"items" binding:"required,min=1,dive"`
}

func (h *InventoryHandler) CreateCheck(c *gin.Context) {
	var req CreateCheckReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	billDate, _ := time.Parse("2006-01-02", req.BillDate)
	billNo := generateOrderNo("IC")

	check := model.InventoryCheck{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		WarehouseID:          req.WarehouseID,
		BillNo:               billNo,
		BillDate:             billDate,
		Status:               "pending",
		OperatorID:           middleware.GetUserID(c),
		Remark:               req.Remark,
	}

	var amount float64
	items := make([]model.InventoryCheckItem, len(req.Items))
	for i, item := range req.Items {
		diffQty := item.ActualQty - item.BookQty
		itemAmount := diffQty * item.Price
		amount += itemAmount

		items[i] = model.InventoryCheckItem{
			ProductID:  item.ProductID,
			PositionID: item.PositionID,
			BookQty:    item.BookQty,
			ActualQty:  item.ActualQty,
			DiffQty:    diffQty,
			Price:      item.Price,
			Amount:     itemAmount,
			Remark:     item.Remark,
		}
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&check).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].CheckID = check.ID
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("create inventory check failed")
		response.ServerError(c, "创建失败")
		return
	}

	response.Ok(c, check)
}

func (h *InventoryHandler) UpdateCheck(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req CreateCheckReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var check model.InventoryCheck
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&check).Error; err != nil {
		response.NotFound(c, "盘点单不存在")
		return
	}
	if check.Status != "pending" {
		response.Fail(c, response.CodeBadRequest, "只有待审核状态的盘点单可编辑")
		return
	}

	billDate, _ := time.Parse("2006-01-02", req.BillDate)
	check.WarehouseID = req.WarehouseID
	check.BillDate = billDate
	check.Remark = req.Remark

	var amount float64
	items := make([]model.InventoryCheckItem, len(req.Items))
	for i, item := range req.Items {
		diffQty := item.ActualQty - item.BookQty
		itemAmount := diffQty * item.Price
		amount += itemAmount

		items[i] = model.InventoryCheckItem{
			CheckID:    uint(id),
			ProductID:  item.ProductID,
			PositionID: item.PositionID,
			BookQty:    item.BookQty,
			ActualQty:  item.ActualQty,
			DiffQty:    diffQty,
			Price:      item.Price,
			Amount:     itemAmount,
			Remark:     item.Remark,
		}
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&check).Error; err != nil {
			return err
		}
		if err := tx.Where("check_id = ?", id).Delete(&model.InventoryCheckItem{}).Error; err != nil {
			return err
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("update inventory check failed")
		response.ServerError(c, "更新失败")
		return
	}

	response.Ok(c, check)
}

func (h *InventoryHandler) DeleteCheck(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var check model.InventoryCheck
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&check).Error; err != nil {
		response.NotFound(c, "盘点单不存在")
		return
	}
	if check.Status != "pending" {
		response.Fail(c, response.CodeBadRequest, "只有待审核状态的盘点单可删除")
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("check_id = ?", id).Delete(&model.InventoryCheckItem{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&check).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("delete inventory check failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

func (h *InventoryHandler) CompleteCheck(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var check model.InventoryCheck
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Preload("Items").First(&check).Error; err != nil {
		response.NotFound(c, "盘点单不存在")
		return
	}
	if check.Status != "pending" {
		response.Fail(c, response.CodeBadRequest, "只有待审核状态的盘点单可完成")
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		check.Status = "completed"
		if err := tx.Save(&check).Error; err != nil {
			return err
		}
		for _, item := range check.Items {
			if item.DiffQty != 0 {
				if err := database.ChangeStock(tx, companyID, check.WarehouseID, item.ProductID, item.DiffQty); err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		if errors.Is(err, database.ErrStockNotEnough) {
			response.Fail(c, response.CodeBadRequest, err.Error())
			return
		}
		log.Error().Err(err).Msg("complete inventory check failed")
		response.ServerError(c, "完成盘点失败")
		return
	}
	response.OkWithMessage(c, "盘点完成", nil)
}

// ==================== 库存调拨 ====================

type TransferListReq struct {
	Page           int    `form:"page,default=1"`
	PageSize       int    `form:"pageSize,default=20"`
	Keyword        string `form:"keyword"`
	FromWarehouseID uint `form:"fromWarehouseId"`
	ToWarehouseID   uint `form:"toWarehouseId"`
	Status         string `form:"status"`
	StartDate      string `form:"startDate"`
	EndDate        string `form:"endDate"`
}

func (h *InventoryHandler) ListTransfers(c *gin.Context) {
	var req TransferListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.InventoryTransfer{}).Where("company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("bill_no LIKE ?", "%"+req.Keyword+"%")
	}
	if req.FromWarehouseID > 0 {
		query = query.Where("from_warehouse_id = ?", req.FromWarehouseID)
	}
	if req.ToWarehouseID > 0 {
		query = query.Where("to_warehouse_id = ?", req.ToWarehouseID)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.StartDate != "" {
		query = query.Where("bill_date >= ?", req.StartDate)
	}
	if req.EndDate != "" {
		query = query.Where("bill_date <= ?", req.EndDate)
	}

	var total int64
	query.Count(&total)

	var list []model.InventoryTransfer
	query.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *InventoryHandler) GetTransfer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var transfer model.InventoryTransfer
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Preload("Items.Product").First(&transfer).Error; err != nil {
		response.NotFound(c, "调拨单不存在")
		return
	}
	response.Ok(c, transfer)
}

type TransferItemReq struct {
	ProductID      uint    `json:"productId" binding:"required"`
	FromPositionID uint    `json:"fromPositionId"`
	ToPositionID   uint    `json:"toPositionId"`
	Quantity       float64 `json:"quantity" binding:"required,gt=0"`
	Price          float64 `json:"price"`
	Remark         string  `json:"remark"`
}

type CreateTransferReq struct {
	FromWarehouseID uint              `json:"fromWarehouseId" binding:"required"`
	ToWarehouseID   uint              `json:"toWarehouseId" binding:"required"`
	BillDate        string            `json:"billDate" binding:"required"`
	Remark          string            `json:"remark"`
	Items           []TransferItemReq `json:"items" binding:"required,min=1,dive"`
}

func (h *InventoryHandler) CreateTransfer(c *gin.Context) {
	var req CreateTransferReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	billDate, _ := time.Parse("2006-01-02", req.BillDate)
	billNo := generateOrderNo("IT")

	transfer := model.InventoryTransfer{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		FromWarehouseID:      req.FromWarehouseID,
		ToWarehouseID:        req.ToWarehouseID,
		BillNo:               billNo,
		BillDate:             billDate,
		Status:               "pending",
		OperatorID:           middleware.GetUserID(c),
		Remark:               req.Remark,
	}

	var amount float64
	items := make([]model.InventoryTransferItem, len(req.Items))
	for i, item := range req.Items {
		itemAmount := item.Quantity * item.Price
		amount += itemAmount

		items[i] = model.InventoryTransferItem{
			ProductID:      item.ProductID,
			FromPositionID: item.FromPositionID,
			ToPositionID:   item.ToPositionID,
			Quantity:       item.Quantity,
			Price:          item.Price,
			Amount:         itemAmount,
			Remark:         item.Remark,
		}
	}
	transfer.Amount = amount

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&transfer).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].TransferID = transfer.ID
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("create inventory transfer failed")
		response.ServerError(c, "创建失败")
		return
	}

	response.Ok(c, transfer)
}

func (h *InventoryHandler) UpdateTransfer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req CreateTransferReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var transfer model.InventoryTransfer
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&transfer).Error; err != nil {
		response.NotFound(c, "调拨单不存在")
		return
	}
	if transfer.Status != "pending" {
		response.Fail(c, response.CodeBadRequest, "只有待审核状态的调拨单可编辑")
		return
	}

	billDate, _ := time.Parse("2006-01-02", req.BillDate)
	transfer.FromWarehouseID = req.FromWarehouseID
	transfer.ToWarehouseID = req.ToWarehouseID
	transfer.BillDate = billDate
	transfer.Remark = req.Remark

	var amount float64
	items := make([]model.InventoryTransferItem, len(req.Items))
	for i, item := range req.Items {
		itemAmount := item.Quantity * item.Price
		amount += itemAmount

		items[i] = model.InventoryTransferItem{
			TransferID:     uint(id),
			ProductID:      item.ProductID,
			FromPositionID: item.FromPositionID,
			ToPositionID:   item.ToPositionID,
			Quantity:       item.Quantity,
			Price:          item.Price,
			Amount:         itemAmount,
			Remark:         item.Remark,
		}
	}
	transfer.Amount = amount

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&transfer).Error; err != nil {
			return err
		}
		if err := tx.Where("transfer_id = ?", id).Delete(&model.InventoryTransferItem{}).Error; err != nil {
			return err
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("update inventory transfer failed")
		response.ServerError(c, "更新失败")
		return
	}

	response.Ok(c, transfer)
}

func (h *InventoryHandler) DeleteTransfer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var transfer model.InventoryTransfer
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&transfer).Error; err != nil {
		response.NotFound(c, "调拨单不存在")
		return
	}
	if transfer.Status != "pending" {
		response.Fail(c, response.CodeBadRequest, "只有待审核状态的调拨单可删除")
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("transfer_id = ?", id).Delete(&model.InventoryTransferItem{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&transfer).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("delete inventory transfer failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

func (h *InventoryHandler) CompleteTransfer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var transfer model.InventoryTransfer
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Preload("Items").First(&transfer).Error; err != nil {
		response.NotFound(c, "调拨单不存在")
		return
	}
	if transfer.Status != "pending" {
		response.Fail(c, response.CodeBadRequest, "只有待审核状态的调拨单可完成")
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		transfer.Status = "completed"
		if err := tx.Save(&transfer).Error; err != nil {
			return err
		}
		for _, item := range transfer.Items {
			if err := database.ChangeStock(tx, companyID, transfer.FromWarehouseID, item.ProductID, -item.Quantity); err != nil {
				return err
			}
			if err := database.ChangeStock(tx, companyID, transfer.ToWarehouseID, item.ProductID, item.Quantity); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		if errors.Is(err, database.ErrStockNotEnough) {
			response.Fail(c, response.CodeBadRequest, err.Error())
			return
		}
		log.Error().Err(err).Msg("complete inventory transfer failed")
		response.ServerError(c, "完成调拨失败")
		return
	}
	response.OkWithMessage(c, "调拨完成", nil)
}

// ==================== 库存预警 ====================

type WarningListReq struct {
	Page        int    `form:"page,default=1"`
	PageSize    int    `form:"pageSize,default=20"`
	WarehouseID uint   `form:"warehouseId"`
	ProductID   uint   `form:"productId"`
	Status      string `form:"status"`
}

func (h *InventoryHandler) ListWarnings(c *gin.Context) {
	var req WarningListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.InventoryWarning{}).Where("company_id = ?", companyID)
	if req.WarehouseID > 0 {
		query = query.Where("warehouse_id = ?", req.WarehouseID)
	}
	if req.ProductID > 0 {
		query = query.Where("product_id = ?", req.ProductID)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	var total int64
	query.Count(&total)

	var list []model.InventoryWarning
	query.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	// 补充商品/仓库名称
	productIDs := make([]uint, 0, len(list))
	warehouseIDs := make([]uint, 0, len(list))
	for _, w := range list {
		productIDs = append(productIDs, w.ProductID)
		if w.WarehouseID > 0 {
			warehouseIDs = append(warehouseIDs, w.WarehouseID)
		}
	}
	var products []model.Product
	if len(productIDs) > 0 {
		h.db.Select("id", "name", "code").Where("id IN ?", productIDs).Find(&products)
	}
	var warehouses []model.Warehouse
	if len(warehouseIDs) > 0 {
		h.db.Select("id", "name").Where("id IN ?", warehouseIDs).Find(&warehouses)
	}
	productMap := make(map[uint]model.Product, len(products))
	for _, p := range products {
		productMap[p.ID] = p
	}
	warehouseMap := make(map[uint]string, len(warehouses))
	for _, wh := range warehouses {
		warehouseMap[wh.ID] = wh.Name
	}

	type warningItem struct {
		model.InventoryWarning
		ProductName  string `json:"productName"`
		ProductCode  string `json:"productCode"`
		WarehouseName string `json:"warehouseName"`
	}
	items := make([]warningItem, 0, len(list))
	for _, w := range list {
		item := warningItem{InventoryWarning: w}
		if p, ok := productMap[w.ProductID]; ok {
			item.ProductName = p.Name
			item.ProductCode = p.Code
		}
		if name, ok := warehouseMap[w.WarehouseID]; ok {
			item.WarehouseName = name
		}
		items = append(items, item)
	}

	response.OkWithPage(c, items, req.Page, req.PageSize, int(total))
}

func (h *InventoryHandler) GetWarning(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var warning model.InventoryWarning
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&warning).Error; err != nil {
		response.NotFound(c, "预警不存在")
		return
	}
	response.Ok(c, warning)
}

// RefreshWarnings 库存预警刷新：按当前总库存自动生成/更新/解决预警
func (h *InventoryHandler) RefreshWarnings(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	// 汇总各商品总库存
	type stockAgg struct {
		ProductID uint
		Total     float64
	}
	var aggs []stockAgg
	if err := h.db.Table("stocks").
		Select("product_id, SUM(quantity) AS total").
		Where("company_id = ?", companyID).
		Group("product_id").
		Scan(&aggs).Error; err != nil {
		log.Error().Err(err).Msg("aggregate stock failed")
		response.ServerError(c, "刷新失败")
		return
	}
	stockMap := make(map[uint]float64, len(aggs))
	for _, a := range aggs {
		stockMap[a.ProductID] = a.Total
	}

	// 所有启用商品
	var products []model.Product
	if err := h.db.Where("company_id = ? AND status = 1", companyID).Find(&products).Error; err != nil {
		log.Error().Err(err).Msg("query products failed")
		response.ServerError(c, "刷新失败")
		return
	}

	created, updated := 0, 0
	warned := make(map[uint]bool)
	for _, p := range products {
		total := stockMap[p.ID]
		warningType := ""
		if total < p.MinStock {
			warningType = "low"
		} else if p.MaxStock > 0 && total > p.MaxStock {
			warningType = "high"
		}
		if warningType == "" {
			continue
		}
		warned[p.ID] = true

		var warning model.InventoryWarning
		err := h.db.Where("company_id = ? AND product_id = ? AND status <> 'resolved'", companyID, p.ID).
			Order("created_at DESC").First(&warning).Error
		if err == nil {
			// 已有未解决预警，更新
			warning.CurrentStock = total
			warning.WarningType = warningType
			warning.Status = "active"
			if err := h.db.Save(&warning).Error; err != nil {
				log.Error().Err(err).Uint("warningID", warning.ID).Msg("update warning failed")
				response.ServerError(c, "刷新失败")
				return
			}
			updated++
		} else if err == gorm.ErrRecordNotFound {
			warning = model.InventoryWarning{
				BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
				WarehouseID:          0,
				ProductID:            p.ID,
				MinStock:             p.MinStock,
				MaxStock:             p.MaxStock,
				CurrentStock:         total,
				WarningType:          warningType,
				Status:               "active",
			}
			if err := h.db.Create(&warning).Error; err != nil {
				log.Error().Err(err).Uint("productID", p.ID).Msg("create warning failed")
				response.ServerError(c, "刷新失败")
				return
			}
			created++
		} else {
			log.Error().Err(err).Msg("query warning failed")
			response.ServerError(c, "刷新失败")
			return
		}
	}

	// 已不再满足预警条件且仍为 active 的记录标记为 resolved
	resolved := 0
	var actives []model.InventoryWarning
	if err := h.db.Where("company_id = ? AND status = 'active'", companyID).Find(&actives).Error; err != nil {
		log.Error().Err(err).Msg("query active warnings failed")
		response.ServerError(c, "刷新失败")
		return
	}
	for _, w := range actives {
		if !warned[w.ProductID] {
			if err := h.db.Model(&model.InventoryWarning{}).Where("id = ?", w.ID).UpdateColumn("status", "resolved").Error; err != nil {
				log.Error().Err(err).Uint("warningID", w.ID).Msg("resolve warning failed")
				response.ServerError(c, "刷新失败")
				return
			}
			resolved++
		}
	}

	response.Ok(c, gin.H{"created": created, "updated": updated, "resolved": resolved})
}

type CreateWarningReq struct {
	WarehouseID uint    `json:"warehouseId" binding:"required"`
	ProductID   uint    `json:"productId" binding:"required"`
	MinStock    float64 `json:"minStock"`
	MaxStock    float64 `json:"maxStock"`
	Remark      string  `json:"remark"`
}

func (h *InventoryHandler) CreateWarning(c *gin.Context) {
	var req CreateWarningReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	warning := model.InventoryWarning{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		WarehouseID:          req.WarehouseID,
		ProductID:            req.ProductID,
		MinStock:             req.MinStock,
		MaxStock:             req.MaxStock,
		Status:               "active",
		Remark:               req.Remark,
	}

	if err := h.db.Create(&warning).Error; err != nil {
		log.Error().Err(err).Msg("create inventory warning failed")
		response.ServerError(c, "创建失败")
		return
	}

	response.Ok(c, warning)
}

func (h *InventoryHandler) UpdateWarning(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req CreateWarningReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var warning model.InventoryWarning
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&warning).Error; err != nil {
		response.NotFound(c, "预警不存在")
		return
	}

	warning.WarehouseID = req.WarehouseID
	warning.ProductID = req.ProductID
	warning.MinStock = req.MinStock
	warning.MaxStock = req.MaxStock
	warning.Remark = req.Remark

	if err := h.db.Save(&warning).Error; err != nil {
		log.Error().Err(err).Msg("update inventory warning failed")
		response.ServerError(c, "更新失败")
		return
	}

	response.Ok(c, warning)
}

func (h *InventoryHandler) DeleteWarning(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var warning model.InventoryWarning
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&warning).Error; err != nil {
		response.NotFound(c, "预警不存在")
		return
	}

	if err := h.db.Delete(&warning).Error; err != nil {
		log.Error().Err(err).Msg("delete inventory warning failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}
