package service

import (
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/asset/model"
	"gorm.io/gorm"
)

// asset 插件错误码补充（云同步）
const (
	ErrCodeSyncCredInvalid = 1018
	ErrCodeSyncRegionEmpty = 1019
)

// ECSSyncResult 阿里云 ECS 同步结果
type ECSSyncResult struct {
	Total   int      `json:"total"`   // 云端实例总数
	Created int      `json:"created"` // 新建
	Updated int      `json:"updated"` // 更新（按 SN=InstanceId 匹配）
	Failed  []string `json:"failed"`
}

// SyncFromAliyunECS 从阿里云 ECS 同步实例到资产台账
// 规则：SN 字段存 InstanceID 作为去重键；vendor=阿里云；hostname 用云侧 HostName（空则 InstanceName/InstanceId）
func (s *AssetHostService) SyncFromAliyunECS(credentialID uint, region, operator string) (*ECSSyncResult, error) {
	if region == "" {
		return nil, newServiceErr(ErrCodeSyncRegionEmpty, "请填写地域（RegionId，如 cn-beijing）")
	}
	credSvc := new(CredCredentialService)
	accessKeySecret, credType, accessKeyID, err := credSvc.GetPlaintext(credentialID)
	if err != nil {
		return nil, fmt.Errorf("凭据读取失败: %w", err)
	}
	if credType != model.CredTypeCloudAK {
		return nil, newServiceErr(ErrCodeSyncCredInvalid, "同步需要云平台 AccessKey 类型凭据")
	}

	instances, err := fetchECSInstances(accessKeyID, accessKeySecret, region, 100)
	if err != nil {
		return nil, err
	}

	res := &ECSSyncResult{Total: len(instances)}
	// 预取 SN→存量主机与 IP 占用（含软删行，唯一索引对二者都生效）：
	// 原逐实例 First+Updates 的 N+1，且更新 ip 不查占用——命中唯一索引冲突时整台进 Failed
	snSet := make([]string, 0, len(instances))
	for _, it := range instances {
		if it.InstanceID != "" {
			snSet = append(snSet, it.InstanceID)
		}
	}
	existingBySN := map[string]model.AssetHost{}
	ipOwner := map[string]uint{} // ip → 持有主机 ID（含软删行）
	if len(snSet) > 0 {
		var rows []model.AssetHost
		if err := global.GVA_DB.Unscoped().Where("sn IN ?", snSet).Find(&rows).Error; err == nil {
			for i := range rows {
				if rows[i].DeletedAt.Time.IsZero() {
					existingBySN[rows[i].SN] = rows[i]
				}
				if rows[i].IP != "" {
					ipOwner[rows[i].IP] = rows[i].ID
				}
			}
		}
	}
	var allHosts []model.AssetHost
	global.GVA_DB.Unscoped().Select("id, ip").Where("ip <> ''").Find(&allHosts)
	for i := range allHosts {
		ipOwner[allHosts[i].IP] = allHosts[i].ID
	}
	for _, it := range instances {
		created, err := s.upsertFromECS(it, operator, existingBySN, ipOwner)
		if err != nil {
			res.Failed = append(res.Failed, fmt.Sprintf("%s: %v", it.InstanceID, err))
			continue
		}
		if created {
			res.Created++
		} else {
			res.Updated++
		}
	}
	return res, nil
}

// upsertFromECS 按 SN（InstanceID）upsert 单个实例，返回是否为新建
// existingBySN/ipOwner 为预取的存量映射（含软删行占用）；新建后回写映射防同批 SN/IP 撞车
func (s *AssetHostService) upsertFromECS(it ecsInstance, operator string,
	existingBySN map[string]model.AssetHost, ipOwner map[string]uint) (bool, error) {
	hostname := it.Hostname
	if hostname == "" {
		hostname = it.InstanceID
	}
	exist, found := existingBySN[it.InstanceID]
	switch {
	case found:
		updates := map[string]any{
			"hostname": hostname,
			"os":       osNameToID(it.OSName),
			"os_version": it.OSName,
			"cpu_cores": it.CPU,
			"mem_gb":    it.MemoryMB / 1024,
			"public_ip": it.PublicIP,
			"status":    string(mapECSStatus(it.Status)),
		}
		if it.InnerIP != "" && it.InnerIP != exist.IP {
			if owner, taken := ipOwner[it.InnerIP]; taken && owner != exist.ID {
				// IP 被其他主机占用（含软删残留行）：跳过 ip 字段更新并显式报失败原因，
				// 否则唯一索引冲突的报错整台进 Failed 且原因晦涩
				return false, fmt.Errorf("内网IP %s 已被主机 id=%d 占用，跳过 ip 更新", it.InnerIP, owner)
			}
			updates["ip"] = it.InnerIP
			ipOwner[it.InnerIP] = exist.ID
		}
		if err := global.GVA_DB.Model(&model.AssetHost{}).Where("id = ?", exist.ID).Updates(updates).Error; err != nil {
			return false, err
		}
		recordHostHistory(global.GVA_DB, exist.ID, model.HostHistoryUpdate,
			map[string]any{"action": "阿里云ECS同步", "instanceId": it.InstanceID}, operator)
		return false, nil
	default: // 按 SN 未找到 → 新建
		ip := it.InnerIP
		if ip == "" {
			ip = it.PublicIP
		}
		h := &model.AssetHost{
			Hostname: hostname,
			IP:       ip,
			OS:       osNameToID(it.OSName),
			OSVersion: it.OSName,
			CPUCores: it.CPU,
			MemGB:    it.MemoryMB / 1024,
			SN:       it.InstanceID,
			Vendor:   "阿里云",
			PublicIP: it.PublicIP,
			Status:   mapECSStatus(it.Status),
		}
		if err := validateHost(h); err != nil {
			return false, err
		}
		if owner, taken := ipOwner[h.IP]; taken {
			return false, fmt.Errorf("内网IP %s 已被主机 id=%d 占用，无法导入", h.IP, owner)
		}
		if err := global.GVA_DB.Create(h).Error; err != nil {
			return false, err
		}
		ipOwner[h.IP] = h.ID
		recordHostHistory(global.GVA_DB, h.ID, model.HostHistoryCreate,
			map[string]any{"action": "阿里云ECS同步", "instanceId": it.InstanceID}, operator)
		return true, nil
	}
}

// osNameToID 云侧 OSName → 系统 ID（ubuntu/centos/windows 等，取首词小写）
func osNameToID(osName string) string {
	fields := strings.Fields(osName)
	if len(fields) == 0 {
		return ""
	}
	return strings.ToLower(fields[0])
}

// isRecordNotFound 判断 gorm not found
func isRecordNotFound(err error) bool {
	return err != nil && err.Error() == gorm.ErrRecordNotFound.Error()
}
