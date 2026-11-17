<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import QRCode from 'qrcode'
import { showSuccessToast, showToast } from 'vant'
import http from '@/api/http'
import { useSession } from '@/stores/session'

const session = useSession()
const route = useRoute()
const router = useRouter()
const bootstrapped = ref(false)
const step = ref(1)
const showCustomerPick = ref(false)
const showProjectPick = ref(false)
const q = ref('')
const searching = ref(false)
const searched = ref(false)
const customers = ref<any[]>([])
const projects = ref<any[]>([])
const selectedCustomer = ref<any>(null)
const amountCard = ref<any>(null)
const selectedProject = ref<any>(null)
const consumeAmount = ref(0)
const welfareList = ref<any[]>([])
const welfareId = ref(0)
const draft = ref<any>(null)
const loading = ref(false)
const canvasRef = ref<HTMLCanvasElement | null>(null)
const verifyMode = ref<'staff' | 'notify'>('staff')
let pollTimer: number | undefined
let searchTimer: number | undefined

const isNotify = computed(() => verifyMode.value === 'notify')
const balance = computed(() => Number(amountCard.value?.amount || 0))
const selectedWelfare = computed(() => welfareList.value.find((w) => w.id === welfareId.value))
const welfareOffset = computed(() => {
  if (!selectedWelfare.value) return 0
  return Math.min(Number(selectedWelfare.value.remain_amount || 0), Number(consumeAmount.value || 0))
})
const payAmount = computed(() => Math.max(0, Number(consumeAmount.value || 0) - welfareOffset.value))

const steps = computed(() => {
  if (isNotify.value) {
    return [
      { n: 1, label: '客户' },
      { n: 2, label: '发码' },
      { n: 3, label: '扫码' },
    ]
  }
  return [
    { n: 1, label: '客户' },
    { n: 2, label: '抵扣' },
    { n: 3, label: '确认' },
  ]
})

const modeLabel = computed(() => (isNotify.value ? '发通知 · 客户亮码' : '不发通知 · 员工自助'))
const stepTitle = computed(() => {
  if (step.value === 1) return '选择客户'
  if (step.value === 2) return isNotify.value ? '选择项目并生成核销码' : '选择项目 · 福利自动抵扣'
  return isNotify.value ? '等待客户亮码确认' : '核对信息并确认扣款'
})

function normalizeMode(v: string): 'staff' | 'notify' {
  if (v === 'notify' || v === 'real') return 'notify'
  return 'staff'
}

async function loadSettings() {
  try {
    const s: any = await http.get('/api/settings')
    verifyMode.value = normalizeMode(s.verify_mode || 'staff')
  } catch {
    verifyMode.value = 'staff'
  }
}

async function searchCustomers(keyword?: string) {
  const key = (keyword ?? q.value).trim()
  if (!key) {
    customers.value = []
    searched.value = false
    return
  }
  searching.value = true
  try {
    const data: any = await http.get('/api/customers', { params: { pageSize: 30, q: key } })
    customers.value = data.list || []
    searched.value = true
  } catch (e: any) {
    showToast(e.message || '搜索失败')
    customers.value = []
  } finally {
    searching.value = false
  }
}

function onSearchInput() {
  if (searchTimer) window.clearTimeout(searchTimer)
  searchTimer = window.setTimeout(() => searchCustomers(), 320)
}

function openCustomerPick() {
  q.value = ''
  customers.value = []
  searched.value = false
  showCustomerPick.value = true
}

async function loadAmountCard(customer: any) {
  const data: any = await http.get('/api/cards', {
    params: { customer_id: customer.id, pageSize: 50 },
  })
  const list = (data.list || []).filter(
    (c: any) => c.status === '使用中' && c.card_kind === 'amount' && Number(c.amount) > 0,
  )
  // 优先本店余额最多的储值卡
  const sid = session.user?.store_id
  list.sort((a: any, b: any) => {
    const as = sid && a.store_id === sid ? 1 : 0
    const bs = sid && b.store_id === sid ? 1 : 0
    if (as !== bs) return bs - as
    return Number(b.amount || 0) - Number(a.amount || 0)
  })
  amountCard.value = list[0] || null
}

async function loadAmountProjects() {
  const data: any = await http.get('/api/projects', { params: { pageSize: 100, status: '上架' } })
  projects.value = (data.list || []).filter((p: any) => {
    const m = p.card_mode || 'amount'
    return m === 'amount' || m === 'both'
  })
}

