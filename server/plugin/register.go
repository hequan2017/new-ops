package plugin

import (
	_ "github.com/hequan2017/new-ops/server/plugin/announcement"
	_ "github.com/hequan2017/new-ops/server/plugin/auto"
	_ "github.com/hequan2017/new-ops/server/plugin/asset"
	_ "github.com/hequan2017/new-ops/server/plugin/term"
	_ "github.com/hequan2017/new-ops/server/plugin/job"
	_ "github.com/hequan2017/new-ops/server/plugin/pipeline"
	_ "github.com/hequan2017/new-ops/server/plugin/container"
	_ "github.com/hequan2017/new-ops/server/plugin/k8s"
	_ "github.com/hequan2017/new-ops/server/plugin/gpu"
	_ "github.com/hequan2017/new-ops/server/plugin/dbops"
	_ "github.com/hequan2017/new-ops/server/plugin/monitor"
	_ "github.com/hequan2017/new-ops/server/plugin/workflow"
	_ "github.com/hequan2017/new-ops/server/plugin/agent"
	_ "github.com/hequan2017/new-ops/server/plugin/aiops"
)
