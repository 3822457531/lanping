<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useSession } from '@/stores/session'

const route = useRoute()
const router = useRouter()
const session = useSession()

const menus = [
  { path: '/dashboard', label: '数据看板' },
  { path: '/customers', label: '客户管理' },
  { path: '/employees', label: '员工管理' },
  { path: '/stores', label: '门店管理' },
  { path: '/audits', label: '审核管理' },
  { path: '/projects', label: '项目管理' },
  { path: '/cards', label: '开卡充值' },
  { path: '/verifications', label: '核销记录' },
  { path: '/appointments', label: '预约管理' },
  { path: '/inventory', label: '出入库' },
  { path: '/assessments', label: '考核情况' },
  { path: '/attendance', label: '考勤管理' },
  { path: '/salary', label: '薪资管理' },
  { path: '/welfare', label: '福利管理' },
  { path: '/settings', label: '系统设置' },
  { path: '/logs', label: '操作日志' },
]

const active = computed(() => route.path)

function logout() {
  session.logout()
  router.push('/login')
}
</script>

<template>
  <div class="shell">
    <aside class="side">
      <div class="brand">
        <div class="mark">兰</div>
        <div>
          <div class="name">兰瓶美业</div>
          <div class="sub">连锁管理后台</div>
        </div>
      </div>
      <nav>
        <router-link
          v-for="m in menus"
          :key="m.path"
          :to="m.path"
          class="nav-item"
          :class="{ active: active === m.path }"
        >{{ m.label }}</router-link>
      </nav>
    </aside>
    <div class="main">
      <header class="top">
        <div class="crumb">{{ menus.find(m => m.path === active)?.label || '工作台' }}</div>
        <div class="user">
          <span>{{ session.user?.name }} · {{ session.user?.role_name }}</span>
          <button class="link" @click="logout">退出</button>
        </div>
      </header>
      <main class="content">
        <router-view />
      </main>
    </div>
  </div>
</template>

<style scoped>
.shell { display: flex; min-height: 100vh; background: var(--lp-bg); }
.side {
  width: 220px; background: var(--lp-sidebar); color: var(--lp-sidebar-text);
  display: flex; flex-direction: column; flex-shrink: 0;
}
.brand { display: flex; gap: 12px; align-items: center; padding: 22px 18px 18px; }
.mark {
  width: 36px; height: 36px; border-radius: 10px; background: var(--lp-accent);
  color: #fff; display: grid; place-items: center; font-weight: 700;
}
.name { color: #fff; font-weight: 600; font-size: 15px; }
.sub { font-size: 11px; opacity: .7; margin-top: 2px; }
nav { padding: 8px 10px 24px; display: flex; flex-direction: column; gap: 2px; overflow: auto; }
.nav-item {
  padding: 10px 12px; border-radius: 8px; font-size: 13.5px; color: var(--lp-sidebar-text);
}
.nav-item:hover { background: rgba(255,255,255,.06); color: #fff; }
.nav-item.active { background: rgba(255,255,255,.12); color: var(--lp-sidebar-active); font-weight: 600; }
.main { flex: 1; min-width: 0; display: flex; flex-direction: column; }
.top {
  height: 56px; background: var(--lp-surface); border-bottom: 1px solid var(--lp-border);
  display: flex; align-items: center; justify-content: space-between; padding: 0 22px;
}
.crumb { font-size: 15px; font-weight: 600; }
.user { display: flex; gap: 14px; align-items: center; color: var(--lp-muted); font-size: 13px; }
.link { border: 0; background: none; color: var(--lp-accent); cursor: pointer; font-size: 13px; }
.content { padding: 20px 22px 32px; }
</style>
