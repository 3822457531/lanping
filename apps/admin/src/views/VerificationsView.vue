<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import http from '@/api/http'
import PageHeader from '@/components/PageHeader.vue'
import StatusTag from '@/components/StatusTag.vue'

const rows = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    const data: any = await http.get('/api/verifications', { params: { page: page.value, pageSize: 20 } })
    rows.value = data.list || []
    total.value = data.total || 0
  } finally {
    loading.value = false
  }
}

async function refund(row: any) {
  await ElMessageBox.confirm(`确认对 ${row.verify_no} 退次？`, '退次')
  await http.post(`/api/verifications/${row.id}/refund`)
  ElMessage.success('已退次')
  load()
}

onMounted(load)
</script>

<template>
  <div>
    <PageHeader title="核销记录" subtitle="支持已完成记录退次" />
    <el-table v-loading="loading" :data="rows" stripe border>
      <el-table-column prop="verify_no" label="单号" width="160" />
      <el-table-column prop="customer_name" label="客户" width="100" />
      <el-table-column prop="project_name" label="项目" />
      <el-table-column prop="staff_name" label="员工" width="100" />
      <el-table-column prop="store_name" label="门店" width="120" />
      <el-table-column prop="times" label="次数" width="70" />
      <el-table-column prop="amount" label="金额" width="100">
        <template #default="{ row }">¥{{ row.amount }}</template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }"><StatusTag :value="row.status" /></template>
      </el-table-column>
      <el-table-column prop="created_at" label="时间" width="170">
        <template #default="{ row }">{{ String(row.created_at||'').replace('T',' ').slice(0,19) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="100">
        <template #default="{ row }">
          <el-button v-if="row.status==='已完成'" link type="danger" @click="refund(row)">退次</el-button>
        </template>
      </el-table-column>
    </el-table>
    <div class="pager"><el-pagination v-model:current-page="page" layout="total, prev, pager, next" :total="total" @current-change="load" /></div>
  </div>
</template>
<style scoped>.pager{margin-top:14px;display:flex;justify-content:flex-end}</style>
