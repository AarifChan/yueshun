package handler

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/response"
)

// MarketingHandler 营销模块（促销/优惠券/积分/分销/储值）
type MarketingHandler struct {
	db *gorm.DB
}

func NewMarketingHandler(db *gorm.DB) *MarketingHandler {
	return &MarketingHandler{db: db}
}

func (h *MarketingHandler) RegisterRoutes(r *gin.RouterGroup) {
	p := r.Group("/promotions")
	{
		p.GET("", h.PromotionList)
		p.POST("", h.PromotionCreate)
		p.PUT("/:id", h.PromotionUpdate)
		p.DELETE("/:id", h.PromotionDelete)
	}
	cp := r.Group("/coupons")
	{
		cp.GET("", h.CouponList)
		cp.POST("", h.CouponCreate)
		cp.PUT("/:id", h.CouponUpdate)
		cp.DELETE("/:id", h.CouponDelete)
		cp.POST("/:id/grant", h.CouponGrantCreate)
		cp.GET("/grants", h.CouponGrantList)
		cp.PUT("/grants/:id/use", h.CouponGrantUse)
	}
	pr := r.Group("/point-rules")
	{
		pr.GET("", h.PointRuleList)
		pr.POST("", h.PointRuleCreate)
		pr.PUT("/:id", h.PointRuleUpdate)
		pr.DELETE("/:id", h.PointRuleDelete)
	}
	pe := r.Group("/point-exchanges")
	{
		pe.GET("", h.PointExchangeList)
		pe.POST("", h.PointExchangeCreate)
		pe.PUT("/:id", h.PointExchangeUpdate)
		pe.DELETE("/:id", h.PointExchangeDelete)
	}
	r.GET("/point-flows", h.PointFlowList)
	r.POST("/point-flows", h.PointFlowCreate)
	r.GET("/customer-points", h.CustomerPointList)

	d := r.Group("/distributors")
	{
		d.GET("", h.DistributorList)
		d.POST("", h.DistributorCreate)
		d.PUT("/:id", h.DistributorUpdate)
		d.PUT("/:id/approve", h.DistributorApprove)
		d.PUT("/:id/disable", h.DistributorDisable)
		d.DELETE("/:id", h.DistributorDelete)
	}
	dw := r.Group("/distributor-withdrawals")
	{
		dw.GET("", h.WithdrawalList)
		dw.POST("", h.WithdrawalCreate)
		dw.PUT("/:id/pay", h.WithdrawalPay)
		dw.PUT("/:id/reject", h.WithdrawalReject)
	}
	sr := r.Group("/stored-value-rules")
	{
		sr.GET("", h.StoredValueRuleList)
		sr.POST("", h.StoredValueRuleCreate)
		sr.PUT("/:id", h.StoredValueRuleUpdate)
		sr.DELETE("/:id", h.StoredValueRuleDelete)
	}
	sv := r.Group("/stored-value-records")
	{
		sv.GET("", h.StoredValueRecordList)
		sv.POST("", h.StoredValueRecordCreate)
	}
}

// ==================== 促销活动 ====================

// promoStatusOf 计算活动状态：disabled已停用 pending未开始 active进行中 finished已结束
func promoStatusOf(p *model.Promotion, now time.Time) string {
	if p.Status == 0 {
		return "disabled"
	}
	if now.Before(p.StartAt) {
		return "pending"
	}
	if now.After(p.EndAt) {
		return "finished"
	}
	return "active"
}

func promoStatusLabel(s string) string {
	switch s {
	case "disabled":
		return "已停用"
	case "pending":
		return "未开始"
	case "finished":
		return "已结束"
	default:
		return "进行中"
	}
}

type promoStats struct {
	Customers    int     `gorm:"column:customers"`
	OrderQty     float64 `gorm:"column:order_qty"`
	OrderAmount  float64 `gorm:"column:order_amount"`
	OrderBills   int     `gorm:"column:order_bills"`
	OutQty       float64 `gorm:"column:out_qty"`
	OutAmount    float64 `gorm:"column:out_amount"`
	ReturnQty    float64 `gorm:"column:return_qty"`
	ReturnAmount float64 `gorm:"column:return_amount"`
}

