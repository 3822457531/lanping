<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import http from '@/api/http'
import PageHeader from '@/components/PageHeader.vue'
import StatusTag from '@/components/StatusTag.vue'

const rows = ref<any[]>([])
const customers = ref<any[]>([])
const projects = ref<any[]>([])
const stores = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const loading = ref(false)
const dialog = ref(false)
const form = reactive({
  customer_mode: 'old' as 'old' | 'new',
  customer_id: undefined as number | undefined,
  new_name: '',
  new_phone: '',
  new_level: '银卡会员',
  project_id: undefined as number | undefined,
  store_id: undefined as number | undefined,
  pay_method: '微信',
  card_kind: 'amount' as 'times' | 'amount',
  times: 1,
  amount: 0,
})

const selectedProject = computed(() => projects.value.find((p) => p.id === form.project_id))
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
const filteredProjects = computed(() =>
  projects.value.filter((p) => {
    const m = projectMode(p)
    if (form.card_kind === 'times') return m === 'times' || m === 'both'
    return m === 'amount' || m === 'both'
  }),
)

function syncFromProject() {
  const p = selectedProject.value
  if (!p) return
  const m = projectMode(p)
  if (m === 'times' && form.card_kind !== 'times') form.card_kind = 'times'
  if (m === 'amount' && form.card_kind !== 'amount') form.card_kind = 'amount'
  if (form.card_kind === 'times') {
    form.times = p.total_times || 1
    form.amount = Number(((p.price || 0) * form.times).toFixed(2))
  }
}

function onKindChange() {
  if (form.project_id && !filteredProjects.value.some((p) => p.id === form.project_id)) {
    form.project_id = filteredProjects.value[0]?.id
  }
  syncFromProject()
}

watch(() => form.project_id, syncFromProject)
watch(() => form.times, () => {
  if (form.card_kind !== 'times') return
  const price = selectedProject.value?.price || 0
  form.amount = Number((price * Number(form.times || 0)).toFixed(2))
})

async function load() {
  loading.value = true
  try {
    const data: any = await http.get('/api/cards', { params: { page: page.value, pageSize: 20 } })
    rows.value = data.list || []
    total.value = data.total || 0
  } finally {
    loading.value = false
  }
}

async function refreshCustomers() {
  const c: any = await http.get('/api/customers', { params: { pageSize: 200 } })
  customers.value = c.list || []
}

async function openCreate() {
  const [p, s]: any[] = await Promise.all([
    http.get('/api/projects', { params: { pageSize: 100, status: '上架' } }),
    http.get('/api/stores'),
  ])
  await refreshCustomers()
  projects.value = p.list || []
  stores.value = s.list || []
  Object.assign(form, {
    customer_mode: customers.value.length ? 'old' : 'new',
    customer_id: customers.value[0]?.id,
    new_name: '',
    new_phone: '',
    new_level: '银卡会员',
    store_id: stores.value[0]?.id,
    project_id: undefined,
    card_kind: 'amount',
    pay_method: '微信',
  })
  const firstAmount = projects.value.find((p) => projectMode(p) === 'amount' || projectMode(p) === 'both')
  const firstAny = projects.value[0]
  form.project_id = firstAmount?.id || firstAny?.id
  if (form.project_id) {
    const m = projectMode(projects.value.find((p) => p.id === form.project_id))
    if (m === 'times') form.card_kind = 'times'
    else form.card_kind = 'amount'
  }
  syncFromProject()
  dialog.value = true
}

async function ensureCustomerId(): Promise<number> {
  if (form.customer_mode === 'old') {
    if (!form.customer_id) throw new Error('请选择老客户')
    return form.customer_id
  }
  if (!form.new_name || !form.new_phone) throw new Error('请填写新客户姓名和手机号')
  if (!form.store_id) throw new Error('请选择门店')
  const created: any = await http.post('/api/customers', {
    name: form.new_name,
    phone: form.new_phone,
    mobile: form.new_phone,
    level_name: form.new_level,
    store_id: form.store_id,
    source: '到店',
  })
  await refreshCustomers()
  form.customer_id = created.id
  form.customer_mode = 'old'
  return created.id
}

async function save() {
  try {
    if (!form.store_id) return ElMessage.warning('请选择门店')
    if (form.card_kind === 'times' && !form.project_id) return ElMessage.warning('请选择项目')
    if (form.card_kind === 'amount' && !(form.amount > 0)) return ElMessage.warning('请输入充值金额')
    if (selectedProject.value) {
      const m = projectMode(selectedProject.value)
      if (m === 'times' && form.card_kind !== 'times') return ElMessage.warning('该项目仅支持次卡模式')
      if (m === 'amount' && form.card_kind !== 'amount') return ElMessage.warning('该项目仅支持金额模式')
    }
    const customerId = await ensureCustomerId()
    await http.post('/api/cards', {
      customer_id: customerId,
      project_id: form.card_kind === 'times' ? form.project_id : (form.project_id || 0),
      store_id: form.store_id,
      pay_method: form.pay_method,
      card_kind: form.card_kind,
      times: form.card_kind === 'times' ? form.times : 0,
      amount: form.amount,
    })
    ElMessage.success('开卡成功')
    dialog.value = false
    load()
  } catch (e: any) {
    ElMessage.error(e?.message || e || '开卡失败')
  }
}

