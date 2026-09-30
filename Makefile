SHELL = /bin/bash

# 白泽 BaiZe · 统一运维开发平台
REPO_DIR        = $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
SERVER_DIR      = $(REPO_DIR)/server
WEB_DIR         = $(REPO_DIR)/web
# 目标机器（可被环境变量覆盖，同 scripts/deploy-test.sh）
NEW_OPS_TEST_HOST ?= root@192.168.112.138

.DEFAULT_GOAL := help

.PHONY: help
help: ## 显示帮助
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

.PHONY: dev-server
dev-server: ## 本地启动后端 (:8888)
	cd $(SERVER_DIR) && go run main.go

.PHONY: dev-web
dev-web: ## 本地启动前端 (:8080)
	cd $(WEB_DIR) && npm run serve

.PHONY: test
test: ## 后端快速验证：build + vet + 单测(-short，不触碰源码)
	cd $(SERVER_DIR) && go build ./... && go vet ./... && go test -short ./...

.PHONY: build-web
build-web: ## 构建前端 dist
	cd $(WEB_DIR) && npm run build --silent

.PHONY: build-server
build-server: ## 交叉编译 linux/amd64 后端
	cd $(SERVER_DIR) && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o new-ops-server .

.PHONY: deploy
deploy: ## 一键部署到测试环境（构建+上传+初始化+冒烟）
	bash $(REPO_DIR)/scripts/deploy-test.sh

.PHONY: deploy-host
deploy-host: ## 指定目标机部署：make deploy-host NEW_OPS_TEST_HOST=root@1.2.3.4
	NEW_OPS_TEST_HOST=$(NEW_OPS_TEST_HOST) bash $(REPO_DIR)/scripts/deploy-test.sh

.PHONY: doc
doc: ## 生成 swagger 文档（需安装 swag）
	cd $(SERVER_DIR) && swag init

.PHONY: clean
clean: ## 清理构建产物
	rm -rf $(WEB_DIR)/dist $(SERVER_DIR)/new-ops-server