// statsForPromotion 统计活动时间窗内的订/出/退（指定商品则按明细过滤，否则整单统计）
func (h *MarketingHandler) statsForPromotion(companyID uint, p *model.Promotion) promoStats {
	var st promoStats
	prodIDs := parseIDList(p.ProductIDs)
	start := p.StartAt.Format("2006-01-02")
	end := p.EndAt.Format("2006-01-02")

	type agg struct {
		Customers int     `gorm:"column:customers"`
		Bills     int     `gorm:"column:bills"`
		Qty       float64 `gorm:"column:qty"`
		Amount    float64 `gorm:"column:amount"`
	}
	query := func(headTable, itemTable, fk, dateCol, statusCond string, ids []uint) agg {
		var a agg
		if len(ids) > 0 {
			h.db.Raw(`SELECT COUNT(DISTINCT b.customer_id) AS customers, COUNT(DISTINCT b.id) AS bills,
				COALESCE(SUM(i.quantity),0) AS qty, COALESCE(SUM(i.amount),0) AS amount
				FROM `+headTable+` b JOIN `+itemTable+` i ON i.`+fk+` = b.id
				WHERE b.company_id = ? AND b.deleted_at IS NULL AND `+statusCond+`
				AND b.`+dateCol+` >= ? AND b.`+dateCol+` <= ? AND i.product_id IN ?`,
				companyID, start, end, ids).Scan(&a)
		} else {
			h.db.Raw(`SELECT COUNT(DISTINCT b.customer_id) AS customers, COUNT(DISTINCT b.id) AS bills,
				COALESCE((SELECT SUM(i.quantity) FROM `+itemTable+` i WHERE i.`+fk+` = b.id),0) AS qty,
				COALESCE(SUM(b.total_amount),0) AS amount
				FROM `+headTable+` b
				WHERE b.company_id = ? AND b.deleted_at IS NULL AND `+statusCond+`
				AND b.`+dateCol+` >= ? AND b.`+dateCol+` <= ?`,
				companyID, start, end).Scan(&a)
		}
		return a
	}

	ord := query("sales_orders", "sales_order_items", "order_id", "order_date", "b.status NOT IN ('draft','cancelled')", prodIDs)
	out := query("sales_out_stocks", "sales_out_stock_items", "out_stock_id", "bill_date", "b.status = 'completed'", prodIDs)
	ret := query("sales_returns", "sales_return_items", "return_id", "bill_date", "b.status = 'completed'", prodIDs)

	st.Customers = ord.Customers
	st.OrderBills = ord.Bills
	st.OrderQty = ord.Qty
	st.OrderAmount = ord.Amount
	st.OutQty = out.Qty
	st.OutAmount = out.Amount
	st.ReturnQty = ret.Qty
	st.ReturnAmount = ret.Amount
	return st
}

func parseIDList(s string) []uint {
	ids := []uint{}
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if id, err := strconv.Atoi(part); err == nil && id > 0 {
			ids = append(ids, uint(id))
		}
	}
	return ids
}

func (h *MarketingHandler) PromotionList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "30"))
	companyID := middleware.GetCompanyID(c)
	promoType := c.Query("type")
	name := c.Query("name")
	statusFilter := c.DefaultQuery("status", "all")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	productKeyword := c.Query("productKeyword")

	q := h.db.Model(&model.Promotion{}).Where("company_id = ?", companyID)
	if promoType != "" {
		q = q.Where("type = ?", promoType)
	}
	if name != "" {
		q = q.Where("name ILIKE ?", "%"+name+"%")
	}
	// 活动时间筛选：与所选区间有交集
	if startDate != "" {
		q = q.Where("end_at >= ?", startDate)
	}
	if endDate != "" {
		q = q.Where("start_at <= ?", endDate+" 23:59:59")
	}
	if productKeyword != "" {
		var ids []uint
		h.db.Model(&model.Product{}).Where("company_id = ? AND (name ILIKE ? OR code ILIKE ?)", companyID, "%"+productKeyword+"%", "%"+productKeyword+"%").Pluck("id", &ids)
		if len(ids) == 0 {
			response.Ok(c, paginateSaleRows([]gin.H{}, page, pageSize))
			return
		}
		likeConds := make([]string, 0, len(ids))
		args := make([]interface{}, 0, len(ids))
		for _, id := range ids {
			likeConds = append(likeConds, "(',' || COALESCE(product_ids,'') || ',') ILIKE ?")
			args = append(args, "%,"+strconv.Itoa(int(id))+",%")
		}
		q = q.Where(strings.Join(likeConds, " OR "), args...)
	}

	var total int64
	q.Count(&total)
	var list []model.Promotion
	q.Order("sort, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	now := time.Now()
	rows := make([]gin.H, 0, len(list))
	for i := range list {
		p := &list[i]
		st := h.statsForPromotion(companyID, p)
		status := promoStatusOf(p, now)
		remainDays := 0
		if status == "active" {
			remainDays = int(math.Ceil(time.Until(p.EndAt).Hours() / 24))
			if remainDays < 0 {
				remainDays = 0
			}
		}
		rows = append(rows, gin.H{
			"id": p.ID, "type": p.Type, "name": p.Name,
			"createdAt": p.CreatedAt, "startAt": p.StartAt, "endAt": p.EndAt,
			"status": status, "statusLabel": promoStatusLabel(status), "remainDays": remainDays,
			"productIds": p.ProductIDs, "productCount": len(parseIDList(p.ProductIDs)),
			"giftQty": p.GiftQty, "sort": p.Sort, "remark": p.Remark,
			"customers": st.Customers,
			"orderQty":  st.OrderQty, "orderAmount": st.OrderAmount, "orderBills": st.OrderBills,
			"outQty": st.OutQty, "outAmount": st.OutAmount,
			"returnQty": st.ReturnQty, "returnAmount": st.ReturnAmount,
			"actualQty": st.OutQty - st.ReturnQty, "actualAmount": st.OutAmount - st.ReturnAmount,
		})
	}
	// 活动状态过滤（内存过滤，因状态是计算值）
	if statusFilter != "" && statusFilter != "all" {
		filtered := make([]gin.H, 0, len(rows))
		for _, r := range rows {
			if r["status"] == statusFilter {
				filtered = append(filtered, r)
			}
		}
		rows = filtered
	}
	response.OkWithPage(c, rows, page, pageSize, int(total))
}

