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
    return data.user as User
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('lp_token')
    localStorage.removeItem('lp_user')
  }

  return { token, user, login, logout }
})
