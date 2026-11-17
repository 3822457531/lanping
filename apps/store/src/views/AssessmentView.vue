<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import http from '@/api/http'
import { useSession } from '@/stores/session'

const session = useSession()
const rows = ref<any[]>([])
const manager = computed(() => session.user?.role === 'manager')

onMounted(async () => {
  const data: any = await http.get('/api/assessments')
  const list = data.list || []
  rows.value = manager.value ? list : list.filter((x: any) => x.user_id === session.user?.id)
})
</script>

<template>
  <div class="page-wrap">
    <van-nav-bar :title="manager ? '考核情况' : '我的考核'" left-arrow @click-left="$router.back()" />
    <div class="list">
      <div v-for="r in rows" :key="r.id" class="row soft-card">
        <div class="name">{{ r.user_name }} · {{ r.month }}</div>
        <div class="sub">{{ r.name }} · 目标 {{ r.target }} · 完成 {{ r.done }}</div>
        <div class="badge" :class="r.pass ? 'ok' : 'bad'">{{ r.pass ? '合格' : '未合格' }}</div>
      </div>
      <van-empty v-if="!rows.length" description="暂无考核数据" />
    </div>
  </div>
</template>

<style scoped>
.list { padding:12px 14px 24px; display:grid; gap:10px; }
.row {
  padding:15px 14px; position:relative;
  border-radius: 14px;
  background: rgba(255,255,255,.95);
  box-shadow: var(--lp-shadow-card);
  border: 1px solid rgba(255,245,248,.95);
}
.name { font-weight:720; }
.sub { margin-top:6px; font-size:12px; color:var(--lp-muted); }
.badge {
  margin-top:12px; display:inline-flex; padding:4px 10px; border-radius:999px; font-size:12px; font-weight:700;
}
.ok { background:#eaf7f1; color:#2f9e74; }
.bad { background:#fdeeee; color:#d45454; }
</style>