type promotionReq struct {
	Type       string  `json:"type" binding:"required"`
	Name       string  `json:"name" binding:"required"`
	StartAt    string  `json:"startAt" binding:"required"`
	EndAt      string  `json:"endAt" binding:"required"`
	ProductIDs string  `json:"productIds"`
	GiftQty    float64 `json:"giftQty"`
	Sort       int     `json:"sort"`
	Status     *int8   `json:"status"`
	Remark     string  `json:"remark"`
}

func parsePromoTime(s string) (time.Time, error) {
	if t, err := time.ParseInLocation("2006-01-02 15:04:05", s, time.Local); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation("2006-01-02", s, time.Local); err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339, s)
}

func (h *MarketingHandler) PromotionCreate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var req promotionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误: "+err.Error())
		return
	}
	startAt, err1 := parsePromoTime(req.StartAt)
	endAt, err2 := parsePromoTime(req.EndAt)
	if err1 != nil || err2 != nil || endAt.Before(startAt) {
		response.Fail(c, 4000, "活动时间无效")
		return
	}
	status := int8(1)
	if req.Status != nil {
		status = *req.Status
	}
	p := model.Promotion{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		Type:                 req.Type, Name: req.Name, StartAt: startAt, EndAt: endAt,
		ProductIDs: req.ProductIDs, GiftQty: req.GiftQty, Sort: req.Sort, Status: status, Remark: req.Remark,
	}
	if err := h.db.Create(&p).Error; err != nil {
		response.Fail(c, 4000, "创建失败: "+err.Error())
		return
	}
	response.Ok(c, p)
}

func (h *MarketingHandler) PromotionUpdate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	var p model.Promotion
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&p).Error; err != nil {
		response.Fail(c, 4004, "活动不存在")
		return
	}
	var req promotionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误: "+err.Error())
		return
	}
	startAt, err1 := parsePromoTime(req.StartAt)
	endAt, err2 := parsePromoTime(req.EndAt)
	if err1 != nil || err2 != nil || endAt.Before(startAt) {
		response.Fail(c, 4000, "活动时间无效")
		return
	}
	p.Name = req.Name
	p.StartAt = startAt
	p.EndAt = endAt
	p.ProductIDs = req.ProductIDs
	p.GiftQty = req.GiftQty
	p.Sort = req.Sort
	if req.Status != nil {
		p.Status = *req.Status
	}
	p.Remark = req.Remark
	h.db.Save(&p)
	response.Ok(c, p)
}

func (h *MarketingHandler) PromotionDelete(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.Promotion{})
	response.Ok(c, nil)
}

// ==================== 优惠券 ====================

func (h *MarketingHandler) CouponList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "30"))
	companyID := middleware.GetCompanyID(c)
	name := c.Query("name")
	status := c.DefaultQuery("status", "")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	q := h.db.Model(&model.Coupon{}).Where("company_id = ?", companyID)
	if name != "" {
		q = q.Where("name ILIKE ?", "%"+name+"%")
	}
	if status == "enabled" {
		q = q.Where("status = 1")
	} else if status == "disabled" {
		q = q.Where("status = 0")
	}
	if startDate != "" {
		q = q.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		q = q.Where("created_at <= ?", endDate+" 23:59:59")
	}
	var total int64
	q.Count(&total)
	var list []model.Coupon
	q.Order("sort, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	// 已领取/已使用统计
	type cnt struct {
		CouponID uint `gorm:"column:coupon_id"`
		Received int  `gorm:"column:received"`
		Used     int  `gorm:"column:used"`
	}
	counts := map[uint]cnt{}
	var cnts []cnt
	h.db.Raw(`SELECT coupon_id, COUNT(*) AS received, COALESCE(SUM(CASE WHEN status='used' THEN 1 ELSE 0 END),0) AS used
		FROM coupon_grants WHERE company_id = ? AND deleted_at IS NULL GROUP BY coupon_id`, companyID).Scan(&cnts)
	for _, x := range cnts {
		counts[x.CouponID] = x
	}

	rows := make([]gin.H, 0, len(list))
	for _, cp := range list {
		cc := counts[cp.ID]
		rows = append(rows, gin.H{
			"id": cp.ID, "name": cp.Name, "createdAt": cp.CreatedAt,
			"faceValue": cp.FaceValue, "minAmount": cp.MinAmount,
			"totalQty": cp.TotalQty, "validStart": cp.ValidStart, "validEnd": cp.ValidEnd,
			"status": cp.Status, "sort": cp.Sort, "remark": cp.Remark,
			"received": cc.Received, "used": cc.Used,
		})
	}
	response.OkWithPage(c, rows, page, pageSize, int(total))
}

