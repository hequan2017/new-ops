import service from '@/utils/request'

// 文本语料
export const getTextCorpora = (params) => service({ url: '/ziqiang/textCorpus/list', method: 'get', params })
export const saveTextCorpus = (data) => service({ url: '/ziqiang/textCorpus', method: 'post', data })
export const importTextCorpora = (data) => service({ url: '/ziqiang/textCorpus/import', method: 'post', data, timeout: 120000 })
export const deleteTextCorpus = (params) => service({ url: '/ziqiang/textCorpus', method: 'delete', params })

// AIGC 对比样本对
export const getComparePairs = (params) => service({ url: '/ziqiang/comparePair/list', method: 'get', params })
export const saveComparePair = (data) => service({ url: '/ziqiang/comparePair', method: 'post', data })

// 图像语料与 aHash 比对
export const getImageCorpora = (params) => service({ url: '/ziqiang/imageCorpus/list', method: 'get', params })
export const saveImageCorpus = (data) => service({ url: '/ziqiang/imageCorpus', method: 'post', data })
export const compareImageHash = (params) => service({ url: '/ziqiang/imageCorpus/compare', method: 'get', params })

// 法律工具
export const termCheck = (data) => service({ url: '/ziqiang/termCheck', method: 'post', data })
export const dpMask = (data) => service({ url: '/ziqiang/dpMask', method: 'post', data })
export const legalTokenize = (data) => service({ url: '/ziqiang/legalTokenize', method: 'post', data })
export const plagiarismCheck = (data) => service({ url: '/ziqiang/plagiarismCheck', method: 'post', data })

// 知识产权图谱
export const getEntities = (params) => service({ url: '/ziqiang/entity/list', method: 'get', params })
export const saveEntity = (data) => service({ url: '/ziqiang/entity', method: 'post', data })
export const deleteEntity = (params) => service({ url: '/ziqiang/entity', method: 'delete', params })
export const saveEdge = (data) => service({ url: '/ziqiang/edge', method: 'post', data })
export const deleteEdge = (params) => service({ url: '/ziqiang/edge', method: 'delete', params })
export const getGraph = () => service({ url: '/ziqiang/graph', method: 'get' })
export const graphBfs = (params) => service({ url: '/ziqiang/graph/bfs', method: 'get', params })

// 模型鉴权知识库
export const getRegulations = (params) => service({ url: '/ziqiang/regulation/list', method: 'get', params })
export const searchRegulations = (params) => service({ url: '/ziqiang/regulation/search', method: 'get', params })
export const getCases = (params) => service({ url: '/ziqiang/case/list', method: 'get', params })

// 区块链存证
export const appendEvidence = (data) => service({ url: '/ziqiang/evidence', method: 'post', data })
export const verifyEvidence = () => service({ url: '/ziqiang/evidence/verify', method: 'get' })
export const getEvidences = (params) => service({ url: '/ziqiang/evidence/list', method: 'get', params })

// 提示词库与渲染
export const getPrompts = (params) => service({ url: '/ziqiang/prompt/list', method: 'get', params })
export const renderPrompt = (data) => service({ url: '/ziqiang/prompt/render', method: 'post', data })

// 知识产权问答
export const qaPatent = (data) => service({ url: '/ziqiang/qa/patent', method: 'post', data, timeout: 60000 })
