<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import http from '@/api/http'
import PageHeader from '@/components/PageHeader.vue'
import StatusTag from '@/components/StatusTag.vue'

const rows = ref<any[]>([])
const loading = ref(false)
const dialog = ref(false)
const editDoneDialog = ref(false)
const editingId = ref<number | null>(null)
const month = ref(new Date().toISOString().slice(0, 7))

const presetNames = ['月度业绩（元）', '新增客户（人）', '考勤天数', '核销次数', '预约到店率（%）']

const form = reactive({
  month: month.value,
  name: '月度业绩（元）',
  weight: 1,
  target: 8000,
})

const doneForm = reactive({
  id: 0,
  user_name: '',
  name: '',
  target: 0,
  done: 0,
  user_id: 0,
  store_id: 0,
  month: '',
  weight: 1,
})

/** 按指标汇总（统一口径） */
const groups = computed(() => {
  const map = new Map<string, { name: string; target: number; weight: number; count: number; pass: number }>()
  for (const r of rows.value) {
    const key = r.name
    const g = map.get(key) || { name: r.name, target: r.target, weight: r.weight, count: 0, pass: 0 }
    g.count += 1
    if (r.pass) g.pass += 1
    g.target = r.target
    g.weight = r.weight
    map.set(key, g)
  }
  return [...map.values()]
})

