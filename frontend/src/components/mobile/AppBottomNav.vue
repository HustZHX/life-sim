<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

function isActive(path: string) {
  return route.path === path || route.path.startsWith(path + '/')
}
</script>

<template>
  <nav class="bottom-nav" aria-label="mobile bottom navigation">
    <button class="nav-item" :class="{ active: isActive('/') }" @click="router.push('/')">首页</button>
    <button
      class="nav-item"
      :class="{ active: isActive('/characters') }"
      @click="router.push('/characters')"
    >
      人物
    </button>
  </nav>
</template>

<style scoped>
.bottom-nav {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  height: var(--layout-bottom-nav-h);
  padding: 8px 12px calc(8px + env(safe-area-inset-bottom));
  display: flex;
  gap: 10px;
  justify-content: center;
  align-items: center;
  background: rgba(255, 255, 255, 0.92);
  border-top: 1px solid #e4e7ed;
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  z-index: 1000;
}

.nav-item {
  flex: 1;
  max-width: 220px;
  height: var(--layout-touch-min);
  border-radius: 12px;
  border: 1px solid #dcdfe6;
  background: #fff;
  color: #303133;
  font-weight: 600;
  letter-spacing: 0.2px;
}

.nav-item.active {
  border-color: #409eff;
  color: #409eff;
  background: #ecf5ff;
}
</style>
