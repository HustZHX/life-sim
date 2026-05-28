import axios from 'axios'
import type { AIModelId } from '@/constants/models'
import { DEFAULT_AI_MODEL } from '@/constants/models'

export type { AIModelId }

const http = axios.create({ baseURL: '', timeout: 60000, withCredentials: true })

export interface AuthStatus {
  enabled: boolean
  gate_passed: boolean
  logged_in: boolean
}

export interface AuthUser {
  username: string
  display_name: string
}

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
  current_timeline_id?: string
}

export interface TimelineActiveJob {
  id: string
  type: string
  status: string
  progress: number
  stage_text?: string
}

export interface Timeline {
  id: string
  character_id: string
  title: string
  current_version_id?: string
  node_count?: number
  version_count?: number
  world_line?: WorldLine
  generation_status?: 'ready' | 'generating' | 'failed'
  generation_error?: string
  active_job?: TimelineActiveJob
  created_at: string
  updated_at: string
}

export interface WorldLineEvent {
  year: number
  name: string
  description?: string
  impact?: string
  divergence_note?: string
  caused_by_node_sequence?: number | null
}

export interface WorldLine {
  timeline_id: string
  era_summary?: string
  historical_trend: string
  daily_life_context?: string
  events: WorldLineEvent[]
  start_year?: number
  end_year?: number
  updated_at?: string
}

export interface WorldLineUpdatePayload {
  world_line: WorldLine
  model?: AIModelId
  apply_to_nodes?: boolean
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
  timeline_count?: number
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
  timeline_id?: string
  parent_version_id?: string
  trigger_node_id?: string
  change_summary?: string
  branch_label?: string
  fork_sequence?: number
  fork_node_id?: string
  death_year_snapshot?: number
  death_cause_snapshot?: string
  node_count?: number
  created_at: string
}

export interface BranchNode {
  id: string
  parent_id?: string
  label: string
  change_summary?: string
  fork_sequence?: number
  fork_node_id?: string
  node_count: number
  death_year_snapshot?: number
  death_cause_snapshot?: string
  is_active: boolean
  creates_branch?: boolean
  created_at: string
  children?: BranchNode[]
}

export interface BranchTreeResponse {
  timeline_id: string
  active_version_id: string
  display_active_branch_id?: string
  roots: BranchNode[]
}

export interface BranchOverviewNodeLite {
  sequence: number
  year: number
  title: string
  trait_changes?: TraitChange[]
}

export interface BranchOverviewEntry {
  version_id: string
  label: string
  is_active: boolean
  creates_branch?: boolean
  fork_sequence?: number
  death_year_snapshot?: number
  nodes: BranchOverviewNodeLite[]
}

export interface BranchOverviewResponse {
  timeline_id: string
  active_version_id: string
  branches: BranchOverviewEntry[]
}

