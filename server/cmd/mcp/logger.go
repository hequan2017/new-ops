package main

import (
	"github.com/hequan2017/new-ops/server/global"
	"go.uber.org/zap"
)

func initializeStandaloneLogger() error {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return err
	}

	global.GVA_LOG = logger
	zap.ReplaceGlobals(logger)
	return nil
}
