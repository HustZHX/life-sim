<script setup lang="ts">
import { computed } from 'vue'
import type { LightNovelJobEntry } from '@/composables/useLightNovelJobs'
import { lightNovelPersonLabel } from '@/constants/lightNovelPerson'
import { modelDisplayLabel } from '@/constants/models'

const props = defineProps<{
  modelValue: boolean
  jobs: LightNovelJobEntry[]
  loading?: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  refresh: []
  view: [entry: LightNovelJobEntry]
}>()

const activeCount = computed(
  () => props.jobs.filter((j) => j.status === 'pending' || j.status === 'running').length
)

function onClose() {
  emit('update:modelValue', false)
}

function formatTime(iso?: string): string {
  if (!iso) return '—'
  const d = new Date(iso)
  return Number.isNaN(d.getTime()) ? iso : d.toLocaleString()
}

function statusTagType(status: string): 'success' | 'danger' | 'warning' | 'info' {
  if (status === 'completed') return 'success'
  if (status === 'failed') return 'danger'
  if (status === 'running') return 'warning'
  return 'info'
}

function statusText(entry: LightNovelJobEntry): string {
  if (entry.status === 'completed') return '已完成'
  if (entry.status === 'failed') return '失败'
  if (entry.status === 'running') return '生成中'
  return '排队中'
}

function canView(entry: LightNovelJobEntry): boolean {
  return entry.status === 'completed' && !!entry.content
}
</script>

<template>
  <el-drawer
    :model-value="modelValue"
    title="轻小说生成队列"
    size="min(520px, 92vw)"
    destroy-on-close
    @update:model-value="onClose"
  >
    <div class="drawer-head">
      <p class="hint">
        生成任务在后台运行，可关闭此面板继续编辑时间轴。超过 5 个节点会自动分章撰写。
      </p>
      <div class="head-actions">
        <el-tag v-if="activeCount" type="warning" effect="plain">{{ activeCount }} 个进行中</el-tag>
        <el-button size="small" :loading="loading" @click="emit('refresh')">刷新</el-button>
      </div>
    </div>

    <div v-loading="loading" class="job-list">
      <el-empty v-if="!jobs.length && !loading" description="暂无生成任务，点击「生成轻小说」提交" />

      <div v-for="entry in jobs" :key="entry.id" class="job-card">
        <div class="job-card-head">
          <div class="job-meta">
            <span class="job-range">节点 {{ entry.fromSequence }}–{{ entry.toSequence }}</span>
            <el-tag size="small" effect="plain">{{ lightNovelPersonLabel(entry.person) }}</el-tag>
            <el-tag size="small" type="info" effect="plain">{{ modelDisplayLabel(entry.model) }}</el-tag>
          </div>
          <el-tag size="small" :type="statusTagType(entry.status)">{{ statusText(entry) }}</el-tag>
        </div>

        <el-progress
          v-if="entry.status === 'pending' || entry.status === 'running'"
          :percentage="Math.max(entry.progress, 5)"
          :stroke-width="8"
          striped
          striped-flow
        />
        <el-progress
          v-else-if="entry.status === 'completed'"
          :percentage="100"
          status="success"
          :stroke-width="8"
        />
        <el-progress v-else :percentage="entry.progress" status="exception" :stroke-width="8" />

        <p v-if="entry.stageText && entry.status === 'running'" class="stage">{{ entry.stageText }}</p>
        <p v-if="entry.error" class="error">{{ entry.error }}</p>
        <p class="time">{{ formatTime(entry.createdAt) }}</p>

        <div v-if="canView(entry)" class="job-actions">
          <el-button type="primary" link @click="emit('view', entry)">查看正文</el-button>
        </div>
      </div>
    </div>
  </el-drawer>
</template>

<style scoped>
.drawer-head {
  margin-bottom: 16px;
}

.hint {
  margin: 0 0 10px;
  font-size: 0.85rem;
  color: #909399;
  line-height: 1.5;
}

.head-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.job-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
  min-height: 120px;
}

.job-card {
  padding: 14px;
  border: 1px solid #ebeef5;
  border-radius: 10px;
  background: #fafafa;
}

.job-card-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
  margin-bottom: 10px;
}

.job-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
}

.job-range {
  font-weight: 600;
  font-size: 0.92rem;
  color: #303133;
}

.stage {
  margin: 8px 0 0;
  font-size: 0.82rem;
  color: #606266;
}

.error {
  margin: 8px 0 0;
  font-size: 0.82rem;
  color: #f56c6c;
}

.time {
  margin: 6px 0 0;
  font-size: 0.78rem;
  color: #a8abb2;
}

.job-actions {
  margin-top: 8px;
}
</style>
