package request

import (
	common "github.com/hequan2017/new-ops/server/model/common"
	commonReq "github.com/hequan2017/new-ops/server/model/common/request"
	system "github.com/hequan2017/new-ops/server/model/system"
)

type SysAIWorkflowSessionUpsert struct {
	ID             uint                       `json:"id"`
	Tab            string                     `json:"tab"`
	Title          string                     `json:"title"`
	Summary        string                     `json:"summary"`
	ConversationID string                     `json:"conversationId"`
	MessageID      string                     `json:"messageId"`
	CurrentNodeID  string                     `json:"currentNodeId"`
	Settings       common.JSONMap             `json:"settings"`
	FormData       common.JSONMap             `json:"formData"`
	ResultData     common.JSONMap             `json:"resultData"`
	Messages       []system.AIWorkflowMessage `json:"messages"`
}

type SysAIWorkflowSessionSearch struct {
	commonReq.PageInfo
	Tab string `json:"tab" form:"tab"`
}

type SysAIWorkflowMarkdownDump struct {
	ID             uint                       `json:"id"`
	Tab            string                     `json:"tab"`
	Title          string                     `json:"title"`
	Summary        string                     `json:"summary"`
	ConversationID string                     `json:"conversationId"`
	MessageID      string                     `json:"messageId"`
	CurrentNodeID  string                     `json:"currentNodeId"`
	Settings       common.JSONMap             `json:"settings"`
	FormData       common.JSONMap             `json:"formData"`
	ResultData     common.JSONMap             `json:"resultData"`
	Messages       []system.AIWorkflowMessage `json:"messages"`
}
