import type { CapacitorConfig } from '@capacitor/cli'

const config: CapacitorConfig = {
  appId: 'com.lanping.store',
  appName: '兰瓶门店',
  webDir: 'dist',
  server: {
    // 调试期用 http，否则 WebView(https://localhost) 请求局域网 http API 会被混合内容拦截
    // 正式上线 HTTPS 后再改回 androidScheme: 'https' 并关掉 cleartext
    cleartext: true,
    androidScheme: 'http',
    allowNavigation: ['192.168.1.4', 'localhost', '127.0.0.1'],
  },
  plugins: {
    SplashScreen: {
      launchAutoHide: true,
      backgroundColor: '#f25575',
    },
    StatusBar: {
      style: 'LIGHT',
      backgroundColor: '#f25575',
    },
  },
}

export default config
