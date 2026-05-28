<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { ArrowDown } from '@element-plus/icons-vue'
import JobProgress from '@/components/JobProgress.vue'
import { copyTextToClipboard } from '@/utils/exportTimelineText'
import {
  downloadNarrative,
  type NarrativeExportFormat,
} from '@/utils/downloadNarrative'

const props = defineProps<{
  modelValue: boolean
  title: string
  content: string
  loading?: boolean
  progress?: number
  statusText?: string
  failed?: boolean
  errorText?: string
  retrying?: boolean
  /** 轻小说等场景：显示导出下载 */
  enableExport?: boolean
  /** 下载文件名（不含扩展名） */
  exportBaseName?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  regenerate: []
  retry: []
  dismissJob: []
}>()

const copying = ref(false)
const exporting = ref(false)

const exportLabels: Record<NarrativeExportFormat, string> = {
  txt: 'TXT 文本',
  md: 'Markdown',
  pdf: 'PDF',
}

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

async function onExport(format: NarrativeExportFormat) {
  if (!props.content.trim()) {
    ElMessage.warning('暂无内容可导出')
    return
  }
  exporting.value = true
  try {
    const base = props.exportBaseName?.trim() || props.title.trim() || '轻小说'
    await downloadNarrative(format, props.title, props.content, base)
    ElMessage.success(`已下载 ${exportLabels[format]} 文件`)
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '导出失败')
  } finally {
    exporting.value = false
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
      <el-dropdown
        v-if="enableExport"
        trigger="click"
        :disabled="!content || exporting"
        @command="onExport"
      >
        <el-button type="primary" :loading="exporting">
          导出下载
          <el-icon class="el-icon--right"><ArrowDown /></el-icon>
        </el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="txt">TXT 文本</el-dropdown-item>
            <el-dropdown-item command="md">Markdown (.md)</el-dropdown-item>
            <el-dropdown-item command="pdf">PDF 文档</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
      <el-button type="warning" plain :loading="loading" @click="emit('regenerate')">
        重新生成
      </el-button>
    </template>
  </el-dialog>

  <JobProgress
    :visible="!!loading || !!failed"
    :progress="progress ?? 0"
    :status-text="statusText ?? ''"
    title="叙事生成"
    :failed="failed"
    :error-text="errorText"
    :retrying="retrying"
    @retry="emit('retry')"
    @dismiss="emit('dismissJob')"
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
