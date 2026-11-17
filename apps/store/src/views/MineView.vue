<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useSession } from '@/stores/session'

const session = useSession()
const router = useRouter()
const manager = computed(() => session.user?.role === 'manager')
const avatarText = computed(() => String(session.user?.name || '员').slice(0, 1))

const staffMenus = [
  { title: '客户管理', desc: '新增客户', path: '/customers', tone: 'pink', icon: 'friends-o' },
  { title: '开卡充值', desc: '办卡 / 储值', path: '/cards', tone: 'gold', icon: 'gold-coin-o' },
  { title: '发起核销', desc: '选卡发码', path: '/verify', tone: 'pink', icon: 'upgrade' },
  { title: '预约服务', desc: '新建预约', path: '/appointments', tone: 'blue', icon: 'calendar-o' },
  { title: '我的薪资', desc: '提成明细', path: '/salary', tone: 'orange', icon: 'balance-list-o' },
  { title: '我的考核', desc: '是否合格', path: '/assessment', tone: 'purple', icon: 'star-o' },
  { title: '打卡上班', desc: '出勤签到', path: '/clockin', tone: 'sky', icon: 'contact' },
  { title: '消息中心', desc: '通知提醒', path: '/messages', tone: 'red', icon: 'bell' },
]
const managerExtra = [
  { title: '核销列表', desc: '记录补退', path: '/verifications', tone: 'green', icon: 'orders-o' },
  { title: '员工管理', desc: '入职审核', path: '/employees', tone: 'blue', icon: 'manager-o' },
  { title: '出入库', desc: '库存调度', path: '/inventory', tone: 'sky', icon: 'logistics' },
  { title: '设置', desc: '门店偏好', path: '/settings', tone: 'purple', icon: 'setting-o' },
]
const cards = computed(() => (manager.value ? [...staffMenus, ...managerExtra] : staffMenus))

function logout() {
  session.logout()
  router.replace('/login')
}
</script>

<template>
  <div class="mine">
    <header class="hero lp-hero">
      <div class="avatar">{{ avatarText }}</div>
      <div class="who">
        <div class="name">{{ session.user?.name }}</div>
        <div class="meta">
          <van-icon name="shop-o" />
          <span>{{ session.user?.role_name || '美业员工' }} · {{ session.user?.store_name || '本店' }}</span>
        </div>
      </div>
      <div class="badge">{{ manager ? '店长' : '员工' }}</div>
    </header>

    <section class="sheet">
      <div class="sec-title">常用功能</div>
      <div class="grid">
        <button
          v-for="c in cards"
          :key="c.path + c.title"
          type="button"
          class="item"
          @click="router.push(c.path)"
        >
          <div class="ico" :class="c.tone">
            <van-icon :name="c.icon" />
          </div>
          <div class="txt">
            <div class="t">{{ c.title }}</div>
            <div class="d">{{ c.desc }}</div>
          </div>
          <van-icon name="arrow" class="go" />
        </button>
      </div>

      <button type="button" class="logout" @click="logout">
        <van-icon name="revoke" />
        <span>退出登录</span>
      </button>
    </section>
  </div>
</template>

<style scoped>
.mine {
  min-height: 100%;
  padding: 8px var(--lp-page-x) 4px;
}
.hero {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 12px;
  margin-bottom: 8px;
}
.avatar {
  position: relative;
  z-index: 1;
  width: 42px;
  height: 42px;
  border-radius: 50%;
  background: #fff;
  color: var(--lp-accent);
  display: grid;
  place-items: center;
  font-size: 16px;
  font-weight: 760;
  flex-shrink: 0;
}
.who { position: relative; z-index: 1; flex: 1; min-width: 0; }
.name { font-size: 17px; font-weight: 760; line-height: 1.15; }
.meta {
  margin-top: 4px;
  display: flex;
  align-items: center;
  gap: 3px;
  font-size: 11px;
  opacity: 0.95;
}
.badge {
  position: relative;
  z-index: 1;
  font-size: 10px;
  font-weight: 700;
  padding: 3px 8px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.22);
  border: 1px solid rgba(255, 255, 255, 0.28);
}

.sheet {
  background: rgba(255, 255, 255, 0.94);
  border-radius: 16px;
  padding: 10px 10px 12px;
  box-shadow: var(--lp-shadow-card);
  border: 1px solid rgba(255, 255, 255, 0.9);
}
.sec-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 740;
  color: #2a2124;
  margin: 0 2px 8px;
}
.sec-title::before {
  content: '';
  width: 3px;
  height: 12px;
  border-radius: 99px;
  background: linear-gradient(180deg, #ff8fa6, #e24b6e);
}
.grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 6px;
}
.item {
  display: grid;
  grid-template-columns: 32px minmax(0, 1fr) auto;
  align-items: center;
  gap: 6px;
  text-align: left;
  border-radius: 12px;
  padding: 8px 8px;
  background: linear-gradient(180deg, #fff, #fffafb);
  border: 1px solid #f6eef1;
}
.ico {
  width: 32px;
  height: 32px;
  border-radius: 10px;
  display: grid;
  place-items: center;
  color: #fff;
  font-size: 15px;
  flex-shrink: 0;
}
.ico.pink { background: linear-gradient(145deg, #ff7d97, #f04e6e); }
.ico.blue { background: linear-gradient(145deg, #6eb6ff, #3d8dff); }
.ico.orange { background: linear-gradient(145deg, #ffb06a, #f08a3a); }
.ico.purple { background: linear-gradient(145deg, #c39bff, #9b6bff); }
.ico.sky { background: linear-gradient(145deg, #7ec8ff, #4aa3ef); }
.ico.red { background: linear-gradient(145deg, #ff7b7b, #ef4d4d); }
.ico.green { background: linear-gradient(145deg, #7ed6b0, #3cae84); }
.ico.gold { background: linear-gradient(145deg, #ffd27a, #f0b040); }
.txt { min-width: 0; }
.t {
  font-size: 12px;
  font-weight: 720;
  color: #2a2124;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.d {
  margin-top: 1px;
  font-size: 10px;
  color: #a3999d;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.go { color: #d5c8cc; font-size: 11px; }
.logout {
  margin-top: 10px;
  width: 100%;
  height: 40px;
  border-radius: 999px;
  border: 1.5px solid rgba(226, 75, 110, 0.55);
  background: #fff;
  color: var(--lp-accent);
  font-size: 13px;
  font-weight: 720;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}
.logout .van-icon { font-size: 15px; }
</style>
