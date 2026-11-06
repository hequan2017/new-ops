import service from '@/utils/request'

// 训练任务
export const getTrainJobs = (params) => service({ url: '/llmops/train/list', method: 'get', params })
export const saveTrainJob = (data) => service({ url: '/llmops/train', method: 'post', data })
export const transitionTrain = (data) => service({ url: '/llmops/train/transition', method: 'post', data })
export const dryRunTrain = (id) => service({ url: '/llmops/train/dryRun', method: 'post', data: { id } })
export const startTrain = (id) => service({ url: '/llmops/train/start', method: 'post', data: { id } })
export const stopTrain = (data) => service({ url: '/llmops/train/stop', method: 'post', data })
export const getTrainLog = (params) => service({ url: '/llmops/train/log', method: 'get', params })
export const deleteTrain = (params) => service({ url: '/llmops/train', method: 'delete', params })

// 模型评估
export const getEvalRuns = (params) => service({ url: '/llmops/eval/list', method: 'get', params })
export const saveEvalRun = (data) => service({ url: '/llmops/eval', method: 'post', data })
export const runEval = (id) => service({ url: '/llmops/eval/run', method: 'post', data: { id }, timeout: 140000 })
export const getEvalResults = (params) => service({ url: '/llmops/eval/results', method: 'get', params })

// 产物
export const getArtifacts = () => service({ url: '/llmops/artifact/list', method: 'get' })
export const registerArtifact = (data) => service({ url: '/llmops/artifact', method: 'post', data })
export const registerArtifactModel = (data) => service({ url: '/llmops/artifact/registerModel', method: 'post', data })
export const deleteArtifact = (params) => service({ url: '/llmops/artifact', method: 'delete', params })