async function loadWelfare(customer: any, card: any) {
  welfareList.value = []
  welfareId.value = 0
  try {
    // 本人指定福利跨店可用；门店通用福利用员工门店（其次卡门店）
    const storeId = session.user?.store_id || card?.store_id || 0
    const data: any = await http.get('/api/welfare/available', {
      params: {
        customer_id: customer.id,
        store_id: storeId,
      },
    })
    welfareList.value = data.list || []
    // 自动选剩余最多的可用福利
    if (welfareList.value.length) {
      const sorted = [...welfareList.value].sort(
        (a, b) => Number(b.remain_amount || 0) - Number(a.remain_amount || 0),
      )
      welfareId.value = sorted[0].id
    }
  } catch (e: any) {
    welfareList.value = []
    showToast(e.message || '福利加载失败')
  }
}

async function pickCustomer(c: any, opts?: { projectId?: number; skipProjectPick?: boolean }) {
  loading.value = true
  try {
    selectedCustomer.value = c
    showCustomerPick.value = false
    await Promise.all([loadAmountCard(c), loadAmountProjects()])
    if (!amountCard.value) {
      showToast('该客户暂无可用储值余额，请先开金额卡')
      selectedCustomer.value = null
      return
    }
    await loadWelfare(c, amountCard.value)
    selectedProject.value = null
    consumeAmount.value = 0
    const pid = Number(opts?.projectId || 0)
    if (pid > 0) {
      const p = projects.value.find((x) => Number(x.id) === pid)
      if (p) {
        selectedProject.value = p
        syncAmountFromProject()
      }
    }
    step.value = 2
  } catch (e: any) {
    showToast(e.message || '加载失败')
  } finally {
    loading.value = false
  }
}

async function bootstrapFromQuery() {
  if (bootstrapped.value) return
  bootstrapped.value = true
  const draftId = Number(route.query.draft_id || 0)
  const customerId = Number(route.query.customer_id || 0)
  const projectId = Number(route.query.project_id || 0)

  if (draftId > 0) {
    loading.value = true
    try {
      const v: any = await http.get(`/api/verifications/${draftId}`)
      if (!v || v.status !== '待确认') {
        showToast(v?.status ? `该单已${v.status}` : '待确认核销单不存在')
        return
      }
      draft.value = v
      // 补齐展示用客户/项目
      try {
        const cust: any = await http.get(`/api/customers/${v.customer_id}`)
        selectedCustomer.value = cust
        await loadAmountCard(cust)
      } catch { /* ignore */ }
      if (v.project_id) {
        await loadAmountProjects()
        selectedProject.value = projects.value.find((p) => Number(p.id) === Number(v.project_id)) || {
          id: v.project_id,
          name: v.project_name || '项目',
        }
      }
      consumeAmount.value = Number(v.amount || v.total_amount || 0)
      step.value = 3
      if (isNotify.value) {
        await drawQR()
        startPoll()
      }
      // 清掉 query，避免返回再触发
      router.replace({ path: '/verify' })
    } catch (e: any) {
      showToast(e.message || '加载核销单失败')
    } finally {
      loading.value = false
    }
    return
  }

  if (customerId > 0) {
    loading.value = true
    try {
      const cust: any = await http.get(`/api/customers/${customerId}`)
      if (!cust?.id) {
        showToast('客户不存在')
        return
      }
      await pickCustomer(cust, { projectId })
      router.replace({ path: '/verify' })
    } catch (e: any) {
      showToast(e.message || '加载客户失败')
      loading.value = false
    }
  }
}

function syncAmountFromProject() {
  const p = selectedProject.value
  if (!p) {
    consumeAmount.value = 0
    return
  }
  const price = Number(p.member_price || p.price || 0)
  consumeAmount.value = Math.min(price, balance.value)
}

function openProjectPick() {
  if (!projects.value.length) {
    showToast('暂无金额/兼容模式项目')
    return
  }
  showProjectPick.value = true
}

function pickProject(p: any) {
  selectedProject.value = p
  showProjectPick.value = false
  syncAmountFromProject()
}

function verifyLink(token: string) {
  return `${location.origin.replace('5174', '5175')}/verify/${token}`
}

