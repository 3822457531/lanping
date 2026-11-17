<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import http from '@/api/http'
import PageHeader from '@/components/PageHeader.vue'
import StatusTag from '@/components/StatusTag.vue'

const rows = ref<any[]>([])
const issued = ref<any[]>([])
const customers = ref<any[]>([])
const stores = ref<any[]>([])
const issueDialog = ref(false)
const editDialog = ref(false)
const current = ref<any>(null)
const saving = ref(false)
const issuing = ref(false)
const issueForm = reactive({
  scope: 'store' as 'store' | 'all',
  store_id: undefined as number | undefined,
  customer_id: undefined as number | undefined,
})
const form = reactive({
  id: 0,
  name: '',
  amount: 0,
  valid_days: 30,
  description: '',
  status: '启用',
})

async function load() {
  const [w, i]: any[] = await Promise.all([
    http.get('/api/welfare'),
    http.get('/api/welfare/issued', { params: { pageSize: 50 } }),
  ])
  rows.value = w.list || []
  issued.value = i.list || []
}

function openCreate() {
  Object.assign(form, { id: 0, name: '', amount: 100, valid_days: 30, description: '', status: '启用' })
  editDialog.value = true
}

function openEdit(row: any) {
  Object.assign(form, {
    id: row.id,
    name: row.name,
    amount: Number(row.amount || 0),
    valid_days: Number(row.valid_days || 30),
    description: row.description || '',
    status: row.status || '启用',
  })
  editDialog.value = true
}

