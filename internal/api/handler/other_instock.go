package handler

import (
	"errors"
	"fmt"
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

// OtherInStockHandler 其他入库单处理器
type OtherInStockHandler struct {
	db *gorm.DB
}

func NewOtherInStockHandler(db *gorm.DB) *OtherInStockHandler {
	return &OtherInStockHandler{db: db}
}

func (h *OtherInStockHandler) RegisterRoutes(r *gin.RouterGroup) {
	g := r.Group("/other-in-stocks")
	{
		g.GET("", h.List)
		g.POST("", h.Create)
		g.GET("/:id", h.Get)
		g.PUT("/:id", h.Update)
		g.DELETE("/:id", h.Delete)
		g.PUT("/:id/complete", h.Complete)
	}
}

type OtherInStockListReq struct {
	Page        int    `form:"page,default=1"`
	PageSize    int    `form:"pageSize,default=20"`
	Keyword     string `form:"keyword"`     // 单号
	ProductKw   string `form:"productKw"`   // 商品名称/编号
	WarehouseID uint   `form:"warehouseId"`
	Status      string `form:"status"`
	InType      string `form:"inType"`
	StartDate   string `form:"startDate"` // 录单时间起
	EndDate     string `form:"endDate"`   // 录单时间止
}

type OtherInStockListResp struct {
	model.OtherInStock
	WarehouseName string `json:"warehouseName"`
	OperatorName  string `json:"operatorName"`
	HandlerName   string `json:"handlerName"`
	DeptName      string `json:"deptName"`
}

// List 其他入库单列表
// @Summary 其他入库单列表
// @Tags 库存
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param keyword query string false "单号"
// @Param productKw query string false "商品名称/编号"
// @Param warehouseId query int false "入库仓库"
// @Param status query string false "状态 draft/completed"
// @Param inType query string false "入库类型"
// @Param startDate query string false "录单时间起 YYYY-MM-DD"
// @Param endDate query string false "录单时间止 YYYY-MM-DD"
// @Success 200 {object} response.Response
// @Router /other-in-stocks [get]
func (h *OtherInStockHandler) List(c *gin.Context) {
	var req OtherInStockListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	applyFilters := func(q *gorm.DB) *gorm.DB {
		q = q.Where("other_in_stocks.company_id = ?", companyID)
		if req.Keyword != "" {
			q = q.Where("other_in_stocks.bill_no LIKE ?", "%"+req.Keyword+"%")
		}
		if req.WarehouseID > 0 {
			q = q.Where("other_in_stocks.warehouse_id = ?", req.WarehouseID)
		}
		if req.Status != "" {
			q = q.Where("other_in_stocks.status = ?", req.Status)
		}
		if req.InType != "" {
			q = q.Where("other_in_stocks.in_type = ?", req.InType)
		}
		if req.StartDate != "" {
			q = q.Where("other_in_stocks.created_at >= ?", req.StartDate+" 00:00:00")
		}
		if req.EndDate != "" {
			q = q.Where("other_in_stocks.created_at <= ?", req.EndDate+" 23:59:59")
		}
		if req.ProductKw != "" {
			q = q.Where(`EXISTS (
				SELECT 1 FROM other_in_stock_items oi
				JOIN products p ON p.id = oi.product_id
				WHERE oi.in_stock_id = other_in_stocks.id
				  AND (p.name LIKE ? OR p.code LIKE ?))`, "%"+req.ProductKw+"%", "%"+req.ProductKw+"%")
		}
		return q
	}

	var total int64
	applyFilters(h.db.Model(&model.OtherInStock{})).Count(&total)

	var list []OtherInStockListResp
	if err := applyFilters(h.db.Model(&model.OtherInStock{})).
		Select(`other_in_stocks.*,
			COALESCE(w.name, '') AS warehouse_name,
			COALESCE(op.name, '') AS operator_name,
			COALESCE(hd.name, '') AS handler_name,
			COALESCE(d.name, '') AS dept_name`).
		Joins("LEFT JOIN warehouses w ON w.id = other_in_stocks.warehouse_id").
		Joins("LEFT JOIN employees op ON op.id = other_in_stocks.operator_id").
		Joins("LEFT JOIN employees hd ON hd.id = other_in_stocks.handler_id").
		Joins("LEFT JOIN departments d ON d.id = other_in_stocks.dept_id").
		Order("other_in_stocks.created_at DESC").
		Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).
		Scan(&list).Error; err != nil {
		log.Error().Err(err).Msg("list other in-stocks failed")
		response.ServerError(c, "查询失败")
		return
	}

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

// Get 其他入库单详情
// @Summary 其他入库单详情
// @Tags 库存
// @Param id path int true "单据ID"
// @Success 200 {object} response.Response
// @Router /other-in-stocks/{id} [get]
func (h *OtherInStockHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var bill model.OtherInStock
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).
		Preload("Items.Product").First(&bill).Error; err != nil {
		response.NotFound(c, "入库单不存在")
		return
	}
	response.Ok(c, bill)
}

