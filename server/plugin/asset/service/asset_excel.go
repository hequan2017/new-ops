package service

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/model/common/request"
	"github.com/hequan2017/new-ops/server/plugin/asset/model"
	"github.com/xuri/excelize/v2"
)

// hostExportHeaders 导出/导入共用的列定义（顺序即 Excel 列顺序）
var hostExportHeaders = []string{"主机名", "内网IP", "公网IP", "操作系统", "系统版本", "CPU核数", "内存GB", "磁盘GB", "SN", "厂商", "负责人", "状态", "备注"}

// ExportAssetHosts 生成主机资产 Excel（全量）
func (s *AssetHostService) ExportAssetHosts() (*excelize.File, error) {
	f := excelize.NewFile()
	sheet := "主机资产"
	f.SetSheetName("Sheet1", sheet)
	for i, h := range hostExportHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	list, _, err := s.GetAssetHostList(request.PageInfo{Page: 1, PageSize: 100000}, "", nil, nil)
	if err != nil {
		return nil, err
	}
	for r, host := range list {
		row := []any{host.Hostname, host.IP, host.PublicIP, host.OS, host.OSVersion,
			host.CPUCores, host.MemGB, host.DiskGB, host.SN, host.Vendor,
			host.Owner, string(host.Status), host.Notes}
		for i, v := range row {
			cell, _ := excelize.CoordinatesToCellName(i+1, r+2)
			f.SetCellValue(sheet, cell, v)
		}
	}
	return f, nil
}

// ImportResult 导入结果统计
type ImportResult struct {
	Created int      `json:"created"`
	Updated int      `json:"updated"`
	Failed  []string `json:"failed"`
}

// ImportAssetHosts 从 Excel 导入主机资产（按 IP upsert；表头行须与导出格式一致）
func (s *AssetHostService) ImportAssetHosts(file io.Reader, operator string) (res ImportResult, err error) {
	f, err := excelize.OpenReader(file)
	if err != nil {
		return res, fmt.Errorf("无法解析 Excel 文件: %w", err)
	}
	defer f.Close()
	rows, err := f.GetRows(f.GetSheetName(0))
	if err != nil || len(rows) < 2 {
		return res, fmt.Errorf("表格为空或缺少数据行")
	}

	// 表头 → 列号
	colIndex := map[string]int{}
	for i, h := range rows[0] {
		colIndex[strings.TrimSpace(h)] = i
	}
	for _, h := range []string{"主机名", "内网IP"} {
		if _, ok := colIndex[h]; !ok {
			return res, fmt.Errorf("缺少必需表头列: %s", h)
		}
	}
	cellStr := func(row []string, name string) string {
		if i, ok := colIndex[name]; ok && i < len(row) {
			return strings.TrimSpace(row[i])
		}
		return ""
	}
	cellInt := func(row []string, name string) int {
		v, _ := strconv.Atoi(cellStr(row, name))
		return v
	}

	for line, row := range rows[1:] {
		h := &model.AssetHost{
			Hostname:  cellStr(row, "主机名"),
			IP:        cellStr(row, "内网IP"),
			PublicIP:  cellStr(row, "公网IP"),
			OS:        cellStr(row, "操作系统"),
			OSVersion: cellStr(row, "系统版本"),
			CPUCores:  cellInt(row, "CPU核数"),
			MemGB:     cellInt(row, "内存GB"),
			DiskGB:    cellInt(row, "磁盘GB"),
			SN:        cellStr(row, "SN"),
			Vendor:    cellStr(row, "厂商"),
			Owner:     cellStr(row, "负责人"),
			Status:    model.AssetStatus(cellStr(row, "状态")),
			Notes:     cellStr(row, "备注"),
		}
		if h.Hostname == "" && h.IP == "" {
			continue // 空行
		}
		if err := validateHost(h); err != nil {
			res.Failed = append(res.Failed, fmt.Sprintf("第%d行: %s", line+2, err.Error()))
			continue
		}
		var exist model.AssetHost
		dbErr := global.GVA_DB.Where("ip = ?", h.IP).First(&exist).Error
		switch {
		case dbErr == nil:
			h.ID = exist.ID
			if err := s.UpdateAssetHost(h, operator); err != nil {
				res.Failed = append(res.Failed, fmt.Sprintf("第%d行: 更新失败 %s", line+2, err.Error()))
				continue
			}
			res.Updated++
		case strings.Contains(dbErr.Error(), "record not found"):
			if err := s.CreateAssetHost(h, operator); err != nil {
				res.Failed = append(res.Failed, fmt.Sprintf("第%d行: 创建失败 %s", line+2, err.Error()))
				continue
			}
			res.Created++
		default:
			res.Failed = append(res.Failed, fmt.Sprintf("第%d行: 查询失败 %s", line+2, dbErr.Error()))
		}
	}
	return res, nil
}
