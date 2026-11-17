<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { DatePicker, showConfirmDialog, showSuccessToast, showToast, TimePicker } from 'vant'
import http from '@/api/http'
import { useSession } from '@/stores/session'

const router = useRouter()

const VanDatePicker = DatePicker
const VanTimePicker = TimePicker

const session = useSession()
const rows = ref<any[]>([])
const tab = ref<'active' | 'history'>('active')
const show = ref(false)
const showDate = ref(false)
const showTime = ref(false)
const customers = ref<any[]>([])
const projects = ref<any[]>([])
const staffList = ref<any[]>([])
const saving = ref(false)
const loadingStaff = ref(false)
const editingId = ref<number | null>(null)

const now = new Date()
const form = reactive({
  customer_id: undefined as number | undefined,
  project_id: undefined as number | undefined,
  staff_id: undefined as number | undefined,
  date: '',
  time: '',
  note: '',
})

const dateValue = ref([now.getFullYear(), now.getMonth() + 1, now.getDate()])
const timeValue = ref([pad(now.getHours()), pad(now.getMinutes())])

function pad(n: number) {
  return String(n).padStart(2, '0')
}
function fmtDate(arr: (string | number)[]) {
  return `${arr[0]}-${pad(Number(arr[1]))}-${pad(Number(arr[2]))}`
}
function fmtTime(arr: (string | number)[]) {
  return `${pad(Number(arr[0]))}:${pad(Number(arr[1]))}`
}
function syncNowDefaults() {
  const n = new Date()
  dateValue.value = [n.getFullYear(), n.getMonth() + 1, n.getDate()]
  timeValue.value = [pad(n.getHours()), pad(n.getMinutes())]
  form.date = fmtDate(dateValue.value)
  form.time = fmtTime(timeValue.value)
}

const selectedStaff = computed(() => staffList.value.find((s) => s.id === form.staff_id))
const isEdit = computed(() => !!editingId.value)

/** 进行中：待履约 / 已改约（兼容旧待确认） */
const activeRows = computed(() =>
  rows.value.filter((a) => ['待履约', '已改约', '待确认', '已确认'].includes(a.status)),
)
/** 历史：已履约 / 已取消 / 已失效 */
const historyRows = computed(() =>
  rows.value.filter((a) =>
    ['已履约', '已取消', '已失效', '失效', '已完成'].includes(a.status),
  ),
)
const displayRows = computed(() => (tab.value === 'active' ? activeRows.value : historyRows.value))

function statusClass(st: string) {
  if (st === '已履约' || st === '已完成') return 'ok'
  if (st === '待履约' || st === '待确认' || st === '已确认') return 'wait'
  if (st === '已改约') return 'resched'
  if (st === '已取消') return 'cancel'
  if (st === '已失效' || st === '失效') return 'expire'
  return ''
}

function statusLabel(st: string) {
  if (st === '待确认' || st === '已确认') return '待履约'
  if (st === '失效') return '已失效'
  if (st === '已完成') return '已履约'
  return st
}

async function load() {
  const data: any = await http.get('/api/appointments', { params: { pageSize: 100 } })
  rows.value = data.list || []
}

async function loadStaff() {
  if (!form.date || !form.time) return
  loadingStaff.value = true
  try {
    const data: any = await http.get('/api/staff/availability', {
      params: {
        date: form.date,
        time: form.time,
        project_id: form.project_id || 0,
        store_id: session.user?.store_id,
        exclude_id: editingId.value || 0,
      },
    })
    staffList.value = data.list || []
    const cur = staffList.value.find((s) => s.id === form.staff_id)
    if (!cur || cur.busy) {
      const free = staffList.value.find((s) => !s.busy)
      form.staff_id = free?.id
    }
  } finally {
    loadingStaff.value = false
  }
}

async function ensureOptions() {
  const [c, p]: any[] = await Promise.all([
    http.get('/api/customers', { params: { pageSize: 200 } }),
    http.get('/api/projects', { params: { pageSize: 100, status: '上架' } }),
  ])
  customers.value = c.list || []
  projects.value = p.list || []
}

