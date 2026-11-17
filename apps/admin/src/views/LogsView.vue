<script setup lang="ts">
import { onMounted, ref } from 'vue'
import http from '@/api/http'
import PageHeader from '@/components/PageHeader.vue'

const rows = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    const data: any = await http.get('/api/operate-logs', { params: { page: page.value, pageSize: 20 } })
    rows.value = data.list || []
    total.value = data.total || 0
  } finally {
    loading.value = false
  }
}
onMounted(load)
</script>

<template>
  <div>
    <PageHeader title="操作日志" subtitle="模块与动作已中文化，操作人显示账号姓名" />
    <el-table v-loading="loading" :data="rows" stripe border>
      <el-table-column prop="created_at" label="时间" width="180" />
      <el-table-column prop="user_name" label="操作人" width="120" />
      <el-table-column prop="module" label="模块" width="120" />
      <el-table-column prop="action" label="动作" width="100" />
      <el-table-column prop="detail" label="详情" min-width="180" />
      <el-table-column prop="ip" label="IP" width="140" />
    </el-table>
    <div class="pager"><el-pagination v-model:current-page="page" layout="total, prev, pager, next" :total="total" @current-change="load" /></div>
  </div>
</template>
<style scoped>.pager{margin-top:14px;display:flex;justify-content:flex-end}</style>
