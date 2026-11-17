import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: () => import('@/views/LandingView.vue') },
    { path: '/verify/:token', component: () => import('@/views/VerifyView.vue') },
  ],
})

export default router
