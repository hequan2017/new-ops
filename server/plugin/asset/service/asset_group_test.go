package service

import (
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/asset/model"
)

func Test_validateGroup(t *testing.T) {
	if err := validateGroup(&model.AssetGroup{Name: "华东资产组"}); err != nil {
		t.Fatalf("合法资产组不应报错: %v", err)
	}
	err := validateGroup(&model.AssetGroup{Notes: "x"})
	if err == nil || !strings.Contains(err.Error(), "资产组名称") {
		t.Fatalf("应提示资产组名称必填, got %v", err)
	}
	err = validateGroup(nil)
	if err == nil || !strings.Contains(err.Error(), "资产组名称") {
		t.Fatalf("nil 应提示资产组名称必填, got %v", err)
	}
}
