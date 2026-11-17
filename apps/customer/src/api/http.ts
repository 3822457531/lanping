import axios from 'axios'
import { showToast } from 'vant'

const http = axios.create({ baseURL: '/', timeout: 20000 })

http.interceptors.request.use((config) => {
  const token = localStorage.getItem('lp_customer_token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

http.interceptors.response.use(
  (res) => {
    const body = res.data
    if (body && body.ok === false) return Promise.reject(new Error(body.error || '请求失败'))
    return body?.data !== undefined ? body.data : body
  },
  (err) => {
    showToast(err.response?.data?.error || err.message || '网络错误')
    return Promise.reject(err)
  },
)

export default http
