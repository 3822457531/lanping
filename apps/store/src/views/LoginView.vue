<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import { useSession } from '@/stores/session'

const phone = ref('13800000002')
const password = ref('123456')
const loading = ref(false)
const router = useRouter()
const session = useSession()

async function submit() {
  loading.value = true
  try {
    const user = await session.login(phone.value, password.value)
    if (!['manager', 'staff'].includes(user.role)) {
      showToast('请使用店长或员工账号')
      session.logout()
      return
    }
    router.replace('/home')
  } catch (e: any) {
    showToast(e.message || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login lp-bg">
    <div class="hero lp-hero">
      <div class="petal" aria-hidden="true" />
      <div class="brand">
        <span class="mark">兰</span>
        <div>
          <div class="brand-name">兰瓶美业</div>
          <div class="brand-sub">LANPING BEAUTY</div>
        </div>
      </div>
      <h1>门店工作台</h1>
      <p>店长与员工的日常经营入口</p>
    </div>

    <div class="panel">
      <div class="panel-hd">账号登录</div>
      <van-field v-model="phone" label="手机号" placeholder="请输入手机号" clearable />
      <van-field v-model="password" type="password" label="密码" placeholder="请输入密码" />
      <p class="tip">体验账号：店长 13800000002 · 员工 13800000004 · 密码 123456</p>
      <van-button block round type="primary" :loading="loading" @click="submit">进入工作台</van-button>
      <van-button block round plain type="primary" class="ghost" @click="$router.push('/register')">新员工注册</van-button>
    </div>
  </div>
</template>

<style scoped>
.login {
  min-height: 100%;
  padding: 14px 18px calc(28px + env(safe-area-inset-bottom, 0px));
  animation: lp-fade-up 0.5s ease both;
}
.hero {
  padding: 14px 20px 24px;
  margin-bottom: 10px;
}
.petal {
  position: absolute;
  right: -10px;
  top: 10px;
  width: 140px;
  height: 140px;
  border-radius: 46% 54% 60% 40% / 48% 42% 58% 52%;
  background: rgba(255, 255, 255, 0.14);
  pointer-events: none;
}
.brand {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  gap: 8px;
}
.mark {
  width: 44px;
  height: 44px;
  border-radius: 16px;
  display: grid;
  place-items: center;
  font-size: 20px;
  font-weight: 760;
  background: rgba(255, 255, 255, 0.22);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.4);
}
.brand-name {
  font-size: 13px;
  font-weight: 720;
  letter-spacing: 0.12em;
}
.brand-sub {
  margin-top: 2px;
  font-family: var(--lp-display);
  font-size: 10px;
  letter-spacing: 0.18em;
  opacity: 0.85;
}
h1 {
  position: relative;
  z-index: 1;
  margin: 22px 0 8px;
  font-size: 22px;
  font-weight: 760;
  letter-spacing: 0.04em;
  line-height: 1.15;
}
p {
  position: relative;
  z-index: 1;
  margin: 0;
  opacity: 0.92;
  font-size: 13px;
}
.panel {
  padding: 10px 14px 20px;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.94);
  box-shadow: var(--lp-shadow-card);
  border: 1px solid rgba(255, 255, 255, 0.9);
  animation: lp-pop 0.45s ease 0.08s both;
}
.panel-hd {
  font-size: 13px;
  font-weight: 720;
  margin: 0 8px 8px;
  color: var(--lp-text);
}
.tip {
  color: var(--lp-muted);
  font-size: 12px;
  line-height: 1.65;
  padding: 6px 10px 16px;
  margin: 0;
  opacity: 1;
}
.ghost { margin-top: 10px; }
</style>