async function load() {
  loading.value = true
  try {
    const data: any = await http.get('/api/assessments', { params: { month: month.value } })
    rows.value = data.list || []
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = null
  Object.assign(form, {
    month: month.value,
    name: '月度业绩（元）',
    weight: 1,
    target: 8000,
  })
  dialog.value = true
}

function openEditUnified(g: { name: string; target: number; weight: number }) {
  editingId.value = -1
  Object.assign(form, {
    month: month.value,
    name: g.name,
    weight: g.weight || 1,
    target: g.target,
  })
  dialog.value = true
}

function openEditDone(row: any) {
  Object.assign(doneForm, {
    id: row.id,
    user_name: row.user_name,
    name: row.name,
    target: row.target,
    done: row.done,
    user_id: row.user_id,
    store_id: row.store_id,
    month: row.month,
    weight: row.weight || 1,
  })
  editDoneDialog.value = true
}

async function saveUnified() {
  if (!form.name) return ElMessage.warning('请填写指标名称')
  if (editingId.value === -1) {
    await http.put('/api/assessments/0', {
      unified: true,
      month: form.month,
      name: form.name,
      weight: form.weight,
      target: form.target,
    })
    ElMessage.success('已统一更新该指标目标/权重')
  } else {
    await http.post('/api/assessments', {
      unified: true,
      month: form.month,
      name: form.name,
      weight: form.weight,
      target: form.target,
    })
    ElMessage.success('已统一下发给全体店长/员工')
  }
  dialog.value = false
  month.value = form.month
  load()
}

async function saveDone() {
  await http.put(`/api/assessments/${doneForm.id}`, {
    user_id: doneForm.user_id,
    store_id: doneForm.store_id,
    month: doneForm.month,
    name: doneForm.name,
    weight: doneForm.weight,
    target: doneForm.target,
    done: doneForm.done,
  })
  ElMessage.success('已更新完成值')
  editDoneDialog.value = false
  load()
}

async function removeRow(row: any) {
  await ElMessageBox.confirm(`删除「${row.user_name} · ${row.name}」这条记录？`, '提示', { type: 'warning' })
  await http.delete(`/api/assessments/${row.id}`)
  ElMessage.success('已删除')
  load()
}

async function removeUnified(g: { name: string }) {
  await ElMessageBox.confirm(`删除本月统一指标「${g.name}」下的全部人员记录？`, '提示', { type: 'warning' })
  await http.delete(`/api/assessments/0`, { params: { unified: 1, month: month.value, name: g.name } })
  ElMessage.success('已删除该统一指标')
  load()
}

onMounted(load)
</script>

<template>
  <div>
    <PageHeader title="考核情况" subtitle="统一配置指标并下发全员；业绩/新客/考勤/核销按开卡与业务数据自动统计，其它指标可手工填完成">
      <el-date-picker v-model="month" type="month" value-format="YYYY-MM" style="width:150px" @change="load" />
      <el-button @click="load">刷新</el-button>
      <el-button type="primary" @click="openCreate">新增统一指标</el-button>
    </PageHeader>

    <div class="groups" v-if="groups.length">
      <div v-for="g in groups" :key="g.name" class="g-card">
        <div class="g-main">
          <div class="g-name">{{ g.name }}</div>
          <div class="g-meta">目标 {{ g.target }} · 权重 {{ g.weight }} · {{ g.pass }}/{{ g.count }} 人合格</div>
        </div>
        <div class="g-ops">
          <el-button link type="primary" @click="openEditUnified(g)">改目标</el-button>
          <el-button link type="danger" @click="removeUnified(g)">删指标</el-button>
        </div>
      </div>
    </div>

    <el-table v-loading="loading" :data="rows" stripe border style="margin-top:14px">
      <el-table-column prop="user_name" label="员工" width="120" />
      <el-table-column prop="month" label="月份" width="100" />
      <el-table-column prop="name" label="指标" min-width="160" />
      <el-table-column prop="weight" label="权重" width="80" />
      <el-table-column prop="target" label="目标" width="110" />
      <el-table-column prop="done" label="完成" width="140">
        <template #default="{ row }">
          {{ row.done }}
          <el-tag v-if="row.auto" size="small" type="success" effect="plain" style="margin-left:6px">自动</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="pass" label="是否合格" width="110">
        <template #default="{ row }"><StatusTag :value="row.pass" /></template>
      </el-table-column>
      <el-table-column label="操作" width="160" fixed="right">
        <template #default="{ row }">
          <el-button v-if="!row.auto" link type="primary" @click="openEditDone(row)">填完成</el-button>
          <el-button link type="danger" @click="removeRow(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-empty v-if="!loading && !rows.length" description="本月暂无考核，点击「新增统一指标」" />

    <el-dialog v-model="dialog" :title="editingId===-1 ? '修改统一指标' : '新增统一指标'" width="520px" align-center destroy-on-close>
      <el-alert type="info" :closable="false" show-icon style="margin-bottom:14px"
        title="不选员工：保存后自动下发给全体在职店长/员工，目标与权重统一。" />
      <el-form label-width="90px">
        <el-form-item label="月份">
          <el-date-picker v-model="form.month" type="month" value-format="YYYY-MM" style="width:100%" :disabled="editingId===-1" />
        </el-form-item>
        <el-form-item label="指标名称">
          <el-select v-model="form.name" filterable allow-create default-first-option style="width:100%" teleported :disabled="editingId===-1">
            <el-option v-for="n in presetNames" :key="n" :label="n" :value="n" />
          </el-select>
        </el-form-item>
        <el-form-item label="权重"><el-input-number v-model="form.weight" :min="0.1" :step="0.1" :precision="2" /></el-form-item>
        <el-form-item label="统一目标"><el-input-number v-model="form.target" :min="0" :precision="2" style="width:100%" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog=false">取消</el-button>
        <el-button type="primary" @click="saveUnified">保存并下发</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="editDoneDialog" title="填写完成值" width="440px" align-center destroy-on-close>
      <el-form label-width="90px">
        <el-form-item label="员工"><el-input :model-value="doneForm.user_name" disabled /></el-form-item>
        <el-form-item label="指标"><el-input :model-value="doneForm.name" disabled /></el-form-item>
        <el-form-item label="目标"><el-input-number v-model="doneForm.target" :min="0" :precision="2" style="width:100%" /></el-form-item>
        <el-form-item label="完成值"><el-input-number v-model="doneForm.done" :min="0" :precision="2" style="width:100%" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDoneDialog=false">取消</el-button>
        <el-button type="primary" @click="saveDone">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.groups { display:grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap:10px; }
.g-card {
  display:flex; align-items:center; justify-content:space-between; gap:12px;
  padding:14px 16px; border:1px solid var(--lp-border); border-radius:12px; background:#fff;
}
.g-name { font-weight:720; }
.g-meta { margin-top:4px; font-size:12px; color:#667085; }
.g-ops { flex-shrink:0; }
</style>