function extractToken(raw: string) {
  const text = (raw || '').trim()
  if (!text) return ''
  const m = text.match(/\/verify\/([a-fA-F0-9]{16,})/)
  if (m) return m[1]
  if (/^[a-fA-F0-9]{16,}$/.test(text)) return text
  try {
    const u = new URL(text)
    const parts = u.pathname.split('/').filter(Boolean)
    const i = parts.indexOf('verify')
    if (i >= 0 && parts[i + 1]) return parts[i + 1]
  } catch { /* ignore */ }
  return ''
}

async function drawQR() {
  await nextTick()
  if (!canvasRef.value || !draft.value?.token || !isNotify.value) return
  await QRCode.toCanvas(canvasRef.value, verifyLink(draft.value.token), {
    width: 200,
    margin: 2,
    color: { dark: '#2a2124', light: '#ffffff' },
  })
}

function stopPoll() {
  if (pollTimer) {
    window.clearInterval(pollTimer)
    pollTimer = undefined
  }
}

function startPoll() {
  stopPoll()
  if (!isNotify.value || !draft.value?.id) return
  pollTimer = window.setInterval(async () => {
    try {
      const v: any = await http.get(`/api/verifications/${draft.value.id}`)
      if (v?.status === '已完成') {
        stopPoll()
        showSuccessToast('核销完成')
        resetFlow()
      } else if (v?.status === '已取消') {
        stopPoll()
        showToast('核销已取消')
        resetFlow()
      }
    } catch { /* ignore */ }
  }, 2500)
}

function resetFlow() {
  stopPoll()
  step.value = 1
  selectedCustomer.value = null
  amountCard.value = null
  selectedProject.value = null
  draft.value = null
  showCustomerPick.value = false
  showProjectPick.value = false
  q.value = ''
  customers.value = []
  searched.value = false
  welfareList.value = []
  welfareId.value = 0
  consumeAmount.value = 0
}

const canGoNext = computed(() =>
  !!selectedProject.value && Number(consumeAmount.value) > 0 && Number(consumeAmount.value) <= balance.value,
)

function onAmountInput() {
  const n = Number(consumeAmount.value)
  if (!Number.isFinite(n) || n < 0) {
    consumeAmount.value = 0
    return
  }
  if (n > balance.value) consumeAmount.value = balance.value
}

async function createDraft() {
  if (!amountCard.value) return showToast('缺少储值卡')
  if (!selectedProject.value) return showToast('请先选择消费项目')
  if (!(consumeAmount.value > 0)) return showToast('请输入消费金额')
  if (consumeAmount.value > balance.value) return showToast('超过储值余额')
  loading.value = true
  try {
    draft.value = await http.post('/api/verifications/draft', {
      card_id: amountCard.value.id,
      project_id: selectedProject.value.id,
      amount: Number(consumeAmount.value),
      times: 0,
      welfare_issued_id: welfareId.value || 0,
    })
    step.value = 3
    if (isNotify.value) {
      showToast('核销码已生成')
      await drawQR()
      startPoll()
    }
  } catch (e: any) {
    showToast(e.message || '创建失败')
  } finally {
    loading.value = false
  }
}

async function confirmStaff() {
  if (!draft.value) return
  loading.value = true
  try {
    await http.post(`/api/verifications/${draft.value.id}/confirm`, { token: draft.value.token })
    showSuccessToast('核销完成')
    resetFlow()
  } catch (e: any) {
    showToast(e.message || '确认失败')
  } finally {
    loading.value = false
  }
}

async function confirmByToken(token: string) {
  if (!draft.value) return
  loading.value = true
  try {
    await http.post(`/api/verifications/${draft.value.id}/confirm`, { token })
    stopPoll()
    showSuccessToast('扫码核销完成')
    resetFlow()
  } catch (e: any) {
    showToast(e.message || '确认失败')
  } finally {
    loading.value = false
  }
}

async function pasteAndConfirm() {
  const raw = window.prompt('粘贴客户核销链接或 token')
  if (raw == null) return
  const token = extractToken(raw)
  if (!token) return showToast('无法识别核销码')
  if (draft.value?.token && token !== draft.value.token) {
    return showToast('核销码与当前单据不匹配')
  }
  await confirmByToken(token)
}

async function copyLink() {
  if (!draft.value?.token) return
  const link = verifyLink(draft.value.token)
  try {
    await navigator.clipboard.writeText(link)
    showSuccessToast('链接已复制')
  } catch {
    showToast(link)
  }
}

