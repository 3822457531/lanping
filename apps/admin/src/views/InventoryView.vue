<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import http from '@/api/http'
import PageHeader from '@/components/PageHeader.vue'

const tab = ref<'stock' | 'orders' | 'product'>('stock')
const stocks = ref<any[]>([])
const products = ref<any[]>([])
const orders = ref<any[]>([])
const stores = ref<any[]>([])
const loading = ref(false)

const orderDialog = ref(false)
const productDialog = ref(false)
const editingId = ref<number | null>(null)
const saving = ref(false)

const filter = reactive({
  store_id: undefined as number | undefined,
  channel: '' as string,
  status: '' as string,
})

const orderForm = reactive({
  store_id: undefined as number | undefined,
  to_store_id: undefined as number | undefined,
  item_id: undefined as number | undefined,
  qty: 10,
  direction: '入库',
  channel: '系统分发',
})

const productForm = reactive({
  name: '',
  sku: '',
  unit: '件',
  store_id: undefined as number | undefined,
  qty: 0,
})

const filteredOrders = computed(() => {
  return orders.value.filter((o) => {
    if (filter.store_id) {
      const sid = filter.store_id
      if (o.store_id !== sid && o.to_store_id !== sid) return false
    }
    if (filter.channel && o.channel !== filter.channel) return false
    if (filter.status && o.status !== filter.status) return false
    return true
  })
})

const otherStores = computed(() =>
  stores.value.filter((s) => s.id !== orderForm.store_id),
)

async function load() {
  loading.value = true
  try {
    const [stock, items, storeData, orderData]: any[] = await Promise.all([
      http.get('/api/inventory/stock'),
      http.get('/api/inventory/items'),
      http.get('/api/stores'),
      http.get('/api/inventory/orders'),
    ])
    stocks.value = stock.list || []
    products.value = items.list || []
    stores.value = storeData.list || []
    orders.value = orderData.list || []
    if (!orderForm.store_id) orderForm.store_id = stores.value[0]?.id
    if (!orderForm.item_id) orderForm.item_id = products.value[0]?.id
    if (!productForm.store_id) productForm.store_id = stores.value[0]?.id
  } finally {
    loading.value = false
  }
}

function openOrder() {
  if (!products.value.length) {
    ElMessage.warning('请先新增产品')
    tab.value = 'product'
    return
  }
  orderForm.store_id = stores.value[0]?.id
  orderForm.to_store_id = undefined
  orderForm.item_id = products.value[0]?.id
  orderForm.qty = 10
  orderForm.direction = '入库'
  orderForm.channel = '系统分发'
  orderDialog.value = true
}

function onChannelChange() {
  if (orderForm.channel === '门店调度') {
    orderForm.direction = '出库'
    orderForm.to_store_id = otherStores.value[0]?.id
  } else {
    orderForm.to_store_id = undefined
    if (!orderForm.direction) orderForm.direction = '入库'
  }
}

function openCreateProduct() {
  editingId.value = null
  productForm.name = ''
  productForm.sku = ''
  productForm.unit = '瓶'
  productForm.store_id = stores.value[0]?.id
  productForm.qty = 0
  productDialog.value = true
}

function openEditProduct(row: any) {
  editingId.value = row.id
  productForm.name = row.name
  productForm.sku = row.sku
  productForm.unit = row.unit
  productForm.store_id = stores.value[0]?.id
  const stock = stocks.value.find((s) => s.item_id === row.id && s.store_id === productForm.store_id)
  productForm.qty = stock?.qty ?? 0
  productDialog.value = true
}

function onStoreChange() {
  if (!editingId.value) return
  const stock = stocks.value.find((s) => s.item_id === editingId.value && s.store_id === productForm.store_id)
  productForm.qty = stock?.qty ?? 0
}

async function saveOrder() {
  if (!orderForm.store_id || !orderForm.item_id) return ElMessage.warning('请选择门店和产品')
  if (orderForm.channel === '门店调度' && !orderForm.to_store_id) {
    return ElMessage.warning('请选择调入门店')
  }
  saving.value = true
  try {
    await http.post('/api/inventory/orders', { ...orderForm })
    ElMessage.success(orderForm.channel === '门店调度' ? '已出库，等待对方确认收货' : '已记账')
    orderDialog.value = false
    tab.value = 'orders'
    await load()
  } catch (e: any) {
    ElMessage.error(e.message || '提交失败')
  } finally {
    saving.value = false
  }
}

