<script setup lang="ts">
import { useRouter } from 'vue-router'
import { User, MagicStick } from '@element-plus/icons-vue'

const router = useRouter()

function pick(mode: 'famous' | 'random') {
  router.push(mode === 'famous' ? '/game/famous' : '/game/random')
}
</script>

<template>
  <div class="page-card">
    <h1>人生游戏</h1>
    <p class="subtitle">选择主角类型，通过抉择逐步推演人生</p>
    <div class="cards">
      <el-card class="mode-card" shadow="hover" @click="pick('famous')">
        <el-icon :size="48" color="#409eff"><User /></el-icon>
        <h2>历史名人</h2>
        <p>从某一历史时期开始，在史实与分支间抉择</p>
      </el-card>
      <el-card class="mode-card" shadow="hover" @click="pick('random')">
        <el-icon :size="48" color="#67c23a"><MagicStick /></el-icon>
        <h2>随机普通人</h2>
        <p>从出生背景出发，书写属于自己的一生</p>
      </el-card>
    </div>
    <p class="back-link">
      <router-link to="/">← 返回首页</router-link>
    </p>
  </div>
</template>

<style scoped>
h1 {
  text-align: center;
  margin-bottom: 8px;
}
.subtitle {
  text-align: center;
  color: #666;
  margin-bottom: 28px;
}
.cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 20px;
}
.mode-card {
  cursor: pointer;
  padding: 20px;
  text-align: center;
}
.mode-card h2 {
  margin: 12px 0 8px;
}
.mode-card p {
  color: #666;
  font-size: 0.9rem;
}
.back-link {
  text-align: center;
  margin-top: 24px;
}
.back-link a {
  color: #409eff;
  text-decoration: none;
}
</style>
