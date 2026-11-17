<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { showConfirmDialog, showFailToast, showSuccessToast, showToast } from 'vant'
import http from '@/api/http'
import { useSession } from '@/stores/session'
import { getGeoPosition, isNativeApp } from '@/utils/cap'

const session = useSession()
const today = ref<any>({})
const loading = ref(false)
const locating = ref(false)
const coords = ref<{ lat: number; lng: number; accuracy?: number } | null>(null)
const locateError = ref('')

const nextAction = computed(() => today.value.next_action || 'check_in')
const isDone = computed(() => nextAction.value === 'done')
const isCheckout = computed(() => nextAction.value === 'check_out')

const nextLabel = computed(() => {
  if (isCheckout.value) return '下班打卡'
  if (isDone.value) return '今日已完成'
  return '上班打卡'
})

const phaseTip = computed(() => {
  if (isDone.value) return '今天的出勤已记录完整'
  if (isCheckout.value) return '已上班，离店前请打下班卡'
  return '到店后请先完成上班打卡'
})

function distanceMeters(lat1: number, lng1: number, lat2: number, lng2: number) {
  const R = 6371000
  const toRad = Math.PI / 180
  const dLat = (lat2 - lat1) * toRad
  const dLng = (lng2 - lng1) * toRad
  const a =
    Math.sin(dLat / 2) ** 2 +
    Math.cos(lat1 * toRad) * Math.cos(lat2 * toRad) * Math.sin(dLng / 2) ** 2
  return 2 * R * Math.asin(Math.min(1, Math.sqrt(a)))
}

const liveDistance = computed(() => {
  if (!coords.value || !today.value.geo_ready) return null
  return Math.round(
    distanceMeters(
      coords.value.lat,
      coords.value.lng,
      Number(today.value.store_lat),
      Number(today.value.store_lng),
    ),
  )
})

const rangeM = computed(() => Number(today.value.range_m || 200))

const inRange = computed(() => {
  if (liveDistance.value == null) return null
  return liveDistance.value <= rangeM.value
})

const rangePercent = computed(() => {
  if (liveDistance.value == null) return 0
  return Math.min(100, Math.round((liveDistance.value / Math.max(rangeM.value, 1)) * 100))
})

const canPunch = computed(() => {
  return !isDone.value && !!today.value.geo_ready && inRange.value === true && !!coords.value
})

const geoStatusText = computed(() => {
  if (locating.value) return '定位中…'
  if (!today.value.geo_ready) return '门店未配置坐标'
  if (locateError.value) return '定位失败'
  if (!coords.value) return '等待定位'
  if (inRange.value) return '在打卡范围内'
  return '超出打卡范围'
})

async function load() {
  today.value = await http.get('/api/attendance/today')
}

async function locate(): Promise<{ lat: number; lng: number; accuracy?: number }> {
  locateError.value = ''
  locating.value = true
  const started = Date.now()
  try {
    // App 内走 WebView 网络定位（与浏览器同路径）；超时 8s 必须给出结果或错误
    const next = await getGeoPosition({
      enableHighAccuracy: false,
      timeout: 8000,
      maximumAge: 60000,
    })
    coords.value = next
    locateError.value = ''
    return next
  } catch (e: any) {
    locateError.value = e?.message || '定位失败'
    throw e
  } finally {
    locating.value = false
    if (import.meta.env.DEV || isNativeApp()) {
      console.info('[clockin] locate done', Date.now() - started, 'ms', coords.value, locateError.value)
    }
  }
}

async function refreshLocate() {
  try {
    await locate()
    showToast('定位已更新')
  } catch (e: any) {
    showFailToast(e.message || '定位失败')
  }
}

