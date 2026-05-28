<script setup lang="ts">
import { computed } from 'vue'
import { RouterView, useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const showHeader = computed(() => !route.meta.authPage && (!auth.enabled || !!auth.user))

async function logout() {
  await auth.logout()
  auth.clearGate()
  await router.replace('/gate')
}
</script>

<template>
  <div class="app-shell">
    <header v-if="showHeader" class="app-header">
      <router-link to="/" class="logo">人生模拟</router-link>
      <nav class="nav">
        <router-link to="/characters">人物列表</router-link>
      </nav>
      <div class="header-right">
        <span v-if="auth.user" class="user-label">{{ auth.user.display_name || auth.user.username }}</span>
        <el-button v-if="auth.enabled && auth.user" size="small" text @click="logout">退出登录</el-button>
        <span class="tagline">AI 演绎 · 非学术考证</span>
      </div>
    </header>
    <main class="app-main">
      <RouterView />
    </main>
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