async function cancelDraft() {
  if (!draft.value?.id) {
    step.value = 2
    return
  }
  try {
    await http.post(`/api/verifications/${draft.value.id}/cancel`)
  } catch { /* ignore */ }
  stopPoll()
  draft.value = null
  // 取消后余额/福利需重新拉取
  if (selectedCustomer.value) {
    await loadAmountCard(selectedCustomer.value)
    await loadWelfare(selectedCustomer.value, amountCard.value)
    syncAmountFromProject()
  }
  step.value = 2
}

function initial(name: string) {
  return (name || '客').slice(0, 1)
}

watch(step, (v) => { if (v === 3 && isNotify.value) drawQR() })
watch(
  () => [route.query.draft_id, route.query.customer_id, route.query.project_id],
  ([draftId, customerId]) => {
    if (!draftId && !customerId) return
    bootstrapped.value = false
    bootstrapFromQuery()
  },
)
onMounted(async () => {
  await loadSettings()
  await bootstrapFromQuery()
})
onUnmounted(() => {
  stopPoll()
  if (searchTimer) window.clearTimeout(searchTimer)
})
</script>

<template>
  <div class="page-wrap verify-page">
    <div class="stage">
      <header class="top">
        <div class="top-row">
          <div>
            <div class="eyebrow">核销</div>
            <h1>{{ isNotify ? '扫码核销' : '自助核销' }}</h1>
          </div>
          <span class="mode" :class="verifyMode">
            <i class="dot" />
            {{ modeLabel }}
          </span>
        </div>

        <div class="progress">
          <div
            v-for="(s, i) in steps"
            :key="s.n"
            class="node"
            :class="{ on: step === s.n, done: step > s.n }"
          >
            <div class="num">
              <van-icon v-if="step > s.n" name="success" />
              <span v-else>{{ s.n }}</span>
            </div>
            <div class="lab">{{ s.label }}</div>
            <div v-if="i < steps.length - 1" class="rail" :class="{ filled: step > s.n }" />
          </div>
        </div>

        <div class="step-copy">
          <div class="st">{{ stepTitle }}</div>
          <div class="sd">金额模式：选客户与项目后，自动带出福利并抵扣储值</div>
        </div>
      </header>

      <!-- Step 1 -->
      <div v-if="step===1" class="panel center-panel">
        <div class="hero-pick">
          <div class="hero-icon">
            <van-icon name="gold-coin-o" size="30" />
          </div>
          <div class="hero-title">金额核销</div>
          <p class="hero-sub">选择有储值余额的客户，再选项目抵扣</p>
          <van-button round block type="primary" class="pick-btn" :loading="loading" @click="openCustomerPick">选择客户</van-button>
        </div>
      </div>

      <!-- Step 2 -->
      <div v-else-if="step===2" class="panel">
        <div class="sheet picked">
          <div class="avatar lg">{{ initial(selectedCustomer?.name) }}</div>
          <div class="picked-body">
            <div class="name">{{ selectedCustomer?.name }}</div>
            <div class="sub">{{ selectedCustomer?.phone }} · {{ amountCard?.card_no }}</div>
            <div class="remain-line">储值余额 <b>¥{{ balance.toFixed(2) }}</b></div>
          </div>
        </div>

        <van-cell
          title="消费项目"
          is-link
          :value="selectedProject?.name || '请选择项目'"
          :value-class="selectedProject ? '' : 'cell-placeholder'"
          @click="openProjectPick"
        />

        <div class="sheet times-card">
          <div>
            <div class="tlabel">消费金额</div>
            <div class="thint">{{ selectedProject ? '默认项目价，可改且不超过余额' : '请先选择项目' }}</div>
          </div>
          <van-field
            v-model.number="consumeAmount"
            type="number"
            input-align="right"
            :disabled="!selectedProject"
            style="padding:0;width:120px"
            @update:model-value="onAmountInput"
          />
        </div>

        <div class="sheet welfare-card">
          <div class="tlabel">福利抵扣</div>
          <div class="thint" style="margin-bottom:10px">
            {{ welfareList.length ? '已自动匹配可用福利，可改选' : '暂无可用福利' }}
          </div>
          <van-radio-group v-model="welfareId" direction="vertical">
            <van-radio :name="0">不使用福利</van-radio>
            <van-radio v-for="w in welfareList" :key="w.id" :name="w.id">
              {{ w.name }} · 剩¥{{ Number(w.remain_amount||0).toFixed(0) }}
              <span v-if="w.expire_at" class="w-exp">至 {{ String(w.expire_at).slice(0,10) }}</span>
            </van-radio>
          </van-radio-group>
          <div class="pay-line">
            消费 ¥{{ Number(consumeAmount||0).toFixed(2) }}
            − 福利 ¥{{ welfareOffset.toFixed(2) }}
            = 实扣 <b>¥{{ payAmount.toFixed(2) }}</b>
          </div>
        </div>

        <div class="actions">
          <van-button block round plain type="primary" @click="resetFlow">重选客户</van-button>
          <van-button
            block
            round
            type="primary"
            :loading="loading"
            :disabled="!canGoNext"
            @click="createDraft"
          >
            {{ isNotify ? '生成核销码' : '下一步' }}
          </van-button>
        </div>
      </div>

      <!-- Step 3 -->
      <div v-else class="panel">
        <template v-if="!isNotify">
          <div class="sheet summary">
            <div class="avatar lg">{{ initial(selectedCustomer?.name) }}</div>
            <div class="cname">{{ selectedCustomer?.name }}</div>
            <div class="pname">{{ selectedProject?.name || draft?.project_name }}</div>
            <div class="pay-sum">
              消费 ¥{{ Number(draft?.amount||0).toFixed(2) }}
              <template v-if="Number(draft?.welfare_offset||0)>0"> · 福利抵 ¥{{ Number(draft.welfare_offset).toFixed(2) }}</template>
            </div>
            <div class="pname">实扣 ¥{{ Number(draft?.pay_amount ?? 0).toFixed(2) }}</div>
            <div class="token">单号 {{ draft?.verify_no || '—' }}</div>
            <p class="hint">确认后从储值余额扣款（已预扣），福利同步核销。</p>
          </div>
          <div class="actions">
            <van-button block round plain type="primary" @click="cancelDraft">上一步</van-button>
            <van-button block round type="primary" :loading="loading" @click="confirmStaff">确认完成核销</van-button>
          </div>
        </template>
        <template v-else>
          <div class="sheet qrbox">
            <div class="qr-wrap"><canvas ref="canvasRef" class="qr" /></div>
            <div class="cname">{{ selectedCustomer?.name }}</div>
            <div class="pname">实扣 ¥{{ Number(draft?.pay_amount ?? 0).toFixed(2) }}</div>
            <div class="token">单号 {{ draft?.verify_no || '—' }}</div>
            <p class="hint">请客户亮码；你扫码或粘贴链接后完成。</p>
            <div class="link-row">
              <van-button size="small" round plain type="primary" @click="copyLink">复制给客户</van-button>
              <van-button size="small" round type="primary" :loading="loading" @click="pasteAndConfirm">扫码 / 粘贴确认</van-button>
            </div>
          </div>
          <div class="actions">
            <van-button block round plain type="primary" @click="cancelDraft">取消并返回</van-button>
          </div>
        </template>
      </div>
    </div>

    <!-- 选客户 -->
    <van-popup
      v-model:show="showCustomerPick"
      class="lp-center-popup"
      round
      position="center"
      teleport="body"
      closeable
      destroy-on-close
      :z-index="3000"
      :close-on-click-overlay="true"
    >
      <div class="lp-popup-body pick-body">
        <div class="pop-title">选择客户</div>
        <p class="pop-hint">搜索后仅展示客户；需有储值余额才能核销</p>
        <van-search
          v-model="q"
          placeholder="姓名 / 手机号"
          shape="round"
          background="transparent"
          show-action
          @update:model-value="onSearchInput"
          @search="searchCustomers()"
          @clear="customers=[]; searched=false"
        >
          <template #action>
            <span class="search-act" @click="searchCustomers()">搜索</span>
          </template>
        </van-search>
        <div v-if="searching" class="hint-box">正在搜索…</div>
        <div v-else-if="!searched" class="hint-box">输入关键词搜索客户</div>
        <div class="list">
          <button v-for="c in customers" :key="c.id" type="button" class="card" @click="pickCustomer(c)">
            <div class="avatar">{{ initial(c.name) }}</div>
            <div class="left">
              <div class="name">{{ c.name }}</div>
              <div class="sub">{{ c.phone }} · 余额¥{{ Number(c.balance||0).toFixed(0) }}</div>
            </div>
          </button>
          <van-empty v-if="searched && !customers.length" description="未找到客户" :image-size="72" />
        </div>
      </div>
    </van-popup>

    <!-- 选项目 -->
    <van-popup v-model:show="showProjectPick" position="bottom" round teleport="body" :z-index="3100">
      <div class="proj-sheet">
        <div class="pop-title" style="padding:14px 0 8px">选择金额项目</div>
        <button
          v-for="p in projects"
          :key="p.id"
          type="button"
          class="proj-row"
          @click="pickProject(p)"
        >
          <span>{{ p.name }}</span>
          <strong>¥{{ p.member_price || p.price }}</strong>
        </button>
        <van-empty v-if="!projects.length" description="暂无金额/兼容项目" />
      </div>
    </van-popup>
  </div>