export interface Job {
  id: string
  character_id: string
  type: string
  status: string
  progress: number
  stage_text?: string
  model?: string
  request_json?: string
  result?: string
  error?: string
  created_at?: string
  updated_at?: string
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

export type NarrativeDensity = 'standard' | 'rich'

export interface TimelineConfig {
  title?: string
  instructions?: string
  /** 时代背景与大事记；留空则由 AI 按生成区间自动整理 */
  era_events?: string
  target_node_count: number
  start_year: number
  end_year: number
  /** 由前端/后端按跨度推算，供展示 */
  step_years?: number
  /** standard=单次生成；rich=骨架+叙事扩写（默认） */
  narrative_density?: NarrativeDensity
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
  mode: 'full_cascade' | 'append_next' | 'inner_current' | 'inner_subsequent'
  model: AIModelId
  target_node_count?: number
  next_node_title?: string
  confirmed_death_year?: number
  confirmed_death_cause?: string
  lifespan_reasoning?: string
}

export interface LifespanPreview {
  death_year: number
  death_cause: string
  reasoning: string
  health_notes?: string
}

export interface LifespanPreviewResponse {
  current_death_year: number
  anchor_year: number
  preview: LifespanPreview
  context_token_hint?: number
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

export type NarrativeKind = 'light_novel' | 'diary' | 'letter' | 'archive'

export interface NarrativeArtifact {
  id: string
  character_id: string
  version_id: string
  node_id?: string
  kind: NarrativeKind
  from_sequence?: number
  to_sequence?: number
  content: string
  model?: string
  created_at?: string
}

export type LightNovelPerson = 'first' | 'second' | 'third'

export interface LightNovelRequest {
  model?: AIModelId
  version_id: string
  from_sequence: number
  to_sequence: number
  person?: LightNovelPerson
  force?: boolean
}

export interface SavedLightNovelMeta {
  id: string
  character_id: string
  display_name?: string
  version_id: string
  from_sequence: number
  to_sequence: number
  person?: LightNovelPerson
  model?: string
  branch?: string
  created_at?: string
}

export interface SavedLightNovel extends SavedLightNovelMeta {
  content: string
}

export interface NodeNarrativeRequest {
  model?: AIModelId
  force?: boolean
}

export interface DialogueIdentityOption {
  label: string
  description: string
}

export interface DialogueMessage {
  id: string
  session_id: string
  role: 'user' | 'assistant'
  content: string
  created_at: string
}

export interface DialogueSession {
  id: string
  character_id: string
  version_id: string
  node_id: string
  node_sequence: number
  speaker_identity: string
  model: AIModelId
  created_at: string
  updated_at: string
  messages?: DialogueMessage[]
}

export interface DialogueSessionSummary {
  id: string
  character_id: string
  version_id: string
  node_id: string
  node_sequence: number
  node_year?: number
  node_age?: number
  node_title?: string
  speaker_identity: string
  model: AIModelId
  message_count: number
  last_message?: string
  created_at: string
  updated_at: string
}

export interface PersonalityImpact {
  has_impact: boolean
  summary?: string
  thoughts?: string
  personality_snapshot?: string
  trait_changes?: TraitChange[]
}

export interface SendDialogueMessageResponse {
  user_message: DialogueMessage
  assistant_message: DialogueMessage
  impact?: PersonalityImpact
}

export interface CharacterMemory {
  id: string
  character_id: string
  version_id: string
  source_node_id: string
  source_sequence: number
  speaker_identity: string
  content: string
  created_at: string
  updated_at: string
}

export interface ApplyDialogueImpactResponse {
  version_id: string
  node: LifeNode
}

export const DIALOGUE_IDENTITY_READER = '命运读者'

export interface NarrativeCachedResponse {
  artifact: NarrativeArtifact
  cached: true
}

export interface NarrativeChangeRequest {
  timeline_id: string
  instruction: string
  model?: AIModelId
  target_node_count?: number
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

let refreshPromise: Promise<void> | null = null

async function refreshAuthSession() {
  if (!refreshPromise) {
    refreshPromise = unwrap<{ ok: boolean }>(http.post('/api/v1/auth/refresh'))
      .then(() => undefined)
      .finally(() => {
        refreshPromise = null
      })
  }
  await refreshPromise
}

http.interceptors.response.use(
  (response) => response,
  async (error: unknown) => {
    if (!axios.isAxiosError(error) || !error.response) {
      throw error
    }
    const status = error.response.status
    const body = error.response.data as ApiResponse<unknown> | undefined
    const msg = body?.message
    const config = error.config as (typeof error.config & { _retry?: boolean }) | undefined
    const url = config?.url ?? ''

    if (status === 401 && msg === 'token_expired' && config && !config._retry && !url.includes('/auth/')) {
      config._retry = true
      try {
        await refreshAuthSession()
        return http(config)
      } catch {
        if (typeof window !== 'undefined') window.location.href = '/login'
        throw error
      }
    }

    if (status === 401 && msg === 'gate_required' && !url.includes('/auth/gate')) {
      if (typeof window !== 'undefined') window.location.href = '/gate'
    }
    if (
      status === 401 &&
      (msg === 'login_required' || msg === 'token_expired') &&
      !url.includes('/auth/login') &&
      !url.includes('/auth/gate')
    ) {
      if (typeof window !== 'undefined') window.location.href = '/login'
    }

    throw error
  }
)

export const api = {
  authStatus: () => unwrap<AuthStatus>(http.get('/api/v1/auth/status')),

  authGate: (code: string) => unwrap<{ ok: boolean }>(http.post('/api/v1/auth/gate', { code })),

  authLogin: (username: string, password: string) =>
    unwrap<AuthUser>(http.post('/api/v1/auth/login', { username, password })),

  authRefresh: () => unwrap<{ ok: boolean }>(http.post('/api/v1/auth/refresh')),

  authMe: () => unwrap<AuthUser>(http.get('/api/v1/auth/me')),

  authLogout: () => unwrap<{ ok: boolean }>(http.post('/api/v1/auth/logout')),

  authChangePassword: (oldPassword: string, newPassword: string) =>
    unwrap<{ ok: boolean }>(
      http.post('/api/v1/auth/change-password', {
        old_password: oldPassword,
        new_password: newPassword,
      })
    ),

  listModels: () =>
    unwrap<{ models: AIModel[] }>(http.get('/api/v1/models')),

  suggestNames: (body: {
    model?: AIModelId
    background?: string
    introduction?: string
  }) =>
    unwrap<{ names: string[] }>(http.post('/api/v1/suggest-names', body)),

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

  generateProfile: (
    id: string,
    options?: {
      model?: AIModelId
      display_name?: string
      background?: string
      introduction?: string
    }
  ) =>
    unwrap<Profile>(
      http.post(`/api/v1/characters/${id}/profile/generate`, {
        model: options?.model,
        display_name: options?.display_name,
        background: options?.background,
        introduction: options?.introduction,
      })
    ),

  updateProfile: (id: string, profile: Profile) =>
    unwrap<Profile>(http.patch(`/api/v1/characters/${id}/profile`, profile)),

  randomizeProfileField: (id: string, field: string, model: AIModelId = DEFAULT_AI_MODEL) =>
    unwrap<Profile>(
      http.post(`/api/v1/characters/${id}/profile/randomize-field`, { field, model })
    ),

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
        era_events: config?.era_events,
        model,
        title: config?.title,
        instructions: config?.instructions,
        target_node_count: config?.target_node_count,
        start_year: config?.start_year,
        end_year: config?.end_year,
        narrative_density: config?.narrative_density,
      })
    ),

