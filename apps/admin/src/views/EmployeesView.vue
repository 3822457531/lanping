<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import http from '@/api/http'
import PageHeader from '@/components/PageHeader.vue'
import StatusTag from '@/components/StatusTag.vue'

const rows = ref<any[]>([])
const stores = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const loading = ref(false)
const show = ref(false)
const saving = ref(false)

const form = reactive({
  name: '',
  phone: '',
  password: '123456',
  role: 'staff',
  store_id: undefined as number | undefined,
  position: '',
})

async function loadStores() {
  const data: any = await http.get('/api/stores')
  stores.value = data.list || []
}

async function load() {
  loading.value = true
  try {
    const data: any = await http.get('/api/users', { params: { page: page.value, pageSize: 50 } })
    rows.value = (data.list || []).filter((x: any) => x.role !== 'admin')
    total.value = data.total || 0
  } finally {
    loading.value = false
  }
}

function openCreate() {
  Object.assign(form, {
    name: '',
    phone: '',
    password: '123456',
    role: 'staff',
    store_id: stores.value[0]?.id,
    position: '',
  })
  show.value = true
}

async function save() {
  if (!form.name || !form.phone) return ElMessage.warning('请填写姓名和手机号')
  if (!form.store_id) return ElMessage.warning('请选择所属门店')
  if (!['manager', 'staff'].includes(form.role)) return ElMessage.warning('请选择角色')
  saving.value = true
  try {
    await http.post('/api/users', {
      name: form.name,
      phone: form.phone,
      password: form.password || '123456',
      role: form.role,
      store_id: form.store_id,
      position: form.position || (form.role === 'manager' ? '店长' : '美容师'),
    })
    ElMessage.success(form.role === 'manager' ? '店长已创建，并已同步为该门店负责人' : '员工已创建')
    show.value = false
    load()
  } catch (e: any) {
    ElMessage.error(e.message || '创建失败')
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  await loadStores()
  await load()
})
</script>

<template>
  <div>
    <PageHeader title="员工管理" subtitle="新增店长会同步到门店管理的负责人；门店侧指定负责人也会同步为店长">
      <el-button type="primary" @click="openCreate">新增员工</el-button>
    </PageHeader>
    <el-table v-loading="loading" :data="rows" stripe border>
      <el-table-column prop="name" label="姓名" width="110" />
      <el-table-column prop="phone" label="手机号" width="130" />
      <el-table-column label="角色" width="120">
        <template #default="{ row }"><StatusTag kind="role" :value="row.role" /></template>
      </el-table-column>
      <el-table-column prop="store_name" label="所属门店" />
      <el-table-column prop="position" label="岗位" width="120" />
      <el-table-column prop="base_salary" label="底薪" width="100">
        <template #default="{ row }">¥{{ row.base_salary }}</template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }"><StatusTag :value="row.status" /></template>
      </el-table-column>
    </el-table>
    <div class="pager"><el-pagination v-model:current-page="page" layout="total, prev, pager, next" :total="total" @current-change="load" /></div>

    <el-dialog v-model="show" title="新增员工 / 店长" width="480px" destroy-on-close>
      <el-form label-width="96px">
        <el-form-item label="姓名" required><el-input v-model="form.name" placeholder="员工姓名" /></el-form-item>
        <el-form-item label="手机号" required><el-input v-model="form.phone" maxlength="11" placeholder="登录手机号" /></el-form-item>
        <el-form-item label="初始密码"><el-input v-model="form.password" placeholder="默认 123456" /></el-form-item>
        <el-form-item label="角色" required>
          <el-radio-group v-model="form.role">
            <el-radio value="staff">员工</el-radio>
            <el-radio value="manager">店长</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="所属门店" required>
          <el-select v-model="form.store_id" placeholder="选择门店" style="width:100%">
            <el-option v-for="s in stores" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="岗位"><el-input v-model="form.position" placeholder="如：美容师 / 店长" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="show=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>
<style scoped>.pager{margin-top:14px;display:flex;justify-content:flex-end}</style>
