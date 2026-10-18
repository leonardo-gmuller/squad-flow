import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = createRouter({
  history: createWebHistory(),
  scrollBehavior: () => ({ top: 0 }),
  routes: [
    { path: '/login', component: () => import('../views/LoginView.vue'), meta: { public: true } },
    {
      path: '/reset-password',
      component: () => import('../views/ResetPasswordView.vue'),
      meta: { public: true },
    },
    {
      path: '/',
      component: () => import('../components/AppShell.vue'),
      children: [
        { path: '', redirect: '/dashboard' },
        {
          path: '/dashboard',
          component: () => import('../views/DashboardView.vue'),
          meta: { title: 'Visão geral' },
        },
        {
          path: '/tickets',
          component: () => import('../views/TicketsView.vue'),
          meta: { title: 'Demandas' },
        },
        {
          path: '/board',
          component: () => import('../views/BoardView.vue'),
          meta: { title: 'Quadro Kanban' },
        },
        {
          path: '/releases',
          component: () => import('../views/ReleasesView.vue'),
          meta: { title: 'Versões' },
        },
        {
          path: '/packages',
          component: () => import('../views/PackagesView.vue'),
          meta: { title: 'OS e iniciativas' },
        },
        {
          path: '/qa',
          component: () => import('../views/QaView.vue'),
          meta: { title: 'Central de QA' },
        },
        {
          path: '/reports',
          component: () => import('../views/ReportsView.vue'),
          meta: { title: 'Relatórios' },
        },
        {
          path: '/workstreams',
          component: () => import('../views/WorkstreamsView.vue'),
          meta: { title: 'Esteiras', admin: true },
        },
        {
          path: '/users',
          component: () => import('../views/UsersView.vue'),
          meta: { title: 'Usuários', admin: true },
        },
        {
          path: '/account',
          component: () => import('../views/AccountView.vue'),
          meta: { title: 'Minha conta' },
        },
      ],
    },
    { path: '/:pathMatch(.*)*', redirect: '/dashboard' },
  ],
  linkActiveClass: 'active-route',
})
router.beforeEach(async (to) => {
  const auth = useAuthStore()
  await auth.restore()
  if (!to.meta.public && !auth.user) return { path: '/login', query: { redirect: to.fullPath } }
  if (to.meta.admin && auth.user?.role !== 'admin') return '/dashboard'
  if (to.path === '/login' && auth.user) return '/dashboard'
})
window.addEventListener('squadflow:session-expired', () => {
  const auth = useAuthStore()
  if (!auth.initialized || !auth.user) return
  auth.expire()
  if (router.currentRoute.value.path !== '/login') void router.replace('/login')
})
export default router
