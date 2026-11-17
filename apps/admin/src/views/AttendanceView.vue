<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import http from '@/api/http'
import PageHeader from '@/components/PageHeader.vue'
import StatusTag from '@/components/StatusTag.vue'

type DayRec = {
  date: string
  check_in?: string
  check_out?: string
  status?: string
  method?: string
}

type Emp = {
  user_id: number
  user_name: string
  role: string
  role_name: string
  store_id: number
  store_name: string
  present_days: number
  days: DayRec[]
}

type DayItem = { emp: Emp; rec: DayRec }

const mode = ref<'employee' | 'calendar'>('employee')
const month = ref(new Date().toISOString().slice(0, 7))
const storeId = ref<number | undefined>()
const stores = ref<any[]>([])
const employees = ref<Emp[]>([])
const daysInMonth = ref(31)
const loading = ref(false)

const keyword = ref('')
const roleFilter = ref('')
const presentFilter = ref('') // '' | 'present' | 'absent'
const page = ref(1)
const pageSize = 20

const dayDrawer = ref(false)
const dayFocus = ref('')
const dayKeyword = ref('')

const empDrawer = ref(false)
const empFocus = ref<Emp | null>(null)

const weekLabels = ['一', '二', '三', '四', '五', '六', '日']

/** 预计算：避免模板里反复扫全员 */
const empDayMaps = computed(() => {
  const map = new Map<number, Map<string, DayRec>>()
  for (const emp of employees.value) {
    const days = new Map<string, DayRec>()
    for (const d of emp.days || []) days.set(d.date, d)
    map.set(emp.user_id, days)
  }
  return map
})

const dateIndex = computed(() => {
  const map = new Map<string, DayItem[]>()
  for (const emp of employees.value) {
    for (const rec of emp.days || []) {
      const list = map.get(rec.date) || []
      list.push({ emp, rec })
      map.set(rec.date, list)
    }
  }
  return map
})

const filteredEmployees = computed(() => {
  const q = keyword.value.trim().toLowerCase()
  return employees.value.filter((e) => {
    if (roleFilter.value && e.role !== roleFilter.value) return false
    if (presentFilter.value === 'present' && !(e.present_days > 0)) return false
    if (presentFilter.value === 'absent' && e.present_days > 0) return false
    if (!q) return true
    return (
      e.user_name.toLowerCase().includes(q) ||
      (e.store_name || '').toLowerCase().includes(q) ||
      String(e.user_id).includes(q)
    )
  })
})

const pagedEmployees = computed(() => {
  const start = (page.value - 1) * pageSize
  return filteredEmployees.value.slice(start, start + pageSize)
})

const summary = computed(() => {
  const total = employees.value.length
  const present = employees.value.filter((e) => e.present_days > 0).length
  const absent = total - present
  return { total, present, absent }
})

const calendarCells = computed(() => {
  const [y, m] = month.value.split('-').map(Number)
  const first = new Date(y, m - 1, 1)
  let lead = first.getDay() - 1
  if (lead < 0) lead = 6
  const total = daysInMonth.value
  const cells: { key: string; day: number | null; date: string | null; count: number }[] = []
  for (let i = 0; i < lead; i++) cells.push({ key: `e${i}`, day: null, date: null, count: 0 })
  for (let d = 1; d <= total; d++) {
    const date = `${month.value}-${String(d).padStart(2, '0')}`
    cells.push({ key: date, day: d, date, count: dateIndex.value.get(date)?.length || 0 })
  }
  while (cells.length % 7) cells.push({ key: `t${cells.length}`, day: null, date: null, count: 0 })
  return cells
})

const dayItems = computed(() => {
  if (!dayFocus.value) return [] as DayItem[]
  const all = dateIndex.value.get(dayFocus.value) || []
  const q = dayKeyword.value.trim().toLowerCase()
  if (!q) return all
  return all.filter(
    (x) =>
      x.emp.user_name.toLowerCase().includes(q) ||
      (x.emp.store_name || '').toLowerCase().includes(q),
  )
})

const dayStats = computed(() => {
  const all = dateIndex.value.get(dayFocus.value) || []
  let ok = 0
  let partial = 0
  for (const x of all) {
    if (x.rec.check_in && x.rec.check_out) ok++
    else if (x.rec.check_in) partial++
  }
  return { total: all.length, ok, partial, none: Math.max(0, employees.value.length - all.length) }
})

