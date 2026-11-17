<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import QRCode from 'qrcode'
import { showToast } from 'vant'
import http from '@/api/http'

const route = useRoute()
const data = ref<any>(null)
const loading = ref(true)
const left = ref(0)
const canvasRef = ref<HTMLCanvasElement | null>(null)
let timer: number | undefined

const expired = computed(() => left.value <= 0 && data.value?.status === '待确认')
const pageUrl = computed(() => location.href)

async function drawQR() {
  await nextTick()
  if (!canvasRef.value || !data.value) return
  await QRCode.toCanvas(canvasRef.value, pageUrl.value, {
    width: 220,
    margin: 1,
    color: { dark: '#2a2422', light: '#ffffff' },
  })
}

async function load() {
  loading.value = true
  try {
    data.value = await http.get(`/api/public/verification/${route.params.token}`)
    if (data.value.expire_at) {
      left.value = Math.max(0, Math.floor((new Date(data.value.expire_at).getTime() - Date.now()) / 1000))
    }
    await drawQR()
  } catch (e: any) {
    showToast(e.message || '核销码无效')
  } finally {
    loading.value = false
  }
}

watch(() => data.value?.token, () => drawQR())

onMounted(() => {
  load()
  timer = window.setInterval(() => {
    if (left.value > 0) left.value -= 1
  }, 1000)
})
onUnmounted(() => { if (timer) clearInterval(timer) })
</script>

<template>
  <div class="page">
    <div v-if="data" class="card">
      <div class="brand">兰瓶美业 · 核销码</div>
      <div class="status" :class="{ done: data.status==='已完成', wait: data.status==='待确认' }">{{ data.status }}</div>

      <canvas ref="canvasRef" class="qr" />

      <div class="name">{{ data.customer_name || '会员' }}</div>
      <div class="project">{{ data.project_name }} · {{ data.times }} 次</div>
      <div class="meta">{{ data.store_name }} · ¥{{ data.amount }}</div>
      <div class="meta" v-if="data.status==='待确认' && !expired">
        有效倒计时 {{ Math.floor(left/60) }}:{{ String(left%60).padStart(2,'0') }}
      </div>
      <div class="warn" v-if="expired">已过期，请让员工重新发起核销</div>
      <div class="tip">请将此二维码出示给店员扫描完成核销</div>
    </div>
    <van-loading v-else-if="loading" vertical>加载核销码…</van-loading>
  </div>
</template>

<style scoped>
.page {
  min-height: 100vh; padding: 28px 16px;
  background:
    radial-gradient(circle at 20% 0%, #f8d5de, transparent 42%),
    #f6f3f1;
  display: flex; justify-content: center;
}
.card {
  width: min(420px, 100%); background:#fff; border-radius:22px; padding:26px 18px 28px;
  text-align:center; border:1px solid #efe4e0; box-shadow: 0 12px 36px rgba(80,40,50,.06);
  height: fit-content;
}
.brand { color:#d45a72; font-weight:700; font-size:13px; letter-spacing:.5px; }
.status { margin-top: 10px; font-size:13px; color:#8f857e; }
.status.wait { color:#c9822d; }
.status.done { color:#2f9e74; }
.qr { margin: 18px auto 8px; width:220px; height:220px; }
.name { margin-top: 10px; font-size: 22px; font-weight: 750; color:#2a2422; }
.project { margin-top: 8px; font-size: 15px; font-weight: 650; color:#d45a72; }
.meta { margin-top: 6px; font-size: 12.5px; color:#8f857e; }
.warn { margin-top: 12px; color:#d45454; font-size: 13px; }
.tip { margin-top: 18px; font-size: 12px; color:#a39992; line-height: 1.6; }
</style>
