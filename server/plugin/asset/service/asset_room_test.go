package service

import (
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/asset/model"
)

func Test_validateRoom(t *testing.T) {
	if err := validateRoom(&model.AssetRoom{Name: "北京一机房"}); err != nil {
		t.Fatalf("合法机房不应报错: %v", err)
	}
	err := validateRoom(&model.AssetRoom{Region: "北京"})
	if err == nil || !strings.Contains(err.Error(), "机房名称") {
		t.Fatalf("应提示机房名称必填, got %v", err)
	}
	err = validateRoom(nil)
	if err == nil || !strings.Contains(err.Error(), "机房名称") {
		t.Fatalf("nil 应提示机房名称必填, got %v", err)
	}
}

func Test_validateRack(t *testing.T) {
	roomID := uint(1)
	if err := validateRack(&model.AssetRack{Name: "A-01", RoomID: roomID, TotalU: 42}); err != nil {
		t.Fatalf("合法机柜不应报错: %v", err)
	}
	err := validateRack(&model.AssetRack{RoomID: roomID})
	if err == nil || !strings.Contains(err.Error(), "机柜名称") {
		t.Fatalf("应提示机柜名称必填, got %v", err)
	}
	err = validateRack(&model.AssetRack{Name: "A-01"})
	if err == nil || !strings.Contains(err.Error(), "所属机房") {
		t.Fatalf("应提示选择机房, got %v", err)
	}
}

func Test_validateProductLine(t *testing.T) {
	if err := validateProductLine(&model.AssetProductLine{Name: "核心业务"}); err != nil {
		t.Fatalf("合法产品线不应报错: %v", err)
	}
	err := validateProductLine(&model.AssetProductLine{Owner: "ops"})
	if err == nil || !strings.Contains(err.Error(), "产品线名称") {
		t.Fatalf("应提示产品线名称必填, got %v", err)
	}
}
