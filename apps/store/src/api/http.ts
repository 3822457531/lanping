import axios from 'axios'
import { showToast } from 'vant'

// H5 开发默认走 Vite 代理 `/`；Capacitor APK 需配置 VITE_API_BASE=http://电脑局域网IP:8080
const apiBase = (import.meta.env.VITE_API_BASE as string | undefined)?.replace(/\/$/, '') || '/'

const http = axios.create({
  baseURL: apiBase,
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
    if (body && body.ok === false) return Promise.reject(new Error(body.error || '请求失败'))
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
      // 高于业务弹窗（3000），避免错误提示被挡住
      showToast({ message: msg, zIndex: 4000 })
    }
    return Promise.reject(new Error(msg))
  },
)

export default http
