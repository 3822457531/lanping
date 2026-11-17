<script setup lang="ts">
import { computed } from 'vue'
import { labelAudit, labelRole } from '@/utils/labels'

const props = defineProps<{ value: string | number | boolean; kind?: 'default' | 'audit' | 'role' }>()

const text = computed(() => {
  const raw = String(props.value)
  if (props.kind === 'audit') return labelAudit(raw)
  if (props.kind === 'role') return labelRole(raw)
  if (raw === 'true') return '是'
  if (raw === 'false') return '否'
  if (raw === 'active') return '在职'
  if (raw === 'disabled') return '停用'
  return labelAudit(raw) === raw ? raw : labelAudit(raw)
})

const tone = computed(() => {
  const v = String(props.value)
  const t = text.value
  if (['已完成', '已履约', '使用中', '营业中', '在职', '启用', '已通过', '正常', '是', '合格'].includes(t) || props.value === true) return 'ok'
  if (['待履约', '待确认', '已改约', '待审核', '已计算', 'pending'].includes(t) || v === 'pending') return 'warn'
  if (['已取消', '已失效', '失效', '已退次', '已拒绝', '停用', '下架', '否', '未合格'].includes(t) || props.value === false) return 'bad'
  return 'neutral'
})
</script>

<template>
  <span class="tag" :class="tone">{{ text }}</span>
</template>

<style scoped>
.tag {
  display: inline-flex; align-items: center; height: 22px; padding: 0 8px; border-radius: 999px;
  font-size: 12px; border: 1px solid var(--lp-border); background: #f7f8fa; color: var(--lp-muted);
}
.ok { background: #eaf7f1; color: var(--lp-success); border-color: #cfead9; }
.warn { background: #fff6e8; color: var(--lp-warning); border-color: #f0dfc0; }
.bad { background: #fdeeee; color: var(--lp-danger); border-color: #f2d0d0; }
</style>
