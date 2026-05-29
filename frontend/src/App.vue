<script setup lang="ts">
import { computed } from 'vue'
import { RouterView, useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useLayoutStore } from '@/stores/layout'
import LayoutModeToggle from '@/components/LayoutModeToggle.vue'
import GlobalModelSwitcher from '@/components/GlobalModelSwitcher.vue'
import AppBottomNav from '@/components/mobile/AppBottomNav.vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const layout = useLayoutStore()

const showHeader = computed(() => !route.meta.authPage && (!auth.enabled || !!auth.user))

async function logout() {
  await auth.logout()
  auth.clearGate()
  await router.replace('/gate')
}
</script>

<template>
  <div class="app-shell" :class="{ 'app-shell--mobile': layout.isMobile }">
    <header v-if="showHeader" class="app-header">
      <router-link to="/" class="logo">人生模拟</router-link>
      <nav class="nav">
        <router-link to="/characters">人物列表</router-link>
      </nav>
      <div class="header-right">
        <GlobalModelSwitcher />
        <LayoutModeToggle />
        <span v-if="auth.user" class="user-label">{{ auth.user.display_name || auth.user.username }}</span>
        <el-button v-if="auth.enabled && auth.user" size="small" text @click="logout">退出登录</el-button>
        <span class="tagline">AI 演绎 · 非学术考证</span>
      </div>
    </header>
    <main class="app-main">
      <RouterView />
    </main>
    <AppBottomNav v-if="layout.isMobile && showHeader" />
  </div>
</template>

<style scoped>
.header-right {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 8px;
}
.user-label {
  font-size: 0.85rem;
  color: #606266;
}
</style>
