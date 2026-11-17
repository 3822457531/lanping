import { createRouter, createWebHistory } from 'vue-router'
import { useSession } from '@/stores/session'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/login', component: () => import('@/views/LoginView.vue'), meta: { public: true } },
    { path: '/register', component: () => import('@/views/RegisterView.vue'), meta: { public: true } },
    {
      path: '/',
      component: () => import('@/layouts/StoreLayout.vue'),
      children: [
        { path: '', redirect: '/home' },
        { path: 'home', component: () => import('@/views/HomeView.vue') },
        { path: 'customers', component: () => import('@/views/CustomersView.vue') },
        { path: 'employees', component: () => import('@/views/EmployeesView.vue'), meta: { manager: true } },
        { path: 'projects', component: () => import('@/views/ProjectsView.vue') },
        { path: 'cards', component: () => import('@/views/CardsView.vue') },
        { path: 'verifications', component: () => import('@/views/VerificationsView.vue'), meta: { manager: true } },
        { path: 'verify', component: () => import('@/views/VerifyFlowView.vue') },
        { path: 'appointments', component: () => import('@/views/AppointmentsView.vue') },
        { path: 'inventory', component: () => import('@/views/InventoryView.vue'), meta: { manager: true } },
        { path: 'clockin', component: () => import('@/views/ClockinView.vue') },
        { path: 'assessment', component: () => import('@/views/AssessmentView.vue') },
        { path: 'salary', component: () => import('@/views/SalaryView.vue') },
        { path: 'settings', component: () => import('@/views/SettingsView.vue'), meta: { manager: true } },
        { path: 'mine', component: () => import('@/views/MineView.vue') },
        { path: 'messages', component: () => import('@/views/MessagesView.vue') },
      ],
    },
  ],
})

router.beforeEach((to) => {
  const session = useSession()
  if (!to.meta.public && !session.token) return '/login'
  if ((to.path === '/login' || to.path === '/register') && session.token) return '/home'
  if (to.meta.manager && session.user?.role !== 'manager') return '/home'
  return true
})

export default router
