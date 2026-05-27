/** 用户可选的 AI 模型（与 GET /api/v1/models 一致） */
export const AI_MODEL_IDS = ['flash', 'pro'] as const
export type AIModelId = (typeof AI_MODEL_IDS)[number]

export const DEFAULT_AI_MODEL: AIModelId = 'flash'

const VALID = new Set<string>(AI_MODEL_IDS)

export function isAIModelId(id: string): id is AIModelId {
  return VALID.has(id)
}

/** 将任意历史/非法值规范为 flash 或 pro */
export function normalizeAIModel(id?: string | null): AIModelId {
  if (id === 'pro') return 'pro'
  return DEFAULT_AI_MODEL
}

export function modelDisplayLabel(id?: string | null): string {
  return normalizeAIModel(id) === 'pro' ? 'Pro' : 'Flash'
}
