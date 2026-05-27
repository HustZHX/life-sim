import axios from 'axios'
import type { AIModelId } from '@/constants/models'
import { DEFAULT_AI_MODEL } from '@/constants/models'

export type { AIModelId }

const http = axios.create({ baseURL: '' })

export interface ApiResponse<T> {
  code: number
  message: string
  data?: T
}

export interface Character {
  id: string
  mode: 'famous' | 'random'
  display_name: string
  status: string
  current_version_id?: string
}

export interface HistoryItem {
  id: string
  mode: 'famous' | 'random'
  display_name: string
  status: string
  resolve_query?: string
  era?: string
  birth_year?: number
  death_year?: number
  node_count: number
  created_at: string
  updated_at: string
}

export interface ResolveCandidate {
  name: string
  birth_year: number
  death_year: number
  nationality: string
  summary: string
  confidence: number
}

export interface Profile {
  character_id: string
  display_name: string
  birth_year: number
  death_year: number
  era: string
  era_background: string
  personality_initial: string
  beliefs_motto: string
  experiences: string
  literature: string
  birth_place: string
  nationality: string
  social_class: string
  occupation: string
  education: string
  family_relations: string
  beliefs_politics: string
  appearance: string
  major_works: string
  controversies: string
  death_cause: string
  sources_note: string
  template_source?: string
}

export interface TraitChange {
  field: string
  before: string
  after: string
  reason: string
}

export interface NodeEntities {
  protagonist?: string[]
  persons?: string[]
  places?: string[]
}

export interface NodeScene {
  datetime?: string
  time_of_day?: string
  season?: string
  weather?: string
  scene?: string
}

export interface LifeNode {
  id: string
  character_id: string
  version_id: string
  sequence: number
  year: number
  age: number
  title: string
  events: string
  thoughts: string
  personality_snapshot: string
  trait_changes?: TraitChange[]
  entities?: NodeEntities
  scene?: NodeScene
}

export interface TimelineVersion {
  id: string
  character_id: string
  parent_version_id?: string
  trigger_node_id?: string
  change_summary?: string
  created_at: string
}

export interface Job {
  id: string
  character_id: string
  type: string
  status: string
  progress: number
  model?: string
  result?: string
  error?: string
}

export interface AIModel {
  id: string
  api_model: string
  label: string
  description: string
  est_time: string
  speed: string
  quality: string
}

export interface TimelineConfig {
  target_node_count: number
  start_year: number
  end_year: number
  /** 由前端/后端按跨度推算，供展示 */
  step_years?: number
}

export interface TimelineRecommendation {
  label: string
  description: string
  target_node_count: number
  step_years?: number
  start_year: number
  end_year: number
  focus_phase?: string
}

export interface PatchNodePayload {
  title: string
  events: string
  thoughts: string
  personality_snapshot: string
  mode: 'full_cascade' | 'inner_current' | 'inner_subsequent'
  model: AIModelId
  target_node_count?: number
}

export interface NodeFieldChange {
  node_id: string
  sequence: number
  year: number
  field: string
  before: string
  after: string
}

export interface VersionDiff {
  version_id: string
  parent_version_id?: string
  changes: NodeFieldChange[]
}

async function unwrap<T>(p: Promise<{ data: ApiResponse<T> }>): Promise<T> {
  try {
    const { data } = await p
    if (data.code !== 0) throw new Error(data.message || '请求失败')
    return data.data as T
  } catch (e: unknown) {
    if (axios.isAxiosError(e) && e.response?.data) {
      const body = e.response.data as ApiResponse<unknown>
      if (body.message) throw new Error(body.message)
    }
    throw e
  }
}

export const api = {
  listModels: () =>
    unwrap<{ models: AIModel[] }>(http.get('/api/v1/models')),

  listHistory: (limit = 100) =>
    unwrap<{ items: HistoryItem[]; total: number }>(
      http.get('/api/v1/history', { params: { limit } })
    ),

  createCharacter: (mode: string) =>
    unwrap<Character>(http.post('/api/v1/characters', { mode })),

  getCharacter: (id: string) =>
    unwrap<Character>(http.get(`/api/v1/characters/${id}`)),

  resolve: (id: string, query: string) =>
    unwrap<{ candidates: ResolveCandidate[] }>(
      http.post(`/api/v1/characters/${id}/resolve`, { query })
    ),

  confirm: (id: string, candidateIndex: number) =>
    unwrap<Character>(
      http.post(`/api/v1/characters/${id}/confirm`, { candidate_index: candidateIndex })
    ),

  generateProfile: (id: string) =>
    unwrap<Profile>(http.post(`/api/v1/characters/${id}/profile/generate`)),

  importTemplate: (id: string, famousQuery: string) =>
    unwrap<Profile>(
      http.post(`/api/v1/characters/${id}/profile/import-template`, {
        famous_query: famousQuery,
      })
    ),

  getProfile: (id: string) =>
    unwrap<Profile>(http.get(`/api/v1/characters/${id}/profile`)),

  generateTimeline: (
    id: string,
    model: AIModelId = DEFAULT_AI_MODEL,
    config?: Partial<TimelineConfig>
  ) =>
    unwrap<Job>(
      http.post(`/api/v1/characters/${id}/timeline/generate`, {
        model,
        target_node_count: config?.target_node_count,
        start_year: config?.start_year,
        end_year: config?.end_year,
      })
    ),

  recommendTimeline: (id: string, model: AIModelId = DEFAULT_AI_MODEL) =>
    unwrap<{ recommendations: TimelineRecommendation[] }>(
      http.post(`/api/v1/characters/${id}/timeline/recommendations`, { model })
    ),

  getTimeline: (id: string, version = 'latest') =>
    unwrap<{ version: TimelineVersion; nodes: LifeNode[] }>(
      http.get(`/api/v1/characters/${id}/timeline`, { params: { version } })
    ),

  patchNode: (charId: string, nodeId: string, body: PatchNodePayload) =>
    unwrap<Job>(http.patch(`/api/v1/characters/${charId}/nodes/${nodeId}`, body)),

  getJob: (jobId: string) => unwrap<Job>(http.get(`/api/v1/jobs/${jobId}`)),

  listVersions: (id: string) =>
    unwrap<{ versions: TimelineVersion[] }>(
      http.get(`/api/v1/characters/${id}/versions`)
    ),

  getVersionDiff: (charId: string, vid: string) =>
    unwrap<VersionDiff>(http.get(`/api/v1/characters/${charId}/versions/${vid}/diff`)),

  rollback: (charId: string, vid: string) =>
    unwrap<Character>(http.post(`/api/v1/characters/${charId}/versions/${vid}/rollback`)),
}

export async function pollJob(
  jobId: string,
  onProgress?: (j: Job) => void,
  intervalMs = 2000,
  maxAttempts = 120
): Promise<Job> {
  for (let i = 0; i < maxAttempts; i++) {
    try {
      const job = await api.getJob(jobId)
      onProgress?.(job)
      if (job.status === 'completed' || job.status === 'failed') return job
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : ''
      if (msg.includes('404') || msg.includes('任务不存在')) {
        throw new Error('任务已丢失（可能因服务重启），请重新提交生成')
      }
      throw e
    }
    await new Promise((r) => setTimeout(r, intervalMs))
  }
  throw new Error('任务超时，请稍后重试')
}
