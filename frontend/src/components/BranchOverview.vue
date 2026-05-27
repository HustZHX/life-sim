<script setup lang="ts">
import type { BranchNode, BranchOverviewEntry } from '@/api/client'
import type { ForkPoint } from '@/utils/forkTimelineLayout'
import ForkTimelineGraph from '@/components/ForkTimelineGraph.vue'

defineProps<{
  branches: BranchOverviewEntry[]
  branchRoots: BranchNode[]
  activeVersionId: string
  loading?: boolean
  activatingId?: string
  forkSequenceOverrides?: Record<string, number>
}>()

const emit = defineEmits<{
  activate: [versionId: string]
  back: []
  showDiff: [fork: ForkPoint]
  select: [payload: { versionId: string; sequence: number }]
}>()
</script>

<template>
  <div class="branch-overview" v-loading="loading">
    <div class="overview-toolbar">
      <div>
        <h3 class="overview-title">分支总览</h3>
        <p class="overview-sub">
          流程图展示共享主干、分叉点与「原本时间轴后续 / 新时间轴」双路径。点击节点可查看详情；点击分支可切换。
        </p>
      </div>
      <el-button type="primary" plain @click="emit('back')">← 返回时间轴</el-button>
    </div>

    <ForkTimelineGraph
      :branches="branches"
      :branch-roots="branchRoots"
      :active-version-id="activeVersionId"
      mode="expanded"
      :loading="loading"
      :fork-sequence-overrides="forkSequenceOverrides"
      @activate="emit('activate', $event)"
      @show-diff="emit('showDiff', $event)"
      @select="emit('select', $event)"
    />
  </div>
</template>

<style scoped>
.branch-overview {
  min-height: 320px;
  padding: 4px 0;
}
.overview-toolbar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 20px;
}
.overview-title {
  margin: 0 0 6px;
  font-size: 1.15rem;
}
.overview-sub {
  margin: 0;
  font-size: 0.88rem;
  color: #606266;
  line-height: 1.5;
  max-width: 52rem;
}
</style>
