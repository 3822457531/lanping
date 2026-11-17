import { defineStore } from 'pinia'
import { ref } from 'vue'
import http from '@/api/http'

export type User = {
  id: number
  name: string
  phone: string
  role: string
  role_name: string
  store_id?: number
  store_name?: string
}

export const useSession = defineStore('session', () => {
  const token = ref(localStorage.getItem('lp_token') || '')
  const user = ref<User | null>(JSON.parse(localStorage.getItem('lp_user') || 'null'))

  async function login(phone: string, password: string) {
    const data: any = await http.post('/api/auth/login', { phone, password })
    token.value = data.token
    user.value = data.user
    localStorage.setItem('lp_token', data.token)
    localStorage.setItem('lp_user', JSON.stringify(data.user))
    return data.user
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('lp_token')
    localStorage.removeItem('lp_user')
  }

  async function refresh() {
    if (!token.value) return null
    const me: any = await http.get('/api/auth/me')
    user.value = me
    localStorage.setItem('lp_user', JSON.stringify(me))
    return me
  }

  return { token, user, login, logout, refresh }
})
