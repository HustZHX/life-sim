<script setup lang="ts">
import type { TimelineVersion } from '@/api/client'

defineProps<{
  versions: TimelineVersion[]
  currentVersionId?: string
  loading?: boolean
  diffLoading?: boolean
  diffVersionId?: string
}>()

const emit = defineEmits<{
  select: [version: TimelineVersion]
  rollback: [version: TimelineVersion]
  diff: [version: TimelineVersion]
}>()
</script>

<template>
  <div class="history">
    <h4>版本历史</h4>
    <el-timeline>
      <el-timeline-item
        v-for="v in versions"
        :key="v.id"
        :type="v.id === currentVersionId ? 'primary' : v.id === diffVersionId ? 'warning' : 'info'"
      >
        <div class="ver-row">
          <span class="summary">{{ v.change_summary || '版本' }}</span>
          <span class="time">{{ new Date(v.created_at).toLocaleString() }}</span>
        </div>
        <div class="actions">
          <el-button size="small" :loading="diffLoading && diffVersionId === v.id" @click="emit('diff', v)">
            对比
          </el-button>
          <el-button
            v-if="v.id !== currentVersionId"
            size="small"
            type="warning"
            :loading="loading"
            @click="emit('rollback', v)"
          >
            回溯
          </el-button>
          <el-tag v-else size="small">当前</el-tag>
        </div>
      </el-timeline-item>
    </el-timeline>
  </div>
</template>

<style scoped>
.history h4 {
  margin: 0 0 12px;
}
.ver-row {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.summary {
  font-weight: 500;
}
.time {
  font-size: 0.8rem;
  color: #999;
}
.actions {
  margin-top: 8px;
  display: flex;
  gap: 8px;
  align-items: center;
}
</style>
