<script setup lang="ts">
import { RouterView } from 'vue-router'
</script>

<template>
  <div class="app-shell">
    <header class="app-header">
      <router-link to="/" class="logo">人生模拟</router-link>
      <nav class="nav">
        <router-link to="/history">生成历史</router-link>
      </nav>
      <span class="tagline">AI 演绎 · 非学术考证</span>
    </header>
    <main class="app-main">
      <RouterView />
    </main>
  </div>
</template>
