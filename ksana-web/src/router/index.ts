import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      redirect: '/jobs'
    },
    {
      path: '/jobs',
      name: 'jobs',
      component: () => import('@/views/JobsList.vue'),
      meta: { title: '任务列表' }
    },
    {
      path: '/jobs/new',
      name: 'job-new',
      component: () => import('@/views/JobForm.vue'),
      meta: { title: '新建任务' }
    },
    {
      path: '/jobs/:id',
      name: 'job-detail',
      component: () => import('@/views/JobDetail.vue'),
      meta: { title: '任务详情' }
    },
    {
      path: '/jobs/:id/edit',
      name: 'job-edit',
      component: () => import('@/views/JobForm.vue'),
      meta: { title: '编辑任务' }
    },
    {
      path: '/health',
      name: 'health',
      component: () => import('@/views/Health.vue'),
      meta: { title: '健康检查' }
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import('@/views/Settings.vue'),
      meta: { title: '设置' }
    }
  ],
})

export default router
