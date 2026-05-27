<script setup lang="ts">
import type { BranchOverviewEntry } from '@/api/client'

defineProps<{
  branches: BranchOverviewEntry[]
  loading?: boolean
  activatingId?: string
}>()

const emit = defineEmits<{
  activate: [versionId: string]
  back: []
}>()

function deathHint(b: BranchOverviewEntry): string {
  if (!b.death_year_snapshot || b.death_year_snapshot <= 0) return '在世'
  return `约 ${b.death_year_snapshot} 年`
}

function onBranchClick(b: BranchOverviewEntry) {
  if (b.is_active) return
  emit('activate', b.version_id)
}
</script>

<template>
  <div class="branch-overview" v-loading="loading">
    <div class="overview-toolbar">
      <div>
        <h3 class="overview-title">分支总览</h3>
        <p class="overview-sub">仅展示各分支节点年份与标题，点击分支可切换并返回时间轴</p>
      </div>
      <el-button type="primary" plain @click="emit('back')">← 返回时间轴</el-button>
    </div>

    <div v-if="!branches.length && !loading" class="overview-empty">暂无分支数据</div>

    <div v-else class="overview-grid">
      <section
        v-for="b in branches"
        :key="b.version_id"
        class="branch-column"
        :class="{ active: b.is_active, activating: activatingId === b.version_id }"
        @click="onBranchClick(b)"
      >
        <header class="column-head">
          <div class="head-row">
            <strong class="branch-name">{{ b.label }}</strong>
            <el-tag v-if="b.is_active" size="small" type="primary" effect="dark">当前</el-tag>
          </div>
          <div class="head-meta">
            <span>{{ b.nodes.length }} 节点</span>
            <span v-if="b.fork_sequence != null && b.fork_sequence >= 0"> · 分叉@#{{ b.fork_sequence + 1 }}</span>
            <span> · {{ deathHint(b) }}</span>
          </div>
        </header>

        <ol class="node-list">
          <li v-for="n in b.nodes" :key="`${b.version_id}-${n.sequence}`" class="node-row">
            <span class="node-year">{{ n.year }}</span>
            <span class="node-title">{{ n.title || '（无标题）' }}</span>
          </li>
        </ol>

        <p v-if="!b.nodes.length" class="no-nodes">（空分支）</p>
      </section>
    </div>
  </div>
</template>

<style scoped>
.branch-overview {
  min-height: 320px;
}
.overview-toolbar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}
.overview-title {
  margin: 0 0 4px;
  font-size: 1.1rem;
  color: #303133;
}
.overview-sub {
  margin: 0;
  font-size: 0.82rem;
  color: #909399;
}
.overview-empty {
  text-align: center;
  color: #c0c4cc;
  padding: 48px 0;
}
.overview-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 16px;
  align-items: start;
}
.branch-column {
  border: 1px solid #e4e7ed;
  border-radius: 10px;
  background: #fff;
  overflow: hidden;
  cursor: pointer;
  transition: border-color 0.2s, box-shadow 0.2s, transform 0.15s;
}
.branch-column:hover:not(.active) {
  border-color: #b3d4ff;
  box-shadow: 0 6px 16px rgba(64, 158, 255, 0.1);
  transform: translateY(-2px);
}
.branch-column.active {
  border-color: #409eff;
  box-shadow: 0 0 0 1px rgba(64, 158, 255, 0.35);
  cursor: default;
}
.branch-column.activating {
  opacity: 0.65;
}
.column-head {
  padding: 12px 14px;
  background: linear-gradient(180deg, #f5f9ff 0%, #fafbfc 100%);
  border-bottom: 1px solid #ebeef5;
}
.branch-column.active .column-head {
  background: linear-gradient(180deg, #ecf5ff 0%, #f0f7ff 100%);
}
.head-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.branch-name {
  font-size: 0.9rem;
  color: #303133;
  word-break: break-word;
}
.head-meta {
  margin-top: 6px;
  font-size: 0.72rem;
  color: #909399;
}
.node-list {
  list-style: none;
  margin: 0;
  padding: 8px 0;
  max-height: 420px;
  overflow-y: auto;
}
.node-row {
  display: flex;
  align-items: baseline;
  gap: 10px;
  padding: 8px 14px;
  border-bottom: 1px solid #f2f4f8;
  font-size: 0.85rem;
  line-height: 1.4;
}
.node-row:last-child {
  border-bottom: none;
}
.node-year {
  flex-shrink: 0;
  min-width: 3.2em;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  color: #409eff;
  font-size: 0.82rem;
}
.node-title {
  color: #303133;
  word-break: break-word;
}
.no-nodes {
  margin: 0;
  padding: 16px 14px;
  font-size: 0.82rem;
  color: #c0c4cc;
  text-align: center;
}
</style>
