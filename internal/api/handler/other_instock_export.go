package handler

import (
	"fmt"
	"math"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/xuri/excelize/v2"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/response"
)

// otherInExportMaxRows 单次导出最大单据数
const otherInExportMaxRows = 5000

// exportExcel 将 excelize 文件流式写入响应
func exportExcel(c *gin.Context, f *excelize.File, filename string) {
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename*=UTF-8''"+url.QueryEscape(filename))
	if err := f.Write(c.Writer); err != nil {
		log.Error().Err(err).Msg("write excel failed")
	}
}

func newExcelSheet(f *excelize.File, headers []string) string {
	const sheet = "Sheet0"
	_ = f.SetSheetName("Sheet1", sheet)
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}
	return sheet
}

// trimFloat 去掉浮点数多余的尾随零（10 -> "10"，2.5 -> "2.5"）
func trimFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func otherInStatusText(status string) string {
	if status == "completed" {
		return "已过账"
	}
	return "草稿"
}

// auxUnitInfo 商品默认辅助单位（用于换算关系/辅助数量）
type auxUnitInfo struct {
	Name          string
	Conversion    float64
	isDefaultFlag bool
}

// ExportList 导出其他入库单列表
// @Summary 导出其他入库单列表 Excel
// @Tags 库存
// @Param keyword query string false "单号"
// @Param productKw query string false "商品名称/编号"
// @Param warehouseId query int false "入库仓库"
// @Param status query string false "状态 draft/completed"
// @Param inType query string false "入库类型"
// @Param startDate query string false "录单时间起 YYYY-MM-DD"
// @Param endDate query string false "录单时间止 YYYY-MM-DD"
// @Success 200 {file} binary
// @Router /other-in-stocks/export [get]
func (h *OtherInStockHandler) ExportList(c *gin.Context) {
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

	var list []OtherInStockListResp
	if err := h.withNames(applyOtherInStockFilters(h.db.Model(&model.OtherInStock{}), companyID, &req)).
		Order("other_in_stocks.created_at DESC").
		Limit(otherInExportMaxRows).
		Scan(&list).Error; err != nil {
		log.Error().Err(err).Msg("export other in-stocks failed")
		response.ServerError(c, "导出失败")
		return
	}

	f := excelize.NewFile()
	defer func() { _ = f.Close() }()
	sheet := newExcelSheet(f, []string{
		"单号", "录单时间", "入库类型", "入库收入", "入库金额", "往来单位",
		"单据状态", "经手人", "制单人", "业务经理", "备注",
	})
	for i, b := range list {
		row := []interface{}{
			b.BillNo,
			b.CreatedAt.Format("2006/01/02 15:04"),
			b.InType,
			b.InType, // 入库收入
			b.Amount,
			b.Counterpart,
			otherInStatusText(b.Status),
			b.HandlerName,
			b.OperatorName,
			b.BusinessManagerName,
			b.Remark,
		}
		cell, _ := excelize.CoordinatesToCellName(1, i+2)
		_ = f.SetSheetRow(sheet, cell, &row)
	}

	exportExcel(c, f, fmt.Sprintf("其他入库单列表_%s.xlsx", time.Now().Format("20060102")))
}

