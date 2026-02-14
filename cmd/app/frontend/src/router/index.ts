import { createRouter, createWebHashHistory } from 'vue-router'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/',
      name: 'dashboard',
      component: () => import('@/views/Dashboard.vue')
    },
    {
      path: '/services',
      name: 'services',
      component: () => import('@/views/Services.vue')
    },
    {
      path: '/services/:name',
      name: 'service-detail',
      component: () => import('@/views/ServiceDetail.vue'),
      props: true
    },
    {
      path: '/logs',
      name: 'logs',
      component: () => import('@/views/Logs.vue')
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import('@/views/Settings.vue')
    }
  ]
})

export default router
