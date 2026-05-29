import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { api, type AIModel } from '@/api/client'
import {
  DEFAULT_AI_MODEL,
  isAIModelId,
  normalizeAIModel,
  type AIModelId,
} from '@/constants/models'

const STORAGE_KEY = 'life-sim-ai-model'

function readStoredModel(): AIModelId {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw && isAIModelId(raw)) return raw
  } catch {
    // ignore
  }
  return DEFAULT_AI_MODEL
}

function writeStoredModel(id: AIModelId) {
  try {
    localStorage.setItem(STORAGE_KEY, id)
  } catch {
    // ignore
  }
}

export const useModelStore = defineStore('model', () => {
  const aiModel = ref<AIModelId>(readStoredModel())
  const models = ref<AIModel[]>([])
  const loading = ref(false)
  const initialized = ref(false)

  const currentLabel = computed(() => {
    const m = models.value.find((x) => x.id === aiModel.value)
    if (m) return m.label
    return aiModel.value === 'pro' ? 'V4 Pro' : 'V4 Flash'
  })

  /** 原 cascade / 世界线等场景与全局模型统一 */
  const cascadeModel = computed(() => aiModel.value)

  function setModel(id: AIModelId) {
    aiModel.value = normalizeAIModel(id)
    writeStoredModel(aiModel.value)
  }

  async function loadModels() {
    if (initialized.value && models.value.length > 0) return
    loading.value = true
    try {
      const res = await api.listModels()
      models.value = res.models.filter((m) => isAIModelId(m.id))
      aiModel.value = normalizeAIModel(aiModel.value)
      initialized.value = true
    } finally {
      loading.value = false
    }
  }

  return {
    aiModel,
    cascadeModel,
    models,
    loading,
    currentLabel,
    initialized,
    setModel,
    loadModels,
  }
})
