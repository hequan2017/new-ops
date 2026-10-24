package system

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/hequan2017/new-ops/server/global"
	model "github.com/hequan2017/new-ops/server/model/system"
	"github.com/hequan2017/new-ops/server/model/system/request"
)

func Test_autoCodePackage_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过依赖全局配置的集成测试")
	}
	if global.GVA_DB == nil {
		// Create 直连 global.GVA_DB——裸测试进程无 DB，不跳过会空指针 panic（auto-code 需完整启动环境）
		t.Skip("全局 DB 未初始化，跳过 auto-code 包创建集成测试")
	}
	type args struct {
		ctx  context.Context
		info *request.SysAutoCodePackageCreate
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "测试 package",
			args: args{
				ctx: context.Background(),
				info: &request.SysAutoCodePackageCreate{
					Template:    "package",
					PackageName: "gva",
				},
			},
			wantErr: false,
		},
		{
			name: "测试 plugin",
			args: args{
				ctx: context.Background(),
				info: &request.SysAutoCodePackageCreate{
					Template:    "plugin",
					PackageName: "gva",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &autoCodePackage{}
			if err := a.Create(tt.args.ctx, tt.args.info); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_autoCodePackage_templates(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过依赖全局配置的集成测试")
	}
	// templates 读 Root/Server/resource/<template> 模版目录——裸测试进程无全局配置（路径塌缩为相对路径）
	if _, err := os.Stat(filepath.Join(global.GVA_CONFIG.AutoCode.Root, global.GVA_CONFIG.AutoCode.Server, "resource", "plugin")); err != nil {
		t.Skip("模版目录不可用（全局配置未初始化），跳过 auto-code 模版渲染集成测试")
	}
	type args struct {
		ctx       context.Context
		entity    model.SysAutoCodePackage
		info      request.AutoCode
		isPackage bool
	}
	tests := []struct {
		name      string
		args      args
		wantCode  map[string]string
		wantEnter map[string]map[string]string
		wantErr   bool
	}{
		{
			name: "测试1",
			args: args{
				ctx: context.Background(),
				entity: model.SysAutoCodePackage{
					Desc:        "描述",
					Label:       "展示名",
					Template:    "plugin",
					PackageName: "preview",
				},
				info: request.AutoCode{
					Abbreviation:    "user",
					HumpPackageName: "user",
				},
				isPackage: false,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &autoCodePackage{}
			gotCode, gotEnter, gotCreates, err := s.templates(tt.args.ctx, tt.args.entity, tt.args.info, tt.args.isPackage)
			if (err != nil) != tt.wantErr {
				t.Errorf("templates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for key, value := range gotCode {
				t.Logf("\n")
				t.Logf("%s", key)
				t.Logf("%s", value)
				t.Logf("\n")
			}
			t.Log(gotCreates)
			if !reflect.DeepEqual(gotEnter, tt.wantEnter) {
				t.Errorf("templates() gotEnter = %v, want %v", gotEnter, tt.wantEnter)
			}
		})
	}
}
