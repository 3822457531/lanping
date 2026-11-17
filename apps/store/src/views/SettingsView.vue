<script setup lang="ts">
import { onMounted, reactive } from 'vue'
import { showSuccessToast } from 'vant'
import http from '@/api/http'

const form = reactive({ birthday_remind: 'true', birthday_days: '7' })

onMounted(async () => {
  const data: any = await http.get('/api/settings')
  form.birthday_remind = data.birthday_remind ?? 'true'
  form.birthday_days = data.birthday_days ?? '7'
})

async function save() {
  await http.put('/api/settings', {
    birthday_remind: form.birthday_remind,
    birthday_days: form.birthday_days,
  })
  showSuccessToast('已保存')
}
</script>

<template>
  <div class="page-wrap">
    <van-nav-bar title="设置" left-arrow @click-left="$router.back()" />
    <div class="box">
      <div class="sec-title">门店偏好</div>
      <van-cell title="生日提醒开关">
        <template #right-icon>
          <van-switch
            :model-value="form.birthday_remind==='true'"
            size="20px"
            active-color="#e24b6e"
            @change="(v:boolean)=>form.birthday_remind=v?'true':'false'"
          />
        </template>
      </van-cell>
      <van-field v-model="form.birthday_days" type="digit" label="提前提醒天数" />
      <div class="pad">
        <van-button block round type="primary" @click="save">保存设置</van-button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.box {
  margin: 8px 14px;
  overflow: hidden;
  border-radius: 16px;
  background: rgba(255,255,255,.95);
  box-shadow: var(--lp-shadow-card);
  border: 1px solid rgba(255,245,248,.95);
}
.sec-title {
  padding: 10px 16px 4px;
  font-size: 13px;
  font-weight: 720;
}
.pad { padding: 10px; }
</style>
