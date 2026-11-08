// Package model ziqiang 自强-理工科大模型：图像鉴权语料（需求 3.5-2）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// ZqImageCorpus 图像鉴权语料（版权元数据 + aHash 感知哈希）
type ZqImageCorpus struct {
	global.GVA_MODEL
	Name               string `json:"name" gorm:"type:varchar(200);comment:图像名称"`
	Path               string `json:"path" gorm:"type:varchar(500);comment:存储路径/URL"`
	HashHex            string `json:"hashHex" gorm:"type:varchar(16);index;comment:aHash 十六进制(64bit)"`
	CreationTime       string `json:"creationTime" gorm:"type:varchar(32);comment:创作时间戳"`
	DeviceFingerprint  string `json:"deviceFingerprint" gorm:"type:varchar(200);comment:设备指纹"`
	Watermark          string `json:"watermark" gorm:"type:varchar(200);comment:数字水印标记"`
	Gallery            string `json:"gallery" gorm:"type:varchar(100);comment:来源图库"`
}

func (ZqImageCorpus) TableName() string { return "ziqiang_image_corpus" }
