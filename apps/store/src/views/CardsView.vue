<script setup lang="ts">
import { computed, nextTick, onMounted, reactive, ref } from 'vue'
import { showDialog, showSuccessToast, showToast } from 'vant'
import http from '@/api/http'
import { useSession } from '@/stores/session'

const session = useSession()
const rows = ref<any[]>([])
const customers = ref<any[]>([])
const projects = ref<any[]>([])
const show = ref(false)
const showCustomerPick = ref(false)
const showProjectPick = ref(false)
const showPayPick = ref(false)
const showLevelPick = ref(false)
const saving = ref(false)
const form = reactive({
  customer_mode: 'old' as 'old' | 'new',
  customer_id: undefined as number | undefined,
  new_name: '',
  new_phone: '',
  new_level: '银卡会员',
  project_id: undefined as number | undefined,
  pay_method: '微信',
  card_kind: 'amount' as 'times' | 'amount',
  times: 1,
  amount: 0,
})

const selectedProject = computed(() => projects.value.find((p) => p.id === form.project_id))
const selectedCustomer = computed(() => customers.value.find((c) => c.id === form.customer_id))
const projectMode = (p: any) => p?.card_mode || 'amount'
const allowTimes = computed(() => {
  const p = selectedProject.value
  if (!p) return true
  const m = projectMode(p)
  return m === 'times' || m === 'both'
})
const allowAmount = computed(() => {
  const p = selectedProject.value
  if (!p) return true
  const m = projectMode(p)
  return m === 'amount' || m === 'both'
})
const customerColumns = computed(() =>
  customers.value.map((c) => ({ text: `${c.name}（${c.phone}）`, value: c.id })),
)
const projectColumns = computed(() => {
  const modeText = (m: string) => (m === 'times' ? '次卡' : m === 'amount' ? '金额' : '兼容')
  const list = projects.value
    .filter((p) => {
      const m = projectMode(p)
      if (form.card_kind === 'times') return m === 'times' || m === 'both'
      return m === 'amount' || m === 'both'
    })
    .map((p) => ({
      text: `${p.name} · ¥${p.price}${p.card_mode !== 'amount' ? `×${p.total_times}次` : ''}（${modeText(projectMode(p))}）`,
      value: p.id,
    }))
  return form.card_kind === 'amount' ? [{ text: '不关联项目', value: 0 }, ...list] : list
})
const payColumns = ['微信', '刷卡', '现金'].map((x) => ({ text: x, value: x }))
const levelColumns = ['银卡会员', '金卡会员', '钻石会员'].map((x) => ({ text: x, value: x }))

function syncFromProject() {
  const p = selectedProject.value
  if (!p) return
  const m = projectMode(p)
  if (m === 'times') form.card_kind = 'times'
  if (m === 'amount') form.card_kind = 'amount'
  if (form.card_kind === 'times') {
    form.times = p.total_times || 1
    form.amount = Number(((p.price || 0) * form.times).toFixed(2))
  }
}

function setCardKind(kind: 'times' | 'amount') {
  if (kind === 'times' && !allowTimes.value) return tip('当前项目仅支持金额模式')
  if (kind === 'amount' && !allowAmount.value) return tip('当前项目仅支持次卡模式')
  form.card_kind = kind
  if (kind === 'times') syncFromProject()
}

async function load() {
  try {
    const data: any = await http.get('/api/cards', { params: { pageSize: 100 } })
    rows.value = data.list || []
  } catch (e: any) {
    showToast(e.message || '加载失败')
  }
}

async function refreshCustomers() {
  const c: any = await http.get('/api/customers', { params: { pageSize: 100 } })
  customers.value = c.list || []
}

async function openCreate() {
  try {
    const p: any = await http.get('/api/projects', { params: { pageSize: 100, status: '上架' } })
    await refreshCustomers()
    projects.value = p.list || []
    Object.assign(form, {
      customer_mode: customers.value.length ? 'old' : 'new',
      customer_id: customers.value[0]?.id,
      new_name: '',
      new_phone: '',
      new_level: '银卡会员',
      project_id: undefined as number | undefined,
      card_kind: 'amount' as 'times' | 'amount',
      pay_method: '微信',
    })
    const firstAmount = projects.value.find((p) => projectMode(p) === 'amount' || projectMode(p) === 'both')
    form.project_id = firstAmount?.id || projects.value[0]?.id
    if (form.project_id) {
      const m = projectMode(projects.value.find((p) => p.id === form.project_id))
      if (m === 'times') form.card_kind = 'times'
      else form.card_kind = 'amount'
    }
    syncFromProject()
    show.value = true
  } catch (e: any) {
    showToast(e.message || '打开失败')
  }
}

function tip(msg: string, type: 'success' | 'text' = 'text') {
  // 弹窗 z-index=3000，Toast 必须更高，否则“成功也看不见”
  const opts = { message: msg, zIndex: 4000, duration: 2000 }
  if (type === 'success') showSuccessToast(opts)
  else showToast(opts)
}

