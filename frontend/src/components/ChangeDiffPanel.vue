<script setup lang="ts">
import type { NodeFieldChange } from '@/api/client'
import { fieldLabel } from '@/constants/fieldLabels'

defineProps<{
  changes: NodeFieldChange[]
  loading?: boolean
}>()
</script>

<template>
  <div v-loading="loading" class="diff-panel">
    <h4>变更对比</h4>
    <template v-if="changes.length">
      <el-card v-for="(c, i) in changes" :key="i" class="diff-item" size="small">
        <div class="meta">第 {{ c.sequence }} 节点 · {{ c.year }} 年 · {{ fieldLabel(c.field) }}</div>
        <div class="row before"><span>原</span>{{ c.before || '（空）' }}</div>
        <div class="row after"><span>新</span>{{ c.after }}</div>
      </el-card>
    </template>
    <el-empty v-else-if="!loading" description="选择版本并点击「对比」查看变更" :image-size="60" />
  </div>
</template>

<style scoped>
.diff-panel {
  min-height: 80px;
}
.diff-panel h4 {
  margin: 0 0 12px;
}
.diff-item {
  margin-bottom: 8px;
}
.meta {
  font-size: 0.85rem;
  color: #888;
  margin-bottom: 8px;
}
.row {
  padding: 6px 8px;
  border-radius: 4px;
  margin-bottom: 4px;
  font-size: 0.9rem;
}
.row span {
  display: inline-block;
  width: 24px;
  font-weight: bold;
  margin-right: 8px;
}
.before {
  background: #fef0f0;
}
.after {
  background: #f0f9eb;
}
</style>
