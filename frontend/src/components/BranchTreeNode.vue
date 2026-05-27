<script setup lang="ts">
import { ArrowDown, ArrowRight } from '@element-plus/icons-vue'
import type { BranchNode } from '@/api/client'

defineProps<{
  node: BranchNode
  depth: number
  collapsedIds: Set<string>
  activatingId?: string
}>()

const emit = defineEmits<{
  toggleCollapse: [id: string]
  select: [node: BranchNode]
}>()

function deathLabel(b: BranchNode): string {
  if (!b.death_year_snapshot || b.death_year_snapshot <= 0) return '在世'
  return `约 ${b.death_year_snapshot} 年`
}
</script>

<template>
  <div class="branch-level" :style="{ '--depth': depth }">
    <div v-if="depth > 0" class="connector-col">
      <span class="fork-line" />
      <span class="fork-dot" />
    </div>
    <div class="node-col">
      <article
        class="branch-card"
        :class="{ active: node.is_active, activating: activatingId === node.id }"
        @click="emit('select', node)"
      >
        <div class="card-stripe" />
        <div class="card-main">
          <div class="card-top">
            <strong class="card-label">{{ node.label }}</strong>
            <el-tag v-if="node.is_active" size="small" type="primary" effect="dark">当前</el-tag>
          </div>
          <p v-if="node.change_summary && node.change_summary !== node.label" class="card-summary">
            {{ node.change_summary }}
          </p>
          <div class="card-meta">
            <span>{{ node.node_count }} 节点</span>
              <span v-if="node.fork_node_id">· 分叉@#{{ (node.fork_sequence ?? 0) + 1 }}</span>
            <span>· {{ deathLabel(node) }}</span>
          </div>
        </div>
        <button
          v-if="node.children?.length"
          type="button"
          class="collapse-btn"
          @click.stop="emit('toggleCollapse', node.id)"
        >
          <el-icon><ArrowDown v-if="!collapsedIds.has(node.id)" /><ArrowRight v-else /></el-icon>
          {{ node.children.length }}
        </button>
      </article>
      <div v-if="node.children?.length && !collapsedIds.has(node.id)" class="children-wrap">
        <BranchTreeNode
          v-for="child in node.children"
          :key="child.id"
          :node="child"
          :depth="depth + 1"
          :collapsed-ids="collapsedIds"
          :activating-id="activatingId"
          @toggle-collapse="emit('toggleCollapse', $event)"
          @select="emit('select', $event)"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
.branch-level {
  display: flex;
  gap: 0;
  padding-left: calc(var(--depth) * 14px);
}
.connector-col {
  width: 18px;
  position: relative;
  flex-shrink: 0;
}
.fork-line {
  position: absolute;
  left: 8px;
  top: 0;
  bottom: 50%;
  width: 2px;
  background: linear-gradient(180deg, #c5d5f0, #a8c4e8);
  border-radius: 1px;
}
.fork-dot {
  position: absolute;
  left: 4px;
  top: 22px;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #fff;
  border: 2px solid #7eb0e8;
}
.node-col {
  flex: 1;
  min-width: 0;
}
.branch-card {
  display: flex;
  align-items: stretch;
  gap: 8px;
  border: 1px solid #dce4f0;
  border-radius: 8px;
  background: #fff;
  cursor: pointer;
  transition: border-color 0.2s, box-shadow 0.2s, transform 0.15s;
  overflow: hidden;
}
.branch-card:hover:not(.active) {
  border-color: #b3d4ff;
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.12);
  transform: translateY(-1px);
}
.branch-card.active {
  border-color: #409eff;
  box-shadow: 0 0 0 1px rgba(64, 158, 255, 0.35);
}
.branch-card.activating {
  opacity: 0.7;
}
.card-stripe {
  width: 4px;
  flex-shrink: 0;
  background: linear-gradient(180deg, #a8c8f0, #6ba3e8);
}
.branch-card.active .card-stripe {
  background: linear-gradient(180deg, #409eff, #2563eb);
}
.card-main {
  flex: 1;
  padding: 8px 4px 8px 0;
  min-width: 0;
}
.card-top {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}
.card-label {
  font-size: 0.86rem;
  color: #303133;
  word-break: break-word;
}
.card-summary {
  margin: 4px 0 0;
  font-size: 0.76rem;
  color: #606266;
  line-height: 1.4;
  word-break: break-word;
}
.card-meta {
  margin-top: 4px;
  font-size: 0.72rem;
  color: #909399;
}
.collapse-btn {
  flex-shrink: 0;
  align-self: center;
  margin-right: 8px;
  display: inline-flex;
  align-items: center;
  gap: 2px;
  padding: 4px 8px;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  background: #f5f7fa;
  font-size: 0.72rem;
  color: #606266;
  cursor: pointer;
}
.children-wrap {
  margin-top: 6px;
  padding-left: 4px;
  border-left: 2px dashed #e4eaf4;
  margin-left: 6px;
}
</style>
