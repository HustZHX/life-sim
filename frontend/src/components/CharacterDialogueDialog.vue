<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { ChatDotRound, Loading } from '@element-plus/icons-vue'
import ModelSelector from '@/components/ModelSelector.vue'
import CharacterMemoryPanel from '@/components/CharacterMemoryPanel.vue'
import type { DialogueSessionSummary, LifeNode } from '@/api/client'
import { useCharacterDialogue } from '@/composables/useCharacterDialogue'
import { fieldLabel } from '@/constants/fieldLabels'

const props = defineProps<{
  allNodes?: LifeNode[]
}>()

const emit = defineEmits<{
  'node-updated': [payload: { versionId: string; node: LifeNode }]
}>()

const {
  visible,
  step,
  loading,
  sending,
  identityLoading,
  identitySaved,
  historyLoading,
  sessionHistory,
  node,
  isActiveBranch,
  model,
  identityOptions,
  selectedIdentity,
  customIdentity,
  identityMode,
  sessionId,
  messages,
  inputText,
  pendingImpact,
  impactDialogVisible,
  applyingImpact,
  memories,
  memoriesLoading,
  resolvedIdentity,
  truncatePreview,
  openDialogue,
  openDialogueHistory,
  regenerateIdentityOptions,
  goToIdentityStep,
  goToHistoryStep,
  returnFromHistory,
  resumeHistoricalSession,
  startSession,
  sendMessage,
  saveMessageToMemory,
  applyImpact,
  dismissImpact,
  updateMemoryContent,
  removeMemory,
  close,
} = useCharacterDialogue()

const chatScrollRef = ref<HTMLElement | null>(null)

const nodesBySequence = computed(() => {
  const map: Record<number, { year: number; age: number }> = {}
  for (const n of props.allNodes ?? []) {
    map[n.sequence] = { year: n.year, age: n.age }
  }
  return map
})

const dialogTitle = computed(() => {
  if (step.value === 'history') return '历史对话'
  const n = node.value
  if (!n) return '与 TA 对话'
  return `与 TA 对话 · ${n.year} 年 · ${n.age} 岁`
})

function findNode(nodeId: string): LifeNode | undefined {
  return props.allNodes?.find((n) => n.id === nodeId)
}

