<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import http from '@/api/http'
import PageHeader from '@/components/PageHeader.vue'
import StatusTag from '@/components/StatusTag.vue'

const rows = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const loading = ref(false)
const dialog = ref(false)
const form = reactive<any>({
  name: '',
  category: '',
  price: 0,
  member_price: 0,
  total_times: 10,
  card_mode: 'amount',
  duration: 60,
  status: '上架',
})

const modeLabel: Record<string, string> = {
  times: '次卡模式',
  amount: '金额模式',
  both: '兼容模式',
}

async function load() {
  loading.value = true
  try {
    const data: any = await http.get('/api/projects', { params: { page: page.value, pageSize: 20 } })
    rows.value = data.list || []
    total.value = data.total || 0
  } finally {
    loading.value = false
  }
}

function openCreate() {
  Object.assign(form, {
    id: undefined,
    name: '',
    category: '面部护理',
    price: 398,
    member_price: 358,
    total_times: 10,
    card_mode: 'amount',
    duration: 60,
    status: '上架',
    commission_rate: 0.02,
    valid_days: 365,
  })
  dialog.value = true
}

async function save() {
  if (!form.name?.trim()) return ElMessage.warning('请填写项目名称')
  if (form.id) {
    const res: any = await http.put(`/api/projects/${form.id}`, form)
    const cv = res?.convert
    if (cv && Number(cv.converted_cards || 0) > 0) {
      ElMessage.success(`已保存，并已将 ${cv.converted_cards} 张次卡折算转入客户余额（¥${Number(cv.converted_amount || 0).toFixed(2)}）`)
    } else {
      ElMessage.success('已保存')
    }
  } else {
    await http.post('/api/projects', form)
    ElMessage.success('已保存')
  }
  dialog.value = false
  load()
}

function edit(row: any) {
  Object.assign(form, { ...row, card_mode: row.card_mode || 'amount' })
  dialog.value = true
}

onMounted(load)
</script>

<template>
  <div>
    <PageHeader title="项目管理" subtitle="默认金额模式（更灵活）；改为金额时会把原次卡剩余价值转入客户余额">
      <el-button type="primary" @click="openCreate">新增项目</el-button>
    </PageHeader>
    <el-table v-loading="loading" :data="rows" stripe border>
      <el-table-column prop="name" label="项目" min-width="140" />
      <el-table-column prop="category" label="分类" width="120" />
      <el-table-column prop="price" label="标价" width="100">
        <template #default="{ row }">¥{{ row.price }}</template>
      </el-table-column>
      <el-table-column prop="member_price" label="会员价" width="100">
        <template #default="{ row }">¥{{ row.member_price }}</template>
      </el-table-column>
      <el-table-column prop="total_times" label="次数" width="80" />
      <el-table-column label="开卡模式" width="110">
        <template #default="{ row }">
          <el-tag size="small" :type="row.card_mode==='times' ? 'warning' : row.card_mode==='amount' ? 'success' : 'info'">
            {{ modeLabel[row.card_mode || 'amount'] || '金额模式' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="duration" label="时长(分)" width="100" />
      <el-table-column prop="status" label="状态" width="90">
        <template #default="{ row }"><StatusTag :value="row.status" /></template>
      </el-table-column>
      <el-table-column label="操作" width="100">
        <template #default="{ row }"><el-button link type="primary" @click="edit(row)">编辑</el-button></template>
      </el-table-column>
    </el-table>
    <div class="pager"><el-pagination v-model:current-page="page" layout="total, prev, pager, next" :total="total" @current-change="load" /></div>
    <el-dialog v-model="dialog" :title="form.id ? '编辑项目' : '新增项目'" width="560px" destroy-on-close>
      <el-form label-width="100px">
        <el-form-item label="名称" required><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="分类"><el-input v-model="form.category" /></el-form-item>
        <el-form-item label="开卡模式" required>
          <el-radio-group v-model="form.card_mode">
            <el-radio-button value="times">次卡模式</el-radio-button>
            <el-radio-button value="amount">金额模式</el-radio-button>
            <el-radio-button value="both">兼容模式</el-radio-button>
          </el-radio-group>
          <div class="mode-tip">
            默认金额模式更灵活。改为金额模式时，该项目下未用完的次卡会按剩余次数折算金额并入客户储值；从金额改回次卡/兼容不会反向转换。
          </div>
        </el-form-item>
        <el-form-item label="标价"><el-input-number v-model="form.price" :min="0" /></el-form-item>
        <el-form-item label="会员价"><el-input-number v-model="form.member_price" :min="0" /></el-form-item>
        <el-form-item v-if="form.card_mode !== 'amount'" label="默认次数">
          <el-input-number v-model="form.total_times" :min="1" />
        </el-form-item>
        <el-form-item label="时长"><el-input-number v-model="form.duration" :min="1" /></el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status" style="width:100%">
            <el-option label="上架" value="上架" />
            <el-option label="下架" value="下架" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog=false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>
<style scoped>
.pager{margin-top:14px;display:flex;justify-content:flex-end}
.mode-tip{margin-top:8px;font-size:12px;color:#909399;line-height:1.5}
</style>