onMounted(load)
</script>

<template>
  <div>
    <PageHeader title="充值开卡" subtitle="可选老客户开卡，或直接新增客户后开卡">
      <el-button type="primary" @click="openCreate">开卡</el-button>
    </PageHeader>
    <el-table v-loading="loading" :data="rows" stripe border>
      <el-table-column prop="card_no" label="卡号" width="140" />
      <el-table-column prop="customer_name" label="客户" width="110" />
      <el-table-column label="类型" width="90">
        <template #default="{ row }">{{ row.card_kind === 'amount' ? '储值' : '次卡' }}</template>
      </el-table-column>
      <el-table-column prop="project_name" label="项目" />
      <el-table-column prop="store_name" label="开卡门店" width="120" />
      <el-table-column label="余额/次数" width="130">
        <template #default="{ row }">
          <span v-if="row.card_kind==='amount'">余额 ¥{{ row.amount }}</span>
          <span v-else>{{ row.remain }}/{{ row.total }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="amount" label="金额" width="110">
        <template #default="{ row }">¥{{ row.amount }}</template>
      </el-table-column>
      <el-table-column prop="pay_method" label="支付方式" width="100" />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }"><StatusTag :value="row.status" /></template>
      </el-table-column>
      <el-table-column label="开卡时间" width="170">
        <template #default="{ row }">{{ String(row.created_at||'').replace('T',' ').slice(0,19) }}</template>
      </el-table-column>
    </el-table>
    <div class="pager"><el-pagination v-model:current-page="page" layout="total, prev, pager, next" :total="total" @current-change="load" /></div>

    <el-dialog v-model="dialog" title="开卡 / 充值" width="560px" align-center destroy-on-close>
      <el-form label-width="100px">
        <el-form-item label="开卡方式">
          <el-radio-group v-model="form.card_kind" @change="onKindChange">
            <el-radio-button value="times" :disabled="!allowTimes">项目次数</el-radio-button>
            <el-radio-button value="amount" :disabled="!allowAmount">直接充值</el-radio-button>
          </el-radio-group>
          <div v-if="selectedProject" style="margin-top:6px;font-size:12px;color:#909399">
            当前项目：{{ selectedProject.card_mode==='times'?'次卡模式':selectedProject.card_mode==='amount'?'金额模式':'兼容模式' }}
          </div>
        </el-form-item>
        <el-form-item label="客户类型">
          <el-radio-group v-model="form.customer_mode">
            <el-radio-button value="old">选择老客户</el-radio-button>
            <el-radio-button value="new">新增新客户</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <template v-if="form.customer_mode==='old'">
          <el-form-item label="老客户">
            <el-select v-model="form.customer_id" filterable style="width:100%" placeholder="搜索姓名/手机" teleported>
              <el-option v-for="c in customers" :key="c.id" :label="`${c.name}（${c.phone}）`" :value="c.id" />
            </el-select>
          </el-form-item>
        </template>
        <template v-else>
          <el-form-item label="客户姓名"><el-input v-model="form.new_name" placeholder="新客户姓名" /></el-form-item>
          <el-form-item label="手机号"><el-input v-model="form.new_phone" maxlength="11" placeholder="11位手机号" /></el-form-item>
          <el-form-item label="会员等级">
            <el-select v-model="form.new_level" style="width:100%" teleported>
              <el-option label="钻石会员" value="钻石会员" />
              <el-option label="金卡会员" value="金卡会员" />
              <el-option label="银卡会员" value="银卡会员" />
            </el-select>
          </el-form-item>
        </template>
        <el-form-item :label="form.card_kind==='times' ? '项目' : '关联项目'">
          <el-select v-model="form.project_id" filterable clearable style="width:100%" :placeholder="form.card_kind==='times' ? '选择项目' : '可选'" teleported>
            <el-option
              v-for="p in filteredProjects"
              :key="p.id"
              :label="`${p.name} · ¥${p.price}${p.card_mode!=='amount' ? ` × ${p.total_times}次` : ''}（${p.card_mode==='times'?'次卡':p.card_mode==='amount'?'金额':'兼容'}）`"
              :value="p.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item v-if="form.card_kind==='times'" label="开卡次数">
          <el-input-number v-model="form.times" :min="1" :max="999" />
        </el-form-item>
        <el-form-item :label="form.card_kind==='amount' ? '充值金额' : '应收金额'">
          <el-input-number v-model="form.amount" :min="0" :precision="2" :step="100" style="width:100%" />
        </el-form-item>
        <el-form-item label="开卡门店">
          <el-select v-model="form.store_id" filterable style="width:100%" placeholder="选择门店" teleported>
            <el-option v-for="s in stores" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="支付方式">
          <el-select v-model="form.pay_method" style="width:100%" teleported>
            <el-option label="微信" value="微信" />
            <el-option label="刷卡" value="刷卡" />
            <el-option label="现金" value="现金" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog=false">取消</el-button>
        <el-button type="primary" @click="save">确认开卡</el-button>
      </template>
    </el-dialog>
  </div>
</template>
<style scoped>.pager{margin-top:14px;display:flex;justify-content:flex-end}</style>
