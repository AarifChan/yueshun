package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/response"
)

// EcosystemHandler 生态互联（供应商连接/商品接收/商城装修接收）
type EcosystemHandler struct {
	db *gorm.DB
}

func NewEcosystemHandler(db *gorm.DB) *EcosystemHandler {
	return &EcosystemHandler{db: db}
}

func (h *EcosystemHandler) RegisterRoutes(r *gin.RouterGroup) {
	sc := r.Group("/supplier-connections")
	{
		sc.GET("", h.ConnectionList)
		sc.POST("", h.ConnectionCreate)
		sc.PUT("/:id", h.ConnectionUpdate)
		sc.PUT("/:id/connect", h.ConnectionConnect)
		sc.PUT("/:id/disconnect", h.ConnectionDisconnect)
		sc.DELETE("/:id", h.ConnectionDelete)
	}
	rg := r.Group("/received-goods")
	{
		rg.GET("", h.GoodsList)
		rg.POST("", h.GoodsCreate)
		rg.PUT("/:id/receive", h.GoodsReceive)
		rg.PUT("/:id/close", h.GoodsClose)
		rg.PUT("/:id/reopen", h.GoodsReopen)
	}
	rd := r.Group("/received-decorations")
	{
		rd.GET("", h.DecorationList)
		rd.POST("", h.DecorationCreate)
		rd.PUT("/:id/receive", h.DecorationReceive)
		rd.PUT("/:id/close", h.DecorationClose)
		rd.POST("/receive-all", h.DecorationReceiveAll)
	}
}

// ==================== 供应商连接 ====================

func (h *EcosystemHandler) ConnectionList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "30"))
	companyID := middleware.GetCompanyID(c)
	keyword := c.Query("keyword")
	status := c.Query("status")

	q := h.db.Model(&model.SupplierConnection{}).Where("company_id = ?", companyID)
	if keyword != "" {
		q = q.Where("corp_name ILIKE ? OR corp_code ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	q.Count(&total)
	var list []model.SupplierConnection
	q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	supplierNames := map[uint]string{}
	var suppliers []model.Supplier
	h.db.Where("company_id = ?", companyID).Find(&suppliers)
	for _, s := range suppliers {
		supplierNames[s.ID] = s.Name
	}

	rows := make([]gin.H, 0, len(list))
	for _, x := range list {
		rows = append(rows, gin.H{
			"id": x.ID, "corpName": x.CorpName, "corpCode": x.CorpCode,
			"supplierId": x.SupplierID, "supplier": supplierNames[x.SupplierID],
			"status": x.Status, "remark": x.Remark, "createdAt": x.CreatedAt,
		})
	}
	response.OkWithPage(c, rows, page, pageSize, int(total))
}

func (h *EcosystemHandler) ConnectionCreate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var x model.SupplierConnection
	if err := c.ShouldBindJSON(&x); err != nil || x.CorpName == "" {
		response.Fail(c, 4000, "上游企业名称不能为空")
		return
	}
	x.ID = 0
	x.CompanyID = companyID
	if x.Status == "" {
		x.Status = "pending"
	}
	if err := h.db.Create(&x).Error; err != nil {
		response.Fail(c, 4000, "创建失败: "+err.Error())
		return
	}
	response.Ok(c, x)
}

func (h *EcosystemHandler) ConnectionUpdate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	var x model.SupplierConnection
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&x).Error; err != nil {
		response.Fail(c, 4004, "连接不存在")
		return
	}
	var req model.SupplierConnection
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误")
		return
	}
	x.CorpName = req.CorpName
	x.CorpCode = req.CorpCode
	x.SupplierID = req.SupplierID
	x.Remark = req.Remark
	h.db.Save(&x)
	response.Ok(c, x)
}

