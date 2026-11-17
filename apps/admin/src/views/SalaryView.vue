<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import http from '@/api/http'
import PageHeader from '@/components/PageHeader.vue'
import StatusTag from '@/components/StatusTag.vue'

const tab = ref<'records' | 'rules'>('records')
const rows = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const loading = ref(false)
const saving = ref(false)
const month = ref(new Date().toISOString().slice(0, 7))

type Tier = { min: number; max: number; rate: number; label?: string }
type Cap = { max_people?: number | null; people?: number | null; min_people?: number | null; cap_ratio: number }

const form = reactive({
  name: '',
  note: '销售提成=开卡提成；消耗提成=核销提成',
  manager: {
    fixed_items: { base_salary: 1600, overtime: 700, non_compete: 100, social_subsidy: 1100 },
    sales_tiers: [] as Tier[],
    verify_rate: 0.05,
  },
  staff: {
    fixed_items: { base_salary: 1200, overtime: 700, non_compete: 0, social_subsidy: 1100 },
    sales_tiers: [] as Tier[],
    verify_rate: 0.05,
  },
  manager_extra: {
    min_store_share: 0.10,
    min_shortfall_penalty: 0.08,
    attend_days_gt: 20,
    cap_by_headcount: [] as Cap[],
  },
})

const mgrFixedTotal = computed(() => {
  const f = form.manager.fixed_items
  return Number(f.base_salary || 0) + Number(f.overtime || 0) + Number(f.non_compete || 0) + Number(f.social_subsidy || 0)
})
const staffFixedTotal = computed(() => {
  const f = form.staff.fixed_items
  return Number(f.base_salary || 0) + Number(f.overtime || 0) + Number(f.social_subsidy || 0)
})

function pct(v: any) {
  return `${(Number(v || 0) * 100).toFixed(0)}%`
}
function money(v: any) {
  return Number(v || 0).toFixed(2)
}
function ratePercent(row: Tier) {
  return Number(((row.rate || 0) * 100).toFixed(2))
}
function setRatePercent(row: Tier, val: number | undefined) {
  row.rate = Number(((val || 0) / 100).toFixed(4))
}

async function loadRecords() {
  loading.value = true
  try {
    const data: any = await http.get('/api/salary/records', { params: { page: page.value, pageSize: 20, month: month.value } })
    rows.value = data.list || []
    total.value = data.total || 0
  } finally {
    loading.value = false
  }
}

async function loadRules() {
  const r: any = await http.get('/api/salary/rules')
  form.name = r.name || ''
  form.note = r.note || form.note
  if (r.manager) {
    form.manager.fixed_items = { ...form.manager.fixed_items, ...(r.manager.fixed_items || {}) }
    form.manager.sales_tiers = (r.manager.sales_tiers || []).map((t: Tier) => ({ ...t }))
    form.manager.verify_rate = r.manager.verify_rate ?? 0.05
  }
  if (r.staff) {
    form.staff.fixed_items = { ...form.staff.fixed_items, ...(r.staff.fixed_items || {}) }
    form.staff.sales_tiers = (r.staff.sales_tiers || []).map((t: Tier) => ({ ...t }))
    form.staff.verify_rate = r.staff.verify_rate ?? 0.05
  }
  if (r.manager_extra) {
    form.manager_extra.min_store_share = r.manager_extra.min_store_share ?? 0.1
    form.manager_extra.min_shortfall_penalty = r.manager_extra.min_shortfall_penalty ?? 0.08
    form.manager_extra.attend_days_gt = r.manager_extra.attend_days_gt ?? 20
    form.manager_extra.cap_by_headcount = (r.manager_extra.cap_by_headcount || []).map((c: Cap) => ({ ...c }))
  }
}