async function askPunch() {
  if (!today.value.geo_ready) {
    return showFailToast('门店未配置打卡坐标，请联系管理员')
  }
  if (isDone.value) return showToast('今日已完成打卡')

  loading.value = true
  try {
    if (!coords.value) await locate()
  } catch (e: any) {
    loading.value = false
    return showFailToast(e.message || '定位失败')
  }
  loading.value = false

  if (inRange.value !== true) {
    return showFailToast(
      `当前距门店约 ${liveDistance.value ?? '—'} 米，超出允许范围 ${rangeM.value} 米`,
    )
  }

  const actionName = isCheckout.value ? '下班打卡' : '上班打卡'
  try {
    await showConfirmDialog({
      title: `确认${actionName}？`,
      message:
        `门店：${today.value.store_name || '本店'}\n` +
        `距离：约 ${liveDistance.value} 米（允许 ${rangeM.value} 米）\n` +
        `员工：${session.user?.name || '—'}\n\n` +
        `确认后将记录本次${isCheckout.value ? '下班' : '上班'}时间，请核对位置无误。`,
      confirmButtonText: `确认${actionName}`,
      confirmButtonColor: '#e24b6e',
      cancelButtonText: '再看看',
    })
  } catch {
    return
  }

  await doPunch()
}

async function doPunch() {
  loading.value = true
  try {
    const pos = coords.value || (await locate())
    const data: any = await http.post('/api/attendance/check', {
      method: '定位',
      lat: pos.lat,
      lng: pos.lng,
    })
    const tip = data?.action === 'check_out' ? '下班打卡成功' : '上班打卡成功'
    showSuccessToast(`${tip} · 距门店 ${data?.distance_m ?? '—'} 米`)
    await load()
  } catch (e: any) {
    showFailToast(e.message || '打卡失败')
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await load()
  try {
    await locate()
  } catch {
    // 权限失败时仍展示页面
  }
})
</script>

<template>
  <div class="page-wrap clockin">
    <van-nav-bar title="出勤打卡" left-arrow @click-left="$router.back()" />

    <header class="hero lp-hero">
      <div class="hero-top">
        <div>
          <div class="label">定位围栏打卡{{ isNativeApp() ? ' · App' : '' }}</div>
          <div class="date">{{ today.date || '—' }}</div>
        </div>
        <div class="phase" :class="nextAction">{{ phaseTip }}</div>
      </div>
      <div class="store-row">
        <van-icon name="shop-o" />
        <span>{{ today.store_name || session.user?.store_name || '未绑定门店' }}</span>
        <em>方圆 {{ rangeM }} 米</em>
      </div>

      <div class="status">
        <div class="st" :class="{ on: today.checked_in }">
          <div class="st-ico"><van-icon name="clock-o" /></div>
          <div>
            <span>上班</span>
            <b>{{ today.check_in || '未打卡' }}</b>
          </div>
        </div>
        <div class="st" :class="{ on: today.checked_out }">
          <div class="st-ico"><van-icon name="passed" /></div>
          <div>
            <span>下班</span>
            <b>{{ today.check_out || '未打卡' }}</b>
          </div>
        </div>
      </div>
    </header>

    <section class="card geo-card">
      <div class="card-hd">
        <div class="sec">位置状态</div>
        <button type="button" class="link-btn" :disabled="locating" @click="refreshLocate">
          {{ locating ? '定位中…' : '重新定位' }}
        </button>
      </div>

      <div class="ring-wrap">
        <div
          class="ring"
          :class="{
            locating,
            ok: inRange === true,
            bad: inRange === false,
            idle: inRange == null,
          }"
        >
          <div class="ring-inner">
            <div class="ring-num" v-if="liveDistance != null">{{ liveDistance }}</div>
            <div class="ring-num muted" v-else-if="locating">···</div>
            <div class="ring-num muted" v-else>—</div>
            <div class="ring-unit">米</div>
            <div class="ring-tag">{{ geoStatusText }}</div>
          </div>
        </div>
        <div class="meter">
          <div class="meter-track">
            <div
              class="meter-fill"
              :class="{ bad: inRange === false, ok: inRange === true }"
              :style="{ width: `${rangePercent}%` }"
            />
          </div>
          <div class="meter-labels">
            <span>0</span>
            <span>允许 {{ rangeM }} 米</span>
          </div>
        </div>
      </div>

      <div class="geo-meta">
        <div v-if="coords" class="meta-line">
          <van-icon name="location-o" />
          精度约 {{ Math.round(coords.accuracy || 0) }} 米
          <span class="dot">·</span>
          {{ coords.lat.toFixed(5) }}, {{ coords.lng.toFixed(5) }}
        </div>
        <div v-else-if="locateError" class="meta-line bad">{{ locateError }}</div>
        <div v-else class="meta-line muted">正在获取位置…</div>
      </div>
    </section>

    <section class="card action-card">
      <div v-if="!today.geo_ready" class="warn">
        门店尚未配置打卡坐标，请管理员在后台「门店管理」用地图选点后再试。
      </div>
      <div v-else-if="inRange === false" class="warn">
        你当前不在打卡范围内。请走到门店附近后点「重新定位」，再打卡。
      </div>
     

      <van-button
        block
        round
        type="primary"
        class="punch-btn"
        :loading="loading || locating"
        :disabled="isDone || !today.geo_ready"
        @click="askPunch"
      >
        {{ nextLabel }}
      </van-button>
      <div class="foot">{{ session.user?.name || '员工' }} · {{ session.user?.role_name || '' }}</div>
    </section>
  </div>
