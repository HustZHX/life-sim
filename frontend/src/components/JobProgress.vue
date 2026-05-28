<script setup lang="ts">
import { CircleCloseFilled, Loading } from '@element-plus/icons-vue'

defineProps<{
  visible: boolean
  progress: number
  statusText: string
  title?: string
  failed?: boolean
  errorText?: string
  retrying?: boolean
}>()

defineEmits<{
  retry: []
  dismiss: []
}>()
</script>

<template>
  <Teleport to="body">
    <Transition name="job-float">
      <div v-if="visible" class="job-progress-float" role="status" aria-live="polite">
        <div class="job-progress-inner" :class="{ 'job-progress-inner--failed': failed }">
          <el-icon v-if="failed" class="fail-icon" :size="22">
            <CircleCloseFilled />
          </el-icon>
          <el-icon v-else class="spinner" :size="22">
            <Loading />
          </el-icon>
          <div class="job-body">
            <div class="job-head">
              <span class="job-title">{{ failed ? '任务失败' : title || '任务进行中' }}</span>
              <span v-if="!failed" class="job-percent">{{ progress }}%</span>
            </div>
            <p v-if="failed && errorText" class="error-text">{{ errorText }}</p>
            <p v-else-if="statusText" class="status">{{ statusText }}</p>
            <el-progress
              v-if="!failed"
              :percentage="progress"
              :stroke-width="6"
              :show-text="false"
              striped
              striped-flow
            />
            <el-progress v-else :percentage="progress" status="exception" :stroke-width="6" :show-text="false" />
            <div v-if="failed" class="job-actions">
              <el-button size="small" type="primary" :loading="retrying" @click="$emit('retry')">
                重试
              </el-button>
              <el-button size="small" @click="$emit('dismiss')">关闭</el-button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.job-progress-float {
  position: fixed;
  top: 64px;
  left: 16px;
  z-index: 2000;
  max-width: min(360px, calc(100vw - 32px));
}

.job-progress-inner {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 14px 16px;
  background: #fff;
  border: 1px solid #dcdfe6;
  border-radius: 10px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.12);
}

.job-progress-inner--failed {
  border-color: #fde2e2;
  background: #fef0f0;
}

.spinner {
  flex-shrink: 0;
  margin-top: 2px;
  color: #409eff;
  animation: spin 1s linear infinite;
}

.fail-icon {
  flex-shrink: 0;
  margin-top: 2px;
  color: #f56c6c;
}

.job-body {
  flex: 1;
  min-width: 0;
}

.job-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 4px;
}

.job-title {
  font-weight: 600;
  font-size: 0.9rem;
  color: #303133;
}

.job-percent {
  font-size: 0.85rem;
  font-weight: 600;
  color: #409eff;
  font-variant-numeric: tabular-nums;
}

.status {
  margin: 0 0 8px;
  font-size: 0.8rem;
  color: #606266;
  line-height: 1.4;
}

.error-text {
  margin: 0 0 8px;
  font-size: 0.8rem;
  color: #f56c6c;
  line-height: 1.45;
  word-break: break-word;
}

.job-actions {
  display: flex;
  gap: 8px;
  margin-top: 10px;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.job-float-enter-active,
.job-float-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.job-float-enter-from,
.job-float-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