async function ensureCustomerId() {
  if (form.customer_mode === 'old') {
    if (!form.customer_id) throw new Error('请选择老客户')
    return form.customer_id
  }
  if (!form.new_name?.trim() || !form.new_phone?.trim()) throw new Error('请填写新客户姓名和手机号')
  if (!/^1\d{10}$/.test(form.new_phone.trim())) throw new Error('请填写正确的11位手机号')
  if (!session.user?.store_id) throw new Error('账号未绑定门店')
  const created: any = await http.post('/api/customers', {
    name: form.new_name.trim(),
    phone: form.new_phone.trim(),
    mobile: form.new_phone.trim(),
    level_name: form.new_level,
    store_id: session.user.store_id,
    source: '到店',
  })
  const id = Number(created?.id || 0)
  if (!id) throw new Error('新建客户失败，未返回客户ID')
  try {
    await refreshCustomers()
  } catch {
    // 列表刷新失败不影响开卡
  }
  form.customer_id = id
  form.customer_mode = 'old'
  return id
}

async function save() {
  if (form.card_kind === 'times' && !form.project_id) return tip('请选择项目')
  if (form.card_kind === 'amount' && !(form.amount > 0)) return tip('请输入充值金额')
  if (selectedProject.value) {
    const m = projectMode(selectedProject.value)
    if (m === 'times' && form.card_kind !== 'times') return tip('该项目仅支持次卡模式')
    if (m === 'amount' && form.card_kind !== 'amount') return tip('该项目仅支持金额模式')
  }
  if (saving.value) return
  saving.value = true
  try {
    const customerId = await ensureCustomerId()
    await http.post('/api/cards', {
      customer_id: customerId,
      project_id: form.card_kind === 'times' ? form.project_id : (form.project_id || 0),
      pay_method: form.pay_method,
      card_kind: form.card_kind,
      times: form.card_kind === 'times' ? form.times : 0,
      amount: form.amount,
      store_id: session.user?.store_id,
    })
    // 先关弹窗，再提示，避免 Toast 被遮罩挡住
    show.value = false
    showCustomerPick.value = false
    showProjectPick.value = false
    showPayPick.value = false
    showLevelPick.value = false
    tip('开卡成功', 'success')
    await load()
  } catch (e: any) {
    tip(e?.message || e?.response?.data?.error || '开卡失败')
  } finally {
    saving.value = false
  }
}

function cardTitle(c: any) {
  if (c.card_kind === 'amount') return `${c.customer_name} · 储值卡`
  return `${c.customer_name} · ${c.project_name || '次卡'}`
}

function onPickCustomer({ selectedOptions }: any) {
  form.customer_id = selectedOptions[0]?.value
  showCustomerPick.value = false
}
function onPickProject({ selectedOptions }: any) {
  form.project_id = selectedOptions[0]?.value || undefined
  showProjectPick.value = false
  syncFromProject()
}
function onPickPay({ selectedOptions }: any) {
  form.pay_method = selectedOptions[0]?.value
  showPayPick.value = false
}
function onPickLevel({ selectedOptions }: any) {
  form.new_level = selectedOptions[0]?.value
  showLevelPick.value = false
}

onMounted(load)
</script>

