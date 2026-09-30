package example

import (
	api "github.com/hequan2017/new-ops/server/api/v1"
)

type RouterGroup struct {
	AttachmentCategoryRouter
	FileUploadAndDownloadRouter
}

var (
	attachmentCategoryApi       = api.ApiGroupApp.ExampleApiGroup.AttachmentCategoryApi
	exaFileUploadAndDownloadApi = api.ApiGroupApp.ExampleApiGroup.FileUploadAndDownloadApi
)
