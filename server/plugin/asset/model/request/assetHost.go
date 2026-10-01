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
}

// AssetRoomSearch 机房列表查询条件
type AssetRoomSearch struct {
	commonrequest.PageInfo
}
