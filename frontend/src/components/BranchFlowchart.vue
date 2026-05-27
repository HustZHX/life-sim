<script setup lang="ts">
import { computed, ref } from 'vue'
import { ArrowDown, ArrowRight } from '@element-plus/icons-vue'
import type { BranchNode } from '@/api/client'
import BranchTreeNode from '@/components/BranchTreeNode.vue'

const props = defineProps<{
  roots: BranchNode[]
  activeVersionId?: string
  loading?: boolean
  activatingId?: string
}>()

const emit = defineEmits<{
  activate: [versionId: string]
}>()

const panelOpen = ref(true)
const collapsedIds = ref<Set<string>>(new Set())

const branchCount = computed(() => countBranches(props.roots))

function countBranches(nodes: BranchNode[]): number {
  let n = 0
  for (const b of nodes) {
    n++
    if (b.children?.length) n += countBranches(b.children)
  }
  return n
}

function togglePanel() {
  panelOpen.value = !panelOpen.value
}

function toggleCollapse(id: string) {
  const next = new Set(collapsedIds.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  collapsedIds.value = next
}

function onSelect(b: BranchNode) {
  if (b.is_active || props.loading || props.activatingId) return
  emit('activate', b.id)
}
</script>

<template>
  <div class="branch-flow" :class="{ collapsed: !panelOpen }">
    <button type="button" class="flow-head" @click="togglePanel">
      <span class="head-title">
        <el-icon class="chev"><ArrowDown v-if="panelOpen" /><ArrowRight v-else /></el-icon>
        分支时间轴
        <span class="count-badge">{{ branchCount }}</span>
      </span>
      <span v-if="!panelOpen" class="head-hint">点击展开</span>
    </button>

    <el-collapse-transition>
      <div v-show="panelOpen" v-loading="loading" class="flow-body">
        <p class="flow-tip">编辑节点会创建新分支并自动切换；点击卡片可查看其他分支。</p>
        <div v-if="!roots.length" class="flow-empty">暂无分支记录</div>
        <div v-else class="flow-canvas">
          <BranchTreeNode
            v-for="root in roots"
            :key="root.id"
            :node="root"
            :depth="0"
            :collapsed-ids="collapsedIds"
            :activating-id="activatingId"
            @toggle-collapse="toggleCollapse"
            @select="onSelect"
          />
        </div>
      </div>
    </el-collapse-transition>
  </div>
</template>

<style scoped>
.branch-flow {
  border: 1px solid #e8ecf4;
  border-radius: 10px;
  background: linear-gradient(165deg, #f8fafc 0%, #f3f6fb 55%, #eef4ff 100%);
  overflow: hidden;
  margin-bottom: 12px;
}
.branch-flow.collapsed {
  background: #fafbfc;
}
.flow-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 10px 14px;
  border: none;
  background: transparent;
  cursor: pointer;
  text-align: left;
}
.flow-head:hover {
  background: rgba(64, 158, 255, 0.06);
}
.head-title {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  font-size: 0.92rem;
  color: #303133;
}
.chev {
  color: #409eff;
}
.count-badge {
  font-size: 0.72rem;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 10px;
  background: rgba(64, 158, 255, 0.12);
  color: #409eff;
}
.head-hint {
  font-size: 0.78rem;
  color: #909399;
}
.flow-body {
  padding: 0 12px 12px;
  min-height: 48px;
}
.flow-tip {
  margin: 0 0 10px;
  font-size: 0.78rem;
  color: #909399;
  line-height: 1.45;
}
.flow-empty {
  font-size: 0.85rem;
  color: #c0c4cc;
  text-align: center;
  padding: 16px 0;
}
.flow-canvas {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
</style>
