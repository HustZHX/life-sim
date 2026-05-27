<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import JobProgress from '@/components/JobProgress.vue'
import { copyTextToClipboard } from '@/utils/exportTimelineText'

defineProps<{
  modelValue: boolean
  title: string
  content: string
  loading?: boolean
  progress?: number
  statusText?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  regenerate: []
}>()

const copying = ref(false)

function onClose() {
  emit('update:modelValue', false)
}

async function copyContent(text: string) {
  if (!text.trim()) {
    ElMessage.warning('暂无内容可复制')
    return
  }
  copying.value = true
  try {
    await copyTextToClipboard(text)
    ElMessage.success('已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败，请手动全选复制')
  } finally {
    copying.value = false
  }
}
</script>

<template>
  <el-dialog
    :model-value="modelValue"
    :title="title"
    width="min(760px, 94vw)"
    class="narrative-dialog"
    destroy-on-close
    @update:model-value="onClose"
  >
    <p v-if="!content && loading" class="narrative-hint">AI 正在撰写，请稍候…</p>
    <p v-else-if="!content" class="narrative-hint">暂无内容</p>
    <el-input
      v-else
      :model-value="content"
      type="textarea"
      :rows="20"
      readonly
      class="narrative-textarea"
    />
    <template #footer>
      <el-button @click="onClose">关闭</el-button>
      <el-button :loading="copying" :disabled="!content" @click="copyContent(content)">
        复制全部
      </el-button>
      <el-button type="warning" plain :loading="loading" @click="emit('regenerate')">
        重新生成
      </el-button>
    </template>
  </el-dialog>

  <JobProgress
    :visible="!!loading"
    :progress="progress ?? 0"
    :status-text="statusText ?? ''"
    title="叙事生成"
  />
</template>

<style scoped>
.narrative-hint {
  margin: 0;
  font-size: 0.9rem;
  color: #909399;
}
.narrative-textarea :deep(textarea) {
  font-family: 'PingFang SC', 'Microsoft YaHei', Georgia, serif;
  font-size: 0.92rem;
  line-height: 1.75;
}
</style>
