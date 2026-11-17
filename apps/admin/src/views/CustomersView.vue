<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import http from '@/api/http'
import PageHeader from '@/components/PageHeader.vue'
import StatusTag from '@/components/StatusTag.vue'

const q = ref('')
const loading = ref(false)
const rows = ref<any[]>([])
const stores = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const dialog = ref(false)
const detail = ref(false)
const detailRow = ref<any>(null)
const form = reactive({ name: '', phone: '', level: '银卡会员', source: '到店', store_id: undefined as number | undefined, remark: '' })

async function loadStores() {
  const data: any = await http.get('/api/stores')
  stores.value = data.list || []
  if (!form.store_id && stores.value.length) form.store_id = stores.value[0].id
}

async function load() {
  loading.value = true
  try {
    const data: any = await http.get('/api/customers', { params: { q: q.value, page: page.value, pageSize: 20 } })
    rows.value = data.list || []
    total.value = data.total || 0
  } finally {
    loading.value = false
  }
}

function openCreate() {
  Object.assign(form, { name: '', phone: '', level: '银卡会员', source: '到店', store_id: stores.value[0]?.id, remark: '' })
  dialog.value = true
}

async function save() {
  if (!form.store_id) return ElMessage.warning('请选择门店')
  await http.post('/api/customers', { ...form, mobile: form.phone, level_name: form.level })
  ElMessage.success('已新增客户')
  dialog.value = false
  load()
}

async function remove(row: any) {
  await ElMessageBox.confirm(`确认删除客户 ${row.name}？`, '提示')
  await http.delete(`/api/customers/${row.id}`)
  ElMessage.success('已删除')
  load()
}

async function openDetail(row: any) {
  const data: any = await http.get(`/api/customers/${row.id}`)
  detailRow.value = data
  detail.value = true
}

function assetText(row: any) {
  const parts: string[] = []
  if (Number(row.balance) > 0) parts.push(`余额 ¥${Number(row.balance).toLocaleString()}`)
  if (Number(row.remain) > 0) parts.push(`次数 ${row.remain}`)
  if (!parts.length) return row.cards_brief || '暂无开卡'
  return parts.join(' · ')
}

onMounted(async () => { await loadStores(); await load() })
</script>

