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

// MallHandler 商城处理器
type MallHandler struct {
	db *gorm.DB
}

func NewMallHandler(db *gorm.DB) *MallHandler {
	return &MallHandler{db: db}
}

func (h *MallHandler) RegisterRoutes(r *gin.RouterGroup) {
	mp := r.Group("/mall-products")
	{
		mp.GET("", h.ListMallProducts)
		mp.POST("", h.CreateMallProduct)
		mp.GET("/:id", h.GetMallProduct)
		mp.PUT("/:id", h.UpdateMallProduct)
		mp.DELETE("/:id", h.DeleteMallProduct)
	}

	mc := r.Group("/mall-carts")
	{
		mc.GET("", h.ListCarts)
		mc.POST("", h.AddCart)
		mc.PUT("/:id", h.UpdateCart)
		mc.DELETE("/:id", h.DeleteCart)
	}

	mo := r.Group("/mall-orders")
	{
		mo.GET("", h.ListMallOrders)
		mo.POST("", h.CreateMallOrder)
		mo.GET("/:id", h.GetMallOrder)
		mo.PUT("/:id/pay", h.PayMallOrder)
		mo.PUT("/:id/ship", h.ShipMallOrder)
		mo.PUT("/:id/complete", h.CompleteMallOrder)
		mo.PUT("/:id/cancel", h.CancelMallOrder)
	}
}

type MallProductListReq struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"pageSize,default=20"`
	Keyword  string `form:"keyword"`
	Status   int8   `form:"status"`
}

func (h *MallHandler) ListMallProducts(c *gin.Context) {
	var req MallProductListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.MallProduct{}).Where("company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("name LIKE ?", "%"+req.Keyword+"%")
	}
	if req.Status != 0 {
		query = query.Where("status = ?", req.Status)
	}

	var total int64
	query.Count(&total)

	var list []model.MallProduct
	query.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *MallHandler) GetMallProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var product model.MallProduct
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&product).Error; err != nil {
		response.NotFound(c, "商品不存在")
		return
	}
	response.Ok(c, product)
}

type CreateMallProductReq struct {
	ProductID   uint    `json:"productId" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	ImageURL    string  `json:"imageUrl"`
	Price       float64 `json:"price" binding:"required,gte=0"`
}

func (h *MallHandler) CreateMallProduct(c *gin.Context) {
	var req CreateMallProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	product := model.MallProduct{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		ProductID:            req.ProductID,
		Name:                 req.Name,
		Description:          req.Description,
		ImageURL:             req.ImageURL,
		Price:                req.Price,
		Status:               1,
	}

	if err := h.db.Create(&product).Error; err != nil {
		log.Error().Err(err).Msg("create mall product failed")
		response.ServerError(c, "创建失败")
		return
	}

	response.Ok(c, product)
}

func (h *MallHandler) UpdateMallProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req CreateMallProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var product model.MallProduct
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&product).Error; err != nil {
		response.NotFound(c, "商品不存在")
		return
	}

	product.ProductID = req.ProductID
	product.Name = req.Name
	product.Description = req.Description
	product.ImageURL = req.ImageURL
	product.Price = req.Price

	if err := h.db.Save(&product).Error; err != nil {
		log.Error().Err(err).Msg("update mall product failed")
		response.ServerError(c, "更新失败")
		return
	}

	response.Ok(c, product)
}

func (h *MallHandler) DeleteMallProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var product model.MallProduct
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&product).Error; err != nil {
		response.NotFound(c, "商品不存在")
		return
	}

	if err := h.db.Delete(&product).Error; err != nil {
		log.Error().Err(err).Msg("delete mall product failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

// ==================== 购物车 ====================

func (h *MallHandler) ListCarts(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}
	customerID, _ := strconv.ParseUint(c.Query("customerId"), 10, 64)

	query := h.db.Model(&model.MallCart{}).Where("company_id = ?", companyID)
	if customerID > 0 {
		query = query.Where("customer_id = ?", customerID)
	}

	var list []model.MallCart
	query.Find(&list)
	response.Ok(c, list)
}

type AddCartReq struct {
	CustomerID uint    `json:"customerId" binding:"required"`
	ProductID  uint    `json:"productId" binding:"required"`
	Quantity   float64 `json:"quantity" binding:"required,gt=0"`
	Price      float64 `json:"price" binding:"required,gte=0"`
}

func (h *MallHandler) AddCart(c *gin.Context) {
	var req AddCartReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	cart := model.MallCart{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		CustomerID:           req.CustomerID,
		ProductID:            req.ProductID,
		Quantity:             req.Quantity,
		Price:                req.Price,
	}

	if err := h.db.Create(&cart).Error; err != nil {
		log.Error().Err(err).Msg("add cart failed")
		response.ServerError(c, "添加失败")
		return
	}

	response.Ok(c, cart)
}

func (h *MallHandler) UpdateCart(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req AddCartReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var cart model.MallCart
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&cart).Error; err != nil {
		response.NotFound(c, "购物车不存在")
		return
	}

	cart.Quantity = req.Quantity
	cart.Price = req.Price

	if err := h.db.Save(&cart).Error; err != nil {
		log.Error().Err(err).Msg("update cart failed")
		response.ServerError(c, "更新失败")
		return
	}

	response.Ok(c, cart)
}