function dayOf(emp: Emp, date: string) {
  return empDayMaps.value.get(emp.user_id)?.get(date)
}

function previewNames(date: string) {
  return (dateIndex.value.get(date) || []).slice(0, 2).map((x) => x.emp.user_name)
}

function dayStatusClass(rec?: DayRec) {
  if (!rec?.check_in) return 'off'
  if (!rec.check_out) return 'partial'
  return 'ok'
}

function fmtRec(rec?: DayRec) {
  if (!rec?.check_in) return '未打卡'
  return `${rec.check_in} — ${rec.check_out || '未下班'}`
}

function openDay(date: string) {
  dayFocus.value = date
  dayKeyword.value = ''
  dayDrawer.value = true
}

function openEmp(emp: Emp) {
  empFocus.value = emp
  empDrawer.value = true
}

function hoverDayItems(date: string) {
  return (dateIndex.value.get(date) || []).slice(0, 8)
}

async function loadStores() {
  const data: any = await http.get('/api/stores')
  stores.value = data.list || []
}

async function load() {
  loading.value = true
  try {
    const data: any = await http.get('/api/attendance', {
      params: { month: month.value, store_id: storeId.value || undefined },
    })
    employees.value = data.list || []
    daysInMonth.value = data.days_in_month || 31
    page.value = 1
  } finally {
    loading.value = false
  }
}

watch([month, storeId], load)
watch([keyword, roleFilter, presentFilter], () => {
  page.value = 1
})

onMounted(async () => {
  await loadStores()
  await load()
})
</script>

