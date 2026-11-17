<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import http from '@/api/http'

const phone = ref('13900000001')
const loading = ref(false)
const router = useRouter()

async function submit() {
  loading.value = true
  try {
    const data: any = await http.post('/api/customer/login', { phone: phone.value })
    localStorage.setItem('lp_customer_token', data.token)
    localStorage.setItem('lp_customer_user', JSON.stringify(data.user))
    router.replace('/home')
  } catch (e: any) {
    showToast(e.message || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="page">
    <h1>会员登录</h1>
    <p>演示手机号 13900000001</p>
    <van-field v-model="phone" label="手机号" />
    <van-button block type="primary" round :loading="loading" style="margin-top:16px" @click="submit">进入</van-button>
  </div>
</template>

<style scoped>
.page { padding: 48px 20px; }
h1 { margin: 0 0 8px; }
p { color: var(--lp-muted); margin: 0 0 24px; font-size: 13px; }
</style>
