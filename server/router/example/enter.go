package example

import (
	api "github.com/hequan2017/new-ops/server/api/v1"
)

type RouterGroup struct {
	CustomerRouter

	AttachmentCategoryRouter
	FileUploadAndDownloadRouter
}

var (
	exaCustomerApi = api.ApiGroupApp.ExampleApiGroup.CustomerApi

	attachmentCategoryApi       = api.ApiGroupApp.ExampleApiGroup.AttachmentCategoryApi
	exaFileUploadAndDownloadApi = api.ApiGroupApp.ExampleApiGroup.FileUploadAndDownloadApi
)
