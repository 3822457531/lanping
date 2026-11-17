<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { showConfirmDialog, showSuccessToast } from 'vant'
import http from '@/api/http'

const router = useRouter()
const rows = ref<any[]>([])

async function load() {
  const data: any = await http.get('/api/verifications', { params: { pageSize: 50 } })
  rows.value = data.list || []
}

async function refund(row: any) {
  await showConfirmDialog({ title: '补退次', message: `确认对 ${row.customer_name} 的 ${row.project_name} 退次？` })
  await http.post(`/api/verifications/${row.id}/refund`)
  showSuccessToast('已退次')
  load()
}

onMounted(load)
</script>

<template>
  <div class="page-wrap">
    <van-nav-bar title="核销列表" right-text="去核销" @click-right="router.push('/verify')" />
    <div class="list">
      <div v-for="v in rows" :key="v.id" class="row soft-card">
        <div class="name">{{ v.customer_name }} · {{ v.project_name }}</div>
        <div class="sub">{{ v.verify_no }} · {{ v.staff_name }} · {{ String(v.created_at||'').replace('T',' ').slice(0,16) }}</div>
        <div class="foot">
          <span class="st">{{ v.status }} · {{ v.times }}次 · ¥{{ v.amount }}</span>
          <van-button v-if="v.status==='已完成'" size="mini" round plain type="danger" @click="refund(v)">补退</van-button>
        </div>
      </div>
      <van-empty v-if="!rows.length" description="暂无核销" />
    </div>
  </div>
</template>

<style scoped>
.list { padding: 12px 14px 24px; display:grid; gap:10px; }
.row {
  padding: 15px 14px;
  border-radius: 14px;
  background: rgba(255,255,255,.95);
  box-shadow: var(--lp-shadow-card);
  border: 1px solid rgba(255,245,248,.95);
}
.name { font-weight: 720; }
.sub { margin-top: 6px; font-size: 12px; color: var(--lp-muted); }
.foot {
  margin-top: 12px; padding-top: 10px; border-top: 1px dashed #f0e2e6;
  display:flex; justify-content:space-between; align-items:center;
}
.st { font-size: 12px; color: var(--lp-accent); font-weight: 700; font-family: var(--lp-display); }
</style>