async function save() {
  if (!form.name.trim()) return ElMessage.warning('请填写福利名称')
  if (!(form.valid_days > 0)) return ElMessage.warning('请设置有效天数')
  saving.value = true
  try {
    const payload = {
      name: form.name.trim(),
      amount: Number(form.amount || 0),
      valid_days: Number(form.valid_days || 30),
      description: form.description,
      status: form.status,
    }
    if (form.id) {
      await http.put(`/api/welfare/${form.id}`, payload)
      ElMessage.success('已更新')
    } else {
      await http.post('/api/welfare', payload)
      ElMessage.success('已新增')
    }
    editDialog.value = false
    load()
  } catch (e: any) {
    ElMessage.error(e.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function remove(row: any) {
  await ElMessageBox.confirm(`确定删除福利「${row.name}」？`, '提示', { type: 'warning' })
  await http.delete(`/api/welfare/${row.id}`)
  ElMessage.success('已删除')
  load()
}

async function openIssue(row: any) {
  current.value = row
  const [u, s]: any[] = await Promise.all([
    http.get('/api/customers', { params: { pageSize: 200 } }),
    http.get('/api/stores'),
  ])
  customers.value = u.list || []
  stores.value = s.list || []
  issueForm.scope = 'store'
  issueForm.store_id = stores.value[0]?.id
  issueForm.customer_id = undefined
  issueDialog.value = true
}

async function issue() {
  if (issueForm.scope === 'store' && !issueForm.store_id) {
    return ElMessage.warning('请选择发放门店')
  }
  issuing.value = true
  try {
    const res: any = await http.post(`/api/welfare/${current.value.id}/issue`, {
      all_stores: issueForm.scope === 'all',
      store_id: issueForm.scope === 'store' ? issueForm.store_id : 0,
      customer_id: issueForm.customer_id || 0,
    })
    const n = res?.issued_stores || 1
    const days = current.value?.valid_days || 30
    ElMessage.success(
      issueForm.scope === 'all'
        ? `已统一发放到 ${n} 家门店（有效 ${days} 天）`
        : `已发放到指定门店（有效 ${days} 天）`,
    )
    issueDialog.value = false
    load()
  } catch (e: any) {
    ElMessage.error(e.message || '发放失败')
  } finally {
    issuing.value = false
  }
}

onMounted(load)
</script>

<template>
  <div>
    <PageHeader title="福利管理" subtitle="设置金额与有效天数；金额模式核销时可抵扣消费">
      <el-button type="primary" @click="openCreate">新增福利</el-button>
    </PageHeader>
    <el-table :data="rows" stripe border>
      <el-table-column prop="name" label="福利名称" min-width="140" />
      <el-table-column prop="amount" label="金额" width="120">
        <template #default="{ row }">¥{{ Number(row.amount||0).toFixed(2) }}</template>
      </el-table-column>
      <el-table-column label="有效期" width="110">
        <template #default="{ row }">{{ row.valid_days || 30 }} 天</template>
      </el-table-column>
      <el-table-column prop="description" label="说明" min-width="180" />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }"><StatusTag :value="row.status" /></template>
      </el-table-column>
      <el-table-column label="操作" width="220" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="primary" @click="openIssue(row)">发放</el-button>
          <el-button link type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div style="margin-top:22px">
      <h3 style="margin:0 0 10px;font-size:15px;font-weight:650">发放记录</h3>
      <el-table :data="issued" stripe border size="small">
        <el-table-column prop="welfare_name" label="福利" min-width="110" />
        <el-table-column prop="store_name" label="门店" width="120" />
        <el-table-column label="客户" width="110">
          <template #default="{ row }">{{ row.customer_name || '门店通用' }}</template>
        </el-table-column>
        <el-table-column label="剩余/总额" width="120">
          <template #default="{ row }">
            ¥{{ Number(row.remain_amount ?? row.amount).toFixed(0) }} / {{ Number(row.amount||0).toFixed(0) }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="90">
          <template #default="{ row }"><StatusTag :value="row.status || '可用'" /></template>
        </el-table-column>
        <el-table-column label="截止时间" width="160">
          <template #default="{ row }">{{ row.expire_at || '—' }}</template>
        </el-table-column>
        <el-table-column prop="issued_at" label="发放时间" width="160" />
      </el-table>
      <el-empty v-if="!issued.length" description="暂无发放记录" :image-size="64" />
    </div>

    <el-dialog v-model="editDialog" :title="form.id ? '编辑福利' : '新增福利'" width="500px" destroy-on-close>
      <el-form label-width="100px">
        <el-form-item label="福利名称" required><el-input v-model="form.name" placeholder="如：生日礼金" /></el-form-item>
        <el-form-item label="金额" required><el-input-number v-model="form.amount" :min="0" :step="50" /></el-form-item>
        <el-form-item label="有效天数" required>
          <el-input-number v-model="form.valid_days" :min="1" :max="3650" :step="1" />
          <div style="margin-top:6px;font-size:12px;color:#909399">发放后 N 天内可在金额模式核销时抵扣</div>
        </el-form-item>
        <el-form-item label="说明"><el-input v-model="form.description" type="textarea" :rows="3" placeholder="用途与发放规则说明" /></el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status" style="width:100%">
            <el-option label="启用" value="启用" />
            <el-option label="停用" value="停用" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialog=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="issueDialog" title="发放福利" width="480px" destroy-on-close>
      <el-alert
        type="info"
        :closable="false"
        show-icon
        style="margin-bottom:14px"
        :title="`发放后有效 ${current?.valid_days || 30} 天；金额模式核销时可抵扣。客户可选，不选则为门店通用。`"
      />
      <el-form label-width="90px">
        <el-form-item label="发放范围" required>
          <el-radio-group v-model="issueForm.scope">
            <el-radio-button value="store">单独门店</el-radio-button>
            <el-radio-button value="all">全部门店</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="issueForm.scope==='store'" label="门店" required>
          <el-select v-model="issueForm.store_id" filterable style="width:100%" placeholder="选择门店">
            <el-option v-for="s in stores" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="客户">
          <el-select
            v-model="issueForm.customer_id"
            filterable
            clearable
            style="width:100%"
            placeholder="可选，不选则门店通用"
          >
            <el-option v-for="c in customers" :key="c.id" :label="`${c.name}（${c.phone}）`" :value="c.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="issueDialog=false">取消</el-button>
        <el-button type="primary" :loading="issuing" @click="issue">确认发放</el-button>
      </template>
    </el-dialog>
  </div>
</template>