<template>
  <div>
    <PageHeader
      title="考勤管理"
      subtitle="人多时先筛门店/姓名；日历悬停看摘要，点击打开完整名单"
    />

    <div class="stats">
      <div class="stat"><span>在职人数</span><b>{{ summary.total }}</b></div>
      <div class="stat"><span>本月有出勤</span><b class="ok">{{ summary.present }}</b></div>
      <div class="stat"><span>本月未打卡</span><b class="warn">{{ summary.absent }}</b></div>
    </div>

    <div class="toolbar">
      <el-radio-group v-model="mode">
        <el-radio-button value="employee">员工维度</el-radio-button>
        <el-radio-button value="calendar">日历维度</el-radio-button>
      </el-radio-group>
      <el-date-picker
        v-model="month"
        type="month"
        value-format="YYYY-MM"
        placeholder="选择月份"
        style="width:150px"
      />
      <el-select v-model="storeId" clearable filterable placeholder="全部门店" style="width:150px">
        <el-option v-for="s in stores" :key="s.id" :label="s.name" :value="s.id" />
      </el-select>
      <el-input v-model="keyword" clearable placeholder="搜姓名 / 门店" style="width:180px" />
      <el-select v-model="roleFilter" clearable placeholder="角色" style="width:110px">
        <el-option label="店长" value="manager" />
        <el-option label="员工" value="staff" />
      </el-select>
      <el-select v-model="presentFilter" clearable placeholder="出勤" style="width:120px">
        <el-option label="本月有出勤" value="present" />
        <el-option label="本月未打卡" value="absent" />
      </el-select>
      <el-button @click="load">刷新</el-button>
    </div>

    <!-- 员工维度 -->
    <template v-if="mode==='employee'">
      <el-table v-loading="loading" :data="pagedEmployees" stripe border>
        <el-table-column label="员工" min-width="160">
          <template #default="{ row }">
            <el-popover
              placement="right-start"
              :width="360"
              trigger="hover"
              :show-after="160"
              :hide-after="100"
              popper-class="att-pop"
            >
              <template #reference>
                <button type="button" class="emp-hot" @click="openEmp(row)">
                  <span class="emp-name">{{ row.user_name }}</span>
                  <span class="hint">悬停预览 · 点击展开</span>
                </button>
              </template>
              <div class="pop-hd">
                <b>{{ row.user_name }}</b>
                <span>{{ month }} · 出勤 {{ row.present_days }} 天</span>
              </div>
              <div class="mini-cal">
                <div v-for="w in weekLabels" :key="w" class="wd">{{ w }}</div>
                <template v-for="c in calendarCells" :key="c.key">
                  <div v-if="!c.day" class="cell empty" />
                  <div
                    v-else
                    class="cell"
                    :class="dayStatusClass(dayOf(row, c.date!))"
                    :title="fmtRec(dayOf(row, c.date!))"
                  >
                    <div class="d">{{ c.day }}</div>
                    <div class="t">{{ dayOf(row, c.date!)?.check_in || '—' }}</div>
                    <div class="t out">{{ dayOf(row, c.date!)?.check_out || '—' }}</div>
                  </div>
                </template>
              </div>
              <div class="legend">
                <span><i class="dot ok" />正常</span>
                <span><i class="dot partial" />仅上班</span>
                <span><i class="dot off" />未打卡</span>
              </div>
            </el-popover>
          </template>
        </el-table-column>
        <el-table-column label="角色" width="100">
          <template #default="{ row }"><StatusTag kind="role" :value="row.role" /></template>
        </el-table-column>
        <el-table-column prop="store_name" label="门店" min-width="120" />
        <el-table-column prop="present_days" label="出勤天数" width="100" />
        <el-table-column label="最近打卡" min-width="200">
          <template #default="{ row }">
            <template v-if="row.days?.length">
              {{ row.days[row.days.length - 1].date }}
              · {{ fmtRec(row.days[row.days.length - 1]) }}
            </template>
            <span v-else class="muted">本月暂无</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openEmp(row)">月历</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="pager">
        <span class="pager-tip">共 {{ filteredEmployees.length }} 人（筛选后）</span>
        <el-pagination
          v-model:current-page="page"
          layout="total, prev, pager, next"
          :total="filteredEmployees.length"
          :page-size="pageSize"
          background
        />
      </div>
    </template>

    <!-- 日历维度 -->
    <div v-else v-loading="loading" class="cal-wrap">
      <div class="cal-tip">格子上悬停看前几人摘要；人数多时请点击格子打开完整名单（可搜索）</div>
      <div class="cal-grid">
        <div v-for="w in weekLabels" :key="'h'+w" class="cal-wd">{{ w }}</div>
        <template v-for="c in calendarCells" :key="c.key">
          <div v-if="!c.day" class="cal-cell empty" />
          <el-popover
            v-else
            placement="bottom"
            :width="300"
            trigger="hover"
            :show-after="160"
            :hide-after="100"
            popper-class="att-pop"
          >
            <template #reference>
              <button
                type="button"
                class="cal-cell"
                :class="{ has: c.count > 0 }"
                @click="openDay(c.date!)"
              >
                <div class="cal-day">{{ c.day }}</div>
                <div class="cal-sum">{{ c.count ? `${c.count} 人打卡` : '无记录' }}</div>
                <div class="cal-preview">
                  <span v-for="n in previewNames(c.date!)" :key="n" class="chip">{{ n }}</span>
                  <span v-if="c.count > 2" class="chip more">+{{ c.count - 2 }}</span>
                </div>
              </button>
            </template>
            <div class="pop-hd">
              <b>{{ c.date }}</b>
              <span>{{ c.count }} 人 · 点击看全部</span>
            </div>
            <div v-if="c.count" class="day-list">
              <div v-for="item in hoverDayItems(c.date!)" :key="item.emp.user_id" class="day-row">
                <div>
                  <div class="n">{{ item.emp.user_name }}</div>
                  <div class="s">{{ item.emp.store_name || '—' }}</div>
                </div>
                <div class="times" :class="dayStatusClass(item.rec)">
                  <div>{{ item.rec.check_in || '—' }}</div>
                  <div>{{ item.rec.check_out || '—' }}</div>
                </div>
              </div>
              <div v-if="c.count > 8" class="more-tip">还有 {{ c.count - 8 }} 人，点击格子查看</div>
            </div>
            <div v-else class="empty-tip">当日暂无打卡记录</div>
          </el-popover>
        </template>
      </div>
    </div>

    <!-- 某日全员明细抽屉 -->
    <el-drawer v-model="dayDrawer" :title="`${dayFocus} 出勤明细`" size="420px" destroy-on-close>
      <div class="drawer-stats">
        <span>打卡 {{ dayStats.total }}</span>
        <span>完整 {{ dayStats.ok }}</span>
        <span>仅上班 {{ dayStats.partial }}</span>
        <span>未打卡 {{ dayStats.none }}</span>
      </div>
      <el-input v-model="dayKeyword" clearable placeholder="在当日名单中搜索" style="margin-bottom:12px" />
      <div v-if="dayItems.length" class="drawer-list">
        <div v-for="item in dayItems" :key="item.emp.user_id" class="drawer-row">
          <div>
            <div class="n">{{ item.emp.user_name }}</div>
            <div class="s">{{ item.emp.store_name || '—' }} · {{ item.emp.role_name }}</div>
          </div>
          <div class="times" :class="dayStatusClass(item.rec)">
            <div>上班 {{ item.rec.check_in || '—' }}</div>
            <div>下班 {{ item.rec.check_out || '—' }}</div>
          </div>
        </div>
      </div>
      <el-empty v-else description="无匹配打卡记录" :image-size="72" />
    </el-drawer>

    <!-- 某员工月历抽屉 -->
    <el-drawer
      v-model="empDrawer"
      :title="empFocus ? `${empFocus.user_name} · ${month}` : '员工月历'"
      size="420px"
      destroy-on-close
    >
      <template v-if="empFocus">
        <div class="drawer-stats">
          <span>{{ empFocus.store_name || '未分店' }}</span>
          <span>{{ empFocus.role_name }}</span>
          <span>出勤 {{ empFocus.present_days }} 天</span>
        </div>
        <div class="mini-cal drawer-cal">
          <div v-for="w in weekLabels" :key="'dw'+w" class="wd">{{ w }}</div>
          <template v-for="c in calendarCells" :key="'dc'+c.key">
            <div v-if="!c.day" class="cell empty" />
            <div
              v-else
              class="cell"
              :class="dayStatusClass(dayOf(empFocus, c.date!))"
              :title="fmtRec(dayOf(empFocus, c.date!))"
            >
              <div class="d">{{ c.day }}</div>
              <div class="t">{{ dayOf(empFocus, c.date!)?.check_in || '—' }}</div>
              <div class="t out">{{ dayOf(empFocus, c.date!)?.check_out || '—' }}</div>
            </div>
          </template>
        </div>
        <div class="legend" style="margin-top:12px">
          <span><i class="dot ok" />正常</span>
          <span><i class="dot partial" />仅上班</span>
          <span><i class="dot off" />未打卡</span>
        </div>
      </template>
    </el-drawer>
  </div>
