<script setup lang="ts">
import { computed, onMounted, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import http from '@/api/http'
import PageHeader from '@/components/PageHeader.vue'

const form = reactive<Record<string, string>>({
  brand_name: '',
  birthday_remind: 'true',
  birthday_days: '7',
  sms_enabled: 'false',
  sms_sign: '兰瓶',
  sms_access_key_id: '',
  sms_access_key_secret: '',
  sms_endpoint: 'dysmsapi.aliyuncs.com',
  sms_tpl_birthday: '',
  sms_tpl_consume: '',
  sms_tpl_appoint: '',
  clockin_wifi: '',
  clockin_range: '200',
  amap_web_key: '',
  verify_mode: 'staff',
  appoint_expire_minutes: '30',
  customer_share: 'false',
})

const rangeMeters = computed({
  get: () => {
    const n = Number(form.clockin_range || 200)
    if (!Number.isFinite(n) || n < 50) return 200
    if (n > 1000) return 1000
    return Math.round(n)
  },
  set: (v: number) => {
    form.clockin_range = String(v)
  },
})

function normalizeVerifyMode(v: string) {
  if (v === 'notify' || v === 'real') return 'notify'
  return 'staff'
}

async function load() {
  const data: any = await http.get('/api/settings')
  Object.assign(form, data)
  form.verify_mode = normalizeVerifyMode(form.verify_mode || 'staff')
  if (!form.appoint_expire_minutes) form.appoint_expire_minutes = '30'
  if (!form.customer_share) form.customer_share = 'false'
  if (!form.sms_sign) form.sms_sign = '兰瓶'
  if (!form.sms_endpoint) form.sms_endpoint = 'dysmsapi.aliyuncs.com'
  if (!form.clockin_range) form.clockin_range = '200'
}
async function save() {
  form.clockin_range = String(rangeMeters.value)
  await http.put('/api/settings', { ...form })
  ElMessage.success('设置已保存；打卡范围立即对员工端生效')
}
onMounted(load)
</script>

<template>
  <div>
    <PageHeader title="系统设置" subtitle="品牌、阿里云短信、核销模式与打卡参数" />
    <div class="card">
      <el-form label-width="150px" style="max-width:720px">
        <div class="sec">基础信息</div>
        <el-form-item label="品牌名"><el-input v-model="form.brand_name" /></el-form-item>
        <el-form-item label="生日提醒"><el-switch v-model="form.birthday_remind" active-value="true" inactive-value="false" /></el-form-item>
        <el-form-item label="提前提醒天数"><el-input v-model="form.birthday_days" /></el-form-item>

        <div class="sec">阿里云通知短信</div>
        <el-form-item label="短信总开关">
          <el-switch v-model="form.sms_enabled" active-value="true" inactive-value="false" active-text="开启" inactive-text="关闭" />
        </el-form-item>
        <el-form-item label="短信签名">
          <el-input v-model="form.sms_sign" placeholder="兰瓶" />
        </el-form-item>
        <el-form-item label="AccessKey ID">
          <el-input v-model="form.sms_access_key_id" placeholder="" clearable />
        </el-form-item>
        <el-form-item label="AccessKey Secret">
          <el-input v-model="form.sms_access_key_secret" type="password" show-password placeholder="" clearable />
        </el-form-item>
        <el-form-item label="Endpoint">
          <el-input v-model="form.sms_endpoint" placeholder="" />
        </el-form-item>
        <el-form-item label="生日模板 Code">
          <el-input v-model="form.sms_tpl_birthday" placeholder="" clearable />
        </el-form-item>
        <el-form-item label="消费模板 Code">
          <el-input v-model="form.sms_tpl_consume" placeholder="" clearable />
        </el-form-item>
        <el-form-item label="预约模板 Code">
          <el-input v-model="form.sms_tpl_appoint" placeholder="" clearable />
        </el-form-item>
        

        <div class="sec">核销模式</div>
        <el-form-item label="客户参与方式">
          <el-radio-group v-model="form.verify_mode">
            <el-radio value="staff">不发通知 — 员工选卡 → 下一步 → 确认即可（客户全程不参与）</el-radio>
            <el-radio value="notify">发通知 — 生成核销码，客户打开亮码，员工扫一扫确认</el-radio>
          </el-radio-group>
        </el-form-item>
   

 
        <div class="sec">预约超时</div>
        <el-form-item label="超时失效(分钟)">
          <el-input v-model="form.appoint_expire_minutes" style="max-width:160px" placeholder="30" />
          <span class="inline-tip">超过预约开始时间多少分钟，仍为「待履约/已改约」→ 系统自动「已失效」</span>
        </el-form-item>

        


        <div class="sec">客户可见范围</div>
        <el-form-item label="门店客户共享">
          <el-switch v-model="form.customer_share" active-value="true" inactive-value="false" active-text="共享（全店可见）" inactive-text="不共享（仅本店）" />
        </el-form-item>
        <el-form-item>
          <div class="tip">关闭共享后，店长/员工预约与核销只能看到本店客户；开启后所有门店客户互相可见。</div>
        </el-form-item>

        <div class="sec">打卡围栏（真实打卡）</div>
        <el-form-item label="打卡范围">
          <div class="range-box">
            <el-slider
              v-model="rangeMeters"
              :min="50"
              :max="1000"
              :step="10"
              show-input
              :show-input-controls="true"
              style="max-width:420px"
            />
            <div class="tip">员工须在门店地图选点的方圆 <b>{{ rangeMeters }}</b> 米内才能上班/下班打卡。室内建议 200–400 米。</div>
          </div>
        </el-form-item>
        <el-form-item label="打卡 Wi-Fi">
          <el-input v-model="form.clockin_wifi" placeholder="仅备注，H5 无法校验 SSID" />
        </el-form-item>
        <el-form-item label="高德 Web Key">
          <el-input v-model="form.amap_web_key" clearable placeholder="用于门店地图地址搜索（Web服务 Key）" />
        </el-form-item>
        
        <el-form-item><el-button type="primary" @click="save">保存</el-button></el-form-item>
      </el-form>
    </div>
  </div>
</template>
<style scoped>
.card { background:#fff; border:1px solid var(--lp-border); border-radius: var(--lp-radius); padding: 22px; }
.sec { font-weight:700; margin: 8px 0 14px; color:#344054; }
.tip { color:#667085; font-size:12px; line-height:1.6; max-width:520px; }
.tip b { color:#e24b6e; }
.inline-tip { margin-left: 12px; color:#667085; font-size:12px; }
.range-box { width: 100%; }
</style>
