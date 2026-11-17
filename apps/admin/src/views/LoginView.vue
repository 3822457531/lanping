<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useSession } from '@/stores/session'

const phone = ref('13800000001')
const password = ref('123456')
const loading = ref(false)
const router = useRouter()
const session = useSession()

async function submit() {
  loading.value = true
  try {
    const user = await session.login(phone.value, password.value)
    if (user.role !== 'admin') {
      ElMessage.warning('管理后台仅限管理员账号')
      session.logout()
      return
    }
    router.push('/dashboard')
  } catch (e: any) {
    ElMessage.error(e.message || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login">
    <div class="panel">
      <div class="brand">兰瓶美业 · 管理后台</div>
      <h1>欢迎回来</h1>
      <p class="hint">演示账号 13800000001 / 123456</p>
      <el-form @submit.prevent="submit" label-position="top">
        <el-form-item label="手机号">
          <el-input v-model="phone" size="large" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="password" type="password" show-password size="large" />
        </el-form-item>
        <el-button type="primary" size="large" style="width:100%" :loading="loading" @click="submit">登录</el-button>
      </el-form>
    </div>
  </div>
</template>

<style scoped>
.login {
  min-height: 100vh; display: grid; place-items: center;
  background:
    radial-gradient(circle at 12% 10%, #f8e9ec 0%, transparent 40%),
    radial-gradient(circle at 90% 90%, #e8eef5 0%, transparent 42%),
    var(--lp-bg);
}
.panel {
  width: 400px; background: #fff; border: 1px solid var(--lp-border);
  border-radius: 16px; padding: 32px 28px;
}
.brand { color: var(--lp-accent); font-size: 13px; font-weight: 600; }
h1 { margin: 8px 0 4px; font-size: 26px; }
.hint { margin: 0 0 22px; color: var(--lp-muted); font-size: 13px; }
</style>