  listTimelines: (id: string) =>
    unwrap<{ timelines: Timeline[] }>(http.get(`/api/v1/characters/${id}/timelines`)),

  recommendTimeline: (id: string, model: AIModelId = DEFAULT_AI_MODEL) =>
    unwrap<{ recommendations: TimelineRecommendation[] }>(
      http.post(`/api/v1/characters/${id}/timeline/recommendations`, { model })
    ),

  getTimeline: (id: string, options?: { timelineId?: string; version?: string }) =>
    unwrap<{ timeline: Timeline; version: TimelineVersion; nodes: LifeNode[] }>(
      http.get(`/api/v1/characters/${id}/timeline`, {
        params: {
          timeline_id: options?.timelineId,
          version: options?.version,
        },
      })
    ),

  updateWorldLine: (
    charId: string,
    timelineId: string,
    body: WorldLineUpdatePayload
  ) =>
    unwrap<{ job?: Job; world_line: WorldLine }>(
      http.patch(`/api/v1/characters/${charId}/timelines/${timelineId}/world-line`, body)
    ),

  refreshWorldLine: (
    charId: string,
    timelineId: string,
    body: { model: AIModelId }
  ) =>
    unwrap<Job>(
      http.post(`/api/v1/characters/${charId}/timelines/${timelineId}/world-line/refresh`, body)
    ),

  patchNode: (charId: string, nodeId: string, body: PatchNodePayload) =>
    unwrap<Job>(http.patch(`/api/v1/characters/${charId}/nodes/${nodeId}`, body)),