async function openCreate() {
  editingId.value = null
  await ensureOptions()
  syncNowDefaults()
  form.customer_id = customers.value[0]?.id
  form.project_id = projects.value[0]?.id
  form.staff_id = session.user?.id
  form.note = ''
  show.value = true
  await loadStaff()
}

async function openEdit(row: any) {
  if (['已履约', '已取消', '已失效', '失效', '已完成'].includes(row.status)) {
    return showToast('已结束的预约不能改，请新建')
  }
  editingId.value = row.id
  await ensureOptions()
  form.customer_id = row.customer_id
  form.project_id = row.project_id
  form.staff_id = row.staff_id
  form.date = row.date
  form.time = String(row.time || '').slice(0, 5)
  form.note = row.note || ''
  const [y, m, d] = String(form.date).split('-').map(Number)
  const [hh, mm] = String(form.time).split(':').map(Number)
  dateValue.value = [y, m, d]
  timeValue.value = [pad(hh || 0), pad(mm || 0)]
  show.value = true
  await loadStaff()
}

watch(() => [form.date, form.time, form.project_id], () => {
  if (show.value) loadStaff()
})

function onConfirmDate({ selectedValues }: any) {
  dateValue.value = selectedValues
  form.date = fmtDate(selectedValues)
  showDate.value = false
}
function onConfirmTime({ selectedValues }: any) {
  timeValue.value = selectedValues
  form.time = fmtTime(selectedValues)
  showTime.value = false
}

async function save() {
  if (!form.customer_id || !form.project_id || !form.date || !form.time) {
    return showToast('请完善客户、项目、日期和时间')
  }
  if (!form.staff_id) return showToast('请选择可预约的美容师')
  const st = staffList.value.find((s) => s.id === form.staff_id)
  if (st?.busy) return showToast(st.conflict_hint || '该时段与已有预约冲突，请更换')
  saving.value = true
  try {
    const payload = {
      customer_id: form.customer_id,
      project_id: form.project_id,
      staff_id: form.staff_id,
      date: form.date,
      time: form.time,
      note: form.note,
      store_id: session.user?.store_id,
    }
    if (editingId.value) {
      await http.put(`/api/appointments/${editingId.value}`, { ...payload, reschedule: true })
      showSuccessToast('已改约（仍待履约）')
    } else {
      await http.post('/api/appointments', payload)
      showSuccessToast('已新建，状态为待履约')
    }
    show.value = false
    editingId.value = null
    tab.value = 'active'
    await load()
  } catch (e: any) {
    showToast(e.message || '保存失败')
  } finally {
    saving.value = false
  }
}

function staffLabel(s: any) {
  if (s.busy) return s.conflict_hint || '该时段冲突'
  if (s.busy_slots?.length) return `可约 · 已排${s.busy_slots.length}档`
  return s.role_name || '空闲可约'
}

async function fulfillAppt(row: any) {
  try {
    await showConfirmDialog({
      title: '来店确认',
      message: `确认客户已到店？\n${row.customer_name || '客户'} · ${row.date} ${String(row.time || '').slice(0, 5)}\n确认后将进入该客户核销。`,
    })
    await http.put(`/api/appointments/${row.id}`, { status: '已履约' })
    showSuccessToast('已来店，去核销')
    router.push({
      path: '/verify',
      query: {
        customer_id: String(row.customer_id || ''),
        project_id: String(row.project_id || ''),
        appointment_id: String(row.id),
      },
    })
  } catch (e: any) {
    if (e !== 'cancel' && e?.message) showToast(e.message)
  }
}

async function cancelAppt(row: any) {
  try {
    await showConfirmDialog({
      title: '取消预约',
      message: `取消后将释放美容师时段。\n${row.customer_name || '客户'} · ${row.date} ${String(row.time || '').slice(0, 5)}`,
      confirmButtonText: '确认取消',
      confirmButtonColor: '#d45454',
    })
    await http.put(`/api/appointments/${row.id}`, { status: '已取消' })
    showSuccessToast('已取消，时段已释放')
    load()
  } catch (e: any) {
    if (e !== 'cancel' && e?.message) showToast(e.message)
  }
}

