import service from '@/utils/request'

// @Tags AssetHost
// @Summary 创建主机资产
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body model.AssetHost true "创建主机资产"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"创建成功"}"
// @Router /asset/host/create [post]
export const createAssetHost = (data) => {
  return service({
    url: '/asset/host/create',
    method: 'post',
    data
  })
}

// @Tags AssetHost
// @Summary 删除主机资产
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body request.GetById true "删除主机资产"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /asset/host/delete [delete]
export const deleteAssetHost = (data) => {
  return service({
    url: '/asset/host/delete',
    method: 'delete',
    data
  })
}

// @Tags AssetHost
// @Summary 批量删除主机资产
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body request.IdsReq true "批量删除主机资产"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"批量删除成功"}"
// @Router /asset/host/deleteByIds [delete]
export const deleteAssetHostByIds = (data) => {
  return service({
    url: '/asset/host/deleteByIds',
    method: 'delete',
    data
  })
}

// @Tags AssetHost
// @Summary 更新主机资产
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body model.AssetHost true "更新主机资产"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"更新成功"}"
// @Router /asset/host/update [put]
export const updateAssetHost = (data) => {
  return service({
    url: '/asset/host/update',
    method: 'put',
    data
  })
}

// @Tags AssetHost
// @Summary 根据ID获取主机资产
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body request.GetById true "根据ID获取主机资产"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"查询成功"}"
// @Router /asset/host/find [get]
export const findAssetHost = (params) => {
  return service({
    url: '/asset/host/find',
    method: 'get',
    params
  })
}

// @Tags AssetHost
// @Summary 分页获取主机资产列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body assetReq.AssetHostSearch true "分页获取主机资产列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"查询成功"}"
// @Router /asset/host/list [post]
export const getAssetHostList = (data) => {
  return service({
    url: '/asset/host/list',
    method: 'post',
    data
  })
}
