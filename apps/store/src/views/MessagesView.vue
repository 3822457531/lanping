<script setup lang="ts">
import { onMounted, ref } from 'vue'
import http from '@/api/http'

const rows = ref<any[]>([])
onMounted(async () => {
  const data: any = await http.get('/api/messages')
  rows.value = data.list || []
})
</script>

<template>
  <div class="page-wrap">
    <van-nav-bar title="消息中心" left-arrow @click-left="$router.back()" />
    <div class="list">
      <div v-for="m in rows" :key="m.id" class="row">
        <div class="ico"><van-icon name="bell" /></div>
        <div class="body">
          <div class="title">{{ m.title }}</div>
          <div class="content">{{ m.content }}</div>
          <div class="time">{{ String(m.created_at||'').replace('T',' ').slice(0,16) }}</div>
        </div>
      </div>
      <van-empty v-if="!rows.length" description="暂无消息" />
    </div>
  </div>
</template>

<style scoped>
.list { padding: 8px 14px 24px; display: grid; gap: 6px; }
.row {
  display: grid;
  grid-template-columns: 42px 1fr;
  gap: 8px;
  padding: 10px;
  border-radius: 14px;
  background: rgba(255,255,255,.95);
  box-shadow: var(--lp-shadow-card);
  border: 1px solid rgba(255,245,248,.95);
}
.ico {
  width: 42px; height: 42px; border-radius: 14px;
  display: grid; place-items: center;
  background: linear-gradient(145deg, #ff8fa6, #e24b6e);
  color: #fff; font-size: 13px;
}
.title { font-weight: 720; font-size: 14px; }
.content { margin-top: 5px; font-size: 12px; color: var(--lp-muted); line-height: 1.5; }
.time { margin-top: 8px; font-size: 11px; color: #b5a8ac; font-family: var(--lp-display); }
</style>