func (h *EcosystemHandler) ConnectionConnect(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		SupplierID uint `json:"supplierId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "请选择关联供应商")
		return
	}
	result := h.db.Model(&model.SupplierConnection{}).
		Where("id = ? AND company_id = ?", id, companyID).
		Updates(map[string]interface{}{"status": "connected", "supplier_id": req.SupplierID})
	if result.RowsAffected == 0 {
		response.Fail(c, 4004, "连接不存在")
		return
	}
	response.Ok(c, nil)
}

func (h *EcosystemHandler) ConnectionDisconnect(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	result := h.db.Model(&model.SupplierConnection{}).
		Where("id = ? AND company_id = ?", id, companyID).
		Update("status", "disconnected")
	if result.RowsAffected == 0 {
		response.Fail(c, 4004, "连接不存在")
		return
	}
	response.Ok(c, nil)
}

func (h *EcosystemHandler) ConnectionDelete(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.SupplierConnection{})
	response.Ok(c, nil)
}

// ==================== 商品接收 ====================

func (h *EcosystemHandler) GoodsList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "30"))
	companyID := middleware.GetCompanyID(c)
	status := c.DefaultQuery("status", "pending")
	supplierID, _ := strconv.Atoi(c.DefaultQuery("supplierId", "0"))
	brand := c.Query("brand")
	keyword := c.Query("keyword")

	q := h.db.Model(&model.ReceivedGoods{}).Where("company_id = ?", companyID)
	if status != "all" {
		q = q.Where("status = ?", status)
	}
	if supplierID > 0 {
		q = q.Where("supplier_id = ?", supplierID)
	}
	if brand != "" {
		q = q.Where("brand ILIKE ?", "%"+brand+"%")
	}
	if keyword != "" {
		q = q.Where("name ILIKE ? OR code ILIKE ? OR barcode ILIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	var total int64
	q.Count(&total)
	var list []model.ReceivedGoods
	q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	supplierNames := map[uint]string{}
	var suppliers []model.Supplier
	h.db.Where("company_id = ?", companyID).Find(&suppliers)
	for _, s := range suppliers {
		supplierNames[s.ID] = s.Name
	}

	rows := make([]gin.H, 0, len(list))
	for _, x := range list {
		rows = append(rows, gin.H{
			"id": x.ID, "image": x.Image, "code": x.Code, "barcode": x.Barcode,
			"name": x.Name, "spec": x.Spec, "brand": x.Brand, "unit": x.Unit,
			"purchasePrice": x.PurchasePrice,
			"supplierId": x.SupplierID, "supplier": supplierNames[x.SupplierID],
			"status": x.Status, "productId": x.ProductID, "createdAt": x.CreatedAt,
		})
	}
	response.OkWithPage(c, rows, page, pageSize, int(total))
}

// GoodsCreate 上游推送商品（对接入口/手工录入）
func (h *EcosystemHandler) GoodsCreate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var x model.ReceivedGoods
	if err := c.ShouldBindJSON(&x); err != nil || x.Name == "" {
		response.Fail(c, 4000, "商品名称不能为空")
		return
	}
	x.ID = 0
	x.CompanyID = companyID
	x.Status = "pending"
	x.ProductID = 0
	if x.Unit == "" {
		x.Unit = "个"
	}
	if err := h.db.Create(&x).Error; err != nil {
		response.Fail(c, 4000, "创建失败: "+err.Error())
		return
	}
	response.Ok(c, x)
}

// GoodsReceive 确认接收：生成商品档案
func (h *EcosystemHandler) GoodsReceive(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	var g model.ReceivedGoods
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&g).Error; err != nil {
		response.Fail(c, 4004, "商品不存在")
		return
	}
	if g.Status == "received" {
		response.Fail(c, 4000, "该商品已接收")
		return
	}

	err := h.db.Transaction(func(tx *gorm.DB) error {
		// 找或建「供应商接收」分类
		var cat model.ProductCategory
		err := tx.Where("company_id = ? AND name = ?", companyID, "供应商接收").First(&cat).Error
		if err != nil {
			cat = model.ProductCategory{
				BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
				Name:                 "供应商接收",
			}
			if err := tx.Create(&cat).Error; err != nil {
				return err
			}
		}
		// 找或建品牌
		brandID := uint(0)
		if g.Brand != "" {
			var b model.Brand
			err := tx.Where("company_id = ? AND name = ?", companyID, g.Brand).First(&b).Error
			if err != nil {
				b = model.Brand{
					BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
					Name:                 g.Brand,
				}
				if err := tx.Create(&b).Error; err != nil {
					return err
				}
			}
			brandID = b.ID
		}
		code := g.Code
		if code == "" {
			code = "JS" + strconv.Itoa(int(g.ID))
		}
		p := model.Product{
			BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
			CategoryID:           cat.ID,
			BrandID:              brandID,
			Name:                 g.Name,
			Code:                 code,
			Barcode:              g.Barcode,
			Specification:        g.Spec,
			Unit:                 g.Unit,
			PurchasePrice:        g.PurchasePrice,
			Image:                g.Image,
			SupplierID:           g.SupplierID,
			Status:               1,
		}
		if err := tx.Create(&p).Error; err != nil {
			return err
		}
		return tx.Model(&g).Updates(map[string]interface{}{"status": "received", "product_id": p.ID}).Error
	})
	if err != nil {
		response.Fail(c, 4000, "接收失败: "+err.Error())
		return
	}
	response.Ok(c, nil)
}

func (h *EcosystemHandler) GoodsClose(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	result := h.db.Model(&model.ReceivedGoods{}).
		Where("id = ? AND company_id = ? AND status = 'pending'", id, companyID).
		Update("status", "closed")
	if result.RowsAffected == 0 {
		response.Fail(c, 4004, "记录不存在或已处理")
		return
	}
	response.Ok(c, nil)
}

func (h *EcosystemHandler) GoodsReopen(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	result := h.db.Model(&model.ReceivedGoods{}).
		Where("id = ? AND company_id = ? AND status = 'closed'", id, companyID).
		Update("status", "pending")
	if result.RowsAffected == 0 {
		response.Fail(c, 4004, "记录不存在或状态不正确")
		return
	}
	response.Ok(c, nil)
}

// ==================== 商城装修接收 ====================

func (h *EcosystemHandler) DecorationList(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	status := c.DefaultQuery("status", "pending")
	q := h.db.Model(&model.ReceivedDecoration{}).Where("company_id = ?", companyID)
	if status != "all" {
		q = q.Where("status = ?", status)
	}
	var list []model.ReceivedDecoration
	q.Order("id DESC").Find(&list)
	response.Ok(c, gin.H{"list": list, "total": len(list)})
}

func (h *EcosystemHandler) DecorationCreate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var x model.ReceivedDecoration
	if err := c.ShouldBindJSON(&x); err != nil || x.Title == "" {
		response.Fail(c, 4000, "装修标题不能为空")
		return
	}
	x.ID = 0
	x.CompanyID = companyID
	x.Status = "pending"
	x.ReceivedAt = nil
	if err := h.db.Create(&x).Error; err != nil {
		response.Fail(c, 4000, "创建失败: "+err.Error())
		return
	}
	response.Ok(c, x)
}

func (h *EcosystemHandler) DecorationReceive(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	now := time.Now()
	err := h.db.Transaction(func(tx *gorm.DB) error {
		var x model.ReceivedDecoration
		if err := tx.Where("id = ? AND company_id = ?", id, companyID).First(&x).Error; err != nil {
			return err
		}
		if x.Status != "pending" {
			return nil
		}
		if err := tx.Model(&x).Updates(map[string]interface{}{"status": "received", "received_at": &now}).Error; err != nil {
			return err
		}
		// 写入当前商城装修设置
		var setting model.CompanySetting
		err := tx.Where("company_id = ? AND key = ?", companyID, "mall.decoration").First(&setting).Error
		if err != nil {
			setting = model.CompanySetting{
				BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
				Key:                  "mall.decoration",
				Value:                x.Content,
			}
			return tx.Create(&setting).Error
		}
		return tx.Model(&setting).Update("value", x.Content).Error
	})
	if err != nil {
		response.Fail(c, 4000, "接收失败: "+err.Error())
		return
	}
	response.Ok(c, nil)
}

func (h *EcosystemHandler) DecorationClose(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	result := h.db.Model(&model.ReceivedDecoration{}).
		Where("id = ? AND company_id = ? AND status = 'pending'", id, companyID).
		Update("status", "closed")
	if result.RowsAffected == 0 {
		response.Fail(c, 4004, "记录不存在或已处理")
		return
	}
	response.Ok(c, nil)
}

// DecorationReceiveAll 一键接收全部待接收装修
func (h *EcosystemHandler) DecorationReceiveAll(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var list []model.ReceivedDecoration
	h.db.Where("company_id = ? AND status = 'pending'", companyID).Find(&list)
	if len(list) == 0 {
		response.Ok(c, gin.H{"received": 0})
		return
	}
	now := time.Now()
	latest := list[0]
	for _, x := range list {
		if x.ID > latest.ID {
			latest = x
		}
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.ReceivedDecoration{}).
			Where("company_id = ? AND status = 'pending'", companyID).
			Updates(map[string]interface{}{"status": "received", "received_at": &now}).Error; err != nil {
			return err
		}
		var setting model.CompanySetting
		err := tx.Where("company_id = ? AND key = ?", companyID, "mall.decoration").First(&setting).Error
		if err != nil {
			setting = model.CompanySetting{
				BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
				Key:                  "mall.decoration",
				Value:                latest.Content,
			}
			return tx.Create(&setting).Error
		}
		return tx.Model(&setting).Update("value", latest.Content).Error
	})
	if err != nil {
		response.Fail(c, 4000, "接收失败: "+err.Error())
		return
	}
	response.Ok(c, gin.H{"received": len(list)})
}
