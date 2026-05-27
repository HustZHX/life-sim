import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import {
  api,
  DIALOGUE_IDENTITY_READER,
  type AIModelId,
  type CharacterMemory,
  type DialogueIdentityOption,
  type DialogueMessage,
  type DialogueSession,
  type DialogueSessionSummary,
  type LifeNode,
  type PersonalityImpact,
} from '@/api/client'
import { DEFAULT_AI_MODEL } from '@/constants/models'

export function useCharacterDialogue() {
  const visible = ref(false)
  const step = ref<'identity' | 'chat' | 'history'>('identity')
  const loading = ref(false)
  const sending = ref(false)
  const identityLoading = ref(false)
  const identitySaved = ref(false)
  const historyLoading = ref(false)
  const sessionHistory = ref<DialogueSessionSummary[]>([])

  const characterId = ref('')
  const versionId = ref('')
  const node = ref<LifeNode | null>(null)
  const isActiveBranch = ref(true)

  const model = ref<AIModelId>(DEFAULT_AI_MODEL)
  const identityOptions = ref<DialogueIdentityOption[]>([])
  const selectedIdentity = ref('')
  const customIdentity = ref('')
  const identityMode = ref<'option' | 'reader' | 'custom'>('option')

  const sessionId = ref('')
  const messages = ref<DialogueMessage[]>([])
  const inputText = ref('')

  const pendingImpact = ref<PersonalityImpact | null>(null)
  const impactDialogVisible = ref(false)
  const applyingImpact = ref(false)

  const memories = ref<CharacterMemory[]>([])
  const memoriesLoading = ref(false)

  const activeVersionId = ref('')

  function truncatePreview(text: string, max = 80): string {
    const t = text.trim()
    if (t.length <= max) return t
    return `${t.slice(0, max)}…`
  }

  function resolvedIdentity(): string {
    if (identityMode.value === 'reader') return DIALOGUE_IDENTITY_READER
    if (identityMode.value === 'custom') return customIdentity.value.trim()
    return selectedIdentity.value.trim()
  }

  function applySessionIdentity(sess: DialogueSession) {
    const identity = sess.speaker_identity
    if (identity === DIALOGUE_IDENTITY_READER) {
      identityMode.value = 'reader'
      return
    }
    if (identityOptions.value.some((o) => o.label === identity)) {
      identityMode.value = 'option'
      selectedIdentity.value = identity
      return
    }
    identityMode.value = 'custom'
    customIdentity.value = identity
  }

  async function loadIdentityOptions(charId: string, nodeId: string, forceGenerate = false) {
    if (!forceGenerate) {
      const saved = await api.getSavedDialogueIdentityOptions(charId, nodeId)
      identitySaved.value = saved.has_saved
      if (saved.options?.length) {
        identityOptions.value = saved.options
        if (!selectedIdentity.value && saved.options[0]) {
          selectedIdentity.value = saved.options[0].label
        }
        return
      }
    }
    const res = await api.generateDialogueIdentityOptions(charId, nodeId, model.value, forceGenerate)
    identityOptions.value = res.options ?? []
    identitySaved.value = identityOptions.value.length > 0
    if (identityOptions.value.length > 0 && !selectedIdentity.value) {
      selectedIdentity.value = identityOptions.value[0].label
    }
  }

  async function loadSessionHistory() {
    if (!characterId.value) return
    historyLoading.value = true
    try {
      const res = await api.listDialogueSessions(characterId.value, versionId.value)
      sessionHistory.value = res.items ?? []
    } catch (e: unknown) {
      sessionHistory.value = []
      ElMessage.error(e instanceof Error ? e.message : '加载历史对话失败')
    } finally {
      historyLoading.value = false
    }
  }

  async function goToHistoryStep() {
    step.value = 'history'
    await loadSessionHistory()
  }

  async function resolveNodeForSession(
    sess: DialogueSession,
    findNode?: (id: string) => LifeNode | undefined
  ): Promise<LifeNode> {
    const found = findNode?.(sess.node_id)
    if (found) return found
    return api.getNode(characterId.value, sess.node_id)
  }

  async function resumeHistoricalSession(
    targetSessionId: string,
    findNode?: (id: string) => LifeNode | undefined
  ) {
    loading.value = true
    try {
      const sess = await api.getDialogueSession(characterId.value, targetSessionId)
      const n = await resolveNodeForSession(sess, findNode)
      node.value = n
      versionId.value = sess.version_id
      isActiveBranch.value = sess.version_id === activeVersionId.value
      sessionId.value = sess.id
      messages.value = sess.messages ?? []
      if (sess.model) model.value = sess.model

      await loadIdentityOptions(characterId.value, n.id, false)
      applySessionIdentity(sess)
      step.value = 'chat'
      await loadMemories()
    } catch (e: unknown) {
      ElMessage.error(e instanceof Error ? e.message : '恢复对话失败')
    } finally {
      loading.value = false
    }
  }

  async function tryResumeLatestSession(charId: string, nodeId: string) {
    const res = await api.getLatestDialogueSession(charId, nodeId)
    const sess = res.session
    if (!sess?.messages?.length) {
      step.value = 'identity'
      sessionId.value = ''
      messages.value = []
      return
    }
    applySessionIdentity(sess)
    sessionId.value = sess.id
    messages.value = sess.messages ?? []
    step.value = 'chat'
  }

  async function openDialogue(
    charId: string,
    n: LifeNode,
    currentVersionId: string,
    activeVersionIdParam: string
  ) {
    characterId.value = charId
    node.value = n
    versionId.value = currentVersionId
    activeVersionId.value = activeVersionIdParam
    isActiveBranch.value = currentVersionId === activeVersionIdParam
    visible.value = true
    pendingImpact.value = null
    identityMode.value = 'option'
    selectedIdentity.value = ''
    customIdentity.value = ''
    inputText.value = ''

    identityLoading.value = true
    try {
      await loadIdentityOptions(charId, n.id, false)
      await tryResumeLatestSession(charId, n.id)
    } catch (e: unknown) {
      ElMessage.error(e instanceof Error ? e.message : '加载对话失败')
      identityOptions.value = []
      step.value = 'identity'
      sessionId.value = ''
      messages.value = []
    } finally {
      identityLoading.value = false
    }

    await loadMemories()
  }

  async function openDialogueHistory(
    charId: string,
    currentVersionId: string,
    activeVersionIdParam: string
  ) {
    characterId.value = charId
    node.value = null
    versionId.value = currentVersionId
    activeVersionId.value = activeVersionIdParam
    isActiveBranch.value = currentVersionId === activeVersionIdParam
    visible.value = true
    pendingImpact.value = null
    inputText.value = ''
    sessionId.value = ''
    messages.value = []
    step.value = 'history'
    await loadSessionHistory()
  }

  async function regenerateIdentityOptions() {
    if (!node.value) return
    identityLoading.value = true
    try {
      await loadIdentityOptions(characterId.value, node.value.id, true)
      ElMessage.success('已重新生成身份选项')
    } catch (e: unknown) {
      ElMessage.error(e instanceof Error ? e.message : '重新生成失败')
    } finally {
      identityLoading.value = false
    }
  }

  function goToIdentityStep() {
    step.value = 'identity'
  }

  function returnFromHistory() {
    if (sessionId.value && messages.value.length) {
      step.value = 'chat'
      return
    }
    if (node.value) {
      step.value = 'identity'
      return
    }
    close()
  }

  async function loadMemories() {
    if (!characterId.value || !versionId.value || !node.value) return
    memoriesLoading.value = true
    try {
      const res = await api.listMemories(characterId.value, versionId.value, node.value.sequence)
      memories.value = res.items ?? []
    } catch {
      memories.value = []
    } finally {
      memoriesLoading.value = false
    }
  }

  async function startSession() {
    const identity = resolvedIdentity()
    if (!identity) {
      ElMessage.warning('请选择或填写身份')
      return
    }
    if (!node.value) return
    loading.value = true
    try {
      const sess = await api.createDialogueSession(characterId.value, node.value.id, {
        identity,
        model: model.value,
        resume: true,
      })
      sessionId.value = sess.id
      messages.value = sess.messages ?? []
      step.value = 'chat'
    } catch (e: unknown) {
      ElMessage.error(e instanceof Error ? e.message : '创建会话失败')
    } finally {
      loading.value = false
    }
  }

  async function sendMessage() {
    const text = inputText.value.trim()
    if (!text || !sessionId.value || sending.value) return

    const tempId = `temp-user-${Date.now()}`
    const optimisticUser: DialogueMessage = {
      id: tempId,
      session_id: sessionId.value,
      role: 'user',
      content: text,
      created_at: new Date().toISOString(),
    }

    messages.value.push(optimisticUser)
    inputText.value = ''
    sending.value = true

    try {
      const res = await api.sendDialogueMessage(
        characterId.value,
        sessionId.value,
        text,
        model.value
      )
      const idx = messages.value.findIndex((m) => m.id === tempId)
      if (idx >= 0) {
        messages.value[idx] = res.user_message
      } else {
        messages.value.push(res.user_message)
      }
      messages.value.push(res.assistant_message)
      if (res.impact?.has_impact) {
        pendingImpact.value = res.impact
        impactDialogVisible.value = true
      }
    } catch (e: unknown) {
      messages.value = messages.value.filter((m) => m.id !== tempId)
      inputText.value = text
      ElMessage.error(e instanceof Error ? e.message : '发送失败')
    } finally {
      sending.value = false
    }
  }

  async function saveMessageToMemory(msg: DialogueMessage, speakerIdentity: string) {
    if (!node.value || !isActiveBranch.value) {
      ElMessage.warning('请在当前激活分支上写入记忆')
      return
    }
    loading.value = true
    try {
      let content = msg.content
      try {
        const summarized = await api.summarizeMemory(characterId.value, {
          text: msg.content,
          identity: speakerIdentity,
          model: model.value,
        })
        if (summarized.content.trim()) content = summarized.content
      } catch {
        // 摘要失败则存原文
      }
      await api.createMemory(characterId.value, {
        version_id: versionId.value,
        source_node_id: node.value.id,
        source_sequence: node.value.sequence,
        speaker_identity: speakerIdentity,
        content,
      })
      ElMessage.success('已写入记忆')
      await loadMemories()
    } catch (e: unknown) {
      ElMessage.error(e instanceof Error ? e.message : '写入记忆失败')
    } finally {
      loading.value = false
    }
  }

  async function applyImpact() {
    if (!pendingImpact.value || !node.value) return
    applyingImpact.value = true
    try {
      const res = await api.applyDialogueImpact(characterId.value, node.value.id, {
        thoughts: pendingImpact.value.thoughts ?? '',
        personality_snapshot: pendingImpact.value.personality_snapshot ?? '',
        trait_changes: pendingImpact.value.trait_changes,
        summary: pendingImpact.value.summary,
      })
      versionId.value = res.version_id
      node.value = res.node
      impactDialogVisible.value = false
      pendingImpact.value = null
      ElMessage.success('已写入当前节点性格/思想变更')
      return res
    } catch (e: unknown) {
      ElMessage.error(e instanceof Error ? e.message : '写入失败')
      return null
    } finally {
      applyingImpact.value = false
    }
  }

  function dismissImpact() {
    impactDialogVisible.value = false
    pendingImpact.value = null
  }

  async function updateMemoryContent(memoryId: string, content: string) {
    if (!isActiveBranch.value) {
      ElMessage.warning('请在当前激活分支上编辑记忆')
      return
    }
    try {
      await api.updateMemory(characterId.value, memoryId, content)
      await loadMemories()
      ElMessage.success('记忆已更新')
    } catch (e: unknown) {
      ElMessage.error(e instanceof Error ? e.message : '更新失败')
    }
  }

  async function removeMemory(memoryId: string) {
    if (!isActiveBranch.value) {
      ElMessage.warning('请在当前激活分支上删除记忆')
      return
    }
    try {
      await api.deleteMemory(characterId.value, memoryId)
      await loadMemories()
      ElMessage.success('已删除')
    } catch (e: unknown) {
      ElMessage.error(e instanceof Error ? e.message : '删除失败')
    }
  }

  function close() {
    visible.value = false
  }

  return {
    visible,
    step,
    loading,
    sending,
    identityLoading,
    identitySaved,
    historyLoading,
    sessionHistory,
    characterId,
    versionId,
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
    loadSessionHistory,
    loadMemories,
    startSession,
    sendMessage,
    saveMessageToMemory,
    applyImpact,
    dismissImpact,
    updateMemoryContent,
    removeMemory,
    close,
  }
}
