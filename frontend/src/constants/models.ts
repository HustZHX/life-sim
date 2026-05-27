/** 用户可选的 AI 模型（与 GET /api/v1/models 一致） */
export const AI_MODEL_IDS = ['flash', 'pro'] as const
export type AIModelId = (typeof AI_MODEL_IDS)[number]

export type NarrativeDensity = 'standard' | 'rich'

export const DEFAULT_AI_MODEL: AIModelId = 'flash'

/** 推演余生默认使用 Pro */
export const DEFAULT_CASCADE_MODEL: AIModelId = 'pro'

const VALID = new Set<string>(AI_MODEL_IDS)

export function isAIModelId(id: string): id is AIModelId {
  return VALID.has(id)
}

/** 将任意历史/非法值规范为 flash 或 pro */
export function normalizeAIModel(id?: string | null): AIModelId {
  if (id === 'pro') return 'pro'
  return DEFAULT_AI_MODEL
}

export const DEFAULT_NARRATIVE_DENSITY: NarrativeDensity = 'rich'

export function normalizeNarrativeDensity(d?: string | null): NarrativeDensity {
  return d === 'standard' ? 'standard' : 'rich'
}

export function narrativeDensityLabel(d: NarrativeDensity): string {
  return d === 'rich' ? '细腻（两阶段叙事）' : '标准（单次生成）'
}

export function estimateTimelineDuration(
  model: AIModelId,
  density: NarrativeDensity,
  targetNodes: number
): string {
  const batches = density === 'rich' ? Math.ceil(targetNodes / 5) + 1 : 1
  if (density === 'rich') {
    if (model === 'pro') {
      return batches >= 5 ? '约 2～4 分钟' : '约 1.5～3 分钟'
    }
    return batches >= 5 ? '约 1.5～3 分钟' : '约 1～2 分钟'
  }
  return model === 'pro' ? '约 1～3 分钟' : '约 30 秒～1 分钟'
}

export function modelDisplayLabel(id?: string | null): string {
  return normalizeAIModel(id) === 'pro' ? 'Pro' : 'Flash'
}
