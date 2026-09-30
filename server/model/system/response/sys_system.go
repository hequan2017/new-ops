package response

import "github.com/hequan2017/new-ops/server/config"

type SysConfigResponse struct {
	Config config.Server `json:"config"`
}
