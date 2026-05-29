import { computed, onUnmounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import {
  api,
  pollJob,
  type AIModelId,
  type ChronicleRequest,
  type Job,
  type NarrativeArtifact,
  type NarrativeCachedResponse,
} from '@/api/client'
import { DEFAULT_AI_MODEL, modelDisplayLabel } from '@/constants/models'

export interface ChronicleJobEntry {
  id: string
  fromSequence: number
  toSequence: number
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

function parseRequestJSON(raw?: string): Partial<ChronicleRequest> {
  if (!raw) return {}
  try {
    return JSON.parse(raw) as Partial<ChronicleRequest>
  } catch {
    return {}
  }
}

function jobToEntry(job: Job): ChronicleJobEntry {
  const req = parseRequestJSON(job.request_json)
  const result = parseJobResult(job.result)
  return {
    id: job.id,
    fromSequence: req.from_sequence ?? 0,
    toSequence: req.to_sequence ?? 0,
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

export function useChronicleJobs(characterId: string) {
  const jobs = ref<ChronicleJobEntry[]>([])
  const listVisible = ref(false)
  const loading = ref(false)
  const pollingIds = new Set<string>()
  let pollTimer: ReturnType<typeof setInterval> | null = null

  const activeCount = computed(
    () => jobs.value.filter((j) => j.status === 'pending' || j.status === 'running').length
  )

  function upsertEntry(entry: ChronicleJobEntry) {
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
      const res = await api.listChronicleJobs(characterId)
      jobs.value = res.jobs.map(jobToEntry)
      for (const job of res.jobs) {
        if (job.status === 'pending' || job.status === 'running') {
          ensurePoll(job.id, false)
        }
      }
    } catch (e: unknown) {
      ElMessage.error(e instanceof Error ? e.message : '加载史书队列失败')
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
        ElMessage.success('史书编纂完成')
      } else if (notify && wasActive && done.status === 'failed') {
        ElMessage.error(done.error || '史书编纂失败')
      }
    } catch (e: unknown) {
      upsertEntry({
        id: jobId,
        fromSequence: 0,
        toSequence: 0,
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
    model: AIModelId
    force?: boolean
    onCached?: (artifact: NarrativeArtifact) => void
  }): Promise<boolean> {
    listVisible.value = true
    try {
      const res = await api.generateChronicle(characterId, {
        model: params.model,
        version_id: params.versionId,
        from_sequence: params.fromSequence,
        to_sequence: params.toSequence,
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
      upsertEntry(entry)
      ensurePoll(res.id)
      startBackgroundPoll()
      ElMessage.success('已提交史书编纂，可在队列查看进度')
      return true
    } catch (e: unknown) {
      ElMessage.error(e instanceof Error ? e.message : '提交失败')
      return false
    }
  }

  async function retryJob(entry: ChronicleJobEntry) {
    listVisible.value = true
    try {
      const job = await api.retryJob(entry.id)
      upsertEntry(jobToEntry(job))
      ensurePoll(job.id)
      startBackgroundPoll()
      ElMessage.success('已重新提交编纂')
    } catch (e: unknown) {
      ElMessage.error(e instanceof Error ? e.message : '重试失败')
    }
  }

  function statusLabel(entry: ChronicleJobEntry): string {
    if (entry.status === 'completed') return '已完成'
    if (entry.status === 'failed') return '失败'
    if (entry.status === 'running') {
      const stage = entry.stageText ? `${entry.stageText} · ` : ''
      return `${stage}${entry.progress}%（${modelDisplayLabel(entry.model)}）`
    }
    return '排队中…'
  }

  function entryTitle(entry: ChronicleJobEntry): string {
    return `节点 ${entry.fromSequence}–${entry.toSequence}`
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
    retryJob,
    statusLabel,
    entryTitle,
    startBackgroundPoll,
    stopBackgroundPoll,
  }
}
