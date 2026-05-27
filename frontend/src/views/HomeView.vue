<script setup lang="ts">
import { useRouter } from 'vue-router'
import { User, MagicStick } from '@element-plus/icons-vue'

const router = useRouter()

function pick(mode: 'famous' | 'random') {
  router.push(mode === 'famous' ? '/famous' : '/random')
}
</script>

<template>
  <div class="home">
    <h1>选择一种人生</h1>
    <p class="subtitle">由 DeepSeek AI 演绎人物档案与人生节点，按寿命自动适配约 20 个关键节点</p>
    <p class="history-link">
      <router-link to="/history">查看生成历史 →</router-link>
    </p>
    <div class="cards">
      <el-card class="mode-card" shadow="hover" @click="pick('famous')">
        <el-icon :size="48" color="#409eff"><User /></el-icon>
        <h2>历史名人</h2>
        <p>输入姓名，AI 消歧确认身份，再生成史实风格的人生时间轴</p>
      </el-card>
      <el-card class="mode-card" shadow="hover" @click="pick('random')">
        <el-icon :size="48" color="#67c23a"><MagicStick /></el-icon>
        <h2>随机普通人</h2>
        <p>AI 随机生成虚构人物，可选导入名人的性格与时代背景</p>
      </el-card>
    </div>
  </div>
</template>

<style scoped>
.home {
  text-align: center;
}
.home h1 {
  margin-bottom: 8px;
  color: #16213e;
}
.subtitle {
  color: #666;
  margin-bottom: 12px;
}
.history-link {
  margin-bottom: 32px;
}
.history-link a {
  color: #409eff;
  text-decoration: none;
}
.history-link a:hover {
  text-decoration: underline;
}
.cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 24px;
}
.mode-card {
  cursor: pointer;
  padding: 24px;
  transition: transform 0.2s;
}
.mode-card:hover {
  transform: translateY(-4px);
}
.mode-card h2 {
  margin: 16px 0 8px;
}
.mode-card p {
  color: #666;
  font-size: 0.9rem;
}
</style>
