<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showConfirmDialog, showSuccessToast, showToast } from 'vant'
import http from '@/api/http'
import { useSession } from '@/stores/session'

const session = useSession()
const route = useRoute()
const router = useRouter()

const tab = ref<'stock' | 'orders' | 'product'>('stock')
const stocks = ref<any[]>([])
const products = ref<any[]>([])
const orders = ref<any[]>([])
const pending = ref<any[]>([])
const stores = ref<any[]>([])
const loading = ref(false)

const showOrder = ref(false)
const showProduct = ref(false)
const editingId = ref<number | null>(null)

const orderForm = reactive({
  item_id: undefined as number | undefined,
  to_store_id: undefined as number | undefined,
  qty: 1,
  direction: '出库',
  channel: '门店调度',
})

const productForm = reactive({
  name: '',
  sku: '',
  unit: '件',
  qty: 0,
})

const storeID = computed(() => session.user?.store_id)
const otherStores = computed(() => stores.value.filter((s) => s.id !== storeID.value))
const productOptions = computed(() => products.value)

async function load() {
  loading.value = true
  try {
    const tasks: Promise<any>[] = [
      http.get('/api/inventory/stock', { params: { store_id: storeID.value } }),
      http.get('/api/inventory/items'),
      http.get('/api/inventory/orders', { params: { store_id: storeID.value } }),
      http.get('/api/stores'),
    ]
    if (session.user?.role === 'manager' || session.user?.role === 'admin') {
      tasks.push(http.get('/api/inventory/transfers/pending'))
    }
    const [stockData, itemData, orderData, storeData, pendingData]: any[] = await Promise.all(tasks)
    stocks.value = stockData.list || []
    products.value = itemData.list || []
    orders.value = orderData.list || []
    stores.value = storeData.list || storeData || []
    pending.value = pendingData?.list || []
    if (!orderForm.item_id && products.value.length) {
      orderForm.item_id = products.value[0].id
    }
  } catch (e: any) {
    showToast(e.message || '加载失败')
  } finally {
    loading.value = false
  }
}

function openOrder() {
  if (!products.value.length) {
    showToast('请先新增产品')
    tab.value = 'product'
    return
  }
  if (!otherStores.value.length) {
    showToast('暂无其他门店可调度')
    return
  }
  orderForm.item_id = products.value[0].id
  orderForm.to_store_id = otherStores.value[0]?.id
  orderForm.qty = 1
  orderForm.direction = '出库'
  orderForm.channel = '门店调度'
  showOrder.value = true
}

function openCreateProduct() {
  editingId.value = null
  productForm.name = ''
  productForm.sku = ''
  productForm.unit = '瓶'
  productForm.qty = 0
  showProduct.value = true
}

function openEditProduct(p: any) {
  editingId.value = p.id
  productForm.name = p.name
  productForm.sku = p.sku
  productForm.unit = p.unit || '件'
  const stock = stocks.value.find((s) => s.item_id === p.id)
  productForm.qty = stock?.qty ?? 0
  showProduct.value = true
}

async function saveOrder() {
  if (!orderForm.item_id) return showToast('请选择产品')
  if (!orderForm.to_store_id) return showToast('请选择调入门店')
  if (!(orderForm.qty > 0)) return showToast('数量需大于 0')
  try {
    await http.post('/api/inventory/orders', {
      item_id: orderForm.item_id,
      qty: orderForm.qty,
      direction: '出库',
      channel: '门店调度',
      store_id: storeID.value,
      to_store_id: orderForm.to_store_id,
    })
    showSuccessToast('已出库，等待对方店长确认收货')
    showOrder.value = false
    tab.value = 'orders'
    await load()
  } catch (e: any) {
    showToast(e.message || '提交失败')
  }
}

async function confirmTransfer(row: any) {
  try {
    await showConfirmDialog({
      title: '确认收货入库',
      message: `确认收到「${row.item_name}」×${row.qty}${row.unit || ''}？\n来自 ${row.from_store_name || '调出门店'}`,
    })
    await http.post(`/api/inventory/orders/${row.id}/confirm`)
    showSuccessToast('已入库')
    await load()
  } catch (e: any) {
    if (e !== 'cancel' && e?.message) showToast(e.message)
  }
}