func (h *MarketingHandler) CouponCreate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var cp model.Coupon
	if err := c.ShouldBindJSON(&cp); err != nil || cp.Name == "" {
		response.Fail(c, 4000, "优惠券名称不能为空")
		return
	}
	cp.ID = 0
	cp.CompanyID = companyID
	if err := h.db.Create(&cp).Error; err != nil {
		response.Fail(c, 4000, "创建失败: "+err.Error())
		return
	}
	response.Ok(c, cp)
}

func (h *MarketingHandler) CouponUpdate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	var cp model.Coupon
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&cp).Error; err != nil {
		response.Fail(c, 4004, "优惠券不存在")
		return
	}
	var req model.Coupon
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误")
		return
	}
	cp.Name = req.Name
	cp.FaceValue = req.FaceValue
	cp.MinAmount = req.MinAmount
	cp.TotalQty = req.TotalQty
	cp.ValidStart = req.ValidStart
	cp.ValidEnd = req.ValidEnd
	cp.Status = req.Status
	cp.Sort = req.Sort
	cp.Remark = req.Remark
	h.db.Save(&cp)
	response.Ok(c, cp)
}

func (h *MarketingHandler) CouponDelete(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.Coupon{})
	response.Ok(c, nil)
}

// CouponGrantCreate 发放优惠券
func (h *MarketingHandler) CouponGrantCreate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	var cp model.Coupon
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&cp).Error; err != nil {
		response.Fail(c, 4004, "优惠券不存在")
		return
	}
	var req struct {
		CustomerIDs []uint `json:"customerIds" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.CustomerIDs) == 0 {
		response.Fail(c, 4000, "请选择客户")
		return
	}
	// 总量限制
	if cp.TotalQty > 0 {
		var received int64
		h.db.Model(&model.CouponGrant{}).Where("company_id = ? AND coupon_id = ?", companyID, cp.ID).Count(&received)
		if int(received)+len(req.CustomerIDs) > cp.TotalQty {
			response.Fail(c, 4000, "超出优惠券发放总量")
			return
		}
	}
	now := time.Now()
	created := 0
	for _, customerID := range req.CustomerIDs {
		var exist int64
		h.db.Model(&model.CouponGrant{}).Where("company_id = ? AND coupon_id = ? AND customer_id = ?", companyID, cp.ID, customerID).Count(&exist)
		if exist > 0 {
			continue
		}
		g := model.CouponGrant{
			BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
			CouponID:             cp.ID,
			CustomerID:           customerID,
			Status:               "unused",
			ReceivedAt:           now,
		}
		if err := h.db.Create(&g).Error; err == nil {
			created++
		}
	}
	response.Ok(c, gin.H{"created": created})
}

func (h *MarketingHandler) CouponGrantList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "30"))
	companyID := middleware.GetCompanyID(c)
	couponID, _ := strconv.Atoi(c.DefaultQuery("couponId", "0"))
	status := c.Query("status")
	keyword := c.Query("keyword")

	q := h.db.Model(&model.CouponGrant{}).Where("coupon_grants.company_id = ?", companyID)
	if couponID > 0 {
		q = q.Where("coupon_grants.coupon_id = ?", couponID)
	}
	if status != "" {
		q = q.Where("coupon_grants.status = ?", status)
	}
	if keyword != "" {
		q = q.Joins("JOIN customers ON customers.id = coupon_grants.customer_id").
			Where("customers.name ILIKE ? OR customers.code ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	var total int64
	q.Count(&total)
	type row struct {
		ID         uint       `gorm:"column:id" json:"id"`
		CouponID   uint       `gorm:"column:coupon_id" json:"couponId"`
		CouponName string     `gorm:"column:coupon_name" json:"couponName"`
		FaceValue  float64    `gorm:"column:face_value" json:"faceValue"`
		CustomerID uint       `gorm:"column:customer_id" json:"customerId"`
		Customer   string     `gorm:"column:customer" json:"customer"`
		Status     string     `gorm:"column:status" json:"status"`
		ReceivedAt time.Time  `gorm:"column:received_at" json:"receivedAt"`
		UsedAt     *time.Time `gorm:"column:used_at" json:"usedAt"`
		RefBillNo  string     `gorm:"column:ref_bill_no" json:"refBillNo"`
	}
	var rows []row
	h.db.Model(&model.CouponGrant{}).
		Select("coupon_grants.id, coupon_grants.coupon_id, coupons.name AS coupon_name, coupons.face_value, coupon_grants.customer_id, customers.name AS customer, coupon_grants.status, coupon_grants.received_at, coupon_grants.used_at, coupon_grants.ref_bill_no").
		Joins("JOIN coupons ON coupons.id = coupon_grants.coupon_id").
		Joins("JOIN customers ON customers.id = coupon_grants.customer_id").
		Where("coupon_grants.company_id = ?", companyID).
		Scopes(func(tx *gorm.DB) *gorm.DB {
			if couponID > 0 {
				tx = tx.Where("coupon_grants.coupon_id = ?", couponID)
			}
			if status != "" {
				tx = tx.Where("coupon_grants.status = ?", status)
			}
			if keyword != "" {
				tx = tx.Where("customers.name ILIKE ? OR customers.code ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")
			}
			return tx
		}).
		Order("coupon_grants.id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&rows)
	response.OkWithPage(c, rows, page, pageSize, int(total))
}

func (h *MarketingHandler) CouponGrantUse(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		RefBillNo string `json:"refBillNo"`
	}
	_ = c.ShouldBindJSON(&req)
	now := time.Now()
	result := h.db.Model(&model.CouponGrant{}).
		Where("id = ? AND company_id = ? AND status = 'unused'", id, companyID).
		Updates(map[string]interface{}{"status": "used", "used_at": &now, "ref_bill_no": req.RefBillNo})
	if result.RowsAffected == 0 {
		response.Fail(c, 4004, "领用记录不存在或已使用")
		return
	}
	response.Ok(c, nil)
}

// ==================== 积分 ====================

func (h *MarketingHandler) PointRuleList(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var list []model.PointRule
	h.db.Where("company_id = ?", companyID).Order("sort, id").Find(&list)
	response.Ok(c, gin.H{"list": list, "total": len(list)})
}

func (h *MarketingHandler) PointRuleCreate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var r model.PointRule
	if err := c.ShouldBindJSON(&r); err != nil || r.Condition == "" {
		response.Fail(c, 4000, "赠送条件不能为空")
		return
	}
	r.ID = 0
	r.CompanyID = companyID
	if err := h.db.Create(&r).Error; err != nil {
		response.Fail(c, 4000, "创建失败: "+err.Error())
		return
	}
	response.Ok(c, r)
}

func (h *MarketingHandler) PointRuleUpdate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	var r model.PointRule
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&r).Error; err != nil {
		response.Fail(c, 4004, "规则不存在")
		return
	}
	var req model.PointRule
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误")
		return
	}
	req.ID = r.ID
	req.CompanyID = companyID
	h.db.Save(&req)
	response.Ok(c, req)
}

func (h *MarketingHandler) PointRuleDelete(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.PointRule{})
	response.Ok(c, nil)
}

func (h *MarketingHandler) PointExchangeList(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	type row struct {
		ID        uint   `gorm:"column:id" json:"id"`
		ProductID uint   `gorm:"column:product_id" json:"productId"`
		Product   string `gorm:"column:product" json:"product"`
		Code      string `gorm:"column:code" json:"code"`
		Spec      string `gorm:"column:spec" json:"spec"`
		Unit      string `gorm:"column:unit" json:"unit"`
		Points    int    `gorm:"column:points" json:"points"`
		Enabled   bool   `gorm:"column:enabled" json:"enabled"`
		Sort      int    `gorm:"column:sort" json:"sort"`
	}
	var rows []row
	h.db.Raw(`SELECT e.id, e.product_id, p.name AS product, p.code, p.specification AS spec, p.unit, e.points, e.enabled, e.sort
		FROM point_exchanges e JOIN products p ON p.id = e.product_id
		WHERE e.company_id = ? AND e.deleted_at IS NULL ORDER BY e.sort, e.id`, companyID).Scan(&rows)
	response.Ok(c, gin.H{"list": rows, "total": len(rows)})
}

func (h *MarketingHandler) PointExchangeCreate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var e model.PointExchange
	if err := c.ShouldBindJSON(&e); err != nil || e.ProductID == 0 {
		response.Fail(c, 4000, "请选择商品")
		return
	}
	e.ID = 0
	e.CompanyID = companyID
	if err := h.db.Create(&e).Error; err != nil {
		response.Fail(c, 4000, "创建失败: "+err.Error())
		return
	}
	response.Ok(c, e)
}

func (h *MarketingHandler) PointExchangeUpdate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	var e model.PointExchange
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&e).Error; err != nil {
		response.Fail(c, 4004, "兑换设置不存在")
		return
	}
	var req model.PointExchange
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误")
		return
	}
	e.Points = req.Points
	e.Enabled = req.Enabled
	e.Sort = req.Sort
	h.db.Save(&e)
	response.Ok(c, e)
}

func (h *MarketingHandler) PointExchangeDelete(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.PointExchange{})
	response.Ok(c, nil)
}

func (h *MarketingHandler) PointFlowList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "30"))
	companyID := middleware.GetCompanyID(c)
	customerID, _ := strconv.Atoi(c.DefaultQuery("customerId", "0"))
	flowType := c.Query("type")

	q := h.db.Model(&model.PointFlow{}).Where("company_id = ?", companyID)
	if customerID > 0 {
		q = q.Where("customer_id = ?", customerID)
	}
	if flowType != "" {
		q = q.Where("type = ?", flowType)
	}
	var total int64
	q.Count(&total)
	var list []model.PointFlow
	q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	custNames := map[uint]string{}
	var customers []model.Customer
	h.db.Where("company_id = ?", companyID).Find(&customers)
	for _, cu := range customers {
		custNames[cu.ID] = cu.Name
	}
	rows := make([]gin.H, 0, len(list))
	for _, f := range list {
		rows = append(rows, gin.H{
			"id": f.ID, "customerId": f.CustomerID, "customer": custNames[f.CustomerID],
			"change": f.Change, "type": f.Type, "remark": f.Remark, "createdAt": f.CreatedAt,
		})
	}
	response.OkWithPage(c, rows, page, pageSize, int(total))
}

// PointFlowCreate 手工调整积分
func (h *MarketingHandler) PointFlowCreate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var req struct {
		CustomerID uint   `json:"customerId" binding:"required"`
		Change     int    `json:"change" binding:"required"`
		Remark     string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误: "+err.Error())
		return
	}
	f := model.PointFlow{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		CustomerID:           req.CustomerID,
		Change:               req.Change,
		Type:                 "manual",
		Remark:               req.Remark,
	}
	if err := h.db.Create(&f).Error; err != nil {
		response.Fail(c, 4000, "保存失败: "+err.Error())
		return
	}
	response.Ok(c, f)
}

func (h *MarketingHandler) CustomerPointList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "30"))
	companyID := middleware.GetCompanyID(c)
	keyword := c.Query("keyword")

	type row struct {
		CustomerID uint   `gorm:"column:customer_id" json:"customerId"`
		Customer   string `gorm:"column:customer" json:"customer"`
		Code       string `gorm:"column:code" json:"code"`
		Points     int    `gorm:"column:points" json:"points"`
	}
	var all []row
	h.db.Raw(`SELECT f.customer_id, cu.name AS customer, cu.code, COALESCE(SUM(f.change),0)::int AS points
		FROM point_flows f JOIN customers cu ON cu.id = f.customer_id
		WHERE f.company_id = ? AND f.deleted_at IS NULL
		GROUP BY f.customer_id, cu.name, cu.code
		ORDER BY points DESC`, companyID).Scan(&all)
	list := make([]row, 0, len(all))
	for _, r := range all {
		if keyword != "" && !strings.Contains(r.Customer, keyword) && !strings.Contains(r.Code, keyword) {
			continue
		}
		list = append(list, r)
	}
	response.Ok(c, paginateSaleRows(list, page, pageSize))
}

// ==================== 分销 ====================

func (h *MarketingHandler) DistributorList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "30"))
	companyID := middleware.GetCompanyID(c)
	status := c.Query("status")
	keyword := c.Query("keyword")

	q := h.db.Model(&model.Distributor{}).Where("company_id = ?", companyID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if keyword != "" {
		q = q.Where("name ILIKE ? OR phone ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	var total int64
	q.Count(&total)
	var list []model.Distributor
	q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	managers := map[uint]string{}
	var emps []model.Employee
	h.db.Where("company_id = ?", companyID).Find(&emps)
	for _, e := range emps {
		managers[e.ID] = e.Name
	}
	type wAgg struct {
		DistributorID uint    `gorm:"column:distributor_id"`
		Pending       float64 `gorm:"column:pending"`
		Paid          float64 `gorm:"column:paid"`
	}
	wMap := map[uint]wAgg{}
	var wAggs []wAgg
	h.db.Raw(`SELECT distributor_id,
		COALESCE(SUM(CASE WHEN status='pending' THEN amount ELSE 0 END),0) AS pending,
		COALESCE(SUM(CASE WHEN status='paid' THEN amount ELSE 0 END),0) AS paid
		FROM distributor_withdrawals WHERE company_id = ? AND deleted_at IS NULL GROUP BY distributor_id`, companyID).Scan(&wAggs)
	for _, w := range wAggs {
		wMap[w.DistributorID] = w
	}

	now := time.Now()
	rows := make([]gin.H, 0, len(list))
	for _, d := range list {
		w := wMap[d.ID]
		undistributedDays := 0
		if d.ApprovedAt != nil {
			undistributedDays = int(now.Sub(*d.ApprovedAt).Hours() / 24)
		}
		rows = append(rows, gin.H{
			"id": d.ID, "name": d.Name, "phone": d.Phone,
			"managerId": d.ManagerID, "manager": managers[d.ManagerID],
			"customerId": d.CustomerID,
			"status":     d.Status, "appliedAt": d.AppliedAt, "approvedAt": d.ApprovedAt,
			"remark":     d.Remark,
			"saleAmount": 0, "unpostedAmount": 0, "postedAmount": 0,
			"withdrawable": 0, "withdrawing": w.Pending, "withdrawn": w.Paid,
			"commissionBalance": 0 - w.Paid,
			"undistributedDays": undistributedDays,
		})
	}
	response.OkWithPage(c, rows, page, pageSize, int(total))
}

func (h *MarketingHandler) DistributorCreate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var d model.Distributor
	if err := c.ShouldBindJSON(&d); err != nil || d.Name == "" {
		response.Fail(c, 4000, "分销商名称不能为空")
		return
	}
	d.ID = 0
	d.CompanyID = companyID
	if d.Status == "" {
		d.Status = "pending"
	}
	d.AppliedAt = time.Now()
	if err := h.db.Create(&d).Error; err != nil {
		response.Fail(c, 4000, "创建失败: "+err.Error())
		return
	}
	response.Ok(c, d)
}

func (h *MarketingHandler) DistributorUpdate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	var d model.Distributor
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&d).Error; err != nil {
		response.Fail(c, 4004, "分销商不存在")
		return
	}
	var req model.Distributor
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误")
		return
	}
	d.Name = req.Name
	d.Phone = req.Phone
	d.ManagerID = req.ManagerID
	d.CustomerID = req.CustomerID
	d.Remark = req.Remark
	h.db.Save(&d)
	response.Ok(c, d)
}

func (h *MarketingHandler) DistributorApprove(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	now := time.Now()
	result := h.db.Model(&model.Distributor{}).
		Where("id = ? AND company_id = ? AND status = 'pending'", id, companyID).
		Updates(map[string]interface{}{"status": "active", "approved_at": &now})
	if result.RowsAffected == 0 {
		response.Fail(c, 4004, "分销商不存在或已审核")
		return
	}
	response.Ok(c, nil)
}

func (h *MarketingHandler) DistributorDisable(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	result := h.db.Model(&model.Distributor{}).
		Where("id = ? AND company_id = ?", id, companyID).
		Update("status", "disabled")
	if result.RowsAffected == 0 {
		response.Fail(c, 4004, "分销商不存在")
		return
	}
	response.Ok(c, nil)
}

func (h *MarketingHandler) DistributorDelete(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.Distributor{})
	response.Ok(c, nil)
}

func (h *MarketingHandler) WithdrawalList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "30"))
	companyID := middleware.GetCompanyID(c)
	status := c.Query("status")

	q := h.db.Model(&model.DistributorWithdrawal{}).Where("company_id = ?", companyID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	q.Count(&total)
	var list []model.DistributorWithdrawal
	q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	names := map[uint]string{}
	var ds []model.Distributor
	h.db.Where("company_id = ?", companyID).Find(&ds)
	for _, d := range ds {
		names[d.ID] = d.Name
	}
	rows := make([]gin.H, 0, len(list))
	for _, w := range list {
		rows = append(rows, gin.H{
			"id": w.ID, "distributorId": w.DistributorID, "distributor": names[w.DistributorID],
			"amount": w.Amount, "status": w.Status, "appliedAt": w.AppliedAt, "paidAt": w.PaidAt, "remark": w.Remark,
		})
	}
	response.OkWithPage(c, rows, page, pageSize, int(total))
}

func (h *MarketingHandler) WithdrawalCreate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var req struct {
		DistributorID uint    `json:"distributorId" binding:"required"`
		Amount        float64 `json:"amount" binding:"required,gt=0"`
		Remark        string  `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误: "+err.Error())
		return
	}
	w := model.DistributorWithdrawal{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		DistributorID:        req.DistributorID,
		Amount:               req.Amount,
		Status:               "pending",
		AppliedAt:            time.Now(),
		Remark:               req.Remark,
	}
	if err := h.db.Create(&w).Error; err != nil {
		response.Fail(c, 4000, "创建失败: "+err.Error())
		return
	}
	response.Ok(c, w)
}

