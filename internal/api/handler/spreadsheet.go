package handler

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

// openSpreadsheetRows 读取上传的 .xlsx 电子表格，返回第一个工作表的全部行。
// 注意：老格式 .xls 不在后端解析（Go 的 xls 解析器对大 SST 文件会静默丢数据），
// 前端上传前会用 SheetJS 自动把 .xls 转成 .xlsx。
func openSpreadsheetRows(src io.Reader, filename string) ([][]string, error) {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".xlsx":
		f, err := excelize.OpenReader(src)
		if err != nil {
			return nil, fmt.Errorf("Excel 文件解析失败")
		}
		defer f.Close()
		rows, err := f.GetRows(f.GetSheetName(0))
		if err != nil {
			return nil, fmt.Errorf("Excel 文件解析失败")
		}
		return rows, nil
	case ".xls":
		return nil, fmt.Errorf("暂不支持 .xls 老格式文件，请在页面上传（页面会自动转换），或另存为 .xlsx 后上传")
	}
	return nil, fmt.Errorf("仅支持 .xlsx 文件")
}

// findHeaderRow 在表格前几行中定位包含指定表头列的行索引（容忍导出文件首行为标题/说明）
func findHeaderRow(rows [][]string, maxScan int, headers ...string) int {
	want := map[string]bool{}
	for _, h := range headers {
		want[h] = true
	}
	if maxScan > len(rows) {
		maxScan = len(rows)
	}
	for i := 0; i < maxScan; i++ {
		for _, cell := range rows[i] {
			if want[strings.TrimSpace(cell)] {
				return i
			}
		}
	}
	return -1
}