async function saveProduct() {
  if (!productForm.name.trim()) return showToast('请填写产品名称')
  if (!productForm.unit.trim()) return showToast('请填写单位')
  try {
    const payload = {
      name: productForm.name.trim(),
      sku: productForm.sku.trim(),
      unit: productForm.unit.trim(),
      store_id: storeID.value,
      qty: Number(productForm.qty) || 0,
    }
    if (editingId.value) {
      await http.put(`/api/inventory/items/${editingId.value}`, payload)
      showSuccessToast('已更新产品')
    } else {
      await http.post('/api/inventory/items', payload)
      showSuccessToast('已新增产品')
    }
    showProduct.value = false
    await load()
  } catch (e: any) {
    showToast(e.message || '保存失败')
  }
}

async function removeProduct(p: any) {
  try {
    await showConfirmDialog({ title: '删除产品', message: `确认删除「${p.name}」？有库存时不可删。` })
    await http.delete(`/api/inventory/items/${p.id}`)
    showSuccessToast('已删除')
    await load()
  } catch (e: any) {
    if (e !== 'cancel' && e?.message) showToast(e.message)
  }
}

function stockOf(itemId: number) {
  return stocks.value.find((s) => s.item_id === itemId)
}

function statusClass(st: string) {
  if (st === '已完成') return 'ok'
  if (st === '待收货') return 'wait'
  return ''
}

function initial(name: string) {
  return (name || '品').slice(0, 1)
}

/** 从首页待办进入：只切到调度单列表，不自动弹确认框 */
function bootstrapFromQuery() {
  const tabQ = String(route.query.tab || '')
  const hasConfirmHint = Boolean(route.query.confirm_id)
  if (tabQ === 'orders' || hasConfirmHint) {
    tab.value = 'orders'
  }
  if (tabQ || hasConfirmHint) {
    router.replace({ path: '/inventory' })
  }
}

onMounted(async () => {
  bootstrapFromQuery()
  await load()
})
</script>