function canOperate(st: string) {
  return ['待履约', '已改约', '待确认', '已确认'].includes(st)
}

onMounted(load)
</script>

<template>
  <div class="page-wrap">
    <van-nav-bar title="预约" right-text="新建" @click-right="openCreate" />

    <div class="tabs">
      <button type="button" :class="{ on: tab==='active' }" @click="tab='active'">
        进行中 <em>{{ activeRows.length }}</em>
      </button>
      <button type="button" :class="{ on: tab==='history' }" @click="tab='history'">
        历史 <em>{{ historyRows.length }}</em>
      </button>
    </div>

    <div class="list">
      <div v-for="a in displayRows" :key="a.id" class="row">
        <div class="appt-main">
          <div class="time">{{ String(a.time||'').slice(0,5) }}</div>
          <div class="divider" />
          <div class="info">
            <div class="name">{{ a.customer_name || ('客户#'+a.customer_id) }}</div>
            <div class="sub">{{ a.date }} · {{ a.project_name }} · {{ a.staff_name || '—' }}</div>
          </div>
          <span class="badge" :class="statusClass(a.status)">{{ statusLabel(a.status) }}</span>
        </div>
        <div v-if="canOperate(a.status)" class="ops">
          <van-button size="mini" round type="primary" @click="fulfillAppt(a)">来店确认</van-button>
          <van-button size="mini" round plain type="primary" @click="openEdit(a)">改约</van-button>
          <van-button size="mini" round plain @click="cancelAppt(a)">取消</van-button>
        </div>
      </div>
      <van-empty
        v-if="!displayRows.length"
        :description="tab==='active' ? '暂无进行中预约，点右上角新建' : '暂无历史记录'"
      />
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
    >
      <div class="lp-popup-body sheet">
        <div class="pop-title">{{ isEdit ? '改约' : '新建预约' }}</div>
        <p class="pop-hint">{{ isEdit ? '保存后为「已改约」，仍待客户到店履约' : '新建后为「待履约」，客户到店后再确认履约' }}</p>
        <van-field label="客户">
          <template #input>
            <select v-model="form.customer_id" class="sel" :disabled="isEdit">
              <option v-for="c in customers" :key="c.id" :value="c.id">{{ c.name }}（{{ c.phone }}）</option>
            </select>
          </template>
        </van-field>
        <van-field label="项目">
          <template #input>
            <select v-model="form.project_id" class="sel">
              <option v-for="p in projects" :key="p.id" :value="p.id">{{ p.name }}</option>
            </select>
          </template>
        </van-field>
        <van-field label="日期" is-link readonly :model-value="form.date" placeholder="选择日期" @click="showDate=true" />
        <van-field label="时间" is-link readonly :model-value="form.time" placeholder="选择时间" @click="showTime=true" />

        <div class="staff-block">
          <div class="staff-title">美容师 <span v-if="loadingStaff">校准中…</span></div>
          <div class="staff-grid">
            <button
              v-for="s in staffList"
              :key="s.id"
              type="button"
              class="staff"
              :class="{ on: form.staff_id===s.id, busy: s.busy }"
              :disabled="s.busy"
              @click="form.staff_id=s.id"
            >
              <div class="sn">{{ s.name }}</div>
              <div class="ss">{{ staffLabel(s) }}</div>
            </button>
          </div>
          <p v-if="selectedStaff?.busy" class="warn">{{ selectedStaff.conflict_hint }}</p>
          <p v-else class="hint">仅展示本店美容师；冲突时段不可选。</p>
        </div>

        <van-field v-model="form.note" label="备注" placeholder="可选" />
        <van-button block round type="primary" class="submit" :loading="saving" @click="save">
          {{ isEdit ? '保存改约' : '确认排期' }}
        </van-button>
      </div>
    </van-popup>

    <van-popup
      v-model:show="showDate"
      position="bottom"
      round
      teleport="body"
      :z-index="3200"
      :lock-scroll="true"
    >
      <VanDatePicker
        title="选择日期"
        :model-value="dateValue"
        @confirm="onConfirmDate"
        @cancel="showDate=false"
      />
    </van-popup>
    <van-popup
      v-model:show="showTime"
      position="bottom"
      round
      teleport="body"
      :z-index="3200"
      :lock-scroll="true"
    >
      <VanTimePicker
        title="选择时间"
        :model-value="timeValue"
        :columns-type="['hour', 'minute']"
        @confirm="onConfirmTime"
        @cancel="showTime=false"
      />
    </van-popup>
  </div>
