import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useSetupStore } from '../stores/setup'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/setup', component: () => import('../views/SetupView.vue'), meta: { public: true } },
    {
      path: '/login',
      redirect: (to) => ({
        path: '/',
        query: {
          signin: '1',
          ...(to.query.redirect ? { redirect: to.query.redirect as string } : {}),
        },
      }),
    },
    { path: '/invite', component: () => import('../views/InviteAcceptView.vue'), meta: { public: true } },
    { path: '/', component: () => import('../views/HomeView.vue'), meta: { public: true } },
    { path: '/launcher', component: () => import('../views/LauncherView.vue') },
    { path: '/admin/catalog', component: () => import('../views/AdminCatalogView.vue'), meta: { admin: true } },
    { path: '/admin/catalog/preview', component: () => import('../views/CatalogPreviewView.vue'), meta: { admin: true } },
    { path: '/admin/users', component: () => import('../views/AdminUsersView.vue'), meta: { admin: true } },
    { path: '/admin/roles', component: () => import('../views/AdminRolesView.vue'), meta: { admin: true } },
    {
      path: '/admin/configuration',
      component: () => import('../views/admin/configuration/AdminConfigurationLayout.vue'),
      meta: { admin: true },
      redirect: '/admin/configuration/identity',
      children: [
        { path: 'identity', component: () => import('../views/admin/configuration/OidcConfigView.vue') },
        { path: 'appearance', component: () => import('../views/admin/configuration/AppearanceConfigView.vue') },
        { path: 'harbor', component: () => import('../views/admin/configuration/HarborConfigView.vue') },
        { path: 'trivy', component: () => import('../views/admin/configuration/TrivyConfigView.vue') },
        { path: 'webhooks', component: () => import('../views/admin/configuration/WebhookLogView.vue') },
        { path: 'email', component: () => import('../views/admin/configuration/EmailConfigView.vue') },
        { path: 'teams', component: () => import('../views/admin/configuration/TeamsConfigView.vue') },
      ],
    },
    { path: '/admin/audit', component: () => import('../views/AdminAuditLogView.vue'), meta: { admin: true } },
    { path: '/admin/notifications', redirect: '/admin/configuration/harbor' },
    { path: '/security', component: () => import('../views/SecurityOverviewView.vue'), meta: { security: true } },
    { path: '/security/cves', component: () => import('../views/CveDashboardView.vue'), meta: { security: true } },
    { path: '/security/reports', component: () => import('../views/DeploymentReportsView.vue'), meta: { security: true } },
    { path: '/security/reports/:id', component: () => import('../views/ReportDetailView.vue'), meta: { security: true } },
    { path: '/profile', component: () => import('../views/ProfileView.vue') },
  ],
})

router.beforeEach(async (to) => {
  const setup = useSetupStore()
  if (setup.complete === null) {
    try {
      await setup.fetchStatus()
    } catch {
      return true
    }
  }

  const auth = useAuthStore()
  if (!auth.loaded) await auth.fetchMe()

  if (!setup.complete && to.path !== '/setup') {
    return '/setup'
  }
  if (setup.complete && to.path === '/setup') {
    return auth.user ? '/launcher' : { path: '/', query: { signin: '1' } }
  }

  if (!to.meta.public && !auth.user) {
    return { path: '/', query: { signin: '1', redirect: to.fullPath } }
  }
  if (to.meta.admin && !auth.isAdmin()) return '/launcher'
  if (to.meta.security && !auth.canViewSecurity()) return '/launcher'
  return true
})

export default router