func (h *MarketingHandler) WithdrawalPay(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	now := time.Now()
	result := h.db.Model(&model.DistributorWithdrawal{}).
		Where("id = ? AND company_id = ? AND status = 'pending'", id, companyID).
		Updates(map[string]interface{}{"status": "paid", "paid_at": &now})
	if result.RowsAffected == 0 {
		response.Fail(c, 4004, "提现已处理或不存在")
		return
	}
	response.Ok(c, nil)
}

func (h *MarketingHandler) WithdrawalReject(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	result := h.db.Model(&model.DistributorWithdrawal{}).
		Where("id = ? AND company_id = ? AND status = 'pending'", id, companyID).
		Update("status", "rejected")
	if result.RowsAffected == 0 {
		response.Fail(c, 4004, "提现已处理或不存在")
		return
	}
	response.Ok(c, nil)
}

// ==================== 预收款储值 ====================

func (h *MarketingHandler) StoredValueRuleList(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var list []model.StoredValueRule
	h.db.Where("company_id = ?", companyID).Order("sort, id").Find(&list)
	response.Ok(c, gin.H{"list": list, "total": len(list)})
}

func (h *MarketingHandler) StoredValueRuleCreate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var r model.StoredValueRule
	if err := c.ShouldBindJSON(&r); err != nil || r.Name == "" {
		response.Fail(c, 4000, "规则名称不能为空")
		return
	}
	r.ID = 0
	r.CompanyID = companyID
	if err := h.db.Create(&r).Error; err != nil {
		response.Fail(c, 4000, "创建失败: "+err.Error())
		return
	}
	response.Ok(c, r)
}

