<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import http from '@/api/http'
import PageHeader from '@/components/PageHeader.vue'
import MapPicker from '@/components/MapPicker.vue'

const rows = ref<any[]>([])
const employees = ref<any[]>([])
const loading = ref(false)
const show = ref(false)
const mapOpen = ref(false)
const editingId = ref<number | null>(null)
const saving = ref(false)
const clockinRange = ref(200)

const form = reactive({
  name: '',
  address: '',
  phone: '',
  lat: undefined as number | undefined,
  lng: undefined as number | undefined,
  manager_id: undefined as number | undefined,
  status: '营业中',
})

const employeeOptions = computed(() =>
  employees.value.map((e) => {
    const isOtherManager =
      e.role === 'manager' &&
      !(editingId.value && e.store_id === editingId.value)
    return {
      ...e,
      disabled: isOtherManager,
      label: isOtherManager
        ? `${e.name}（${e.phone}）· 已是店长不可选`
        : `${e.name}（${e.phone}）`,
    }
  }),
)

const geoReady = computed(() => !!(form.lat && form.lng))

async function loadEmployees() {
  const data: any = await http.get('/api/users', { params: { page: 1, pageSize: 500 } })
  employees.value = (data.list || []).filter(
    (x: any) => x.role !== 'admin' && x.status === 'active',
  )
}

async function loadSettings() {
  try {
    const data: any = await http.get('/api/settings')
    const n = Number(data.clockin_range || 200)
    clockinRange.value = n > 0 ? n : 200
  } catch {
    clockinRange.value = 200
  }
}