</template>

<style scoped>
.clockin {
  padding-bottom: 8px;
}

.hero {
  margin: 4px 12px 8px;
  padding: 12px 12px 10px;
}
.hero-top {
  position: relative;
  z-index: 1;
  display: flex;
  justify-content: space-between;
  gap: 8px;
  align-items: flex-start;
}
.label {
  font-size: 11px;
  opacity: .9;
  letter-spacing: .04em;
}
.date {
  margin-top: 3px;
  font-family: var(--lp-display);
  font-size: 20px;
  font-weight: 700;
}
.phase {
  max-width: 46%;
  padding: 5px 8px;
  border-radius: 10px;
  font-size: 10px;
  line-height: 1.35;
  background: rgba(255,255,255,.16);
  border: 1px solid rgba(255,255,255,.18);
}
.phase.check_in { background: rgba(255, 196, 120, .22); }
.phase.check_out { background: rgba(120, 210, 180, .22); }
.phase.done { background: rgba(255,255,255,.14); }

.store-row {
  position: relative;
  z-index: 1;
  margin-top: 8px;
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  opacity: .92;
}
.store-row em {
  margin-left: auto;
  font-style: normal;
  padding: 1px 6px;
  border-radius: 999px;
  background: rgba(255,255,255,.16);
  font-size: 10px;
}

.status {
  position: relative;
  z-index: 1;
  margin-top: 10px;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 6px;
}
.st {
  display: flex;
  gap: 8px;
  align-items: center;
  padding: 8px;
  border-radius: 12px;
  background: rgba(120, 20, 45, 0.16);
  border: 1px solid rgba(255,255,255,.14);
}
.st.on {
  background: rgba(255,255,255,.18);
  border-color: rgba(255,255,255,.28);
}
.st-ico {
  width: 28px;
  height: 28px;
  border-radius: 9px;
  display: grid;
  place-items: center;
  background: rgba(255,255,255,.16);
  font-size: 14px;
}
.st span {
  display: block;
  font-size: 10px;
  opacity: .88;
}
.st b {
  display: block;
  margin-top: 2px;
  font-size: 13px;
  font-weight: 720;
}

.card {
  margin: 0 12px 8px;
  padding: 10px;
  border-radius: 16px;
  background: rgba(255,255,255,.96);
  box-shadow: var(--lp-shadow-card);
  border: 1px solid rgba(255,245,248,.95);
}
.card-hd {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 4px;
}
.sec {
  font-size: 13px;
  font-weight: 720;
  color: #1f2937;
}
.link-btn {
  border: 0;
  background: none;
  color: #e24b6e;
  font-size: 12px;
  font-weight: 600;
  padding: 0;
}
.link-btn:disabled { opacity: .5; }