type OtherInStockItemReq struct {
	ProductID uint    `json:"productId" binding:"required"`
	Quantity  float64 `json:"quantity" binding:"required,gt=0"`
	Price     float64 `json:"price"`
	Remark    string  `json:"remark"`
}

type OtherInStockSaveReq struct {
	WarehouseID uint                 `json:"warehouseId" binding:"required"`
	BillDate    string               `json:"billDate" binding:"required"`
	InType      string               `json:"inType" binding:"required"`
	Counterpart string               `json:"counterpart"`
	SettleUnit  string               `json:"settleUnit"`
	HandlerID   uint                 `json:"handlerId"`
	DeptID      uint                 `json:"deptId"`
	Remark      string               `json:"remark"`
	Items       []OtherInStockItemReq `json:"items" binding:"required,min=1,dive"`
	Complete    bool                 `json:"complete"` // true=保存并过账
}

// generateOtherInStockBillNo 生成 QR-YYYYMMDD-0001 格式的单号（按日递增）
func (h *OtherInStockHandler) generateBillNo(tx *gorm.DB) string {
	prefix := "QR-" + time.Now().Format("20060102") + "-"
	var count int64
	tx.Model(&model.OtherInStock{}).Where("bill_no LIKE ?", prefix+"%").Count(&count)
	return fmt.Sprintf("%s%04d", prefix, count+1)
}

func buildOtherInStockItems(billID uint, reqItems []OtherInStockItemReq) ([]model.OtherInStockItem, float64, float64) {
	var amount, totalQty float64
	items := make([]model.OtherInStockItem, len(reqItems))
	for i, item := range reqItems {
		itemAmount := item.Quantity * item.Price
		amount += itemAmount
		totalQty += item.Quantity
		items[i] = model.OtherInStockItem{
			InStockID: billID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
			Amount:    itemAmount,
			Remark:    item.Remark,
		}
	}
	return items, amount, totalQty
}

