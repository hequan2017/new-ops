package service

import (
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/dbops/model"
)

// TestValidateInstance 实例入参校验（名称必填/主机格式/端口兜底）
func TestValidateInstance(t *testing.T) {
	if err := validateInstance(nil); err == nil {
		t.Fatalf("nil 实例应报错")
	}
	if err := validateInstance(&model.DbopsInstance{Name: ""}); err == nil {
		t.Fatalf("空名称应报错")
	}
	ok := &model.DbopsInstance{Name: "db1", Host: "10.0.0.8", Port: 0}
	if err := validateInstance(ok); err != nil {
		t.Fatalf("合法实例报错: %v", err)
	}
	if ok.Port != 3306 {
		t.Fatalf("端口缺省应兜底 3306, got %d", ok.Port)
	}
	badHost := &model.DbopsInstance{Name: "db1", Host: "10.0.0.8/x", Port: 3306}
	if err := validateInstance(badHost); err == nil {
		t.Fatalf("非 IP 且含路径分隔的主机应报错")
	}
}
