import service from '@/utils/request'

// 终端公开面（机器码白名单 + 券码凭证）
export const terminalRegister = (data) => service({ url: '/exam/terminal/register', method: 'post', data })
export const terminalHeartbeat = (data) => service({ url: '/exam/terminal/heartbeat', method: 'post', data })
export const terminalExams = (nodeKey) => service({ url: '/exam/terminal/exams', method: 'get', headers: { 'X-Terminal-Key': nodeKey } })
export const terminalResume = (data) => service({ url: '/exam/terminal/resume', method: 'post', data, timeout: 60000 })
export const terminalSave = (data) => service({ url: '/exam/terminal/save', method: 'post', data, timeout: 60000 })
export const terminalSubmit = (data) => service({ url: '/exam/terminal/submit', method: 'post', data, timeout: 60000 })
export const terminalPractice = (data) => service({ url: '/exam/terminal/practice', method: 'post', data, timeout: 180000 })