</template>

<style scoped>
.tabs {
  margin: 4px 12px 0;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 6px;
  padding: 4px;
  border-radius: 16px;
  background: rgba(255,255,255,.78);
}
.tabs button {
  border: 0;
  border-radius: 12px;
  padding: 10px 0;
  background: transparent;
  color: var(--lp-muted);
  font-size: 13px;
  font-weight: 650;
}
.tabs button.on {
  background: #fff;
  color: var(--lp-accent);
  box-shadow: 0 4px 12px rgba(196,88,112,.12);
  font-weight: 720;
}
.tabs em {
  font-style: normal;
  font-family: var(--lp-display);
  margin-left: 2px;
}
.tip {
  margin: 6px 12px 0;
  font-size: 12px;
  color: var(--lp-muted);
  line-height: 1.5;
}
.list { padding: 8px 12px 12px; display: grid; gap: 6px; }
.row {
  padding: 10px;
  border-radius: 20px;
  background: rgba(255,255,255,.95);
  box-shadow: var(--lp-shadow-card);
  border: 1px solid rgba(255,245,248,.95);
}
.appt-main {
  display: grid;
  grid-template-columns: 52px 1px 1fr auto;
  gap: 8px;
  align-items: center;
}
.time { font-family: var(--lp-display); font-weight: 700; color: var(--lp-accent); font-size: 13px; }
.divider { width: 1px; height: 34px; background: linear-gradient(180deg, transparent, #f0e2e6, transparent); }
.name { font-weight: 720; font-size: 14px; }
.sub { margin-top: 4px; font-size: 12px; color: var(--lp-muted); }
.badge {
  font-size: 11px; font-weight: 650;
  border-radius: 999px; padding: 4px 10px; white-space: nowrap;
  color: var(--lp-accent-deep); background: #ffe8ef;
}
.badge.ok { color: #2f9e74; background: #e8f7f0; }
.badge.wait { color: #c9822d; background: #fff3e6; }
.badge.resched { color: #3b6fd9; background: #eaf0ff; }
.badge.cancel { color: #a3999d; background: #f5f0f2; text-decoration: line-through; }
.badge.expire { color: #8a7f84; background: #efe9eb; }
.ops { margin-top: 10px; display: flex; gap: 6px; justify-content: flex-end; flex-wrap: wrap; }
.sheet { padding: 10px 14px 20px; }
.pop-title { font-size: 14px; font-weight: 740; margin: 2px 4px 4px; text-align: center; }
.pop-hint { margin: 0 4px 12px; text-align: center; font-size: 12px; color: var(--lp-muted); line-height: 1.45; }
.staff-block { margin: 10px 4px 4px; }
.staff-title { font-size: 13px; font-weight: 650; color: #444; margin-bottom: 8px; }
.staff-title span { margin-left: 8px; font-weight: 500; color: var(--lp-muted); font-size: 12px; }
.staff-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; }
.staff {
  border: 1px solid #f0e2e6; border-radius: 14px; background: #fff; padding: 10px 12px; text-align: left;
}
.staff.on { border-color: var(--lp-accent); background: #fff5f7; box-shadow: 0 0 0 1px var(--lp-accent) inset; }
.staff.busy { opacity: .45; background: #f6f6f6; }
.sn { font-weight: 650; font-size: 14px; }
.ss { margin-top: 4px; font-size: 11px; color: var(--lp-muted); }
.hint { margin: 8px 0 0; font-size: 12px; color: var(--lp-muted); line-height: 1.45; }
.warn { margin-top: 8px; font-size: 12px; color: #c45656; }
.submit { margin-top: 14px; }
.sel { width: 100%; border: 0; background: transparent; font-size: 14px; outline: none; }
</style>