</template>

<style scoped>
.verify-page {
  min-height: calc(100vh - var(--lp-tabbar-h, 56px) - 24px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 12px 14px 28px;
  animation: lp-fade-up 0.4s ease both;
}
.stage { width: min(100%, 420px); }

.top {
  padding: 10px 16px 14px;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.94);
  box-shadow: var(--lp-shadow-card);
  border: 1px solid rgba(255, 245, 248, 0.95);
}
.top-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 8px;
}
.eyebrow {
  font-size: 11px;
  font-weight: 650;
  letter-spacing: 0.14em;
  color: var(--lp-accent);
}
h1 {
  margin: 4px 0 0;
  font-size: 14px;
  font-weight: 760;
  color: var(--lp-text);
  line-height: 1.2;
}
.mode {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-top: 4px;
  padding: 6px 10px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 700;
  background: #fff0f4;
  color: var(--lp-accent-deep);
  border: 1px solid rgba(226, 75, 110, 0.12);
  max-width: 148px;
  line-height: 1.3;
}
.mode.notify { background: #eaf7f1; color: #2f9e74; border-color: rgba(47, 158, 116, 0.15); }
.mode .dot {
  width: 6px; height: 6px; border-radius: 50%; background: currentColor;
  box-shadow: 0 0 0 3px rgba(226, 75, 110, 0.15);
}

.progress {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  margin-top: 10px;
  position: relative;
}
.node {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  z-index: 1;
}
.num {
  width: 30px; height: 30px; border-radius: 50%;
  display: grid; place-items: center;
  font-family: var(--lp-display); font-size: 13px; font-weight: 700;
  color: #b5a8ac; background: #f3ecef; border: 2px solid #fff;
}
.node.on .num {
  color: #fff;
  background: linear-gradient(145deg, #ff8fa6, #e24b6e);
  box-shadow: 0 8px 18px rgba(226, 75, 110, 0.28);
}
.node.done .num {
  color: #fff;
  background: linear-gradient(145deg, #7ed6b0, #3cae84);
}
.lab { font-size: 12px; font-weight: 650; color: #b5a8ac; }
.node.on .lab { color: var(--lp-accent); font-weight: 740; }
.node.done .lab { color: #3cae84; }
.rail {
  position: absolute; top: 14px; left: calc(50% + 18px);
  width: calc(100% - 36px); height: 2px; border-radius: 99px; background: #efe0e5;
}
.rail.filled { background: linear-gradient(90deg, #7ed6b0, #e24b6e); }

.step-copy {
  margin-top: 8px; padding-top: 14px; border-top: 1px dashed #f0e2e6; text-align: center;
}
.st { font-size: 13px; font-weight: 740; }
.sd { margin-top: 4px; font-size: 12px; color: var(--lp-muted); line-height: 1.45; }

.panel { margin-top: 12px; animation: lp-fade-up 0.35s ease both; }
.center-panel { display: flex; justify-content: center; }
.hero-pick {
  width: 100%; padding: 14px 20px 22px; border-radius: 16px;
  background: rgba(255,255,255,.96); box-shadow: var(--lp-shadow-card);
  border: 1px solid rgba(255,245,248,.95); text-align: center;
}
.hero-icon {
  width: 58px; height: 58px; margin: 0 auto 12px; border-radius: 14px;
  display: grid; place-items: center; color: #fff;
  background: linear-gradient(145deg, #ff8fa6, #e24b6e);
  box-shadow: 0 10px 22px rgba(226, 75, 110, 0.25);
}
.hero-title { font-size: 20px; font-weight: 760; }
.hero-sub { margin: 8px 0 20px; font-size: 13px; color: var(--lp-muted); }
.pick-btn { font-weight: 700; }

.sheet {
  border-radius: 20px;
  background: rgba(255,255,255,.96);
  box-shadow: var(--lp-shadow-card);
  border: 1px solid rgba(255,245,248,.95);
}
.picked {
  display: flex; gap: 8px; align-items: center; padding: 10px; margin-bottom: 10px;
}
.avatar {
  width: 44px; height: 44px; border-radius: 14px; display: grid; place-items: center;
  color: #fff; font-weight: 740; font-size: 14px;
  background: linear-gradient(145deg, #ff8fa6, #e24b6e);
}
.avatar.lg { width: 52px; height: 52px; border-radius: 16px; font-size: 20px; margin: 0; }
.name { font-weight: 720; font-size: 13px; }
.sub { margin-top: 4px; font-size: 12px; color: var(--lp-muted); }
.remain-line { margin-top: 8px; font-size: 13px; color: var(--lp-muted); }
.remain-line b { color: var(--lp-accent); font-family: var(--lp-display); font-size: 14px; }

.times-card {
  padding: 10px; display: flex; justify-content: space-between; align-items: center; margin: 10px 0;
}
.tlabel { font-weight: 740; font-size: 14px; }
.thint { margin-top: 4px; font-size: 11px; color: var(--lp-muted); }
.welfare-card { padding: 10px; margin-bottom: 12px; text-align: left; }
.w-exp { margin-left: 6px; color: var(--lp-muted); font-size: 11px; }
.pay-line {
  margin-top: 12px; padding-top: 10px; border-top: 1px dashed #f0e2e6;
  font-size: 13px; color: var(--lp-muted);
}
.pay-line b { color: var(--lp-accent); font-family: var(--lp-display); font-size: 14px; }
.pay-sum { margin-top: 8px; font-size: 13px; color: var(--lp-muted); }

:deep(.cell-placeholder) { color: #c0c4cc !important; }
.actions { display: grid; gap: 6px; margin-top: 4px; }
.summary, .qrbox { padding: 22px 16px 20px; text-align: center; margin-bottom: 12px; }
.qr-wrap {
  width: 212px; height: 212px; margin: 0 auto; padding: 6px; border-radius: 14px;
  background: #fff; box-shadow: 0 10px 28px rgba(80,40,55,.08); border: 1px solid #f3ecef;
}
.qr { width: 200px; height: 200px; border-radius: 12px; }
.cname { margin-top: 8px; font-size: 13px; font-weight: 740; }
.pname { margin-top: 6px; color: var(--lp-accent); font-size: 14px; font-weight: 650; }
.token { margin-top: 8px; font-size: 12px; color: var(--lp-muted); font-family: var(--lp-display); }
.hint { font-size: 12px; color: var(--lp-muted); line-height: 1.55; margin: 12px 12px 14px; }
.link-row { display: flex; justify-content: center; gap: 8px; flex-wrap: wrap; }

.pick-body { padding-top: 24px !important; }
.pop-title { font-size: 14px; font-weight: 740; text-align: center; }
.pop-hint { margin: 6px 0 10px; text-align: center; font-size: 12px; color: var(--lp-muted); }
.search-act { color: var(--lp-accent); font-weight: 700; padding: 0 4px; }
.hint-box {
  margin: 4px 0 10px; padding: 12px 14px; border-radius: 14px;
  background: #fff7f9; border: 1px dashed #f0e2e6; color: var(--lp-muted);
  font-size: 12px; text-align: center;
}
.list { display: grid; gap: 6px; max-height: 40vh; overflow-y: auto; }
.card {
  display: grid; grid-template-columns: 44px 1fr; gap: 8px; align-items: center;
  padding: 10px; text-align: left; border: 1px solid rgba(255,245,248,.95);
  background: #fff; border-radius: 14px; box-shadow: var(--lp-shadow-card);
}
.proj-sheet { padding: 8px 14px 24px; max-height: 52vh; overflow-y: auto; }
.proj-row {
  width: 100%; display: flex; justify-content: space-between; align-items: center;
  padding: 10px 4px; border: 0; border-bottom: 1px solid #f3ecef; background: transparent;
  font-size: 13px; text-align: left;
}
.proj-row strong { color: var(--lp-accent); font-family: var(--lp-display); }
</style>
