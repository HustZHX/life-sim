import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'home', component: () => import('@/views/HomeView.vue') },
    { path: '/famous', name: 'famous', component: () => import('@/views/FamousCreateView.vue') },
    { path: '/random', name: 'random', component: () => import('@/views/RandomCreateView.vue') },
    { path: '/history', name: 'history', component: () => import('@/views/HistoryView.vue') },
    { path: '/continue/:id', name: 'continue', component: () => import('@/views/ContinueView.vue') },
    { path: '/timeline/:id', name: 'timeline', component: () => import('@/views/TimelineView.vue') },
  ],
})

export default router