<template>
  <div class="page-wrap inv-page">
    <van-nav-bar title="出入库管理">
      <template #right>
        <span class="nav-link" @click="tab==='product' ? openCreateProduct() : openOrder()">
          {{ tab === 'product' ? '新增产品' : '调度出库' }}
        </span>
      </template>
    </van-nav-bar>

    <div v-if="pending.length" class="pending-banner" @click="tab='orders'">
      <van-icon name="logistics" />
      <div class="pb-body">
        <div class="pb-t">{{ pending.length }} 笔待确认入库</div>
        <div class="pb-d">其他门店已出库调拨到本店，请确认收货</div>
      </div>
      <van-icon name="arrow" />
    </div>

    <div class="tabs">
      <button type="button" :class="{ on: tab==='stock' }" @click="tab='stock'">本店库存</button>
      <button type="button" :class="{ on: tab==='orders' }" @click="tab='orders'">
        调度单<em v-if="pending.length">{{ pending.length }}</em>
      </button>
      <button type="button" :class="{ on: tab==='product' }" @click="tab='product'">产品</button>
    </div>
    <p class="tip">门店调度仅出库；对方店长在待办确认后完成本店入库。</p>

    <div v-if="tab==='stock'" class="list">
      <div v-for="r in stocks" :key="r.id" class="row stock-card">
        <div class="avatar">{{ initial(r.name) }}</div>
        <div class="body">
          <div class="top">
            <div class="name">{{ r.name }}</div>
            <div class="qty">{{ r.qty }} <small>{{ r.unit }}</small></div>
          </div>
          <div class="sub">{{ r.sku || '无SKU' }} · {{ r.store_name || '本店' }}</div>
          <div class="meter">
            <div class="meter-fill" :style="{ width: Math.min(100, Number(r.qty||0) * 4) + '%' }" />
          </div>
        </div>
      </div>
      <van-empty v-if="!loading && !stocks.length" description="本店暂无库存，请先新增产品" />
    </div>

    <div v-else-if="tab==='orders'" class="list">
      <div v-if="pending.length" class="sec-label">待我确认</div>
      <div v-for="p in pending" :key="'p'+p.id" class="row order-card pending">
        <div class="top">
          <div>
            <div class="name">{{ p.item_name }}</div>
            <div class="sub">来自 {{ p.from_store_name }} · {{ p.created_at }}</div>
          </div>
          <div class="qty">+{{ p.qty }} {{ p.unit }}</div>
        </div>
        <div class="ops">
          <span class="badge wait">待收货</span>
          <van-button size="mini" round type="primary" @click="confirmTransfer(p)">确认入库</van-button>
        </div>
      </div>

      <div class="sec-label">全部单据</div>
      <div v-for="o in orders" :key="o.id" class="row order-card">
        <div class="top">
          <div>
            <div class="name">{{ o.item_name }}</div>
            <div class="sub">
              {{ o.from_store_name || '本店' }}
              <template v-if="o.to_store_name"> → {{ o.to_store_name }}</template>
              · {{ o.created_at }}
            </div>
          </div>
          <div class="qty" :class="{ out: o.direction==='出库' && o.store_id===storeID }">
            {{ o.direction==='出库' && o.store_id===storeID ? '-' : '+' }}{{ o.qty }} {{ o.unit }}
          </div>
        </div>
        <div class="ops">
          <span class="badge" :class="statusClass(o.status)">{{ o.status }}</span>
          <span class="ch">{{ o.channel }} · {{ o.direction }}</span>
          <van-button
            v-if="o.status==='待收货' && o.to_store_id===storeID"
            size="mini"
            round
            type="primary"
            @click="confirmTransfer(o)"
          >确认入库</van-button>
        </div>
      </div>
      <van-empty v-if="!loading && !orders.length && !pending.length" description="暂无调度单" />
    </div>

    <div v-else class="list">
      <div v-for="p in products" :key="p.id" class="row product">
        <div class="top">
          <div>
            <div class="name">{{ p.name }}</div>
            <div class="sub">{{ p.sku }} · 单位 {{ p.unit }} · 全店合计 {{ p.total_qty }}</div>
            <div class="sub">本店库存 {{ stockOf(p.id)?.qty ?? 0 }} {{ p.unit }}</div>
          </div>
        </div>
        <div class="ops">
          <van-button size="mini" round plain type="primary" @click="openEditProduct(p)">编辑</van-button>
          <van-button size="mini" round plain type="danger" @click="removeProduct(p)">删除</van-button>
        </div>
      </div>
      <van-empty v-if="!loading && !products.length" description="暂无产品，点击右上角新增" />
    </div>

    <!-- 调度出库：居中弹窗 -->
    <van-popup
      v-model:show="showOrder"
      class="lp-center-popup"
      round
      position="center"
      teleport="body"
      closeable
      destroy-on-close
      :z-index="3000"
    >
      <div class="lp-popup-body">
        <div class="pop-title">调度出库</div>
        <p class="pop-hint">本店扣减库存后，由对方店长确认收货入库</p>
        <van-field label="产品">
          <template #input>
            <select v-model.number="orderForm.item_id" class="sel">
              <option v-for="i in productOptions" :key="i.id" :value="i.id">{{ i.name }}（{{ i.unit }}）</option>
            </select>
          </template>
        </van-field>
        <van-field label="调入门店">
          <template #input>
            <select v-model.number="orderForm.to_store_id" class="sel">
              <option v-for="s in otherStores" :key="s.id" :value="s.id">{{ s.name }}</option>
            </select>
          </template>
        </van-field>
        <van-field v-model.number="orderForm.qty" type="digit" label="数量" />
        <van-field label="类型" readonly model-value="出库 · 门店调度" />
        <van-button block round type="primary" style="margin-top:14px" @click="saveOrder">确认出库</van-button>
      </div>
    </van-popup>

    <van-popup
      v-model:show="showProduct"
      class="lp-center-popup"
      round
      position="center"
      teleport="body"
      closeable
      destroy-on-close
      :z-index="3000"
    >
      <div class="lp-popup-body">
        <div class="pop-title">{{ editingId ? '编辑产品' : '新增产品' }}</div>
        <van-field v-model="productForm.name" label="名称" placeholder="如：玻尿酸补水面膜" />
        <van-field v-model="productForm.sku" label="SKU" placeholder="可空，自动生成" />
        <van-field label="单位">
          <template #input>
            <select v-model="productForm.unit" class="sel">
              <option>件</option>
              <option>瓶</option>
              <option>片</option>
              <option>盒</option>
              <option>支</option>
              <option>袋</option>
              <option>套</option>
            </select>
          </template>
        </van-field>
        <van-field v-model.number="productForm.qty" type="digit" :label="editingId ? '本店库存' : '初始库存'" placeholder="本店数量" />
        <div class="hint">门店：{{ session.user?.store_name || '本店' }}</div>
        <van-button block round type="primary" style="margin-top:12px" @click="saveProduct">保存</van-button>
      </div>
    </van-popup>
  </div>
