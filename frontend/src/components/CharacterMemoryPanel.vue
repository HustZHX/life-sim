<script setup lang="ts">
import { ref } from 'vue'
import type { CharacterMemory } from '@/api/client'

const props = defineProps<{
  memories: CharacterMemory[]
  loading?: boolean
  readOnly?: boolean
  nodesBySequence?: Record<number, { year: number; age: number }>
}>()

const emit = defineEmits<{
  update: [id: string, content: string]
  remove: [id: string]
}>()

const editingId = ref<string | null>(null)
const editContent = ref('')

function startEdit(m: CharacterMemory) {
  editingId.value = m.id
  editContent.value = m.content
}

function cancelEdit() {
  editingId.value = null
  editContent.value = ''
}

function saveEdit(id: string) {
  const text = editContent.value.trim()
  if (!text) return
  emit('update', id, text)
  cancelEdit()
}

function nodeLabel(seq: number) {
  const n = props.nodesBySequence?.[seq]
  if (n) return `第 ${seq + 1} 节点 · ${n.year} 年 · ${n.age} 岁`
  return `第 ${seq + 1} 节点`
}
</script>

<template>
  <div class="memory-panel">
    <div class="panel-head">
      <span class="title">他对你的记忆</span>
      <span class="hint">仅显示此刻能想起的条目</span>
    </div>

    <div v-if="loading" class="empty">加载中…</div>
    <div v-else-if="!memories.length" class="empty">暂无记忆，可在对话中写入</div>

    <div v-else class="memory-list">
      <el-card v-for="m in memories" :key="m.id" shadow="never" class="memory-card">
        <div class="meta">
          <span>{{ nodeLabel(m.source_sequence) }}</span>
          <span v-if="m.speaker_identity" class="identity">· {{ m.speaker_identity }}</span>
        </div>

        <template v-if="editingId === m.id">
          <el-input v-model="editContent" type="textarea" :rows="3" />
          <div class="edit-actions">
            <el-button size="small" @click="cancelEdit">取消</el-button>
            <el-button size="small" type="primary" @click="saveEdit(m.id)">保存</el-button>
          </div>
        </template>
        <template v-else>
          <p class="content">{{ m.content }}</p>
          <div v-if="!readOnly" class="card-actions">
            <el-button link type="primary" size="small" @click="startEdit(m)">编辑</el-button>
            <el-button link type="danger" size="small" @click="emit('remove', m.id)">删除</el-button>
          </div>
        </template>
      </el-card>
    </div>
  </div>
</template>

<style scoped>
.memory-panel {
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-height: 120px;
}

.panel-head {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.title {
  font-weight: 600;
  font-size: 14px;
}

.hint {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.empty {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  padding: 12px 0;
}

.memory-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 280px;
  overflow-y: auto;
}

.memory-card {
  --el-card-padding: 10px;
}

.meta {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-bottom: 6px;
}

.identity {
  color: var(--el-color-primary);
}

.content {
  margin: 0;
  font-size: 13px;
  line-height: 1.5;
  white-space: pre-wrap;
}

.edit-actions,
.card-actions {
  margin-top: 8px;
  display: flex;
  gap: 8px;
}
</style>
