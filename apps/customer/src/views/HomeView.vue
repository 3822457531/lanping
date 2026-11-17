<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import http from '@/api/http'

const router = useRouter()
const me = ref<any>(JSON.parse(localStorage.getItem('lp_customer_user') || 'null'))
const cards = ref<any[]>([])

onMounted(async () => {
  if (!localStorage.getItem('lp_customer_token')) {
    router.replace('/login')
    return
  }
  try {
    me.value = await http.get('/api/customer/me')
    const data: any = await http.get('/api/cards', { params: { pageSize: 50 } })
    cards.value = data.list || []
  } catch {
    router.replace('/login')
  }
})
</script>

<template>
  <div>
    <div class="hero">
      <div class="name">{{ me?.name }}</div>
      <div class="meta">{{ me?.level }} · 剩余 {{ me?.remain ?? cards.reduce((s,c)=>s+(c.remain||0),0) }} 次</div>
    </div>
    <van-cell-group inset title="我的卡项">
      <van-cell
        v-for="c in cards"
        :key="c.id"
        :title="c.project_name"
        :label="c.card_no"
        :value="`${c.remain}/${c.total}`"
      />
    </van-cell-group>
  </div>
</template>

<style scoped>
.hero {
  margin-bottom: 14px; padding: 28px 18px; color:#fff;
  background: linear-gradient(150deg,#f291ad,#e2587c); border-radius: 0 0 22px 22px;
}
.name { font-size: 22px; font-weight: 650; }
.meta { margin-top: 8px; opacity: .92; font-size: 13px; }
</style>