  getNode: (charId: string, nodeId: string) =>
    unwrap<LifeNode>(http.get(`/api/v1/characters/${charId}/nodes/${nodeId}`)),

  previewLifespan: (
    charId: string,
    nodeId: string,
    body: Pick<
      PatchNodePayload,
      'title' | 'events' | 'thoughts' | 'personality_snapshot' | 'model'
    >
  ) =>
    unwrap<LifespanPreviewResponse>(
      http.post(`/api/v1/characters/${charId}/nodes/${nodeId}/lifespan/preview`, body)
    ),

  regenerateNodeEvents: (
    charId: string,
    nodeId: string,
    body: { title: string; model: AIModelId }
  ) =>
    unwrap<{
      events: string
      thoughts?: string
      personality_snapshot?: string
      trait_changes?: TraitChange[]
    }>(
      http.post(`/api/v1/characters/${charId}/nodes/${nodeId}/regenerate-events`, body)
    ),

  getJob: (jobId: string) => unwrap<Job>(http.get(`/api/v1/jobs/${jobId}`)),

  retryJob: (jobId: string) => unwrap<Job>(http.post(`/api/v1/jobs/${jobId}/retry`)),

  listVersions: (id: string, timelineId?: string) =>
    unwrap<{ versions: TimelineVersion[] }>(
      timelineId
        ? http.get(`/api/v1/characters/${id}/timelines/${timelineId}/versions`)
        : http.get(`/api/v1/characters/${id}/versions`)
    ),

  listBranches: (charId: string, timelineId: string) =>
    unwrap<BranchTreeResponse>(
      http.get(`/api/v1/characters/${charId}/timelines/${timelineId}/branches`)
    ),

  getBranchOverview: (charId: string, timelineId: string) =>
    unwrap<BranchOverviewResponse>(
      http.get(`/api/v1/characters/${charId}/timelines/${timelineId}/branches/overview`)
    ),

  activateBranch: (charId: string, timelineId: string, versionId: string) =>
    unwrap<Character>(
      http.post(`/api/v1/characters/${charId}/timelines/${timelineId}/branches/${versionId}/activate`)
    ),

  getVersionDiff: (charId: string, vid: string) =>
    unwrap<VersionDiff>(http.get(`/api/v1/characters/${charId}/versions/${vid}/diff`)),

  rollback: (charId: string, vid: string) =>
    unwrap<Character>(http.post(`/api/v1/characters/${charId}/versions/${vid}/rollback`)),

  rollbackToNode: (charId: string, nodeId: string, body?: { change_summary?: string }) =>
    unwrap<Character>(http.post(`/api/v1/characters/${charId}/nodes/${nodeId}/rollback-to`, body ?? {})),

  getNarrative: (
    charId: string,
    params: {
      version_id: string
      kind: NarrativeKind
      node_id?: string
      from_sequence?: number
      to_sequence?: number
    }
  ) =>
    unwrap<NarrativeArtifact>(
      http.get(`/api/v1/characters/${charId}/narratives`, { params })
    ),

  generateLightNovel: (charId: string, body: LightNovelRequest) =>
    unwrap<Job | NarrativeCachedResponse>(
      http.post(`/api/v1/characters/${charId}/narratives/light-novel`, body)
    ),

  listLightNovelJobs: (charId: string, limit = 30) =>
    unwrap<{ jobs: Job[] }>(
      http.get(`/api/v1/characters/${charId}/narratives/light-novel/jobs`, {
        params: { limit },
      })
    ),

  listSavedLightNovels: (params?: { character_id?: string }) =>
    unwrap<{ items: SavedLightNovelMeta[]; branch?: string }>(
      http.get('/api/v1/light-novels', { params })
    ),

  getSavedLightNovel: (id: string) =>
    unwrap<SavedLightNovel>(http.get(`/api/v1/light-novels/${id}`)),

  applyNarrativeChange: (charId: string, body: NarrativeChangeRequest) =>
    unwrap<Job>(http.post(`/api/v1/characters/${charId}/timeline/narrative-change`, body)),