async function saveProduct() {
  if (!productForm.name.trim()) return ElMessage.warning('请填写产品名称')
  saving.value = true
  try {
    const payload = {
      name: productForm.name.trim(),
      sku: productForm.sku.trim(),
      unit: productForm.unit,
      store_id: productForm.store_id || 0,
      qty: Number(productForm.qty) || 0,
    }
    if (editingId.value) {
      await http.put(`/api/inventory/items/${editingId.value}`, payload)
      ElMessage.success('已更新产品')
    } else {
      await http.post('/api/inventory/items', payload)
      ElMessage.success('已新增产品')
    }
    productDialog.value = false
    await load()
  } catch (e: any) {
    ElMessage.error(e.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function removeProduct(row: any) {
  await ElMessageBox.confirm(`确认删除「${row.name}」？有库存时不可删。`, '删除产品')
  await http.delete(`/api/inventory/items/${row.id}`)
  ElMessage.success('已删除')
  load()
}

async function confirmTransfer(row: any) {
  await ElMessageBox.confirm(
    `确认「${row.item_name}」×${row.qty}${row.unit || ''} 已由 ${row.from_store_name || '调出门店'} 调入？`,
    '确认收货入库',
  )
  try {
    await http.post(`/api/inventory/orders/${row.id}/confirm`)
    ElMessage.success('已确认入库')
    await load()
  } catch (e: any) {
    ElMessage.error(e.message || '确认失败')
  }
}

function statusType(st: string) {
  if (st === '已完成') return 'success'
  if (st === '待收货') return 'warning'
  return 'info'
}

onMounted(load)
</script>

<template>
  <div>
    <PageHeader title="出入库 / 产品管理" subtitle="查看全部门店库存、调度与出入库记录（系统分发 / 系统初始化 / 门店调度）">
      <el-button @click="openCreateProduct">新增产品</el-button>
      <el-button type="primary" @click="openOrder">新建出入库</el-button>
    </PageHeader>

    <el-radio-group v-model="tab" style="margin-bottom:14px">
      <el-radio-button value="stock">门店库存</el-radio-button>
      <el-radio-button value="orders">出入库记录</el-radio-button>
      <el-radio-button value="product">产品目录</el-radio-button>
    </el-radio-group>

    <el-table v-if="tab==='stock'" v-loading="loading" :data="stocks" stripe border>
      <el-table-column prop="store_name" label="门店" width="140" />
      <el-table-column prop="name" label="产品" />
      <el-table-column prop="sku" label="SKU" width="120" />
      <el-table-column prop="qty" label="库存数量" width="110" />
      <el-table-column prop="unit" label="单位" width="80" />
    </el-table>

    <div v-else-if="tab==='orders'">
      <div class="filters">
        <el-select v-model="filter.store_id" clearable filterable placeholder="全部门店" style="width:160px">
          <el-option v-for="s in stores" :key="s.id" :label="s.name" :value="s.id" />
        </el-select>
        <el-select v-model="filter.channel" clearable placeholder="全部渠道" style="width:140px">
          <el-option label="系统分发" value="系统分发" />
          <el-option label="系统初始化" value="系统初始化" />
          <el-option label="门店调度" value="门店调度" />
        </el-select>
        <el-select v-model="filter.status" clearable placeholder="全部状态" style="width:120px">
          <el-option label="已完成" value="已完成" />
          <el-option label="待收货" value="待收货" />
        </el-select>
        <el-button @click="load">刷新</el-button>
      </div>
      <el-table v-loading="loading" :data="filteredOrders" stripe border>
        <el-table-column prop="created_at" label="时间" width="170" />
        <el-table-column prop="order_no" label="单号" width="140" />
        <el-table-column prop="item_name" label="产品" min-width="140" />
        <el-table-column label="门店" min-width="180">
          <template #default="{ row }">
            <span>{{ row.from_store_name || '—' }}</span>
            <template v-if="row.to_store_name">
              <span class="arrow"> → </span>
              <span>{{ row.to_store_name }}</span>
            </template>
          </template>
        </el-table-column>
        <el-table-column prop="direction" label="方向" width="80" />
        <el-table-column prop="qty" label="数量" width="90">
          <template #default="{ row }">{{ row.qty }} {{ row.unit }}</template>
        </el-table-column>
        <el-table-column prop="channel" label="渠道" width="110" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="110" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.status === '待收货'"
              link
              type="primary"
              @click="confirmTransfer(row)"
            >确认入库</el-button>
            <span v-else class="muted">—</span>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-table v-else v-loading="loading" :data="products" stripe border>
      <el-table-column prop="name" label="产品名称" />
      <el-table-column prop="sku" label="SKU" width="140" />
      <el-table-column prop="unit" label="单位" width="80" />
      <el-table-column prop="total_qty" label="全店合计" width="110" />
      <el-table-column prop="store_count" label="覆盖门店" width="100" />
      <el-table-column label="操作" width="160">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEditProduct(row)">编辑</el-button>
          <el-button link type="danger" @click="removeProduct(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="orderDialog" title="出入库" width="480px" destroy-on-close>
      <el-form label-width="96px">
        <el-form-item label="渠道">
          <el-select v-model="orderForm.channel" style="width:100%" @change="onChannelChange">
            <el-option label="系统分发" value="系统分发" />
            <el-option label="系统初始化" value="系统初始化" />
            <el-option label="门店调度" value="门店调度" />
          </el-select>
        </el-form-item>
        <el-form-item :label="orderForm.channel === '门店调度' ? '调出门店' : '门店'">
          <el-select v-model="orderForm.store_id" filterable style="width:100%" @change="onChannelChange">
            <el-option v-for="s in stores" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="orderForm.channel === '门店调度'" label="调入门店">
          <el-select v-model="orderForm.to_store_id" filterable style="width:100%" placeholder="选择目标门店">
            <el-option v-for="s in otherStores" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="产品">
          <el-select v-model="orderForm.item_id" filterable style="width:100%">
            <el-option v-for="i in products" :key="i.id" :label="`${i.name}（${i.unit}）`" :value="i.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="数量"><el-input-number v-model="orderForm.qty" :min="1" /></el-form-item>
        <el-form-item v-if="orderForm.channel !== '门店调度'" label="方向">
          <el-select v-model="orderForm.direction" style="width:100%">
            <el-option label="入库" value="入库" />
            <el-option label="出库" value="出库" />
          </el-select>
        </el-form-item>
        <el-form-item v-else label="方向">
          <el-tag type="warning">出库（待对方确认入库）</el-tag>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="orderDialog=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveOrder">提交</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="productDialog" :title="editingId ? '编辑产品' : '新增产品'" width="520px" destroy-on-close>
      <el-form label-width="100px">
        <el-form-item label="产品名称"><el-input v-model="productForm.name" placeholder="如：玻尿酸补水面膜" /></el-form-item>
        <el-form-item label="SKU"><el-input v-model="productForm.sku" placeholder="可空，自动生成" /></el-form-item>
        <el-form-item label="单位">
          <el-select v-model="productForm.unit" style="width:100%">
            <el-option v-for="u in ['件','瓶','片','盒','支','袋','套']" :key="u" :label="u" :value="u" />
          </el-select>
        </el-form-item>
        <el-form-item label="对应门店">
          <el-select v-model="productForm.store_id" filterable style="width:100%" @change="onStoreChange">
            <el-option v-for="s in stores" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item :label="editingId ? '该店库存' : '初始库存'">
          <el-input-number v-model="productForm.qty" :min="0" />
          <div class="tip">初始/调整库存会记入「系统初始化」出入库记录</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="productDialog=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveProduct">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.filters {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 14px;
  align-items: center;
}
.arrow { color: var(--el-text-color-secondary); }
.muted { color: var(--el-text-color-placeholder); }
.tip {
  margin-top: 6px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  line-height: 1.4;
}
</style>
