<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { showSuccessToast } from 'vant'
import http from '@/api/http'

const tab = ref(0)
const users = ref<any[]>([])
const audits = ref<any[]>([])
const loading = ref(false)

const roleMap: Record<string, string> = { manager: '店长', staff: '普通员工', admin: '管理员' }
const statusMap: Record<string, string> = {
  pending: '待审核', approved: '已通过', rejected: '已拒绝',
  待审核: '待审核', 已通过: '已通过', 已拒绝: '已拒绝',
  active: '在职', disabled: '停用',
}

async function load() {
  loading.value = true
  try {
    const [u, a]: any[] = await Promise.all([
      http.get('/api/users', { params: { pageSize: 100 } }),
      http.get('/api/audits'),
    ])
    users.value = (u.list || []).filter((x: any) => x.role !== 'admin')
    audits.value = a.list || []
  } finally {
    loading.value = false
  }
}

async function approve(row: any) {
  await http.post(`/api/audits/${row.id}/approve`)
  showSuccessToast('已通过')
  load()
}
async function reject(row: any) {
  await http.post(`/api/audits/${row.id}/reject`)
  showSuccessToast('已拒绝')
  load()
}

function pending(s: string) {
  return s === 'pending' || s === '待审核'
}

onMounted(load)
</script>

<template>
  <div class="page-wrap">
    <van-nav-bar title="员工管理" left-arrow @click-left="$router.back()" />
    <van-tabs v-model:active="tab" color="#e24b6e" sticky>
      <van-tab title="本店员工">
        <div class="list">
          <div v-for="u in users" :key="u.id" class="row soft-card">
            <div>
              <div class="name">{{ u.name }} <span class="tag">{{ roleMap[u.role] || u.role_name }}</span></div>
              <div class="sub">{{ u.phone }} · {{ u.position || '—' }}</div>
            </div>
            <div class="st">{{ statusMap[u.status] || u.status }}</div>
          </div>
          <van-empty v-if="!users.length" description="暂无员工" />
        </div>
      </van-tab>
      <van-tab title="入职审核">
        <div class="list">
          <div v-for="a in audits" :key="a.id" class="row soft-card">
            <div class="grow">
              <div class="name">{{ a.name }} <span class="tag">{{ roleMap[a.role] || a.role }}</span></div>
              <div class="sub">{{ a.phone }} · {{ a.apply_at }}</div>
              <div class="sub">状态：{{ statusMap[a.status] || a.status }}</div>
              <div v-if="pending(a.status)" class="btns">
                <van-button size="small" round type="primary" @click="approve(a)">通过</van-button>
                <van-button size="small" round plain type="danger" @click="reject(a)">拒绝</van-button>
              </div>
            </div>
          </div>
          <van-empty v-if="!audits.length" description="暂无审核" />
        </div>
      </van-tab>
    </van-tabs>
  </div>
</template>

<style scoped>
.list { padding: 12px 14px 24px; display:grid; gap:10px; }
.row {
  padding: 10px; display:flex; justify-content:space-between; gap:10px;
  border-radius: 14px;
  background: rgba(255,255,255,.95);
  box-shadow: var(--lp-shadow-card);
  border: 1px solid rgba(255,245,248,.95);
}
.grow { flex: 1; }
.name { font-weight: 720; font-size: 13px; display:flex; align-items:center; gap:8px; }
.tag {
  font-size: 11px; font-weight: 650; color: var(--lp-accent);
  background: #fdeef2; padding: 2px 8px; border-radius: 999px;
}
.sub { margin-top: 5px; font-size: 12px; color: var(--lp-muted); }
.st { font-size: 12px; color: var(--lp-muted); white-space: nowrap; }
.btns { display:flex; gap:8px; margin-top: 10px; }
</style>
