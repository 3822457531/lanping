<script setup lang="ts">
import { onMounted, ref } from 'vue'
import http from '@/api/http'
import { useSession } from '@/stores/session'

const session = useSession()
const rows = ref<any[]>([])

function money(v: any) {
  return Number(v || 0).toFixed(2)
}
function pct(v: any) {
  return `${(Number(v || 0) * 100).toFixed(0)}%`
}

onMounted(async () => {
  const data: any = await http.get('/api/salary/records', { params: { pageSize: 50 } })
  rows.value = (data.list || []).filter((x: any) => x.user_id === session.user?.id || session.user?.role === 'manager')
})
</script>

<template>
  <div class="page-wrap">
    <van-nav-bar title="薪资情况" left-arrow @click-left="$router.back()" />
    <p class="tip">销售提成=开卡提成 · 消耗提成=核销提成</p>
    <div class="list">
      <div v-for="r in rows" :key="r.id" class="row soft-card">
        <div class="name">{{ r.month }} · {{ r.user_name }} <span>{{ r.role_name }}</span></div>
        <div class="sub">固定 ¥{{ money(r.fixed_salary ?? r.base_salary) }}</div>
        <div class="sub">开卡 ¥{{ money(r.sales_amount) }} → 提成 ¥{{ money(r.sales_commission) }}（{{ pct(r.sales_rate) }}）</div>
        <div class="sub">核销 ¥{{ money(r.verify_amount) }} → 提成 ¥{{ money(r.verify_commission) }}（5%）</div>
        <div v-if="Number(r.manager_penalty) > 0" class="sub warn">业绩下限扣减 ¥{{ money(r.manager_penalty) }}</div>
        <div class="pay">应发 ¥{{ money(r.total_payable) }} <span>{{ r.status }}</span></div>
      </div>
      <van-empty v-if="!rows.length" description="暂无薪资记录，请管理员先计算" />
    </div>
  </div>
</template>

<style scoped>
.tip { margin: 6px 12px 0; font-size: 12px; color: var(--lp-muted); }
.list { padding:12px 14px 24px; display:grid; gap:10px; }
.row {
  padding:15px 14px;
  border-radius: 14px;
  background: rgba(255,255,255,.95);
  box-shadow: var(--lp-shadow-card);
  border: 1px solid rgba(255,245,248,.95);
}
.name { font-weight:720; }
.name span { margin-left:6px; font-size:12px; font-weight:500; color:var(--lp-muted); }
.sub { margin-top:6px; font-size:12px; color:var(--lp-muted); line-height: 1.45; }
.sub.warn { color:#c45656; }
.pay {
  margin-top:12px; padding-top:10px; border-top:1px dashed #f0e2e6;
  color:var(--lp-accent); font-family: var(--lp-display); font-weight:700; font-size:17px;
}
.pay span { margin-left:8px; font-size:12px; font-weight:500; color:var(--lp-muted); font-family: var(--lp-font); }
</style>
