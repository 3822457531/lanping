import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  // 生产子路径：VITE_BASE=/h5/ npm run build
  base: process.env.VITE_BASE || '/',
  plugins: [vue()],
  resolve: {
    alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) },
  },
  server: {
    host: true, // 局域网可访问（手机浏览器 / 真机调试）
    port: 5174,
    proxy: { '/api': { target: 'http://127.0.0.1:8080', changeOrigin: true } },
  },
})
