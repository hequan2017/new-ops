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

// @Tags AssetHost
// @Summary 分页查询主机资产变更历史
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "主机ID"
// @Router /asset/host/history [get]
export const getAssetHostHistory = (params) => {
  return service({
    url: '/asset/host/history',
    method: 'get',
    params
  })
}

// @Tags AssetHost
// @Summary 导入主机资产 Excel（按内网IP upsert）
// @Security ApiKeyAuth
// @accept multipart/form-data
// @Produce application/json
// @Router /asset/host/import [post]
export const importAssetHost = (file) => {
  const data = new FormData()
  data.append('file', file)
  return service({
    url: '/asset/host/import',
    method: 'post',
    data,
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

// @Tags AssetHost
// @Summary CIDR 网段 SSH 探测
export const discoverHosts = (data) => {
  return service({ url: '/asset/host/discover', method: 'post', data, timeout: 70000 })
}

// @Tags AssetHost
// @Summary 导入网段发现的主机
export const importDiscoveredHosts = (data) => {
  return service({ url: '/asset/host/discover/import', method: 'post', data })
}

// @Tags AssetHost
// @Summary SSH 现场采集主机信息
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body object true "ID(主机ID)/credentialId(SSH凭据ID)"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"采集成功"}"
// @Router /asset/host/collect [post]
export const collectAssetHost = (data) => {
  return service({
    url: '/asset/host/collect',
    method: 'post',
    data,
    timeout: 30000
  })
}

// @Tags AssetHost
// @Summary 阿里云 ECS 实例同步
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Router /asset/sync/aliyun-ecs [post]
export const syncAliyunECS = (data) => {
  return service({
    url: '/asset/sync/aliyun-ecs',
    method: 'post',
    data,
    timeout: 60000
  })
}
