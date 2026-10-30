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

// ExportAssetHostsForUser 按数据权限生成主机资产 Excel
func (s *AssetHostService) ExportAssetHostsForUser(scope *HostUserScope) (*excelize.File, error) {
	f := excelize.NewFile()
	sheet := "主机资产"
	f.SetSheetName("Sheet1", sheet)
	for i, h := range hostExportHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	list, _, err := s.GetAssetHostList(request.PageInfo{Page: 1, PageSize: 100000}, "", nil, nil, scope, "", "")
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

	// 批量预取已存在主机（按 IP）：原逐行 First 存在性查询为 N+1（创建/更新仍逐行——
	// 逐行失败明细与事务语义需要保留）
	ips := make([]string, 0, len(rows))
	for _, row := range rows[1:] {
		if ip := cellStr(row, "内网IP"); ip != "" {
			ips = append(ips, ip)
		}
	}
	existingByIP := map[string]model.AssetHost{}
	if len(ips) > 0 {
		var existing []model.AssetHost
		if err := global.GVA_DB.Unscoped().Where("ip IN ?", ips).Find(&existing).Error; err == nil {
			for i := range existing {
				if existing[i].DeletedAt.Time.IsZero() { // 活行才算"已存在"；软删行走创建（CreateAssetHost 会先物理清理）
					existingByIP[existing[i].IP] = existing[i]
				}
			}
		}
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
		exist, found := existingByIP[h.IP]
		switch {
		case found:
			h.ID = exist.ID
			if err := s.UpdateAssetHost(h, operator); err != nil {
				res.Failed = append(res.Failed, fmt.Sprintf("第%d行: 更新失败 %s", line+2, err.Error()))
				continue
			}
			res.Updated++
		default:
			if err := s.CreateAssetHost(h, operator); err != nil {
				res.Failed = append(res.Failed, fmt.Sprintf("第%d行: 创建失败 %s", line+2, err.Error()))
				continue
			}
			res.Created++
		}
	}
	return res, nil
}
