import { createApp } from 'vue'
import { createPinia } from 'pinia'
import Vant from 'vant'
import 'vant/lib/index.css'
import './styles/tokens.css'
import App from './App.vue'
import router from './router'
import { Capacitor } from '@capacitor/core'

async function bootstrapNative() {
  if (!Capacitor.isNativePlatform()) return
  try {
    const { StatusBar, Style } = await import('@capacitor/status-bar')
    // 不让 WebView 钻到系统状态栏下面，避免标题/「新建」被信号栏挡住
    await StatusBar.setOverlaysWebView({ overlay: false })
    await StatusBar.setStyle({ style: Style.Light })
    await StatusBar.setBackgroundColor({ color: '#f25575' })
  } catch {
    // 插件在 Web 预览时可忽略
  }
}

bootstrapNative()
createApp(App).use(createPinia()).use(router).use(Vant).mount('#app')
