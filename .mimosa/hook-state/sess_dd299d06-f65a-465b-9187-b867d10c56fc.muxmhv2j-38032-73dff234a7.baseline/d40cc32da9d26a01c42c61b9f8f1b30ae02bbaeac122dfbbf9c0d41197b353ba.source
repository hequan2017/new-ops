package main

//go:generate go mod tidy
//go:generate go mod download
//go:generate go run gen.go

import (
	"gorm.io/gen"
	"path/filepath"
	"github.com/hequan2017/new-ops/server/plugin/announcement/model"
)

func main() {
	g := gen.NewGenerator(gen.Config{OutPath: filepath.Join("..", "..", "..", "announcement", "blender", "model", "dao"), Mode: gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface})
	g.ApplyBasic(
		new(model.Info),
	)
	g.Execute()
}