async function saveRules() {
  saving.value = true
  try {
    await http.put('/api/salary/rules', {
      name: form.name,
      note: form.note,
      manager: form.manager,
      staff: form.staff,
      manager_extra: form.manager_extra,
    })
    ElMessage.success('薪资规则已保存')
    await loadRules()
  } catch (e: any) {
    ElMessage.error(e.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function compute() {
  await http.post('/api/salary/compute', { month: month.value })
  ElMessage.success('已按当前规则计算薪资')
  loadRecords()
}

async function markPaid(row: any) {
  await http.post(`/api/salary/records/${row.id}/status`, { status: '已发放' })
  ElMessage.success('已标记发放')
  loadRecords()
}

function addTier(role: 'manager' | 'staff') {
  form[role].sales_tiers.push({ min: 0, max: 0, rate: 0.02, label: '' })
}
function removeTier(role: 'manager' | 'staff', idx: number) {
  form[role].sales_tiers.splice(idx, 1)
}
function syncStaffTiersFromManager() {
  form.staff.sales_tiers = form.manager.sales_tiers.map((t) => ({ ...t }))
  form.staff.verify_rate = form.manager.verify_rate
  ElMessage.success('已同步店长开卡阶梯与核销提成率到店员')
}

onMounted(async () => {
  await Promise.all([loadRecords(), loadRules()])
})
</script>

<template>
  <div>
    <PageHeader title="薪资管理" subtitle="销售提成=开卡提成；消耗提成=核销提成。规则可在「薪资规则」中修改。">
      <el-radio-group v-model="tab" style="margin-right:12px">
        <el-radio-button value="records">薪资记录</el-radio-button>
        <el-radio-button value="rules">薪资规则</el-radio-button>
      </el-radio-group>
      <template v-if="tab==='records'">
        <el-date-picker v-model="month" type="month" value-format="YYYY-MM" style="width:140px;margin-right:8px" @change="loadRecords" />
        <el-button type="primary" @click="compute">计算本月薪资</el-button>
      </template>
      <el-button v-else type="primary" :loading="saving" @click="saveRules">保存规则</el-button>
    </PageHeader>

    <template v-if="tab==='records'">
      <el-table v-loading="loading" :data="rows" stripe border>
        <el-table-column prop="month" label="月份" width="90" />
        <el-table-column prop="user_name" label="员工" width="100" />
        <el-table-column prop="role_name" label="岗位" width="90" />
        <el-table-column label="固定薪酬" width="100">
          <template #default="{ row }">{{ money(row.fixed_salary ?? row.base_salary) }}</template>
        </el-table-column>
        <el-table-column label="开卡业绩" width="100">
          <template #default="{ row }">{{ money(row.sales_amount) }}</template>
        </el-table-column>
        <el-table-column label="开卡提成" width="120">
          <template #default="{ row }">{{ money(row.sales_commission) }}（{{ pct(row.sales_rate) }}）</template>
        </el-table-column>
        <el-table-column label="核销业绩" width="100">
          <template #default="{ row }">{{ money(row.verify_amount) }}</template>
        </el-table-column>
        <el-table-column label="核销提成" width="100">
          <template #default="{ row }">{{ money(row.verify_commission) }}</template>
        </el-table-column>
        <el-table-column label="下限扣减" width="100">
          <template #default="{ row }">{{ money(row.manager_penalty) }}</template>
        </el-table-column>
        <el-table-column label="应发" width="110">
          <template #default="{ row }"><b>{{ money(row.total_payable) }}</b></template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="90">
          <template #default="{ row }"><StatusTag :value="row.status" /></template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button v-if="row.status!=='已发放'" link type="primary" @click="markPaid(row)">发放</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="pager"><el-pagination v-model:current-page="page" layout="total, prev, pager, next" :total="total" @current-change="loadRecords" /></div>
    </template>

    <template v-else>
      <el-alert type="info" :closable="false" show-icon style="margin-bottom:14px"
        title="修改后点「保存规则」立即生效；再点「计算本月薪资」才会按新规则重算已有记录。" />

      <div class="card">
        <el-form label-width="120px">
          <el-form-item label="规则名称"><el-input v-model="form.name" style="max-width:420px" /></el-form-item>
          <el-form-item label="说明"><el-input v-model="form.note" style="max-width:560px" /></el-form-item>
        </el-form>
      </div>

      <div class="grid2">
        <div class="card">
          <div class="sec-title">店长固定薪酬 <span>合计 ¥{{ mgrFixedTotal.toFixed(0) }}</span></div>
          <el-form label-width="110px">
            <el-form-item label="基本工资"><el-input-number v-model="form.manager.fixed_items.base_salary" :min="0" :step="100" /></el-form-item>
            <el-form-item label="固定加班费"><el-input-number v-model="form.manager.fixed_items.overtime" :min="0" :step="100" /></el-form-item>
            <el-form-item label="禁业补贴"><el-input-number v-model="form.manager.fixed_items.non_compete" :min="0" :step="50" /></el-form-item>
            <el-form-item label="社保补贴"><el-input-number v-model="form.manager.fixed_items.social_subsidy" :min="0" :step="100" /></el-form-item>
            <el-form-item label="核销提成率">
              <el-input-number :model-value="Number((form.manager.verify_rate*100).toFixed(2))" :min="0" :max="100" :step="1"
                @update:model-value="(v:number)=> form.manager.verify_rate=Number(((v||0)/100).toFixed(4))" />
              <span class="unit">%</span>
            </el-form-item>
          </el-form>
        </div>
        <div class="card">
          <div class="sec-title">店员固定薪酬 <span>合计 ¥{{ staffFixedTotal.toFixed(0) }}</span></div>
          <el-form label-width="110px">
            <el-form-item label="基本工资"><el-input-number v-model="form.staff.fixed_items.base_salary" :min="0" :step="100" /></el-form-item>
            <el-form-item label="固定加班费"><el-input-number v-model="form.staff.fixed_items.overtime" :min="0" :step="100" /></el-form-item>
            <el-form-item label="社保补贴"><el-input-number v-model="form.staff.fixed_items.social_subsidy" :min="0" :step="100" /></el-form-item>
            <el-form-item label="核销提成率">
              <el-input-number :model-value="Number((form.staff.verify_rate*100).toFixed(2))" :min="0" :max="100" :step="1"
                @update:model-value="(v:number)=> form.staff.verify_rate=Number(((v||0)/100).toFixed(4))" />
              <span class="unit">%</span>
            </el-form-item>
          </el-form>
        </div>
      </div>

      <div class="card">
        <div class="sec-title">
          开卡（销售）提成阶梯
          <div class="ops">
            <el-button size="small" @click="syncStaffTiersFromManager">店长阶梯同步到店员</el-button>
            <el-button size="small" @click="addTier('manager')">加店长档</el-button>
            <el-button size="small" @click="addTier('staff')">加店员档</el-button>
          </div>
        </div>
        <div class="tier-grid">
          <div>
            <div class="sub">店长</div>
            <el-table :data="form.manager.sales_tiers" border size="small">
              <el-table-column label="下限(元)" width="120">
                <template #default="{ row }"><el-input-number v-model="row.min" :min="0" :step="1000" controls-position="right" /></template>
              </el-table-column>
              <el-table-column label="上限(元,0无上限)" width="150">
                <template #default="{ row }"><el-input-number v-model="row.max" :min="0" :step="1000" controls-position="right" /></template>
              </el-table-column>
              <el-table-column label="提成率%" width="120">
                <template #default="{ row }">
                  <el-input-number :model-value="ratePercent(row)" :min="0" :max="100" :step="1" controls-position="right"
                    @update:model-value="(v:number)=>setRatePercent(row,v)" />
                </template>
              </el-table-column>
              <el-table-column label="说明">
                <template #default="{ row }"><el-input v-model="row.label" /></template>
              </el-table-column>
              <el-table-column label="" width="70">
                <template #default="{ $index }"><el-button link type="danger" @click="removeTier('manager',$index)">删</el-button></template>
              </el-table-column>
            </el-table>
          </div>
          <div>
            <div class="sub">店员</div>
            <el-table :data="form.staff.sales_tiers" border size="small">
              <el-table-column label="下限(元)" width="120">
                <template #default="{ row }"><el-input-number v-model="row.min" :min="0" :step="1000" controls-position="right" /></template>
              </el-table-column>
              <el-table-column label="上限(元,0无上限)" width="150">
                <template #default="{ row }"><el-input-number v-model="row.max" :min="0" :step="1000" controls-position="right" /></template>
              </el-table-column>
              <el-table-column label="提成率%" width="120">
                <template #default="{ row }">
                  <el-input-number :model-value="ratePercent(row)" :min="0" :max="100" :step="1" controls-position="right"
                    @update:model-value="(v:number)=>setRatePercent(row,v)" />
                </template>
              </el-table-column>
              <el-table-column label="说明">
                <template #default="{ row }"><el-input v-model="row.label" /></template>
              </el-table-column>
              <el-table-column label="" width="70">
                <template #default="{ $index }"><el-button link type="danger" @click="removeTier('staff',$index)">删</el-button></template>
              </el-table-column>
            </el-table>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="sec-title">店长个人提成附加规则</div>
        <el-form label-width="180px" inline>
          <el-form-item label="业绩下限占门店比例">
            <el-input-number :model-value="Number((form.manager_extra.min_store_share*100).toFixed(2))" :min="0" :max="100" :step="1"
              @update:model-value="(v:number)=>form.manager_extra.min_store_share=Number(((v||0)/100).toFixed(4))" />
            <span class="unit">%</span>
          </el-form-item>
          <el-form-item label="不足部分扣减比例">
            <el-input-number :model-value="Number((form.manager_extra.min_shortfall_penalty*100).toFixed(2))" :min="0" :max="100" :step="1"
              @update:model-value="(v:number)=>form.manager_extra.min_shortfall_penalty=Number(((v||0)/100).toFixed(4))" />
            <span class="unit">%</span>
          </el-form-item>
          <el-form-item label="人数核定出勤天数>">
            <el-input-number v-model="form.manager_extra.attend_days_gt" :min="1" :max="31" :step="1" />
            <span class="unit">天</span>
          </el-form-item>
        </el-form>
        <div class="sub" style="margin:8px 0 10px">业绩上限（按有效上班人数）</div>
        <el-table :data="form.manager_extra.cap_by_headcount" border size="small" style="max-width:720px">
          <el-table-column label="人数≤" width="120">
            <template #default="{ row }"><el-input-number v-model="row.max_people" :min="0" controls-position="right" /></template>
          </el-table-column>
          <el-table-column label="人数=" width="120">
            <template #default="{ row }"><el-input-number v-model="row.people" :min="0" controls-position="right" /></template>
          </el-table-column>
          <el-table-column label="人数≥" width="120">
            <template #default="{ row }"><el-input-number v-model="row.min_people" :min="0" controls-position="right" /></template>
          </el-table-column>
          <el-table-column label="可占门店业绩%" width="160">
            <template #default="{ row }">
              <el-input-number :model-value="Number(((row.cap_ratio||0)*100).toFixed(2))" :min="0" :max="100" :step="1" controls-position="right"
                @update:model-value="(v:number)=>row.cap_ratio=Number(((v||0)/100).toFixed(4))" />
            </template>
          </el-table-column>
        </el-table>
        <div style="margin-top:16px">
          <el-button type="primary" :loading="saving" @click="saveRules">保存规则</el-button>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.pager{margin-top:14px;display:flex;justify-content:flex-end}
.card{background:#fff;border:1px solid var(--lp-border);border-radius:var(--lp-radius);padding:18px 20px;margin-bottom:14px}
.grid2{display:grid;grid-template-columns:1fr 1fr;gap:14px}
.sec-title{font-weight:700;margin-bottom:12px;display:flex;align-items:center;justify-content:space-between;gap:10px}
.sec-title span{font-weight:600;color:var(--lp-accent)}
.ops{display:flex;gap:8px;flex-wrap:wrap}
.sub{font-size:13px;color:#667085;margin-bottom:8px}
.unit{margin-left:6px;color:#667085}
.tier-grid{display:grid;grid-template-columns:1fr 1fr;gap:14px}
@media (max-width:1100px){.grid2,.tier-grid{grid-template-columns:1fr}}
</style>