<template>
  <div>
    <PageHeader title="客户管理" subtitle="余额与次数来自开卡充值数据；多项目可进详情查看">
      <el-input v-model="q" placeholder="姓名/手机/编号" clearable style="width:220px" @keyup.enter="load" />
      <el-button @click="load">搜索</el-button>
      <el-button type="primary" @click="openCreate">新增客户</el-button>
    </PageHeader>
    <el-table v-loading="loading" :data="rows" stripe border>
      <el-table-column prop="code" label="编号" width="140" />
      <el-table-column label="姓名" width="120">
        <template #default="{ row }">
          <el-popover placement="top" :width="320" trigger="hover">
            <template #reference>
              <el-button link type="primary" @click="openDetail(row)">{{ row.name }}</el-button>
            </template>
            <div class="tip-box">
              <div class="tip-name">{{ row.name }} · {{ row.phone }}</div>
              <div class="tip-line">门店：{{ row.store_name || '—' }}</div>
              <div class="tip-line">余额：¥{{ Number(row.balance||0).toLocaleString() }}</div>
              <div class="tip-line">剩余次数：{{ row.remain || 0 }}</div>
              <div class="tip-line">开卡：{{ row.cards_brief || '暂无' }}</div>
              <div class="tip-line">累计消费：¥{{ Number(row.spent||0).toLocaleString() }}</div>
            </div>
          </el-popover>
        </template>
      </el-table-column>
      <el-table-column prop="phone" label="手机号" width="130" />
      <el-table-column prop="level" label="等级" width="110" />
      <el-table-column prop="store_name" label="所属门店" min-width="120" />
      <el-table-column label="余额" width="110">
        <template #default="{ row }">¥{{ Number(row.balance||0).toLocaleString() }}</template>
      </el-table-column>
      <el-table-column label="剩余次数" width="100">
        <template #default="{ row }">{{ row.remain || 0 }}</template>
      </el-table-column>
      <el-table-column label="开卡概要" min-width="200">
        <template #default="{ row }">
          <span class="brief">{{ row.cards_brief || assetText(row) }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="spent" label="累计消费" width="110">
        <template #default="{ row }">¥{{ Number(row.spent||0).toLocaleString() }}</template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }"><StatusTag :value="row.status === 1 ? '正常' : '停用'" /></template>
      </el-table-column>
      <el-table-column label="操作" width="140" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openDetail(row)">详情</el-button>
          <el-button link type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <div class="pager">
      <el-pagination v-model:current-page="page" layout="total, prev, pager, next" :total="total" @current-change="load" />
    </div>

    <el-dialog v-model="dialog" title="新增客户" width="480px" align-center>
      <el-form label-width="80px">
        <el-form-item label="姓名"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="手机号"><el-input v-model="form.phone" /></el-form-item>
        <el-form-item label="等级">
          <el-select v-model="form.level" style="width:100%" teleported>
            <el-option label="钻石会员" value="钻石会员" />
            <el-option label="金卡会员" value="金卡会员" />
            <el-option label="银卡会员" value="银卡会员" />
          </el-select>
        </el-form-item>
        <el-form-item label="所属门店">
          <el-select v-model="form.store_id" filterable style="width:100%" placeholder="请选择门店" teleported>
            <el-option v-for="s in stores" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="来源">
          <el-select v-model="form.source" style="width:100%" teleported>
            <el-option label="到店" value="到店" />
            <el-option label="转介绍" value="转介绍" />
            <el-option label="线上" value="线上" />
            <el-option label="活动" value="活动" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" type="textarea" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog=false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>

    <el-drawer v-model="detail" title="客户详情 · 开卡资产" size="480px">
      <template v-if="detailRow">
        <div class="d-head">
          <div class="d-name">{{ detailRow.name }}</div>
          <div class="d-sub">{{ detailRow.phone }} · {{ detailRow.level }} · {{ detailRow.store_name }}</div>
          <div class="d-kpi">
            <div><em>¥{{ Number(detailRow.balance||0).toLocaleString() }}</em><span>储值余额</span></div>
            <div><em>{{ detailRow.remain || 0 }}</em><span>剩余次数</span></div>
            <div><em>{{ detailRow.cards_count || (detailRow.cards||[]).length }}</em><span>使用中卡</span></div>
          </div>
        </div>
        <el-table :data="detailRow.cards || []" size="small" border>
          <el-table-column label="类型" width="70">
            <template #default="{ row }">{{ row.card_kind==='amount' ? '储值' : '次卡' }}</template>
          </el-table-column>
          <el-table-column prop="project_name" label="项目" min-width="110">
            <template #default="{ row }">{{ row.project_name || (row.card_kind==='amount'?'储值卡':'—') }}</template>
          </el-table-column>
          <el-table-column label="余额/次数" width="110">
            <template #default="{ row }">
              <span v-if="row.card_kind==='amount'">¥{{ row.amount }}</span>
              <span v-else>{{ row.remain }}/{{ row.total }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="80" />
        </el-table>
        <p v-if="!(detailRow.cards||[]).length" class="empty">暂无开卡记录，请到「充值开卡」办理</p>
      </template>
    </el-drawer>
  </div>
</template>

<style scoped>
.pager { margin-top: 14px; display: flex; justify-content: flex-end; }
.brief { color:#667085; font-size:12px; }
.tip-box { font-size:13px; line-height:1.7; color:#344054; }
.tip-name { font-weight:700; margin-bottom:6px; }
.tip-line { color:#667085; }
.d-head { margin-bottom:16px; }
.d-name { font-size:18px; font-weight:740; }
.d-sub { margin-top:4px; color:#667085; font-size:13px; }
.d-kpi { margin-top:14px; display:grid; grid-template-columns:repeat(3,1fr); gap:10px; }
.d-kpi div { background:#f8fafc; border:1px solid #eef2f6; border-radius:12px; padding:12px; text-align:center; }
.d-kpi em { display:block; font-style:normal; font-weight:740; font-size:16px; color:#e24b6e; }
.d-kpi span { display:block; margin-top:4px; font-size:12px; color:#667085; }
.empty { margin-top:16px; color:#98a2b3; font-size:13px; text-align:center; }
</style>
