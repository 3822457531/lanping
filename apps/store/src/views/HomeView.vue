<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import http from '@/api/http'
import { useSession } from '@/stores/session'

const session = useSession()
const router = useRouter()
const data = ref<any>({})
const loading = ref(true)

const hourTip = computed(() => {
  const h = new Date().getHours()
  if (h < 11) return '上午好'
  if (h < 14) return '中午好'
  if (h < 18) return '下午好'
  return '晚上好'
})

const brandName = computed(() => session.user?.store_name || '兰瓶美业')
const todos = computed(() => data.value.todos || [])
const todayAppts = computed(() => data.value.today_appointments || [])

onMounted(async () => {
  try {
    data.value = await http.get('/api/workbench')
  } finally {
    loading.value = false
  }
})

function statusClass(st: string) {
  if (st === '已履约') return 'ok'
  if (st === '已改约') return 'wait'
  if (st === '待履约' || st === '待确认' || st === '已确认') return 'wait'
  return ''
}
function statusLabel(st: string) {
  if (st === '待确认' || st === '已确认') return '待履约'
  if (st === '失效') return '已失效'
  return st
}

function goTodo(t: any) {
  const raw = String(t?.path || '/appointments')
  if (raw.includes('?')) {
    const [path, qs] = raw.split('?')
    const query: Record<string, string> = {}
    new URLSearchParams(qs).forEach((v, k) => { query[k] = v })
    router.push({ path, query })
    return
  }
  router.push(raw)
}
</script>

<template>
  <div class="home">
    <header class="hero lp-hero">
      <div class="brand-row">
        <span class="mark">兰</span>
        <div>
          <div class="brand">{{ brandName.replace(/总店$/, '') || '兰瓶美业' }}</div>
          <div class="brand-en">LANPING STORE</div>
        </div>
      </div>

      <div class="greet">{{ hourTip }}，{{ session.user?.name }}</div>
      <div class="meta">{{ session.user?.role_name || '美业员工' }} · {{ session.user?.store_name || '本店' }}</div>

      <div class="kpi">
        <div class="kpi-item">
          <div class="kpi-top">
            <van-icon name="coupon-o" />
            <span class="n">{{ data.verify_count || 0 }}</span>
          </div>
          <div class="l">本月核销</div>
        </div>
        <div class="kpi-item">
          <div class="kpi-top">
            <van-icon name="gold-coin-o" />
            <span class="n">¥{{ Number(data.verify_amount || 0).toLocaleString() }}</span>
          </div>
          <div class="l">核销业绩</div>
        </div>
        <div class="kpi-item">
          <div class="kpi-top">
            <van-icon name="calendar-o" />
            <span class="n">{{ data.today_appt_count || todayAppts.length || 0 }}</span>
          </div>
          <div class="l">今日预约</div>
        </div>
      </div>
    </header>

    <section class="block">
      <div class="lp-sec-hd">
        <h3>今日待办</h3>
        <button type="button" class="lp-sec-link" @click="router.push('/verify')">去核销 ›</button>
      </div>
      <div class="todo-list">
        <button
          v-for="(t, i) in todos"
          :key="i"
          type="button"
          class="todo"
          @click="goTodo(t)"
        >
          <div class="todo-ico" :class="t.tone || 'info'">
            <van-icon :name="t.path === '/verify' ? 'passed' : (t.path === '/employees' ? 'friends-o' : 'calendar-o')" />
          </div>
          <div>
            <div class="todo-t">{{ t.title }}</div>
            <div class="todo-d">{{ t.desc }}</div>
          </div>
          <van-icon name="arrow" class="arrow" />
        </button>
        <div v-if="!todos.length" class="lp-empty soft">暂无待办</div>
      </div>
    </section>

    <section class="block">
      <div class="lp-sec-hd">
        <h3>今日预约</h3>
        <button type="button" class="lp-sec-link" @click="router.push('/appointments')">全部 ›</button>
      </div>
      <div v-if="todayAppts.length" class="appt-list">
        <div v-for="a in todayAppts" :key="a.id" class="appt" @click="router.push('/appointments')">
          <div class="time">{{ String(a.time || '').slice(0, 5) }}</div>
          <div class="divider" />
          <div>
            <div class="name">{{ a.customer_name || '客户' }}</div>
            <div class="sub">{{ a.project_name }} · {{ a.staff_name || '未指定' }}</div>
          </div>
          <span class="lp-pill" :class="statusClass(a.status)">{{ statusLabel(a.status) }}</span>
        </div>
      </div>
      <div v-else class="lp-empty soft">今日暂无预约排期</div>
    </section>

    <section class="block">
      <div class="lp-sec-hd"><h3>本月概况</h3></div>
      <div class="facts">
        <div><span>开卡金额</span><b class="num">¥{{ Number(data.card_amount || 0).toLocaleString() }}</b></div>
        <div><span>新增客户</span><b class="num">{{ data.new_customers || 0 }}</b></div>
        <div><span>待确认核销</span><b class="num">{{ data.pending_verify || 0 }}</b></div>
      </div>
    </section>

    <div v-if="loading" class="loading-mask">加载中…</div>
  </div>
