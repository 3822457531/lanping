<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import http from '@/api/http'
import PageHeader from '@/components/PageHeader.vue'
import StatusTag from '@/components/StatusTag.vue'

const rows = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const loading = ref(false)
const filter = ref<'active' | 'all'>('active')

const displayRows = computed(() => {
  if (filter.value === 'all') return rows.value
  return rows.value.filter((r) =>
    ['待履约', '已改约', '待确认', '已确认'].includes(r.status),
  )
})

function labelStatus(st: string) {
  if (st === '待确认' || st === '已确认') return '待履约'
  if (st === '失效') return '已失效'
  if (st === '已完成') return '已履约'
  return st
}

async function load() {
  loading.value = true
  try {
    const data: any = await http.get('/api/appointments', { params: { page: page.value, pageSize: 50 } })
    rows.value = data.list || []
    total.value = data.total || 0
  } finally {
    loading.value = false
  }
}

async function setStatus(row: any, status: string) {
  const tips: Record<string, string> = {
    已履约: '确认客户已到店，状态变为「已履约」',
    已取消: '取消后释放时段（≠已失效）',
  }
  try {
    await ElMessageBox.confirm(tips[status] || `变更为「${status}」？`, `预约操作 · ${status}`, {
      type: status === '已取消' ? 'warning' : 'info',
      confirmButtonText: '确定',
    })
    await http.put(`/api/appointments/${row.id}`, { status })
    ElMessage.success(`已变为「${status}」`)
    load()
  } catch (e: any) {
    if (e !== 'cancel' && e?.message) ElMessage.error(e.message)
  }
}

onMounted(load)
</script>

<template>
  <div>
    <PageHeader title="预约管理" subtitle="新建=待履约；到店确认=已履约；改约；取消；超时=已失效">
      <el-radio-group v-model="filter" size="small">
        <el-radio-button value="active">进行中</el-radio-button>
        <el-radio-button value="all">全部</el-radio-button>
      </el-radio-group>
    </PageHeader>
    <el-table v-loading="loading" :data="displayRows" stripe border>
      <el-table-column prop="customer_name" label="客户" width="110" />
      <el-table-column prop="project_name" label="项目" />
      <el-table-column prop="staff_name" label="员工" width="100" />
      <el-table-column label="预约时间" width="160">
        <template #default="{ row }">{{ row.date }} {{ row.time }}</template>
      </el-table-column>
      <el-table-column prop="source" label="来源" width="100" />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }"><StatusTag :value="labelStatus(row.status)" /></template>
      </el-table-column>
      <el-table-column label="操作" width="220">
        <template #default="{ row }">
          <template v-if="['待履约','已改约','待确认','已确认'].includes(row.status)">
            <el-button link type="primary" @click="setStatus(row,'已履约')">确认履约</el-button>
            <el-button link type="danger" @click="setStatus(row,'已取消')">取消</el-button>
          </template>
          <span v-else class="muted">终态</span>
        </template>
      </el-table-column>
    </el-table>
    <div class="pager">
      <el-pagination
        v-model:current-page="page"
        layout="total, prev, pager, next"
        :total="total"
        @current-change="load"
      />
    </div>
  </div>
</template>

<style scoped>
.pager { margin-top: 14px; display: flex; justify-content: flex-end; }
.muted { color: #999; font-size: 12px; }
</style>
