import { computed, onUnmounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import {
  api,
  pollJob,
  type AIModelId,
  type Job,
  type LightNovelPerson,
  type LightNovelRequest,
  type NarrativeArtifact,
  type NarrativeCachedResponse,
} from '@/api/client'
import { DEFAULT_AI_MODEL, modelDisplayLabel } from '@/constants/models'
import { lightNovelPersonLabel } from '@/constants/lightNovelPerson'

export interface LightNovelJobEntry {
  id: string
  fromSequence: number
  toSequence: number
  person: LightNovelPerson
  model: AIModelId
  status: Job['status']
  progress: number
  stageText: string
  error?: string
  savedFileId?: string
  content?: string
  createdAt?: string
}

function isCachedResponse(res: unknown): res is NarrativeCachedResponse {
  return (
    typeof res === 'object' &&
    res !== null &&
    'cached' in res &&
    (res as NarrativeCachedResponse).cached === true &&
    'artifact' in res
  )
}

function parseJobResult(result?: string): { content?: string; saved_file_id?: string } {
  if (!result) return {}
  try {
    return JSON.parse(result) as { content?: string; saved_file_id?: string }
  } catch {
    return {}
  }
}

function parseRequestJSON(raw?: string): Partial<LightNovelRequest> {
  if (!raw) return {}
  try {
    return JSON.parse(raw) as Partial<LightNovelRequest>
  } catch {
    return {}
  }
}

function jobToEntry(job: Job): LightNovelJobEntry {
  const req = parseRequestJSON(job.request_json)
  const result = parseJobResult(job.result)
  return {
    id: job.id,
    fromSequence: req.from_sequence ?? 0,
    toSequence: req.to_sequence ?? 0,
    person: (req.person as LightNovelPerson) || 'first',
    model: (job.model || req.model || DEFAULT_AI_MODEL) as AIModelId,
    status: job.status,
    progress: job.progress,
    stageText: job.stage_text || '',
    error: job.error,
    savedFileId: result.saved_file_id,
    content: result.content,
    createdAt: job.created_at,
  }
}

export function useLightNovelJobs(characterId: string) {
  const jobs = ref<LightNovelJobEntry[]>([])
  const listVisible = ref(false)
  const loading = ref(false)
  const pollingIds = new Set<string>()
  let pollTimer: ReturnType<typeof setInterval> | null = null

  const activeCount = computed(
    () => jobs.value.filter((j) => j.status === 'pending' || j.status === 'running').length
  )

  function upsertEntry(entry: LightNovelJobEntry) {
    const idx = jobs.value.findIndex((j) => j.id === entry.id)
    if (idx >= 0) {
      jobs.value[idx] = { ...jobs.value[idx], ...entry }
    } else {
      jobs.value.unshift(entry)
    }
  }

  function updateFromJob(job: Job) {
    upsertEntry(jobToEntry(job))
  }

  async function refreshList() {
    loading.value = true
    try {
      const res = await api.listLightNovelJobs(characterId)
      jobs.value = res.jobs.map(jobToEntry)
      for (const job of res.jobs) {
        if (job.status === 'pending' || job.status === 'running') {
          ensurePoll(job.id, false)
        }
      }
    } catch (e: unknown) {
      ElMessage.error(e instanceof Error ? e.message : '加载生成队列失败')
    } finally {
      loading.value = false
    }
  }

  async function pollOne(jobId: string, notify = true) {
    if (pollingIds.has(jobId)) return
    pollingIds.add(jobId)
    const before = jobs.value.find((j) => j.id === jobId)
    const wasActive = before?.status === 'pending' || before?.status === 'running'
    try {
      const done = await pollJob(jobId, updateFromJob, 2000, 600)
      updateFromJob(done)
      if (notify && wasActive && done.status === 'completed') {
        ElMessage.success('轻小说生成完成')
      } else if (notify && wasActive && done.status === 'failed') {
        ElMessage.error(done.error || '轻小说生成失败')
      }
    } catch (e: unknown) {
      upsertEntry({
        id: jobId,
        fromSequence: 0,
        toSequence: 0,
        person: 'first',
        model: DEFAULT_AI_MODEL,
        status: 'failed',
        progress: 0,
        stageText: '',
        error: e instanceof Error ? e.message : '轮询失败',
      })
    } finally {
      pollingIds.delete(jobId)
    }
  }

  function ensurePoll(jobId: string, notify = true) {
    if (pollingIds.has(jobId)) return
    void pollOne(jobId, notify)
  }

  function startBackgroundPoll() {
    if (pollTimer) return
    pollTimer = setInterval(() => {
      for (const job of jobs.value) {
        if (job.status === 'pending' || job.status === 'running') {
          ensurePoll(job.id, false)
        }
      }
    }, 5000)
  }

  function stopBackgroundPoll() {
    if (pollTimer) {
      clearInterval(pollTimer)
      pollTimer = null
    }
  }

  async function submit(params: {
    versionId: string
    fromSequence: number
    toSequence: number
    person: LightNovelPerson
    model: AIModelId
    force?: boolean
    onCached?: (artifact: NarrativeArtifact) => void
  }): Promise<boolean> {
    listVisible.value = true
    try {
      const res = await api.generateLightNovel(characterId, {
        model: params.model,
        version_id: params.versionId,
        from_sequence: params.fromSequence,
        to_sequence: params.toSequence,
        person: params.person,
        force: params.force,
      })

      if (isCachedResponse(res)) {
        params.onCached?.(res.artifact)
        ElMessage.success('已加载缓存')
        return true
      }

      const entry = jobToEntry(res)
      entry.fromSequence = params.fromSequence
      entry.toSequence = params.toSequence
      entry.person = params.person
      upsertEntry(entry)
      ensurePoll(res.id)
      startBackgroundPoll()
      ElMessage.success('已提交后台生成，可在生成队列查看进度')
      return true
    } catch (e: unknown) {
      ElMessage.error(e instanceof Error ? e.message : '提交失败')
      return false
    }
  }

  function statusLabel(entry: LightNovelJobEntry): string {
    if (entry.status === 'completed') return '已完成'
    if (entry.status === 'failed') return '失败'
    if (entry.status === 'running') {
      const stage = entry.stageText ? `${entry.stageText} · ` : ''
      return `${stage}${entry.progress}%（${modelDisplayLabel(entry.model)}）`
    }
    return '排队中…'
  }

  function entryTitle(entry: LightNovelJobEntry): string {
    return `节点 ${entry.fromSequence}–${entry.toSequence} · ${lightNovelPersonLabel(entry.person)}`
  }

  onUnmounted(() => {
    stopBackgroundPoll()
  })

  return {
    jobs,
    listVisible,
    loading,
    activeCount,
    refreshList,
    submit,
    statusLabel,
    entryTitle,
    startBackgroundPoll,
    stopBackgroundPoll,
  }
}
