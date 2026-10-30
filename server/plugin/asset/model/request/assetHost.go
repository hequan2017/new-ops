// Package request asset 插件请求结构
package request

import (
	commonrequest "github.com/hequan2017/new-ops/server/model/common/request"
)

// AssetHostSearch 主机列表查询条件
type AssetHostSearch struct {
	commonrequest.PageInfo
	Status        string `json:"status" form:"status"`               // 状态过滤
	RoomID        *uint  `json:"roomId" form:"roomId"`               // 机房过滤
	ProductLineID *uint  `json:"productLineId" form:"productLineId"` // 产品线过滤
	SortField     string `json:"sortField" form:"sortField"`         // 排序字段白名单：id|updatedAt|lastCollectAt，白名单外忽略并回退默认
	SortOrder     string `json:"sortOrder" form:"sortOrder"`         // 排序方向：asc|desc（默认 desc）
}

// AssetRoomSearch 机房列表查询条件
type AssetRoomSearch struct {
	commonrequest.PageInfo
}