async function load() {
  loading.value = true
  try {
    const data: any = await http.get('/api/stores')
    rows.value = data.list || []
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = null
  Object.assign(form, {
    name: '',
    address: '',
    phone: '',
    lat: undefined,
    lng: undefined,
    manager_id: undefined,
    status: '营业中',
  })
  show.value = true
}

function openEdit(row: any) {
  editingId.value = row.id
  Object.assign(form, {
    name: row.name || '',
    address: row.address || '',
    phone: row.phone || '',
    lat: row.lat || undefined,
    lng: row.lng || undefined,
    manager_id: row.manager_id || undefined,
    status: row.status || '营业中',
  })
  show.value = true
}

function onMapConfirm(p: { lat: number; lng: number; address?: string }) {
  form.lat = p.lat
  form.lng = p.lng
  if (p.address && !form.address.trim()) form.address = p.address
  ElMessage.success('已选取打卡中心点')
}

function clearGeo() {
  form.lat = undefined
  form.lng = undefined
}

async function save() {
  if (!form.name.trim()) return ElMessage.warning('请填写门店名称')
  if (!form.lat || !form.lng) {
    return ElMessage.warning('请先用地图选点配置打卡坐标，否则员工无法真实打卡')
  }
  saving.value = true
  try {
    const payload = {
      name: form.name.trim(),
      address: form.address,
      phone: form.phone,
      lat: Number(form.lat) || 0,
      lng: Number(form.lng) || 0,
      status: form.status,
      manager_id: form.manager_id || 0,
    }
    if (editingId.value) {
      await http.put(`/api/stores/${editingId.value}`, payload)
      ElMessage.success('门店已更新')
    } else {
      await http.post('/api/stores', payload)
      ElMessage.success('门店已创建')
    }
    show.value = false
    await Promise.all([load(), loadEmployees()])
  } catch (e: any) {
    ElMessage.error(e.message || '保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  await Promise.all([load(), loadEmployees(), loadSettings()])
})
</script>

<template>
  <div>
    <PageHeader title="门店管理" subtitle="地图选点设定打卡中心；员工须在系统设置的方圆 N 米内才能打卡">
      <el-button type="primary" @click="openCreate">新增门店</el-button>
    </PageHeader>

    <div class="banner">
      当前全局打卡范围：<b>{{ clockinRange }} 米</b>
      （可在「系统设置」调节）。未配置坐标的门店，员工端无法打卡。
      <router-link class="link" to="/settings">去调节范围 ›</router-link>
    </div>

    <el-table v-loading="loading" :data="rows" stripe border>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="name" label="门店名称" min-width="140" />
      <el-table-column prop="address" label="地址" min-width="180" />
      <el-table-column label="打卡坐标" width="200">
        <template #default="{ row }">
          <span v-if="row.lat || row.lng" class="ok">{{ Number(row.lat).toFixed(5) }}, {{ Number(row.lng).toFixed(5) }}</span>
          <el-tag v-else type="danger" size="small">未配置 · 无法打卡</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="phone" label="电话" width="130" />
      <el-table-column prop="manager_name" label="负责人" width="110" />
      <el-table-column prop="status" label="状态" width="100" />
      <el-table-column label="操作" width="100" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="show" :title="editingId ? '编辑门店' : '新增门店'" width="560px" destroy-on-close>
      <el-form label-width="96px">
        <el-form-item label="门店名称" required><el-input v-model="form.name" placeholder="如：万象城店" /></el-form-item>
        <el-form-item label="地址"><el-input v-model="form.address" placeholder="门店地址（可地图搜索带入）" /></el-form-item>
        <el-form-item label="打卡位置" required>
          <div class="geo-box">
            <div v-if="geoReady" class="geo-ok">
              已选 {{ Number(form.lat).toFixed(6) }}, {{ Number(form.lng).toFixed(6) }}
              <span class="r">围栏 {{ clockinRange }} 米</span>
            </div>
            <div v-else class="geo-miss">尚未选点 · 员工将无法打卡</div>
            <div class="geo-actions">
              <el-button type="primary" @click="mapOpen = true">打开地图选点</el-button>
              <el-button v-if="geoReady" @click="clearGeo">清除</el-button>
            </div>
          </div>
        </el-form-item>
        <el-form-item label="电话"><el-input v-model="form.phone" placeholder="联系电话" /></el-form-item>
        <el-form-item label="负责人">
          <el-select
            v-model="form.manager_id"
            filterable
            clearable
            style="width:100%"
            placeholder="从员工中选择负责人"
            teleported
          >
            <el-option
              v-for="e in employeeOptions"
              :key="e.id"
              :label="e.label"
              :value="e.id"
              :disabled="e.disabled"
            />
          </el-select>
 
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status" style="width:100%">
            <el-option label="营业中" value="营业中" />
            <el-option label="装修中" value="装修中" />
            <el-option label="已停业" value="已停业" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="show=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>

    <MapPicker
      v-model="mapOpen"
      :lat="form.lat || 0"
      :lng="form.lng || 0"
      :range-m="clockinRange"
      :address-hint="form.address"
      @confirm="onMapConfirm"
    />
  </div>
</template>

<style scoped>
.banner {
  margin-bottom: 14px;
  padding: 10px 14px;
  border-radius: 10px;
  background: #fff7f9;
  border: 1px solid rgba(226, 75, 110, 0.18);
  font-size: 13px;
  color: #667085;
  line-height: 1.5;
}
.banner b { color: #e24b6e; }
.banner .link {
  margin-left: 8px;
  color: var(--el-color-primary);
}
.tip {
  margin-top: 6px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  line-height: 1.4;
}
.ok { color: #039855; font-size: 12px; }
.geo-box {
  width: 100%;
  padding: 12px;
  border: 1px dashed var(--el-border-color);
  border-radius: 10px;
  background: #fafafa;
}
.geo-ok { font-size: 13px; color: #039855; font-weight: 600; }
.geo-ok .r {
  margin-left: 8px;
  font-weight: 500;
  color: #667085;
}
.geo-miss { font-size: 13px; color: #d92d20; }
.geo-actions { margin-top: 10px; display: flex; gap: 8px; }
</style>