</template>

<style scoped>
.home {
  padding: 8px var(--lp-page-x) 4px;
  position: relative;
}
.hero {
  padding: 12px 12px 10px;
  margin-bottom: 8px;
}
.brand-row {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  gap: 8px;
}
.mark {
  width: 28px;
  height: 28px;
  border-radius: 8px;
  display: grid;
  place-items: center;
  font-weight: 760;
  font-size: 13px;
  background: rgba(255, 255, 255, 0.22);
}
.brand { font-size: 12px; font-weight: 700; letter-spacing: 0.06em; }
.brand-en {
  margin-top: 1px;
  font-family: var(--lp-display);
  font-size: 8px;
  letter-spacing: 0.16em;
  opacity: 0.82;
}
.greet {
  position: relative;
  z-index: 1;
  margin-top: 10px;
  font-size: 20px;
  font-weight: 760;
  line-height: 1.2;
}
.meta {
  position: relative;
  z-index: 1;
  margin-top: 3px;
  font-size: 11px;
  opacity: 0.9;
}
.kpi {
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 6px;
  margin-top: 10px;
}
.kpi-item {
  padding: 7px 6px 8px;
  border-radius: 12px;
  background: rgba(120, 20, 45, 0.16);
  border: 1px solid rgba(255, 255, 255, 0.14);
}
.kpi-top { display: flex; align-items: center; gap: 3px; min-height: 18px; }
.kpi-top .van-icon { font-size: 12px; opacity: 0.9; }
.kpi .n {
  font-family: var(--lp-display);
  font-size: 13px;
  font-weight: 700;
}
.kpi .l { margin-top: 3px; font-size: 10px; opacity: 0.88; }

.block {
  margin-bottom: 8px;
  padding: 10px 10px 8px;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.94);
  box-shadow: var(--lp-shadow-card);
  border: 1px solid rgba(255, 245, 248, 0.95);
}
.todo-list { display: grid; gap: 6px; }
.todo {
  display: grid;
  grid-template-columns: 34px 1fr auto;
  gap: 8px;
  align-items: center;
  text-align: left;
  border: 0;
  border-radius: 12px;
  padding: 8px;
  background: linear-gradient(135deg, #fff8fa, #fff5f7);
}
.todo-ico {
  width: 34px;
  height: 34px;
  border-radius: 10px;
  display: grid;
  place-items: center;
  color: #fff;
  font-size: 16px;
  background: linear-gradient(145deg, #ff8fa6, #e24b6e);
}
.todo-ico.warn { background: linear-gradient(145deg, #ffb074, #f08a45); }
.todo-ico.accent { background: linear-gradient(145deg, #f39ab0, #d45a72); }
.todo-ico.ok { background: linear-gradient(145deg, #7ed6b0, #3cae84); }
.todo-t { font-size: 13px; font-weight: 700; color: #2a2124; }
.todo-d { margin-top: 1px; font-size: 11px; color: var(--lp-muted); }
.arrow { color: #d4c4c8; font-size: 12px; }

.appt-list { display: grid; gap: 0; }
.appt {
  display: grid;
  grid-template-columns: 42px 1px 1fr auto;
  gap: 8px;
  align-items: center;
  padding: 8px 2px;
}
.time {
  font-family: var(--lp-display);
  font-weight: 700;
  color: var(--lp-accent);
  font-size: 14px;
}
.divider {
  width: 1px;
  height: 26px;
  background: linear-gradient(180deg, transparent, #f0e2e6, transparent);
}
.name { font-weight: 700; font-size: 13px; }
.sub { margin-top: 2px; font-size: 11px; color: var(--lp-muted); }

.facts > div {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 2px;
  border-bottom: 1px dashed #f0e2e6;
}
.facts > div:last-child { border-bottom: 0; }
.facts span { color: var(--lp-muted); font-size: 12px; }
.facts .num {
  font-family: var(--lp-display);
  font-size: 14px;
  color: #2a2124;
  font-weight: 700;
}

.lp-empty.soft {
  padding: 10px 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-size: 12px;
}
.loading-mask {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  background: rgba(253, 246, 248, 0.55);
  color: var(--lp-muted);
  font-size: 12px;
  pointer-events: none;
}
</style>