function formatHistoryTime(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString('zh-CN', {
    month: 'numeric',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function nodeLabel(item: DialogueSessionSummary): string {
  if (item.node_year && item.node_title) {
    return `${item.node_year} 年 · ${item.node_age ?? '?'} 岁 · ${item.node_title}`
  }
  return `节点 #${item.node_sequence}`
}

async function onResumeSession(id: string) {
  await resumeHistoricalSession(id, findNode)
}

watch(
  () => [messages.value.length, sending.value],
  async () => {
    await nextTick()
    const el = chatScrollRef.value
    if (el) el.scrollTop = el.scrollHeight
  }
)

async function onApplyImpact() {
  const res = await applyImpact()
  if (res) {
    emit('node-updated', { versionId: res.version_id, node: res.node })
  }
}

defineExpose({ open: openDialogue, openHistory: openDialogueHistory })
</script>

<template>
  <el-dialog
    :model-value="visible"
    :title="dialogTitle"
    width="720px"
    class="dialogue-dialog"
    destroy-on-close
    @update:model-value="(v: boolean) => !v && close()"
  >
    <el-alert
      v-if="!isActiveBranch"
      type="info"
      :closable="false"
      show-icon
      title="当前查看的是非激活分支，可以对话；写入记忆或性格变更需切换到当前分支。"
      class="branch-alert"
    />

    <div v-if="step === 'identity'" class="identity-step">
      <ModelSelector v-model="model" class="model-row" />

      <div v-loading="identityLoading" class="identity-body">
        <div class="identity-head">
          <p class="step-label">选择你的身份</p>
          <el-button
            v-if="identityOptions.length"
            link
            type="primary"
            :loading="identityLoading"
            @click="regenerateIdentityOptions()"
          >
            重新生成身份选项
          </el-button>
        </div>
        <p v-if="identitySaved" class="muted saved-hint">已保存身份选项，再次打开无需重新生成</p>

        <el-radio-group v-model="identityMode" class="identity-modes">
          <el-radio value="option">上下文身份</el-radio>
          <el-radio value="reader">命运读者</el-radio>
          <el-radio value="custom">自定义</el-radio>
        </el-radio-group>

        <div v-if="identityMode === 'option'" class="option-list">
          <el-radio-group v-model="selectedIdentity">
            <el-card
              v-for="opt in identityOptions"
              :key="opt.label"
              shadow="never"
              class="opt-card"
            >
              <el-radio :value="opt.label">
                <strong>{{ opt.label }}</strong>
                <p class="opt-desc">{{ opt.description }}</p>
              </el-radio>
            </el-card>
          </el-radio-group>
          <p v-if="!identityOptions.length" class="muted">暂无生成选项，请使用命运读者或自定义身份</p>
        </div>

        <div v-else-if="identityMode === 'reader'" class="reader-box">
          <p>以「命运读者」身份交谈：通晓此人命运，但在当前时间节点与他对话，不主动剧透未来。</p>
        </div>

        <div v-else class="custom-box">
          <el-input
            v-model="customIdentity"
            placeholder="例如：远道而来的商人、少年时的玩伴…"
            maxlength="80"
            show-word-limit
          />
        </div>
      </div>

      <div class="dialog-footer">
        <el-button link type="primary" @click="goToHistoryStep()">历史对话</el-button>
        <div class="dialog-footer-right">
          <el-button @click="close()">取消</el-button>
          <el-button type="primary" :loading="loading" @click="startSession()">开始对话</el-button>
        </div>
      </div>
    </div>

    <div v-else-if="step === 'history'" v-loading="historyLoading" class="history-step">
      <p class="history-hint">当前版本下所有有过消息的对话，按最近活跃排序。</p>
      <div v-if="!sessionHistory.length && !historyLoading" class="empty-history">
        <el-icon><ChatDotRound /></el-icon>
        <span>暂无历史对话</span>
      </div>
      <el-card
        v-for="item in sessionHistory"
        :key="item.id"
        shadow="never"
        class="history-card"
      >
        <div class="history-card-head">
          <div class="history-meta">
            <strong>{{ nodeLabel(item) }}</strong>
            <span class="history-identity">身份：{{ item.speaker_identity }}</span>
            <span class="history-stats">{{ item.message_count }} 条消息 · {{ formatHistoryTime(item.updated_at) }}</span>
          </div>
          <el-button type="primary" plain size="small" :loading="loading" @click="onResumeSession(item.id)">
            继续对话
          </el-button>
        </div>
        <p v-if="item.last_message" class="history-preview">{{ truncatePreview(item.last_message) }}</p>
      </el-card>
      <div class="dialog-footer">
        <el-button @click="returnFromHistory()">
          {{ sessionId && messages.length ? '返回对话' : node ? '返回当前节点' : '关闭' }}
        </el-button>
      </div>
    </div>

    <div v-else-if="step === 'chat'" class="chat-step">
      <div class="chat-layout">
        <div class="chat-main">
          <div class="chat-head">
            <div class="chat-head-left">
              <span class="identity-tag">身份：{{ resolvedIdentity() }}</span>
              <el-button link type="primary" size="small" @click="goToIdentityStep()">
                更换身份
              </el-button>
              <el-button link type="primary" size="small" @click="goToHistoryStep()">
                历史对话
              </el-button>
            </div>
            <ModelSelector v-model="model" />
          </div>

          <div ref="chatScrollRef" class="messages">
            <div v-for="msg in messages" :key="msg.id" class="msg-row" :class="msg.role">
              <div class="bubble">{{ msg.content }}</div>
              <el-button
                v-if="msg.role === 'assistant'"
                link
                type="primary"
                size="small"
                class="mem-btn"
                :disabled="loading"
                @click="saveMessageToMemory(msg, resolvedIdentity())"
              >
                写入记忆
              </el-button>
            </div>
            <div v-if="sending" class="msg-row assistant typing">
              <div class="bubble typing-bubble">
                <el-icon class="typing-icon is-loading"><Loading /></el-icon>
                <span>说话中…</span>
              </div>
            </div>
            <div v-if="!messages.length && !sending" class="empty-chat">
              <el-icon><ChatDotRound /></el-icon>
              <span>向他/她说点什么吧</span>
            </div>
          </div>

          <div class="composer">
            <el-input
              v-model="inputText"
              type="textarea"
              :rows="3"
              placeholder="输入消息…"
              :disabled="sending"
              @keydown.enter.exact.prevent="sendMessage()"
            />
            <el-button type="primary" :loading="sending" :disabled="!inputText.trim()" @click="sendMessage()">
              发送
            </el-button>
          </div>
        </div>

        <aside class="chat-side">
          <CharacterMemoryPanel
            :memories="memories"
            :loading="memoriesLoading"
            :read-only="!isActiveBranch"
            :nodes-by-sequence="nodesBySequence"
            @update="(id, c) => updateMemoryContent(id, c)"
            @remove="(id) => removeMemory(id)"
          />
        </aside>
      </div>
    </div>

    <el-dialog
      v-model="impactDialogVisible"
      title="对话可能改变了他的性格或思想"
      width="480px"
      append-to-body
    >
      <p v-if="pendingImpact?.summary" class="impact-summary">{{ pendingImpact.summary }}</p>
      <div v-if="pendingImpact?.personality_snapshot" class="impact-block">
        <strong>性格快照</strong>
        <p>{{ pendingImpact.personality_snapshot }}</p>
      </div>
      <div v-if="pendingImpact?.thoughts" class="impact-block">
        <strong>内心想法</strong>
        <p>{{ pendingImpact.thoughts }}</p>
      </div>
      <div v-for="(t, i) in pendingImpact?.trait_changes ?? []" :key="i" class="impact-block">
        <strong>{{ fieldLabel(t.field, 'trait') }}</strong>
        <p>{{ t.before }} → {{ t.after }}</p>
        <p class="muted">{{ t.reason }}</p>
      </div>
      <template #footer>
        <el-button @click="dismissImpact()">忽略</el-button>
        <el-button
          type="primary"
          :loading="applyingImpact"
          :disabled="!isActiveBranch"
          @click="onApplyImpact"
        >
          写入当前节点
        </el-button>
      </template>
    </el-dialog>
  </el-dialog>
</template>

<style scoped>
.branch-alert {
  margin-bottom: 12px;
}

.identity-step {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.model-row {
  margin-bottom: 4px;
}

.identity-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 4px;
}

.step-label {
  font-weight: 600;
  margin: 0;
}

.saved-hint {
  margin: 0 0 8px;
}

.identity-modes {
  margin-bottom: 12px;
}

.option-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.opt-card {
  --el-card-padding: 10px;
}

.opt-desc {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--el-text-color-secondary);
  white-space: normal;
}

.reader-box,
.custom-box {
  font-size: 13px;
  color: var(--el-text-color-regular);
}

.muted {
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.dialog-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
}

.dialog-footer-right {
  display: flex;
  gap: 8px;
}

.history-step {
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-height: 320px;
}

.history-hint {
  margin: 0;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.empty-history {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 48px 0;
  color: var(--el-text-color-secondary);
  font-size: 14px;
}

.history-card {
  --el-card-padding: 12px;
}

.history-card-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}

.history-meta {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.history-identity,
.history-stats {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.history-preview {
  margin: 8px 0 0;
  font-size: 13px;
  color: var(--el-text-color-regular);
  white-space: pre-wrap;
  line-height: 1.5;
}

.chat-layout {
  display: grid;
  grid-template-columns: 1fr 220px;
  gap: 12px;
  min-height: 420px;
}

.chat-main {
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.chat-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 8px;
}

.chat-head-left {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-wrap: wrap;
}

.identity-tag {
  font-size: 13px;
  color: var(--el-color-primary);
}

.messages {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
  background: var(--el-fill-color-lighter);
  border-radius: 8px;
  min-height: 260px;
  max-height: 360px;
}

.msg-row {
  display: flex;
  flex-direction: column;
  margin-bottom: 12px;
  max-width: 92%;
}

.msg-row.user {
  align-self: flex-end;
  align-items: flex-end;
}

.msg-row.assistant {
  align-self: flex-start;
  align-items: flex-start;
}

.bubble {
  padding: 8px 12px;
  border-radius: 10px;
  font-size: 14px;
  line-height: 1.55;
  white-space: pre-wrap;
}

.user .bubble {
  background: var(--el-color-primary-light-8);
}

.assistant .bubble {
  background: #fff;
  border: 1px solid var(--el-border-color-lighter);
}

.typing-bubble {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: var(--el-text-color-secondary);
}

.typing-icon {
  font-size: 16px;
}

.mem-btn {
  margin-top: 2px;
}

.empty-chat {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 48px 0;
  color: var(--el-text-color-secondary);
  font-size: 14px;
}

.composer {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 10px;
}

.chat-side {
  border-left: 1px solid var(--el-border-color-lighter);
  padding-left: 10px;
}

.impact-summary {
  font-weight: 500;
  margin-bottom: 12px;
}

.impact-block {
  margin-bottom: 10px;
  font-size: 13px;
}

.impact-block p {
  margin: 4px 0 0;
  white-space: pre-wrap;
}

@media (max-width: 640px) {
  .chat-layout {
    grid-template-columns: 1fr;
  }
  .chat-side {
    border-left: none;
    padding-left: 0;
    border-top: 1px solid var(--el-border-color-lighter);
    padding-top: 10px;
  }
}
</style>
