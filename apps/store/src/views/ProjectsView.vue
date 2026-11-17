<script setup lang="ts">
import { onMounted, ref } from 'vue'
import http from '@/api/http'

const rows = ref<any[]>([])
onMounted(async () => {
  const data: any = await http.get('/api/projects', { params: { pageSize: 100, status: '上架' } })
  rows.value = data.list || []
})
</script>

<template>
  <div class="page-wrap">
    <van-nav-bar title="项目管理" left-arrow @click-left="$router.back()" />
    <p class="tip">总部设定，门店只读</p>
    <div class="list">
      <div v-for="p in rows" :key="p.id" class="row soft-card">
        <div class="name">{{ p.name }}</div>
        <div class="sub">
          {{ p.category }} · {{ p.duration }}分钟
          <template v-if="(p.card_mode||'amount')!=='amount'"> · {{ p.total_times }}次</template>
          · {{ p.card_mode==='times'?'次卡模式':p.card_mode==='both'?'兼容模式':'金额模式' }}
        </div>
        <div class="price">¥{{ p.price }} <span>会员价 ¥{{ p.member_price }}</span></div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.tip { margin: 0; padding: 8px 16px 4px; color: var(--lp-muted); font-size: 12px; }
.list { padding: 8px 14px 24px; display:grid; gap:10px; }
.row {
  padding: 15px 14px;
  border-radius: 14px;
  background: rgba(255,255,255,.95);
  box-shadow: var(--lp-shadow-card);
  border: 1px solid rgba(255,245,248,.95);
}
.name { font-weight: 720; font-size: 13px; }
.sub { margin-top: 6px; color: var(--lp-muted); font-size: 12px; }
.price {
  margin-top: 12px; padding-top: 10px; border-top: 1px dashed #f0e2e6;
  color: var(--lp-accent); font-family: var(--lp-display); font-weight: 700; font-size: 14px;
}
.price span { margin-left: 8px; color: var(--lp-muted); font-size: 12px; font-weight: 500; font-family: var(--lp-font); }
</style>