// ExportItems 导出其他入库单明细
// @Summary 导出其他入库单明细 Excel
// @Tags 库存
// @Param keyword query string false "单号"
// @Param productKw query string false "商品名称/编号"
// @Param warehouseId query int false "入库仓库"
// @Param status query string false "状态 draft/completed"
// @Param inType query string false "入库类型"
// @Param startDate query string false "录单时间起 YYYY-MM-DD"
// @Param endDate query string false "录单时间止 YYYY-MM-DD"
// @Success 200 {file} binary
// @Router /other-in-stocks/export/items [get]
func (h *OtherInStockHandler) ExportItems(c *gin.Context) {
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

	var bills []OtherInStockListResp
	if err := h.withNames(applyOtherInStockFilters(h.db.Model(&model.OtherInStock{}), companyID, &req)).
		Order("other_in_stocks.created_at DESC").
		Limit(otherInExportMaxRows).
		Scan(&bills).Error; err != nil {
		log.Error().Err(err).Msg("export other in-stock items failed")
		response.ServerError(c, "导出失败")
		return
	}

	f := excelize.NewFile()
	defer func() { _ = f.Close() }()
	sheet := newExcelSheet(f, []string{
		"单据编号", "入库仓库", "往来单位", "结算单位", "经手人", "制单人", "部门", "录单时间",
		"入库类型", "单据状态", "入库金额", "单据备注",
		"商品条码", "商品名称", "商品规格", "库存数量", "单位", "换算关系", "数量", "辅助数量", "单价", "金额", "备注",
	})

	if len(bills) == 0 {
		exportExcel(c, f, fmt.Sprintf("其他入库单明细_%s.xlsx", time.Now().Format("20060102")))
		return
	}

	billIDs := make([]uint, len(bills))
	for i, b := range bills {
		billIDs[i] = b.ID
	}

	var items []model.OtherInStockItem
	if err := h.db.Where("in_stock_id IN ?", billIDs).Preload("Product").Order("id").Find(&items).Error; err != nil {
		log.Error().Err(err).Msg("export: query items failed")
		response.ServerError(c, "导出失败")
		return
	}

	// 当前库存（按 仓库+商品）
	type stockKey struct {
		WarehouseID uint
		ProductID   uint
	}
	warehouseIDs := make([]uint, 0, len(bills))
	seenWh := map[uint]bool{}
	for _, b := range bills {
		if !seenWh[b.WarehouseID] {
			seenWh[b.WarehouseID] = true
			warehouseIDs = append(warehouseIDs, b.WarehouseID)
		}
	}
	var stocks []model.Stock
	if err := h.db.Where("company_id = ? AND warehouse_id IN ?", companyID, warehouseIDs).Find(&stocks).Error; err != nil {
		log.Error().Err(err).Msg("export: query stocks failed")
		response.ServerError(c, "导出失败")
		return
	}
	stockMap := make(map[stockKey]float64, len(stocks))
	for _, s := range stocks {
		stockMap[stockKey{s.WarehouseID, s.ProductID}] = s.Quantity
	}

	// 商品辅助单位
	productIDs := make([]uint, 0, len(items))
	seenP := map[uint]bool{}
	for _, it := range items {
		if !seenP[it.ProductID] {
			seenP[it.ProductID] = true
			productIDs = append(productIDs, it.ProductID)
		}
	}
	var units []model.ProductUnit
	if len(productIDs) > 0 {
		if err := h.db.Where("product_id IN ? AND status = 1", productIDs).Find(&units).Error; err != nil {
			log.Error().Err(err).Msg("export: query product units failed")
			response.ServerError(c, "导出失败")
			return
		}
	}
	auxMap := make(map[uint]auxUnitInfo)
	for _, u := range units {
		if u.Conversion <= 1 {
			continue // 基本单位（换算系数 1）不参与换算关系
		}
		cur, ok := auxMap[u.ProductID]
		if !ok || (u.IsDefault && !cur.isDefaultFlag) || (u.IsDefault == cur.isDefaultFlag && u.Conversion > cur.Conversion) {
			auxMap[u.ProductID] = auxUnitInfo{Name: u.Name, Conversion: u.Conversion, isDefaultFlag: u.IsDefault}
		}
	}

	billMap := make(map[uint]*OtherInStockListResp, len(bills))
	for i := range bills {
		billMap[bills[i].ID] = &bills[i]
	}

	rowIdx := 2
	for _, it := range items {
		b, ok := billMap[it.InStockID]
		if !ok {
			continue
		}
		p := it.Product
		aux, hasAux := auxMap[it.ProductID]
		conversion, auxQty := conversionText(p.Unit, aux, hasAux), auxQuantityText(it.Quantity, p.Unit, aux, hasAux)
		row := []interface{}{
			b.BillNo, b.WarehouseName, b.Counterpart, b.SettleUnit, b.HandlerName, b.OperatorName, b.DeptName,
			b.CreatedAt.Format("2006/01/02 15:04"),
			b.InType, otherInStatusText(b.Status), b.Amount, b.Remark,
			p.Barcode, p.Name, p.Specification,
			stockMap[stockKey{b.WarehouseID, it.ProductID}],
			p.Unit, conversion, it.Quantity, auxQty, it.Price, it.Amount, it.Remark,
		}
		cell, _ := excelize.CoordinatesToCellName(1, rowIdx)
		_ = f.SetSheetRow(sheet, cell, &row)
		rowIdx++
	}

	exportExcel(c, f, fmt.Sprintf("其他入库单明细_%s.xlsx", time.Now().Format("20060102")))
}

// conversionText 换算关系，如 "1盒=10支"；无辅助单位时为 "1套"
func conversionText(baseUnit string, aux auxUnitInfo, hasAux bool) string {
	if hasAux && aux.Conversion > 0 && aux.Name != "" && aux.Name != baseUnit {
		return fmt.Sprintf("1%s=%s%s", aux.Name, trimFloat(aux.Conversion), baseUnit)
	}
	return "1" + baseUnit
}

// auxQuantityText 辅助数量，如 "5件2支"/"9盒"/"33套"
func auxQuantityText(qty float64, baseUnit string, aux auxUnitInfo, hasAux bool) string {
	if hasAux && aux.Conversion > 1 && qty >= aux.Conversion {
		n := math.Floor(qty/aux.Conversion + 1e-9)
		rem := qty - n*aux.Conversion
		if rem > 1e-9 {
			return fmt.Sprintf("%s%s%s%s", trimFloat(n), aux.Name, trimFloat(rem), baseUnit)
		}
		return fmt.Sprintf("%s%s", trimFloat(n), aux.Name)
	}
	return trimFloat(qty) + baseUnit
}