</template>

<style scoped>
.stats {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 160px));
  gap: 10px;
  margin-bottom: 14px;
}
.stat {
  border: 1px solid var(--el-border-color);
  border-radius: 10px;
  padding: 10px 12px;
  background: var(--el-bg-color);
}
.stat span {
  display: block;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.stat b {
  display: block;
  margin-top: 4px;
  font-size: 20px;
}
.stat b.ok { color: #67c23a; }
.stat b.warn { color: #e6a23c; }

.toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  align-items: center;
  margin-bottom: 14px;
}
.emp-hot {
  display: inline-flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 2px;
  border: 0;
  background: none;
  padding: 0;
  cursor: pointer;
  text-align: left;
}
.emp-name {
  font-weight: 600;
  color: var(--el-color-primary);
  border-bottom: 1px dashed rgba(64, 158, 255, 0.45);
}
.hint {
  font-size: 11px;
  color: var(--el-text-color-secondary);
}
.muted { color: var(--el-text-color-placeholder); }
.pager {
  margin-top: 14px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
.pager-tip { font-size: 13px; color: var(--el-text-color-secondary); }

.cal-wrap {
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color);
  border-radius: 8px;
  padding: 12px;
  min-height: 320px;
}
.cal-tip {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-bottom: 10px;
}
.cal-grid {
  display: grid;
  grid-template-columns: repeat(7, minmax(0, 1fr));
  gap: 8px;
}
.cal-wd {
  text-align: center;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  padding: 4px 0 8px;
}
.cal-cell {
  min-height: 88px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
  background: var(--el-fill-color-blank);
  padding: 8px;
  text-align: left;
  width: 100%;
  transition: border-color .15s, box-shadow .15s;
}
.cal-cell.empty {
  background: transparent;
  border-color: transparent;
  min-height: 0;
  padding: 0;
}
button.cal-cell { cursor: pointer; }
button.cal-cell:hover {
  border-color: var(--el-color-primary-light-5);
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.06);
}
.cal-cell.has {
  background: linear-gradient(180deg, #f6fbff 0%, #fff 70%);
}
.cal-day { font-size: 15px; font-weight: 700; }
.cal-sum {
  margin-top: 4px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.cal-preview {
  margin-top: 8px;
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}
.chip {
  font-size: 11px;
  padding: 1px 6px;
  border-radius: 999px;
  background: var(--el-fill-color);
}
.chip.more { color: var(--el-color-primary); }

.drawer-stats {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 12px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.drawer-stats span {
  padding: 4px 8px;
  border-radius: 999px;
  background: var(--el-fill-color);
}
.drawer-list { display: grid; gap: 8px; }
.drawer-row {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 10px;
  background: #f7f8fa;
}
.drawer-row .n { font-weight: 600; font-size: 13px; }
.drawer-row .s { font-size: 11px; color: #909399; margin-top: 2px; }
.drawer-row .times {
  text-align: right;
  font-size: 12px;
  line-height: 1.45;
  white-space: nowrap;
}
.drawer-row .times.ok { color: #67c23a; }
.drawer-row .times.partial { color: #e6a23c; }
.drawer-row .times.off { color: #909399; }
.drawer-cal { margin-top: 4px; }
</style>

<style>
.att-pop .pop-hd {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-bottom: 10px;
  gap: 10px;
}
.att-pop .pop-hd b { font-size: 14px; }
.att-pop .pop-hd span {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.att-pop .mini-cal,
.drawer-cal.mini-cal,
.mini-cal {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 4px;
}
.att-pop .wd,
.mini-cal .wd {
  text-align: center;
  font-size: 11px;
  color: var(--el-text-color-secondary);
  padding-bottom: 2px;
}
.att-pop .cell,
.mini-cal .cell {
  border-radius: 6px;
  padding: 4px 2px;
  text-align: center;
  background: #f5f7fa;
  min-height: 46px;
}
.att-pop .cell.empty,
.mini-cal .cell.empty { background: transparent; min-height: 0; padding: 0; }
.att-pop .cell.ok,
.mini-cal .cell.ok { background: #e8f8ef; }
.att-pop .cell.partial,
.mini-cal .cell.partial { background: #fff7e6; }
.att-pop .cell.off,
.mini-cal .cell.off { background: #f5f7fa; color: #909399; }
.att-pop .cell .d,
.mini-cal .cell .d { font-size: 11px; font-weight: 700; }
.att-pop .cell .t,
.mini-cal .cell .t { font-size: 10px; line-height: 1.2; margin-top: 1px; }
.att-pop .legend,
.legend {
  display: flex;
  gap: 12px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.att-pop .legend .dot,
.legend .dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 4px;
  vertical-align: middle;
}
.att-pop .legend .dot.ok,
.legend .dot.ok { background: #67c23a; }
.att-pop .legend .dot.partial,
.legend .dot.partial { background: #e6a23c; }
.att-pop .legend .dot.off,
.legend .dot.off { background: #c0c4cc; }
.att-pop .day-list { display: grid; gap: 8px; max-height: 280px; overflow: auto; }
.att-pop .day-row {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  padding: 8px 10px;
  border-radius: 8px;
  background: #f7f8fa;
}
.att-pop .day-row .n { font-weight: 600; font-size: 13px; }
.att-pop .day-row .s { font-size: 11px; color: #909399; margin-top: 2px; }
.att-pop .day-row .times {
  text-align: right;
  font-size: 12px;
  line-height: 1.45;
  white-space: nowrap;
}
.att-pop .day-row .times.ok { color: #67c23a; }
.att-pop .day-row .times.partial { color: #e6a23c; }
.att-pop .more-tip,
.att-pop .empty-tip {
  padding: 10px 4px 2px;
  text-align: center;
  color: #909399;
  font-size: 12px;
}
</style>