</template>

<style scoped>
.inv-page { padding-bottom: 10px; }
.nav-link { color: var(--lp-accent); font-size: 14px; font-weight: 650; }
.pending-banner {
  margin: 10px 14px 0;
  padding: 12px 14px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  gap: 6px;
  background: linear-gradient(135deg, rgba(226,75,110,.12), rgba(255,255,255,.95));
  border: 1px solid rgba(226,75,110,.18);
  color: var(--lp-text);
}
.pb-body { flex: 1; min-width: 0; }
.pb-t { font-weight: 720; font-size: 14px; }
.pb-d { margin-top: 2px; font-size: 12px; color: var(--lp-muted); }

.tabs {
  margin: 10px 14px 0;
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  gap: 6px;
  padding: 4px;
  border-radius: 16px;
  background: rgba(255,255,255,.78);
  box-shadow: 0 4px 14px rgba(80,40,55,.04);
}
.tabs button {
  border: 0;
  border-radius: 12px;
  padding: 10px 0;
  background: transparent;
  color: var(--lp-muted);
  font-size: 13px;
  font-weight: 650;
  position: relative;
}
.tabs button.on {
  background: #fff;
  color: var(--lp-accent);
  box-shadow: 0 4px 12px rgba(196,88,112,.12);
  font-weight: 720;
}
.tabs em {
  position: absolute;
  top: 4px;
  right: 10px;
  min-width: 16px;
  height: 16px;
  padding: 0 4px;
  border-radius: 8px;
  background: var(--lp-accent);
  color: #fff;
  font-size: 10px;
  font-style: normal;
  line-height: 16px;
}
.tip { margin: 8px 0 0; padding: 0 16px; font-size: 12px; color: var(--lp-muted); line-height: 1.5; }
.sec-label { padding: 4px 2px; font-size: 12px; color: var(--lp-muted); font-weight: 650; }
.list { padding: 8px 12px 12px; display: grid; gap: 6px; }
.row {
  padding: 10px;
  border-radius: 14px;
  background: rgba(255,255,255,.95);
  box-shadow: var(--lp-shadow-card);
  border: 1px solid rgba(255,245,248,.95);
}
.stock-card { display: flex; gap: 8px; align-items: center; }
.avatar {
  width: 44px; height: 44px; border-radius: 14px; flex-shrink: 0;
  display: grid; place-items: center; color: #fff; font-weight: 740;
  background: linear-gradient(145deg, #ff8fa6, #e24b6e);
}
.body { flex: 1; min-width: 0; }
.top { display: flex; justify-content: space-between; gap: 6px; align-items: flex-start; }
.name { font-weight: 720; font-size: 13px; }
.sub { margin-top: 5px; font-size: 12px; color: var(--lp-muted); }
.qty { font-family: var(--lp-display); color: var(--lp-accent); font-weight: 700; white-space: nowrap; }
.qty small { font-size: 12px; font-weight: 600; }
.qty.out { color: #c45c5c; }
.meter {
  margin-top: 10px; height: 6px; border-radius: 999px; background: #f3e8ec; overflow: hidden;
}
.meter-fill { height: 100%; border-radius: inherit; background: linear-gradient(90deg, #ff8fa6, #e24b6e); }
.ops { margin-top: 10px; display: flex; gap: 8px; align-items: center; justify-content: flex-end; flex-wrap: wrap; }
.badge {
  font-size: 11px; padding: 2px 8px; border-radius: 999px; background: #f0f0f0; color: #666;
}
.badge.ok { background: rgba(56,168,120,.12); color: #2d9a6a; }
.badge.wait { background: rgba(230,150,60,.14); color: #c47a20; }
.ch { font-size: 11px; color: var(--lp-muted); margin-right: auto; }
.order-card.pending { border-color: rgba(226,75,110,.22); }
.hint { margin: 8px 4px 0; font-size: 12px; color: var(--lp-muted); }
.pop-title { font-size: 14px; font-weight: 740; text-align: center; }
.pop-hint { margin: 6px 0 12px; text-align: center; font-size: 12px; color: var(--lp-muted); line-height: 1.5; }
.sel {
  width: 100%; border: 0; background: transparent; outline: none;
  font-size: 14px; color: inherit; appearance: none;
}
</style>
