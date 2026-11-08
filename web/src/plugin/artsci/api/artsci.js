import service from '@/utils/request'

// 工具
export const getTools = (params) => service({ url: '/artsci/tool/list', method: 'get', params })
export const runTool = (data) => service({ url: '/artsci/tool/run', method: 'post', data, timeout: 120000 })
export const getRunLogs = (params) => service({ url: '/artsci/tool/runLogs', method: 'get', params })

// 作业批改
export const gradeEssay = (data) => service({ url: '/artsci/grading', method: 'post', data, timeout: 180000 })
export const getGradings = (params) => service({ url: '/artsci/grading/list', method: 'get', params })

// 出题与组卷
export const generateQuestion = (data) => service({ url: '/artsci/question/generate', method: 'post', data, timeout: 180000 })
export const getQuestions = (params) => service({ url: '/artsci/question/list', method: 'get', params })
export const buildPaper = (data) => service({ url: '/artsci/paper', method: 'post', data })
export const getPapers = (params) => service({ url: '/artsci/paper/list', method: 'get', params })

// 文生图场景
export const getSceneTemplates = () => service({ url: '/artsci/scene/list', method: 'get' })
export const submitImageTask = (data) => service({ url: '/artsci/imageTask', method: 'post', data, timeout: 180000 })
export const getImageTasks = (params) => service({ url: '/artsci/imageTask/list', method: 'get', params })

// 图像解读
export const registerImageAsset = (data) => service({ url: '/artsci/imageAsset', method: 'post', data })
export const getImageAssets = (params) => service({ url: '/artsci/imageAsset/list', method: 'get', params })
export const interpretImage = (data) => service({ url: '/artsci/imageAsset/interpret', method: 'post', data, timeout: 180000 })
export const getInterpretations = (params) => service({ url: '/artsci/interpretation/list', method: 'get', params })
