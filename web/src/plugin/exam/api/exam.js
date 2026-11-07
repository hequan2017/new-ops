import service from '@/utils/request'

// 题库
export const getQuestions = (params) => service({ url: '/exam/question/list', method: 'get', params })
export const saveQuestion = (data) => service({ url: '/exam/question', method: 'post', data })
export const deleteQuestion = (params) => service({ url: '/exam/question', method: 'delete', params })
export const importQuestions = (file) => {
  const data = new FormData()
  data.append('file', file)
  return service({ url: '/exam/question/import', method: 'post', data })
}

// 试卷
export const getPapers = () => service({ url: '/exam/paper/list', method: 'get' })
export const savePaper = (data) => service({ url: '/exam/paper', method: 'post', data })
export const composePaper = (id) => service({ url: '/exam/paper/compose', method: 'post', data: { id } })
export const deletePaper = (params) => service({ url: '/exam/paper', method: 'delete', params })

// 考试场次
export const getExams = (params) => service({ url: '/exam/exam/list', method: 'get', params })
export const saveExam = (data) => service({ url: '/exam/exam', method: 'post', data })
export const transitionExam = (data) => service({ url: '/exam/exam/transition', method: 'post', data })
export const deleteExam = (params) => service({ url: '/exam/exam', method: 'delete', params })

// 考生
export const getCandidates = (params) => service({ url: '/exam/candidate/list', method: 'get', params })
export const importCandidates = (data) => service({ url: '/exam/candidate/import', method: 'post', data })
export const deleteCandidate = (params) => service({ url: '/exam/candidate', method: 'delete', params })

// 成绩
export const publishScores = (data) => service({ url: '/exam/score/publish', method: 'post', data, timeout: 300000 })
export const getScores = (params) => service({ url: '/exam/score/list', method: 'get', params })
export const exportScores = (params) => service({ url: '/exam/score/export', method: 'get', params })

// 节点与防作弊
export const getNodes = () => service({ url: '/exam/node/list', method: 'get' })
export const checkSimilarity = (data) => service({ url: '/exam/anticheat/check', method: 'post', data, timeout: 120000 })
