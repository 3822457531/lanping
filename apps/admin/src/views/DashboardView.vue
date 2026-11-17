<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import http from '@/api/http'
import PageHeader from '@/components/PageHeader.vue'

const data = ref<any>({})
const loading = ref(true)
const period = ref<'day' | 'week' | 'month'>('month')

const periodOptions = [
  { value: 'day', label: '日' },
  { value: 'week', label: '周' },
  { value: 'month', label: '月' },
] as const

const maxTrend = computed(() => {
  const arr = (data.value.trend || []).map((x: any) => Number(x.amount || 0))
  return Math.max(...arr, 1)
})

async function load() {
  loading.value = true
  try {
    data.value = await http.get('/api/dashboard', { params: { period: period.value } })
  } finally {
    loading.value = false
  }
}

watch(period, load)
onMounted(load)

function money(n: number) {
  return '¥' + Number(n || 0).toLocaleString('zh-CN', { maximumFractionDigits: 0 })
}
</script>

<template>
  <div v-loading="loading">
    <PageHeader :title="'数据看板'" :subtitle="(data.period_label || '本月') + '经营概览'">
      <el-radio-group v-model="period" size="default">
        <el-radio-button v-for="p in periodOptions" :key="p.value" :value="p.value">{{ p.label }}</el-radio-button>
      </el-radio-group>
    </PageHeader>

    <div class="stats">
      <div class="stat"><div class="l">营收</div><div class="n">{{ money(data.revenue) }}</div></div>
      <div class="stat"><div class="l">开卡金额</div><div class="n">{{ money(data.card_amount) }}</div></div>
      <div class="stat"><div class="l">核销金额</div><div class="n">{{ money(data.verify_amount) }}</div></div>
      <div class="stat"><div class="l">核销笔数</div><div class="n">{{ data.verify_count || 0 }}</div></div>
      <div class="stat"><div class="l">新增客户</div><div class="n">{{ data.new_customers || 0 }}</div></div>
      <div class="stat"><div class="l">客户总数</div><div class="n">{{ data.customer_total || 0 }}</div></div>
      <div class="stat"><div class="l">待确认核销</div><div class="n">{{ data.pending_verify || 0 }}</div></div>
      <div class="stat"><div class="l">待审核入职</div><div class="n">{{ data.pending_audit || 0 }}</div></div>
    </div>

    <div class="card">
      <div class="card-hd">
        <h3>{{ data.trend_label || '核销趋势' }}</h3>
        <span class="range" v-if="data.range_start">自 {{ data.range_start }}</span>
      </div>
      <div class="bars" :class="{ dense: (data.trend||[]).length > 14 }">
        <div v-for="t in (data.trend || [])" :key="t.date" class="bar-col" :title="`${t.date} · ${money(t.amount)}`">
          <div class="bar" :style="{ height: Math.max(4, (Number(t.amount||0) / maxTrend) * 120) + 'px' }"></div>
          <div class="d">{{ String(t.date).slice(5) }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.stats { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; margin-bottom: 16px; }
.stat {
  background: var(--lp-surface); border: 1px solid var(--lp-border); border-radius: var(--lp-radius);
  padding: 16px 18px;
}
.l { color: var(--lp-muted); font-size: 12.5px; }
.n { margin-top: 8px; font-size: 24px; font-weight: 700; font-family: var(--lp-num); }
.card {
  background: var(--lp-surface); border: 1px solid var(--lp-border); border-radius: var(--lp-radius);
  padding: 18px 20px;
}
.card-hd { display: flex; align-items: baseline; justify-content: space-between; gap: 12px; margin-bottom: 16px; }
h3 { margin: 0; font-size: 15px; }
.range { font-size: 12px; color: var(--lp-muted); }
.bars { display: flex; align-items: flex-end; gap: 14px; height: 160px; overflow-x: auto; }
.bars.dense { gap: 6px; }
.bar-col { flex: 1; min-width: 18px; display: flex; flex-direction: column; align-items: center; justify-content: flex-end; gap: 8px; }
.bar { width: 100%; max-width: 36px; background: var(--lp-accent); border-radius: 6px 6px 2px 2px; opacity: .85; }
.d { font-size: 11px; color: var(--lp-muted); white-space: nowrap; }
@media (max-width: 1100px) { .stats { grid-template-columns: repeat(2, 1fr); } }
</style>
