<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { showSuccessToast, showToast } from 'vant'
import axios from 'axios'

const router = useRouter()
const loading = ref(false)
const stores = ref<any[]>([])
const form = reactive({
  name: '',
  phone: '',
  store_id: undefined as number | undefined,
})

async function loadStores() {
  try {
    const res = await axios.get('/api/public/stores')
    const body = res.data
    const data = body?.data !== undefined ? body.data : body
    stores.value = data.list || []
  } catch {
    stores.value = []
  }
}

async function submit() {
  if (!form.name.trim()) return showToast('请填写姓名')
  if (!/^1\d{10}$/.test(form.phone)) return showToast('请填写正确手机号')
  if (!form.store_id) return showToast('请选择所在门店')
  loading.value = true
  try {
    const res = await axios.post('/api/auth/register', {
      name: form.name.trim(),
      phone: form.phone.trim(),
      store_id: form.store_id,
      role: 'staff',
    }, { headers: { 'Content-Type': 'application/json;charset=UTF-8' } })
    const body = res.data
    if (body && body.ok === false) throw new Error(body.error || '提交失败')
    showSuccessToast('已提交，等待审核')
    setTimeout(() => router.replace('/login'), 800)
  } catch (e: any) {
    const msg = e.response?.data?.error || e.message || '提交失败'
    showToast(msg)
  } finally {
    loading.value = false
  }
}

onMounted(loadStores)
</script>

<template>
  <div class="page lp-bg">
    <div class="hero lp-hero">
      <div class="mark">兰</div>
      <div class="brand">兰瓶美业</div>
      <h1>员工注册</h1>
      <p>填写资料后提交，管理员审核通过即可登录</p>
    </div>
    <div class="panel">
      <van-field v-model="form.name" label="姓名" required placeholder="请输入真实姓名" />
      <van-field v-model="form.phone" type="tel" maxlength="11" label="手机号" required placeholder="登录手机号" />
      <van-field label="所在门店" required>
        <template #input>
          <select v-model="form.store_id" class="sel">
            <option :value="undefined" disabled>请选择门店</option>
            <option v-for="s in stores" :key="s.id" :value="s.id">{{ s.name }}</option>
          </select>
        </template>
      </van-field>
      <p class="tip">审核通过后默认密码为 123456，请登录后联系店长修改。</p>
      <van-button block round type="primary" :loading="loading" @click="submit">提交注册</van-button>
      <van-button block round plain type="primary" class="ghost" @click="$router.replace('/login')">已有账号去登录</van-button>
    </div>
  </div>
</template>

<style scoped>
.page {
  min-height: 100%;
  padding: 18px 18px 36px;
  animation: lp-fade-up 0.5s ease both;
}
.hero {
  padding: 24px 20px 22px;
  margin-bottom: 10px;
}
.mark {
  position: relative;
  z-index: 1;
  width: 48px;
  height: 40px;
  border-radius: 16px;
  display: grid;
  place-items: center;
  font-size: 14px;
  font-weight: 760;
  background: rgba(255, 255, 255, 0.22);
}
.brand {
  position: relative;
  z-index: 1;
  margin-top: 14px;
  font-size: 12px;
  font-weight: 650;
  letter-spacing: 0.16em;
  opacity: 0.95;
}
h1 {
  position: relative;
  z-index: 1;
  margin: 10px 0 8px;
  font-size: 28px;
  font-weight: 760;
  letter-spacing: 0.04em;
}
p {
  position: relative;
  z-index: 1;
  margin: 0;
  opacity: 0.92;
  font-size: 13px;
  line-height: 1.5;
}
.panel {
  padding: 10px 12px 20px;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.95);
  box-shadow: var(--lp-shadow-card);
  border: 1px solid rgba(255, 255, 255, 0.9);
}
.tip {
  color: var(--lp-muted);
  font-size: 12px;
  line-height: 1.65;
  padding: 8px 10px 16px;
  margin: 0;
  opacity: 1;
}
.ghost { margin-top: 10px; }
</style>
