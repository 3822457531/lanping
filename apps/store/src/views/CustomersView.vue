<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { showSuccessToast, showToast } from 'vant'
import http from '@/api/http'
import { useSession } from '@/stores/session'

const session = useSession()
const router = useRouter()
const rows = ref<any[]>([])
const q = ref('')
const loading = ref(false)
const show = ref(false)
const form = reactive({ name: '', phone: '', level: '银卡会员', source: '到店' })

async function load() {
  loading.value = true
  try {
    const data: any = await http.get('/api/customers', { params: { q: q.value, pageSize: 100 } })
    rows.value = data.list || []
  } catch (e: any) {
    showToast(e.message || '加载失败')
  } finally {
    loading.value = false
  }
}

async function save() {
  if (!form.name || !form.phone) return showToast('请填写姓名和手机号')
  if (!session.user?.store_id && session.user?.role !== 'admin') {
    return showToast('当前账号未绑定门店，无法新增')
  }
  try {
    await http.post('/api/customers', {
      name: form.name,
      phone: form.phone,
      mobile: form.phone,
      level_name: form.level,
      level: form.level,
      source: form.source,
      store_id: session.user?.store_id,
    })
    showSuccessToast('已新增客户')
    show.value = false
    Object.assign(form, { name: '', phone: '', level: '银卡会员', source: '到店' })
    await load()
  } catch (e: any) {
    showToast(e.message || '新增失败')
  }
}

function levelTone(level: string) {
  if (level?.includes('钻石')) return 'gold'
  if (level?.includes('金')) return 'warm'
  return ''
}

function assetLine(c: any) {
  const parts: string[] = []
  if (Number(c.balance) > 0) parts.push(`余额¥${Number(c.balance).toFixed(0)}`)
  if (Number(c.remain) > 0) parts.push(`次数${c.remain}`)
  if (c.cards_brief) return c.cards_brief
  return parts.length ? parts.join(' · ') : '暂无开卡'
}

onMounted(load)
</script>

<template>
  <div class="page-wrap">
    <van-nav-bar title="客户">
      <template #right>
        <span class="nav-link" @click="show=true">新增</span>
      </template>
    </van-nav-bar>
    <van-search v-model="q" placeholder="搜索姓名 / 手机" shape="round" background="transparent" @search="load" />

    <div class="toolbar">
      <van-button round type="primary" size="small" @click="show=true">新增客户</van-button>
      <van-button round plain type="primary" size="small" @click="router.push('/cards')">开卡充值</van-button>
    </div>

    <div class="list">
      <div v-for="c in rows" :key="c.id" class="row">
        <div class="avatar">{{ (c.name || '客').slice(0,1) }}</div>
        <div class="left">
          <div class="name-row">
            <div class="name">{{ c.name }}</div>
            <span class="lv" :class="levelTone(c.level)">{{ c.level || '会员' }}</span>
          </div>
          <div class="sub">{{ c.phone }} · {{ c.source || '到店' }}</div>
          <div class="asset">{{ assetLine(c) }}</div>
        </div>
        <div class="right">
          <div class="remain">余额 ¥{{ Number(c.balance||0).toFixed(0) }}</div>
          <div class="times">次数 {{ c.remain || 0 }}</div>
          <van-button size="mini" round plain type="primary" @click="router.push('/cards')">开卡</van-button>
        </div>
      </div>
      <van-empty v-if="!loading && !rows.length" description="暂无客户，点击上方新增" />
    </div>

    <van-popup
      v-model:show="show"
      class="lp-center-popup"
      position="center"
      round
      teleport="body"
      closeable
      destroy-on-close
      :z-index="3000"
      :lock-scroll="true"
    >
      <div class="lp-popup-body">
        <div class="pop-title">新增本店客户</div>
        <van-field v-model="form.name" label="姓名" placeholder="客户姓名" />
        <van-field v-model="form.phone" type="tel" maxlength="11" label="手机号" placeholder="11位手机号" />
        <van-field v-model="form.level" label="等级" readonly is-link @click="form.level = form.level==='银卡会员'?'金卡会员':form.level==='金卡会员'?'钻石会员':'银卡会员'" />
        <van-button block round type="primary" style="margin-top:14px" @click="save">保存</van-button>
      </div>
    </van-popup>
  </div>
</template>

<style scoped>
.toolbar { display:flex; gap:10px; padding: 4px 16px 10px; }
.list { padding: 0 14px 20px; display:grid; gap:10px; }
.row {
  padding: 10px;
  display:flex;
  align-items:center;
  gap:12px;
  border-radius: 14px;
  background: rgba(255,255,255,.95);
  box-shadow: var(--lp-shadow-card);
  border: 1px solid rgba(255,245,248,.95);
}
.avatar {
  width: 42px; height: 42px; border-radius: 14px; flex-shrink: 0;
  display: grid; place-items: center;
  background: linear-gradient(145deg, #ff8fa6, #e24b6e);
  color: #fff; font-weight: 740; font-size: 14px;
  box-shadow: 0 6px 14px rgba(226,75,110,.22);
}
.left { flex: 1; min-width: 0; }
.name-row { display:flex; align-items:center; gap:8px; }
.name { font-weight: 720; font-size: 13px; }
.lv {
  font-size: 10px; font-weight: 650; padding: 2px 7px; border-radius: 999px;
  background: #f3ecef; color: #8a7d82;
}
.lv.warm { background: #fff3e6; color: #c9822d; }
.lv.gold { background: #fff6df; color: #b8892c; }
.sub { margin-top: 4px; font-size: 12px; color: var(--lp-muted); }
.asset { margin-top: 4px; font-size: 11px; color: #9a8b90; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.right { text-align: right; display:grid; gap:4px; justify-items:end; }
.remain { font-family: var(--lp-display); font-size: 12px; color: var(--lp-accent); font-weight: 700; }
.times { font-size: 11px; color: var(--lp-muted); }
.pop-title { font-size: 14px; font-weight: 740; text-align: center; margin-bottom: 8px; }
</style>