func (h *MarketingHandler) StoredValueRuleUpdate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	var r model.StoredValueRule
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&r).Error; err != nil {
		response.Fail(c, 4004, "规则不存在")
		return
	}
	var req model.StoredValueRule
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误")
		return
	}
	req.ID = r.ID
	req.CompanyID = companyID
	h.db.Save(&req)
	response.Ok(c, req)
}

func (h *MarketingHandler) StoredValueRuleDelete(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.StoredValueRule{})
	response.Ok(c, nil)
}

func (h *MarketingHandler) StoredValueRecordList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "30"))
	companyID := middleware.GetCompanyID(c)
	keyword := c.Query("keyword")

	q := h.db.Model(&model.StoredValueRecord{}).Where("company_id = ?", companyID)
	var total int64
	q.Count(&total)
	var list []model.StoredValueRecord
	q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	custNames := map[uint]string{}
	var customers []model.Customer
	h.db.Where("company_id = ?", companyID).Find(&customers)
	for _, cu := range customers {
		custNames[cu.ID] = cu.Name
	}
	ruleNames := map[uint]string{}
	var rules []model.StoredValueRule
	h.db.Where("company_id = ?", companyID).Find(&rules)
	for _, r := range rules {
		ruleNames[r.ID] = r.Name
	}
	accNames := map[uint]string{}
	var accs []model.Account
	h.db.Where("company_id = ?", companyID).Find(&accs)
	for _, a := range accs {
		accNames[a.ID] = a.Name
	}

	rows := make([]gin.H, 0, len(list))
	for _, r := range list {
		if keyword != "" && !strings.Contains(custNames[r.CustomerID], keyword) {
			continue
		}
		rows = append(rows, gin.H{
			"id": r.ID, "customerId": r.CustomerID, "customer": custNames[r.CustomerID],
			"ruleId": r.RuleID, "rule": ruleNames[r.RuleID],
			"amount": r.Amount, "giftAmount": r.GiftAmount,
			"accountId": r.AccountID, "account": accNames[r.AccountID],
			"remark": r.Remark, "createdAt": r.CreatedAt,
		})
	}
	response.OkWithPage(c, rows, page, pageSize, int(total))
}

