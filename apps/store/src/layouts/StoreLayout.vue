<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const tabs = [
  { name: 'home', path: '/home', label: '工作台', icon: 'home-o' },
  { name: 'appt', path: '/appointments', label: '预约', icon: 'calendar-o' },
  { name: 'verify', path: '/verify', label: '核销', icon: 'scan' },
  { name: 'mine', path: '/mine', label: '我的', icon: 'user-o' },
]

const active = computed(() => {
  const p = route.path
  if (p.startsWith('/appointments') || p.startsWith('/cards') || p.startsWith('/employees') || p.startsWith('/customers') || p.startsWith('/projects')) return 'appt'
  if (p.startsWith('/verify') || p.startsWith('/verifications')) return 'verify'
  if (p.startsWith('/mine') || p.startsWith('/salary') || p.startsWith('/assessment') || p.startsWith('/clockin') || p.startsWith('/settings') || p.startsWith('/inventory') || p.startsWith('/messages')) return 'mine'
  return 'home'
})

function go(name: string) {
  const tab = tabs.find((t) => t.name === name)
  if (tab) router.push(tab.path)
}
</script>

<template>
  <div class="app lp-bg">
    <div class="page"><router-view /></div>
    <nav class="dock" aria-label="底部导航">
      <button
        v-for="t in tabs"
        :key="t.name"
        type="button"
        class="dock-item"
        :class="{ on: active === t.name }"
        @click="go(t.name)"
      >
        <span class="ico-wrap">
          <van-icon :name="t.icon" class="ico" />
        </span>
        <span class="lab">{{ t.label }}</span>
      </button>
    </nav>
  </div>
</template>

<style scoped>
.app {
  height: 100%;
  max-height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  position: relative;
  /* 底部留给固定底栏，避免内容被挡 */
  padding-bottom: calc(var(--lp-tabbar-h) + env(safe-area-inset-bottom, 0px));
}
.page {
  flex: 1;
  min-height: 0;
  overflow-x: hidden;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
  overscroll-behavior-y: contain;
  padding-bottom: 12px;
  scrollbar-width: none;
}
.page::-webkit-scrollbar {
  width: 0;
  height: 0;
  display: none;
}

/* 固定贴死屏幕最底部（含 Home 指示条安全区），白底一路铺满 */
.dock {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 40;
  box-sizing: border-box;
  height: calc(var(--lp-tabbar-h) + env(safe-area-inset-bottom, 0px));
  padding: 0 2px env(safe-area-inset-bottom, 0px);
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  align-items: stretch;
  background: #fff;
  border: 0;
  border-top: 1px solid rgba(239, 224, 229, 0.95);
  border-radius: 0;
  /* 再保险：把白底向下多铺一点，防止个别机型露边 */
  box-shadow: 0 1px 0 0 #fff, 0 -1px 0 rgba(80, 40, 55, 0.06);
}
.dock-item {
  border: 0;
  background: transparent;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  color: #b7a8ad;
  font-size: 11px;
  font-weight: 600;
  padding: 2px 0 8px;
  height: var(--lp-tabbar-h);
  position: relative;
  transition: color 0.2s ease;
}
.ico-wrap {
  width: 32px;
  height: 22px;
  display: grid;
  place-items: center;
  border-radius: 10px;
  transition: background 0.22s ease, transform 0.22s ease;
}
.dock-item .ico { font-size: 18px; line-height: 1; }
.dock-item .lab {
  letter-spacing: 0;
  line-height: 1.15;
  font-size: 10px;
  white-space: nowrap;
}
.dock-item.on {
  color: var(--lp-accent);
}
.dock-item.on .ico-wrap {
  background: linear-gradient(145deg, rgba(255, 143, 166, 0.22), rgba(226, 75, 110, 0.12));
}
.dock-item.on::after {
  content: '';
  position: absolute;
  bottom: 4px;
  left: 50%;
  transform: translateX(-50%);
  width: 14px;
  height: 3px;
  border-radius: 99px;
  background: linear-gradient(90deg, #ff8fa6, #e24b6e);
  box-shadow: 0 2px 6px rgba(226, 75, 110, 0.4);
}
.dock-item:active .ico-wrap {
  transform: scale(0.92);
}
</style>