func (h *MallHandler) DeleteCart(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.MallCart{}).Error; err != nil {
		log.Error().Err(err).Msg("delete cart failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

// ==================== 商城订单 ====================

type MallOrderListReq struct {
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"pageSize,default=20"`
	Keyword    string `form:"keyword"`
	CustomerID uint   `form:"customerId"`
	Status     string `form:"status"`
}

func (h *MallHandler) ListMallOrders(c *gin.Context) {
	var req MallOrderListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.MallOrder{}).Where("company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("order_no LIKE ?", "%"+req.Keyword+"%")
	}
	if req.CustomerID > 0 {
		query = query.Where("customer_id = ?", req.CustomerID)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	var total int64
	query.Count(&total)

	var list []model.MallOrder
	query.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *MallHandler) GetMallOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var order model.MallOrder
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Preload("Items.Product").First(&order).Error; err != nil {
		response.NotFound(c, "订单不存在")
		return
	}
	response.Ok(c, order)
}

type MallOrderItemReq struct {
	ProductID uint    `json:"productId" binding:"required"`
	Quantity  float64 `json:"quantity" binding:"required,gt=0"`
	Price     float64 `json:"price" binding:"required,gte=0"`
	Remark    string  `json:"remark"`
}

type CreateMallOrderReq struct {
	CustomerID uint               `json:"customerId" binding:"required"`
	Remark     string             `json:"remark"`
	Items      []MallOrderItemReq `json:"items" binding:"required,min=1,dive"`
}

func (h *MallHandler) CreateMallOrder(c *gin.Context) {
	var req CreateMallOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	orderDate := time.Now()
	orderNo := generateOrderNo("MO")

	order := model.MallOrder{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		CustomerID:           req.CustomerID,
		OrderNo:              orderNo,
		OrderDate:            orderDate,
		Status:               "pending",
		OperatorID:           middleware.GetUserID(c),
		Remark:               req.Remark,
	}

	var amount, totalAmount float64
	items := make([]model.MallOrderItem, len(req.Items))
	for i, item := range req.Items {
		itemAmount := item.Quantity * item.Price
		amount += itemAmount
		totalAmount += itemAmount

		items[i] = model.MallOrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
			Amount:    itemAmount,
			Remark:    item.Remark,
		}
	}
	order.Amount = amount
	order.TotalAmount = totalAmount

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&order).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].OrderID = order.ID
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("create mall order failed")
		response.ServerError(c, "创建失败")
		return
	}

	response.Ok(c, order)
}

func (h *MallHandler) PayMallOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var order model.MallOrder
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&order).Error; err != nil {
		response.NotFound(c, "订单不存在")
		return
	}
	if order.Status != "pending" {
		response.Fail(c, response.CodeBadRequest, "当前状态不可付款")
		return
	}

	order.Status = "paid"
	if err := h.db.Save(&order).Error; err != nil {
		log.Error().Err(err).Msg("pay mall order failed")
		response.ServerError(c, "付款失败")
		return
	}
	response.OkWithMessage(c, "付款成功", nil)
}

func (h *MallHandler) ShipMallOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var order model.MallOrder
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&order).Error; err != nil {
		response.NotFound(c, "订单不存在")
		return
	}
	if order.Status != "paid" {
		response.Fail(c, response.CodeBadRequest, "当前状态不可发货")
		return
	}

	order.Status = "shipped"
	if err := h.db.Save(&order).Error; err != nil {
		log.Error().Err(err).Msg("ship mall order failed")
		response.ServerError(c, "发货失败")
		return
	}
	response.OkWithMessage(c, "发货成功", nil)
}

func (h *MallHandler) CompleteMallOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var order model.MallOrder
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&order).Error; err != nil {
		response.NotFound(c, "订单不存在")
		return
	}
	if order.Status != "shipped" {
		response.Fail(c, response.CodeBadRequest, "当前状态不可完成")
		return
	}

	order.Status = "completed"
	if err := h.db.Save(&order).Error; err != nil {
		log.Error().Err(err).Msg("complete mall order failed")
		response.ServerError(c, "完成失败")
		return
	}
	response.OkWithMessage(c, "订单完成", nil)
}

func (h *MallHandler) CancelMallOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var order model.MallOrder
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&order).Error; err != nil {
		response.NotFound(c, "订单不存在")
		return
	}
	if order.Status != "pending" && order.Status != "paid" {
		response.Fail(c, response.CodeBadRequest, "当前状态不可取消")
		return
	}

	order.Status = "cancelled"
	if err := h.db.Save(&order).Error; err != nil {
		log.Error().Err(err).Msg("cancel mall order failed")
		response.ServerError(c, "取消失败")
		return
	}
	response.OkWithMessage(c, "取消成功", nil)
}