<template>
  <div class="page-wrap">
    <van-nav-bar title="开卡充值">
      <template #right>
        <span class="nav-link" @click="openCreate">开卡</span>
      </template>
    </van-nav-bar>

    <div class="toolbar">
      <van-button round block type="primary" @click="openCreate">新增开卡 / 充值</van-button>
    </div>

    <div class="list">
      <div v-for="c in rows" :key="c.id" class="row">
        <div class="name">{{ cardTitle(c) }}</div>
        <div class="sub">{{ c.card_no }} · {{ c.card_kind === 'amount' ? '储值' : '次卡' }} · {{ c.pay_method }} · {{ String(c.created_at||'').replace('T',' ').slice(0,16) }}</div>
        <div class="foot">
          <span v-if="c.card_kind==='amount'">余额 ¥{{ c.amount }}</span>
          <span v-else>余额/次数 {{ c.remain }}/{{ c.total }}</span>
          <strong>¥{{ c.amount }}</strong>
        </div>
      </div>
      <van-empty v-if="!rows.length" description="暂无开卡记录，点击上方开卡" />
    </div>

    <van-popup
      v-model:show="show"
      class="lp-center-popup"
      round
      position="center"
      teleport="body"
      closeable
      destroy-on-close
      :z-index="3000"
      :lock-scroll="true"
      :close-on-click-overlay="true"
      @closed="show = false"
    >
      <div class="lp-popup-body">
        <div class="pop-title">客户开卡</div>
        <p class="pop-hint">可选老客户，或直接新增客户后开卡</p>

        <div class="kind">
          <button type="button" :class="{ on: form.card_kind==='times' }" :disabled="!allowTimes" @click="setCardKind('times')">项目次数</button>
          <button type="button" :class="{ on: form.card_kind==='amount' }" :disabled="!allowAmount" @click="setCardKind('amount')">直接充值</button>
        </div>
        <p v-if="selectedProject" class="mode-tip">
          项目模式：{{ selectedProject.card_mode==='times'?'次卡':selectedProject.card_mode==='amount'?'金额':'兼容' }}
        </p>

        <div class="kind">
          <button type="button" :class="{ on: form.customer_mode==='old' }" @click="form.customer_mode='old'">老客户</button>
          <button type="button" :class="{ on: form.customer_mode==='new' }" @click="form.customer_mode='new'">新客户</button>
        </div>

        <template v-if="form.customer_mode==='old'">
          <van-cell title="选择客户" is-link :value="selectedCustomer ? selectedCustomer.name : '请选择'" @click="showCustomerPick=true" />
        </template>
        <template v-else>
          <van-field v-model="form.new_name" label="姓名" placeholder="新客户姓名" />
          <van-field v-model="form.new_phone" type="tel" maxlength="11" label="手机号" placeholder="11位手机号" />
          <van-cell title="会员等级" is-link :value="form.new_level" @click="showLevelPick=true" />
        </template>

        <van-cell
          :title="form.card_kind==='times' ? '项目' : '关联项目'"
          is-link
          :value="selectedProject?.name || (form.card_kind==='amount' ? '不关联' : '请选择')"
          @click="showProjectPick=true"
        />
        <van-field
          v-if="form.card_kind==='times'"
          v-model.number="form.times"
          type="digit"
          label="次数"
          placeholder="开卡次数"
          @update:model-value="form.amount = Number(((selectedProject?.price||0)*Number(form.times||0)).toFixed(2))"
        />
        <van-field
          v-model.number="form.amount"
          type="number"
          :label="form.card_kind==='amount' ? '充值金额' : '应收金额'"
          placeholder="金额"
        />
        <van-cell title="支付方式" is-link :value="form.pay_method" @click="showPayPick=true" />

        <van-button block round type="primary" class="submit" :loading="saving" @click="save">确认开卡</van-button>
      </div>
    </van-popup>

    <van-popup v-model:show="showCustomerPick" position="bottom" round teleport="body" :z-index="3100">
      <van-picker title="选择老客户" :columns="customerColumns" @confirm="onPickCustomer" @cancel="showCustomerPick=false" />
    </van-popup>
    <van-popup v-model:show="showProjectPick" position="bottom" round teleport="body" :z-index="3100">
      <van-picker title="选择项目" :columns="projectColumns" @confirm="onPickProject" @cancel="showProjectPick=false" />
    </van-popup>
    <van-popup v-model:show="showPayPick" position="bottom" round teleport="body" :z-index="3100">
      <van-picker title="支付方式" :columns="payColumns" @confirm="onPickPay" @cancel="showPayPick=false" />
    </van-popup>
    <van-popup v-model:show="showLevelPick" position="bottom" round teleport="body" :z-index="3100">
      <van-picker title="会员等级" :columns="levelColumns" @confirm="onPickLevel" @cancel="showLevelPick=false" />
    </van-popup>
  </div>
</template>

<style scoped>
.toolbar { padding: 10px 14px 4px; }
.list { padding: 12px 14px 24px; display:grid; gap:10px; }
.row {
  padding: 15px 14px;
  border-radius: 20px;
  background: rgba(255,255,255,.95);
  box-shadow: var(--lp-shadow-card);
  border: 1px solid rgba(255,245,248,.95);
}
.name { font-weight: 720; letter-spacing: .01em; }
.sub { margin-top: 6px; font-size: 12px; color: var(--lp-muted); line-height: 1.45; }
.foot {
  margin-top: 12px; display:flex; justify-content:space-between; align-items:center;
  color: var(--lp-accent); padding-top: 10px; border-top: 1px dashed #f0e2e6;
  font-size: 13px; font-weight: 650;
}
.foot strong { font-family: var(--lp-display); font-size: 14px; }
.pop-title { font-size: 14px; font-weight: 740; text-align: center; }
.pop-hint { margin: 6px 0 14px; text-align: center; font-size: 12px; color: var(--lp-muted); }
.kind { display:flex; gap:8px; margin-bottom: 10px; }
.kind button {
  flex:1; border:1px solid #f0e2e6; border-radius:12px; background:#fff; padding:10px 0; font-size:13px;
}
.kind button.on {
  border-color: var(--lp-accent); background:#fff5f7; color: var(--lp-accent); font-weight:700;
  box-shadow: 0 0 0 1px var(--lp-accent) inset;
}
.kind button:disabled { opacity: .45; }
.mode-tip { margin: 0 0 10px; text-align: center; font-size: 12px; color: var(--lp-muted); }
.submit { margin-top: 8px; }
</style>
