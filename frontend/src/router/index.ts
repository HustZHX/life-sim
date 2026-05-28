import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/gate', name: 'gate', component: () => import('@/views/GateView.vue'), meta: { authPage: true } },
    { path: '/login', name: 'login', component: () => import('@/views/LoginView.vue'), meta: { authPage: true } },
    { path: '/', name: 'home', component: () => import('@/views/HomeView.vue') },
    { path: '/famous', name: 'famous', component: () => import('@/views/FamousCreateView.vue') },
    { path: '/random', name: 'random', component: () => import('@/views/RandomCreateView.vue') },
    { path: '/characters', name: 'characters', component: () => import('@/views/CharactersView.vue') },
    { path: '/characters/:id', name: 'character-timelines', component: () => import('@/views/CharacterTimelinesView.vue') },
    { path: '/history', redirect: '/characters' },
    { path: '/continue/:id', name: 'continue', component: () => import('@/views/ContinueView.vue') },
    { path: '/timeline/:id', name: 'timeline', component: () => import('@/views/TimelineView.vue') },
  ],
})

let authBootstrapped = false

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (!authBootstrapped) {
    await auth.loadStatus()
    authBootstrapped = true
  }

  if (!auth.enabled) {
    return true
  }

  if (to.meta.authPage) {
    if (to.name === 'gate' && auth.gatePassed) {
      return auth.user ? '/' : '/login'
    }
    if (to.name === 'login') {
      if (!auth.gatePassed) return '/gate'
      if (auth.user) return '/'
    }
    return true
  }

  if (!auth.gatePassed) return '/gate'
  if (!auth.user) return '/login'
  return true
})

export default router
