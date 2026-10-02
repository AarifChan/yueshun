package handler

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/database"
	"zhizhang-server/internal/pkg/response"
)

// OtherInStockImportResult 其他入库单导入结果
type OtherInStockImportResult struct {
	Created          int                `json:"created"`          // 成功创建的单据数
	Skipped          int                `json:"skipped"`          // 跳过的明细行数（商品未匹配或整单失败）
	CreatedProducts  []string           `json:"createdProducts"`  // 商品资料不存在而自动建档的商品（名称/规格）
	BillNoMap        map[string]string  `json:"billNoMap"`        // 原单号冲突时 原单号 -> 新单号 的映射
	Errors           []PriceImportError `json:"errors"`
}

// otherInImportColumns 导入文件表头列名
var otherInImportColumns = []string{
	"单据编号", "入库仓库", "往来单位", "结算单位", "经手人", "制单人", "部门", "录单时间",
	"入库类型", "单据状态", "单据备注", "商品名称", "商品规格", "数量", "单价", "金额", "备注",
}

func parseImportTime(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{"2006/01/02 15:04", "2006-01-02 15:04", "2006/01/02 15:04:05", "2006-01-02 15:04:05", "2006/01/02", "2006-01-02"} {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func parseImportFloat(s string) (float64, bool) {
	s = strings.TrimSpace(strings.ReplaceAll(s, ",", ""))
	if s == "" {
		return 0, false
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

// Import 导入其他入库单
// @Summary 导入其他入库单
// @Tags 库存
// @Accept multipart/form-data
// @Param file formData file true "其他入库单明细文件（.xlsx/.xls）"
// @Param applyStock formData string false "传 1 时过账单据同时增加库存（默认不写库存）"
// @Success 200 {object} response.Response
// @Router /other-in-stocks/import [post]
func (h *OtherInStockHandler) Import(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}
	userID := middleware.GetUserID(c)
	applyStock := c.PostForm("applyStock") == "1"

	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请上传文件")
		return
	}
	if file.Size > 10*1024*1024 {
		response.BadRequest(c, "文件大小不能超过 10MB")
		return
	}
	src, err := file.Open()
	if err != nil {
		response.ServerError(c, "读取文件失败")
		return
	}
	defer src.Close()

	rows, err := openSpreadsheetRows(src, file.Filename)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	headerIdx := findHeaderRow(rows, 5, "单据编号", "商品名称")
	if headerIdx < 0 {
		response.BadRequest(c, "未找到表头行（需包含 单据编号/商品名称 列）")
		return
	}

	// 表头列名 -> 列索引
	colIdx := map[string]int{}
	for i, cell := range rows[headerIdx] {
		name := strings.TrimSpace(cell)
		if _, ok := colIdx[name]; !ok {
			colIdx[name] = i
		}
	}
	for _, name := range []string{"单据编号", "商品名称", "数量"} {
		if _, ok := colIdx[name]; !ok {
			response.BadRequest(c, "表头缺少必要列："+name)
			return
		}
	}

	// 预载基础资料：仓库/职员/部门/商品
	var warehouses []model.Warehouse
	h.db.Where("company_id = ?", companyID).Order("id").Find(&warehouses)
	warehouseByName := map[string]uint{}
	for _, w := range warehouses {
		if _, ok := warehouseByName[w.Name]; !ok {
			warehouseByName[w.Name] = w.ID
		}
	}

	var employees []model.Employee
	h.db.Where("company_id = ?", companyID).Select("id", "name").Find(&employees)
	employeeByName := map[string]uint{}
	for _, e := range employees {
		if _, ok := employeeByName[e.Name]; !ok {
			employeeByName[e.Name] = e.ID
		}
	}

	var departments []model.Department
	h.db.Where("company_id = ?", companyID).Select("id", "name").Find(&departments)
	deptByName := map[string]uint{}
	for _, d := range departments {
		if _, ok := deptByName[d.Name]; !ok {
			deptByName[d.Name] = d.ID
		}
	}

	var products []model.Product
	h.db.Where("company_id = ?", companyID).Select("id", "name", "specification", "status").Find(&products)
	productsByName := map[string][]model.Product{}
	for _, p := range products {
		productsByName[p.Name] = append(productsByName[p.Name], p)
	}
	matchProduct := func(name, spec string) (model.Product, bool) {
		candidates := productsByName[name]
		if len(candidates) == 0 {
			return model.Product{}, false
		}
		if len(candidates) == 1 && (spec == "" || candidates[0].Specification == spec) {
			return candidates[0], true
		}
		for _, p := range candidates {
			if p.Specification == spec {
				return p, true
			}
		}
		// 规格为空且有多个候选：取唯一上架的商品
		if spec == "" {
			var onShelf []model.Product
			for _, p := range candidates {
				if p.Status == 1 {
					onShelf = append(onShelf, p)
				}
			}
			if len(onShelf) == 1 {
				return onShelf[0], true
			}
		}
		return model.Product{}, false
	}

	// 已存在的单号（避免冲突）
	var existingNos []string
	h.db.Model(&model.OtherInStock{}).Where("company_id = ?", companyID).Pluck("bill_no", &existingNos)
	existingBillNo := map[string]bool{}
	for _, no := range existingNos {
		existingBillNo[no] = true
	}

	// 商品资料中不存在的商品自动建档案（默认分类），与库存状况导入同一迁移策略
	var defaultCategoryID uint
	createdProducts := make([]string, 0)
	createStubProduct := func(name, spec, unit string, price float64) (model.Product, error) {
		if unit == "" {
			unit = "个"
		}
		p := model.Product{
			BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
			Name:                 name,
			Specification:        spec,
			Unit:                 unit,
			PurchasePrice:        price,
			Status:               1,
		}
		err := h.db.Transaction(func(tx *gorm.DB) error {
			if defaultCategoryID == 0 {
				var cat model.ProductCategory
				cerr := tx.Where("company_id = ? AND name = ?", companyID, "默认分类").First(&cat).Error
				if cerr == gorm.ErrRecordNotFound {
					cat = model.ProductCategory{
						BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
						Name:                 "默认分类",
						Code:                 "DEFAULT",
						Status:               1,
					}
					if cerr = tx.Create(&cat).Error; cerr != nil {
						return cerr
					}
				} else if cerr != nil {
					return cerr
				}
				defaultCategoryID = cat.ID
			}
			p.CategoryID = defaultCategoryID
			code, cerr := nextProductCode(tx, companyID)
			if cerr != nil {
				return cerr
			}
			p.Code = code
			return tx.Create(&p).Error
		})
		if err != nil {
			return model.Product{}, err
		}
		productsByName[name] = append(productsByName[name], p)
		label := name
		if spec != "" {
			label += "/" + spec
		}
		createdProducts = append(createdProducts, label)
		return p, nil
	}

	result := OtherInStockImportResult{BillNoMap: map[string]string{}, Errors: make([]PriceImportError, 0, 50)}
	addError := func(row int, code, reason string) {
		result.Skipped++
		if len(result.Errors) < 50 {
			result.Errors = append(result.Errors, PriceImportError{Row: row, Code: code, Reason: reason})
		}
	}

	// 按单据编号分组（保持出现顺序）
	type billGroup struct {
		billNo string
		rows   []int // rows 切片中的行索引
	}
	groups := make([]billGroup, 0, 32)
	groupIdx := map[string]int{}
	for i := headerIdx + 1; i < len(rows); i++ {
		cell := func(name string) string {
			if idx, ok := colIdx[name]; ok && idx < len(rows[i]) {
				return strings.TrimSpace(rows[i][idx])
			}
			return ""
		}
		billNo := cell("单据编号")
		if billNo == "" {
			continue
		}
		if gi, ok := groupIdx[billNo]; ok {
			groups[gi].rows = append(groups[gi].rows, i)
			continue
		}
		groupIdx[billNo] = len(groups)
		groups = append(groups, billGroup{billNo: billNo, rows: []int{i}})
	}

	// 逐张单导入，每张单一个事务
	for _, g := range groups {
		cell := func(rowIdx int, name string) string {
			if idx, ok := colIdx[name]; ok && idx < len(rows[rowIdx]) {
				return strings.TrimSpace(rows[rowIdx][idx])
			}
			return ""
		}
		first := g.rows[0]
		billNo := g.billNo

		// 解析明细行，商品匹配失败的行跳过并记录
		type parsedItem struct {
			productID uint
			qty       float64
			price     float64
			amount    float64
			remark    string
		}
		items := make([]parsedItem, 0, len(g.rows))
		for _, ri := range g.rows {
			name := cell(ri, "商品名称")
			spec := cell(ri, "商品规格")
			if name == "" {
				continue
			}
			qty, ok := parseImportFloat(cell(ri, "数量"))
			if !ok || qty <= 0 {
				addError(ri+1, billNo, "数量无效: "+billNo+" / "+name)
				continue
			}
			price, _ := parseImportFloat(cell(ri, "单价"))
			product, ok := matchProduct(name, spec)
			if !ok {
				var cerr error
				product, cerr = createStubProduct(name, spec, cell(ri, "单位"), price)
				if cerr != nil {
					log.Error().Err(cerr).Str("product", name).Msg("import: create stub product failed")
					addError(ri+1, billNo, "商品未匹配且建档失败: "+billNo+" / "+name)
					continue
				}
			}
			amount, ok := parseImportFloat(cell(ri, "金额"))
			if !ok {
				amount = qty * price
			}
			items = append(items, parsedItem{productID: product.ID, qty: qty, price: price, amount: amount, remark: cell(ri, "备注")})
		}
		if len(items) == 0 {
			continue
		}

		// 单据头字段
		warehouseID := warehouseByName[cell(first, "入库仓库")]
		handlerID := employeeByName[cell(first, "经手人")]
		operatorID := employeeByName[cell(first, "制单人")]
		if operatorID == 0 {
			operatorID = userID
		}
		deptID := deptByName[cell(first, "部门")]
		createdAt, hasTime := parseImportTime(cell(first, "录单时间"))
		billDate := time.Now()
		if hasTime {
			billDate = time.Date(createdAt.Year(), createdAt.Month(), createdAt.Day(), 0, 0, 0, 0, time.Local)
		}
		status := "draft"
		if cell(first, "单据状态") == "已过账" {
			status = "completed"
		}

		var totalQty, amount float64
		for _, it := range items {
			totalQty += it.qty
			amount += it.amount
		}

		bill := model.OtherInStock{
			BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
			WarehouseID:          warehouseID,
			BillNo:               billNo,
			BillDate:             billDate,
			InType:               cell(first, "入库类型"),
			Counterpart:          cell(first, "往来单位"),
			SettleUnit:           cell(first, "结算单位"),
			HandlerID:            handlerID,
			DeptID:               deptID,
			TotalQty:             totalQty,
			Amount:               amount,
			Status:               status,
			OperatorID:           operatorID,
			Remark:               cell(first, "单据备注"),
		}
		if bill.InType == "" {
			bill.InType = "其他入库"
		}

		err := h.db.Transaction(func(tx *gorm.DB) error {
			// 仓库：按名称查不到则用公司第一个仓库，再没有则创建
			if bill.WarehouseID == 0 {
				var wh model.Warehouse
				werr := tx.Where("company_id = ?", companyID).Order("id").First(&wh).Error
				if werr != nil {
					wh = model.Warehouse{
						BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
						Name:                 cell(first, "入库仓库"),
						Status:               1,
					}
					if wh.Name == "" {
						wh.Name = "默认仓库"
					}
					if cerr := tx.Create(&wh).Error; cerr != nil {
						return cerr
					}
				}
				bill.WarehouseID = wh.ID
				warehouseByName[wh.Name] = wh.ID
			}

			// 单号冲突则生成新单号并记录映射
			if existingBillNo[bill.BillNo] {
				newNo := h.generateBillNo(tx)
				result.BillNoMap[bill.BillNo] = newNo
				bill.BillNo = newNo
			}
			existingBillNo[bill.BillNo] = true

			if cerr := tx.Create(&bill).Error; cerr != nil {
				return cerr
			}
			dbItems := make([]model.OtherInStockItem, len(items))
			for i, it := range items {
				dbItems[i] = model.OtherInStockItem{
					InStockID: bill.ID,
					ProductID: it.productID,
					Quantity:  it.qty,
					Price:     it.price,
					Amount:    it.amount,
					Remark:    it.remark,
				}
			}
			if cerr := tx.Create(&dbItems).Error; cerr != nil {
				return cerr
			}
			// 保留原始录单时间
			if hasTime {
				if uerr := tx.Model(&model.OtherInStock{}).Where("id = ?", bill.ID).
					UpdateColumn("created_at", createdAt).Error; uerr != nil {
					return uerr
				}
			}
			if applyStock && bill.Status == "completed" {
				for _, it := range dbItems {
					if serr := database.ChangeStock(tx, companyID, bill.WarehouseID, it.ProductID, it.Quantity); serr != nil {
						return serr
					}
				}
			}
			return nil
		})
		if err != nil {
			log.Error().Err(err).Str("billNo", billNo).Msg("import other in-stock bill failed")
			for range g.rows {
				addError(first+1, billNo, "单据保存失败: "+billNo)
			}
			continue
		}
		result.Created++
	}

	result.CreatedProducts = createdProducts
	response.Ok(c, result)
}
