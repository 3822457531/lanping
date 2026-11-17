import axios from 'axios'
import { ElMessage } from 'element-plus'

const http = axios.create({
  baseURL: '/',
  timeout: 20000,
  headers: { 'Content-Type': 'application/json;charset=UTF-8' },
})

http.interceptors.request.use((config) => {
  const token = localStorage.getItem('lp_token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  if (!config.headers['Content-Type']) {
    config.headers['Content-Type'] = 'application/json;charset=UTF-8'
  }
  return config
})

http.interceptors.response.use(
  (res) => {
    const body = res.data
    if (body && body.ok === false) {
      return Promise.reject(new Error(body.error || '请求失败'))
    }
    return body?.data !== undefined ? body.data : body
  },
  (err) => {
    const status = err.response?.status
    const msg = err.response?.data?.error || err.message || '网络错误'
    if (status === 401) {
      localStorage.removeItem('lp_token')
      localStorage.removeItem('lp_user')
      if (!location.pathname.includes('/login')) location.href = '/login'
    } else {
      ElMessage.error(msg)
    }
    return Promise.reject(err)
  },
)

export default http
