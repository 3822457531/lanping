<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import http from '@/api/http'
import PageHeader from '@/components/PageHeader.vue'
import StatusTag from '@/components/StatusTag.vue'
import { isPendingAudit, labelRole } from '@/utils/labels'

const rows = ref<any[]>([])
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    const data: any = await http.get('/api/audits')
    rows.value = data.list || []
  } finally {
    loading.value = false
  }
}

async function approve(row: any) {
  await http.post(`/api/audits/${row.id}/approve`)
  ElMessage.success('已通过')
  load()
}
async function reject(row: any) {
  await http.post(`/api/audits/${row.id}/reject`)
  ElMessage.success('已拒绝')
  load()
}
onMounted(load)
</script>

<template>
  <div>
    <PageHeader title="审核管理" subtitle="员工通过手机号注册后进入本列表，通过后获得店长/员工权限" />
    <el-table v-loading="loading" :data="rows" stripe border>
      <el-table-column prop="name" label="姓名" width="110" />
      <el-table-column prop="phone" label="手机号" width="130" />
      <el-table-column label="申请角色" width="120">
        <template #default="{ row }">{{ labelRole(row.role) }}</template>
      </el-table-column>
      <el-table-column prop="store_name" label="门店" />
      <el-table-column prop="apply_at" label="申请时间" width="160" />
      <el-table-column label="状态" width="110">
        <template #default="{ row }"><StatusTag kind="audit" :value="row.status" /></template>
      </el-table-column>
      <el-table-column label="操作" width="160">
        <template #default="{ row }">
          <template v-if="isPendingAudit(row.status)">
            <el-button link type="primary" @click="approve(row)">通过</el-button>
            <el-button link type="danger" @click="reject(row)">拒绝</el-button>
          </template>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>