// Create 新增其他入库单
// @Summary 新增其他入库单（草稿或保存并过账）
// @Tags 库存
// @Param body body OtherInStockSaveReq true "单据内容"
// @Success 200 {object} response.Response
// @Router /other-in-stocks [post]
func (h *OtherInStockHandler) Create(c *gin.Context) {
	var req OtherInStockSaveReq
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
	bill := model.OtherInStock{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		WarehouseID:          req.WarehouseID,
		BillDate:             billDate,
		InType:               req.InType,
		Counterpart:          req.Counterpart,
		SettleUnit:           req.SettleUnit,
		HandlerID:            req.HandlerID,
		DeptID:               req.DeptID,
		Status:               "draft",
		OperatorID:           middleware.GetUserID(c),
		Remark:               req.Remark,
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		bill.BillNo = h.generateBillNo(tx)
		items, amount, totalQty := buildOtherInStockItems(0, req.Items)
		bill.Amount = amount
		bill.TotalQty = totalQty
		if req.Complete {
			bill.Status = "completed"
		}
		if err := tx.Create(&bill).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].InStockID = bill.ID
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		if req.Complete {
			for _, item := range items {
				if err := database.ChangeStock(tx, companyID, bill.WarehouseID, item.ProductID, item.Quantity); err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("create other in-stock failed")
		response.ServerError(c, "保存失败")
		return
	}

	response.Ok(c, bill)
}

// Update 编辑其他入库单（仅草稿）
// @Summary 编辑其他入库单
// @Tags 库存
// @Param id path int true "单据ID"
// @Param body body OtherInStockSaveReq true "单据内容"
// @Success 200 {object} response.Response
// @Router /other-in-stocks/{id} [put]
func (h *OtherInStockHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req OtherInStockSaveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var bill model.OtherInStock
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&bill).Error; err != nil {
		response.NotFound(c, "入库单不存在")
		return
	}
	if bill.Status != "draft" {
		response.Fail(c, response.CodeBadRequest, "已过账的单据不可编辑")
		return
	}

	billDate, _ := time.Parse("2006-01-02", req.BillDate)
	bill.WarehouseID = req.WarehouseID
	bill.BillDate = billDate
	bill.InType = req.InType
	bill.Counterpart = req.Counterpart
	bill.SettleUnit = req.SettleUnit
	bill.HandlerID = req.HandlerID
	bill.DeptID = req.DeptID
	bill.Remark = req.Remark

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		items, amount, totalQty := buildOtherInStockItems(bill.ID, req.Items)
		bill.Amount = amount
		bill.TotalQty = totalQty
		if req.Complete {
			bill.Status = "completed"
		}
		if err := tx.Save(&bill).Error; err != nil {
			return err
		}
		if err := tx.Where("in_stock_id = ?", bill.ID).Delete(&model.OtherInStockItem{}).Error; err != nil {
			return err
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		if req.Complete {
			for _, item := range items {
				if err := database.ChangeStock(tx, companyID, bill.WarehouseID, item.ProductID, item.Quantity); err != nil {
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
		log.Error().Err(err).Msg("update other in-stock failed")
		response.ServerError(c, "保存失败")
		return
	}

	response.Ok(c, bill)
}

// Delete 删除其他入库单（仅草稿）
// @Summary 删除其他入库单
// @Tags 库存
// @Param id path int true "单据ID"
// @Success 200 {object} response.Response
// @Router /other-in-stocks/{id} [delete]
func (h *OtherInStockHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var bill model.OtherInStock
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&bill).Error; err != nil {
		response.NotFound(c, "入库单不存在")
		return
	}
	if bill.Status != "draft" {
		response.Fail(c, response.CodeBadRequest, "已过账的单据不可删除")
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("in_stock_id = ?", bill.ID).Delete(&model.OtherInStockItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&bill).Error
	}); err != nil {
		log.Error().Err(err).Msg("delete other in-stock failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

// Complete 过账（增加库存）
// @Summary 其他入库单过账
// @Tags 库存
// @Param id path int true "单据ID"
// @Success 200 {object} response.Response
// @Router /other-in-stocks/{id}/complete [put]
func (h *OtherInStockHandler) Complete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var bill model.OtherInStock
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Preload("Items").First(&bill).Error; err != nil {
		response.NotFound(c, "入库单不存在")
		return
	}
	if bill.Status != "draft" {
		response.Fail(c, response.CodeBadRequest, "只有草稿状态的单据可过账")
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		bill.Status = "completed"
		if err := tx.Save(&bill).Error; err != nil {
			return err
		}
		for _, item := range bill.Items {
			if err := database.ChangeStock(tx, companyID, bill.WarehouseID, item.ProductID, item.Quantity); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		if errors.Is(err, database.ErrStockNotEnough) {
			response.Fail(c, response.CodeBadRequest, err.Error())
			return
		}
		log.Error().Err(err).Msg("complete other in-stock failed")
		response.ServerError(c, "过账失败")
		return
	}
	response.OkWithMessage(c, "过账成功", nil)
}