  generateNodeNarrative: (
    charId: string,
    nodeId: string,
    kind: Exclude<NarrativeKind, 'light_novel'>,
    body?: NodeNarrativeRequest
  ) =>
    unwrap<Job | NarrativeCachedResponse>(
      http.post(`/api/v1/characters/${charId}/nodes/${nodeId}/narratives/${kind}`, body ?? {})
    ),

  getSavedDialogueIdentityOptions: (charId: string, nodeId: string) =>
    unwrap<{ options: DialogueIdentityOption[]; has_saved: boolean }>(
      http.get(`/api/v1/characters/${charId}/nodes/${nodeId}/dialogue/identity-options`)
    ),

  generateDialogueIdentityOptions: (
    charId: string,
    nodeId: string,
    model: AIModelId = DEFAULT_AI_MODEL,
    regenerate = false
  ) =>
    unwrap<{ options: DialogueIdentityOption[] }>(
      http.post(`/api/v1/characters/${charId}/nodes/${nodeId}/dialogue/identity-options`, {
        model,
        regenerate,
      })
    ),

  getLatestDialogueSession: (charId: string, nodeId: string) =>
    unwrap<{ session: DialogueSession | null }>(
      http.get(`/api/v1/characters/${charId}/nodes/${nodeId}/dialogue/latest-session`)
    ),

  listDialogueSessions: (charId: string, versionId?: string) =>
    unwrap<{ items: DialogueSessionSummary[] }>(
      http.get(`/api/v1/characters/${charId}/dialogue/sessions`, {
        params: versionId ? { version_id: versionId } : undefined,
      })
    ),

  createDialogueSession: (
    charId: string,
    nodeId: string,
    body: { identity: string; model?: AIModelId; resume?: boolean }
  ) =>
    unwrap<DialogueSession>(
      http.post(`/api/v1/characters/${charId}/nodes/${nodeId}/dialogue/sessions`, body)
    ),

  getDialogueSession: (charId: string, sessionId: string) =>
    unwrap<DialogueSession>(http.get(`/api/v1/characters/${charId}/dialogue/sessions/${sessionId}`)),

  sendDialogueMessage: (charId: string, sessionId: string, content: string, model?: AIModelId) =>
    unwrap<SendDialogueMessageResponse>(
      http.post(
        `/api/v1/characters/${charId}/dialogue/sessions/${sessionId}/messages`,
        { content, model },
        { timeout: 180000 }
      )
    ),

  listMemories: (charId: string, versionId: string, upToSequence?: number) =>
    unwrap<{ items: CharacterMemory[] }>(
      http.get(`/api/v1/characters/${charId}/memories`, {
        params: {
          version_id: versionId,
          up_to_sequence: upToSequence,
        },
      })
    ),

  createMemory: (
    charId: string,
    body: {
      version_id: string
      source_node_id: string
      source_sequence: number
      speaker_identity: string
      content: string
    }
  ) => unwrap<CharacterMemory>(http.post(`/api/v1/characters/${charId}/memories`, body)),

  summarizeMemory: (
    charId: string,
    body: { text: string; identity: string; model?: AIModelId }
  ) =>
    unwrap<{ content: string }>(
      http.post(`/api/v1/characters/${charId}/memories/summarize`, body, { timeout: 120000 })
    ),

  updateMemory: (charId: string, memoryId: string, content: string) =>
    unwrap<CharacterMemory>(http.patch(`/api/v1/characters/${charId}/memories/${memoryId}`, { content })),

  deleteMemory: (charId: string, memoryId: string) =>
    unwrap<{ deleted: boolean }>(http.delete(`/api/v1/characters/${charId}/memories/${memoryId}`)),

  applyDialogueImpact: (
    charId: string,
    nodeId: string,
    body: {
      thoughts: string
      personality_snapshot: string
      trait_changes?: TraitChange[]
      summary?: string
    }
  ) =>
    unwrap<ApplyDialogueImpactResponse>(
      http.post(`/api/v1/characters/${charId}/nodes/${nodeId}/dialogue/apply-impact`, body)
    ),
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