.ring-wrap {
  display: grid;
  justify-items: center;
  gap: 8px;
  padding: 4px 0 2px;
}
.ring {
  width: 128px;
  height: 128px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  background:
    radial-gradient(circle at center, rgba(255,255,255,.95) 0 54%, transparent 55%),
    conic-gradient(from 210deg, #f3d6df, #e24b6e, #f7b5c6, #f3d6df);
  box-shadow: 0 10px 28px rgba(226, 75, 110, 0.16);
  position: relative;
}
.ring.ok {
  background:
    radial-gradient(circle at center, #fff 0 54%, transparent 55%),
    conic-gradient(from 210deg, #c6f0d8, #12b76a, #a6f0c6, #c6f0d8);
  box-shadow: 0 10px 28px rgba(18, 183, 106, 0.16);
}
.ring.bad {
  background:
    radial-gradient(circle at center, #fff 0 54%, transparent 55%),
    conic-gradient(from 210deg, #fde2e2, #f04438, #fecaca, #fde2e2);
  box-shadow: 0 10px 28px rgba(240, 68, 56, 0.14);
}
.ring.idle {
  background:
    radial-gradient(circle at center, #fff 0 54%, transparent 55%),
    conic-gradient(from 210deg, #e8eef5, #98a2b3, #d0d5dd, #e8eef5);
  box-shadow: none;
}
.ring.locating {
  animation: pulse 1.2s ease-in-out infinite;
}
.ring-inner {
  width: 90px;
  height: 90px;
  border-radius: 50%;
  background: #fff;
  display: grid;
  place-content: center;
  text-align: center;
  box-shadow: inset 0 0 0 1px rgba(15, 23, 42, 0.04);
}
.ring-num {
  font-family: var(--lp-display);
  font-size: 26px;
  font-weight: 760;
  line-height: 1;
  color: #1f2937;
}
.ring-num.muted { color: #98a2b3; font-size: 22px; }
.ring-unit {
  margin-top: 1px;
  font-size: 11px;
  color: #667085;
}
.ring-tag {
  margin-top: 4px;
  font-size: 10px;
  font-weight: 650;
  color: #e24b6e;
}
.ring.ok .ring-tag { color: #039855; }
.ring.bad .ring-tag { color: #d92d20; }
.ring.idle .ring-tag { color: #667085; }

.meter { width: 100%; max-width: 260px; }
.meter-track {
  height: 8px;
  border-radius: 999px;
  background: #f2f4f7;
  overflow: hidden;
}
.meter-fill {
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, #f7b5c6, #e24b6e);
  transition: width .25s ease;
}
.meter-fill.ok { background: linear-gradient(90deg, #6ce9a6, #12b76a); }
.meter-fill.bad { background: linear-gradient(90deg, #fda29b, #f04438); }
.meter-labels {
  margin-top: 6px;
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: #98a2b3;
}

.geo-meta { margin-top: 8px; }
.meta-line {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 4px;
  font-size: 11px;
  color: #475467;
  line-height: 1.4;
}
.meta-line .dot { opacity: .5; }
.meta-line.muted { color: #98a2b3; }
.meta-line.bad { color: #d92d20; }

.warn {
  padding: 8px 10px;
  border-radius: 10px;
  background: #fff7e6;
  color: #ad6800;
  font-size: 11px;
  line-height: 1.45;
  margin-bottom: 8px;
}
.hint {
  font-size: 11px;
  color: #667085;
  line-height: 1.4;
  margin-bottom: 8px;
}
.punch-btn {
  height: 40px !important;
  font-size: 14px !important;
  font-weight: 720 !important;
}
.foot {
  margin-top: 8px;
  text-align: center;
  font-size: 11px;
  color: #98a2b3;
}

@keyframes pulse {
  0%, 100% { transform: scale(1); opacity: 1; }
  50% { transform: scale(0.98); opacity: .88; }
}
</style>