// StoredValueRecordCreate 新增储值（收款账户余额增加并写账户流水）
func (h *MarketingHandler) StoredValueRecordCreate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	userID := middleware.GetUserID(c)
	var req struct {
		CustomerID uint    `json:"customerId" binding:"required"`
		RuleID     uint    `json:"ruleId"`
		Amount     float64 `json:"amount" binding:"required,gt=0"`
		GiftAmount float64 `json:"giftAmount"`
		AccountID  uint    `json:"accountId"`
		Remark     string  `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误: "+err.Error())
		return
	}
	// 选择规则时带出金额
	if req.RuleID > 0 {
		var rule model.StoredValueRule
		if err := h.db.Where("id = ? AND company_id = ?", req.RuleID, companyID).First(&rule).Error; err == nil {
			req.Amount = rule.Amount
			req.GiftAmount = rule.GiftAmount
		}
	}
	rec := model.StoredValueRecord{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		CustomerID:           req.CustomerID,
		RuleID:               req.RuleID,
		Amount:               req.Amount,
		GiftAmount:           req.GiftAmount,
		AccountID:            req.AccountID,
		OperatorID:           userID,
		Remark:               req.Remark,
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&rec).Error; err != nil {
			return err
		}
		// 收款账户入账
		if req.AccountID > 0 && req.Amount > 0 {
			if err := tx.Model(&model.Account{}).Where("id = ? AND company_id = ?", req.AccountID, companyID).
				UpdateColumn("balance", gorm.Expr("balance + ?", req.Amount)).Error; err != nil {
				return err
			}
			flow := model.AccountFlow{
				BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
				AccountID:            req.AccountID,
				Type:                 "income",
				Amount:               req.Amount,
				RefType:              "stored_value",
				RefID:                rec.ID,
				Remark:               "客户储值",
			}
			if err := tx.Create(&flow).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		response.Fail(c, 4000, "储值失败: "+err.Error())
		return
	}
	response.Ok(c, rec)
}
