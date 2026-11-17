import { createRouter, createWebHistory } from 'vue-router'
import { useSession } from '@/stores/session'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/login', component: () => import('@/views/LoginView.vue'), meta: { public: true } },
    {
      path: '/',
      component: () => import('@/layouts/AdminLayout.vue'),
      children: [
        { path: '', redirect: '/dashboard' },
        { path: 'dashboard', component: () => import('@/views/DashboardView.vue') },
        { path: 'customers', component: () => import('@/views/CustomersView.vue') },
        { path: 'projects', component: () => import('@/views/ProjectsView.vue') },
        { path: 'cards', component: () => import('@/views/CardsView.vue') },
        { path: 'verifications', component: () => import('@/views/VerificationsView.vue') },
        { path: 'employees', component: () => import('@/views/EmployeesView.vue') },
        { path: 'stores', component: () => import('@/views/StoresView.vue') },
        { path: 'audits', component: () => import('@/views/AuditsView.vue') },
        { path: 'appointments', component: () => import('@/views/AppointmentsView.vue') },
        { path: 'inventory', component: () => import('@/views/InventoryView.vue') },
        { path: 'assessments', component: () => import('@/views/AssessmentsView.vue') },
        { path: 'attendance', component: () => import('@/views/AttendanceView.vue') },
        { path: 'salary', component: () => import('@/views/SalaryView.vue') },
        { path: 'welfare', component: () => import('@/views/WelfareView.vue') },
        { path: 'settings', component: () => import('@/views/SettingsView.vue') },
        { path: 'logs', component: () => import('@/views/LogsView.vue') },
      ],
    },
  ],
})

router.beforeEach((to) => {
  const session = useSession()
  if (!to.meta.public && !session.token) return '/login'
  if (to.path === '/login' && session.token) return '/dashboard'
  return true
})

export default router
