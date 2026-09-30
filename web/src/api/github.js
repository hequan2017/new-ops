import axios from 'axios'

const service = axios.create()

export function Commits(page) {
  return service({
    url:
      'https://api.github.com/repos/hequan2017/new-ops/commits?page=' +
      page,
    method: 'get'
  })
}

export function Members() {
  return service({
    url: 'https://api.github.com/orgs/hequan2017/members',
    method: 'get'
  })
}
