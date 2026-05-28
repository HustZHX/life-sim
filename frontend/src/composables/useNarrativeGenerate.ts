import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import {
  api,
  type AIModelId,
  type Job,
  type LifeNode,
  type NarrativeArtifact,
  type NarrativeCachedResponse,
  type NarrativeKind,
} from '@/api/client'
import { DEFAULT_AI_MODEL, modelDisplayLabel } from '@/constants/models'
import { useJobRunner } from '@/composables/useJobRunner'

export function narrativeKindLabel(kind: NarrativeKind): string {
  switch (kind) {
    case 'light_novel':
      return '轻小说'
    case 'diary':
      return '日记'
    case 'letter':
      return '书信'
    case 'archive':
      return '档案'
    default:
      return kind
  }
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

function parseJobContent(result?: string): string {
  if (!result) return ''
  try {
    const parsed = JSON.parse(result) as { content?: string }
    return parsed.content ?? ''
  } catch {
    return result
  }
}

export interface NodeNarrativeContext {
  type: 'node'
  characterId: string
  nodeId: string
  versionId: string
  kind: Exclude<NarrativeKind, 'light_novel'>
  model: AIModelId
  nodeLabel?: string
}

export type NarrativeContext = NodeNarrativeContext

export function useNarrativeGenerate() {
  const visible = ref(false)
  const title = ref('')
  const content = ref('')
  const running = ref(false)
  const context = ref<NarrativeContext | null>(null)
  const jobRunner = useJobRunner('叙事生成')

  function resetProgress() {
    jobRunner.progress.value = 0
    jobRunner.statusText.value = ''
  }

  function close() {
    visible.value = false
    if (!running.value && !jobRunner.visible.value) {
      content.value = ''
      context.value = null
      resetProgress()
    }
  }

  function openWithArtifact(artifact: NarrativeArtifact, ctx: NarrativeContext, dialogTitle: string) {
    context.value = ctx
    title.value = dialogTitle
    content.value = artifact.content
    visible.value = true
  }

  async function generateNodeNarrative(
    ctx: NodeNarrativeContext,
    options?: { force?: boolean }
  ): Promise<boolean> {
    const kindLabel = narrativeKindLabel(ctx.kind)
    const dialogTitle = ctx.nodeLabel ? `${kindLabel} · ${ctx.nodeLabel}` : kindLabel

    context.value = ctx
    title.value = dialogTitle
    visible.value = true
    if (options?.force) content.value = ''

    running.value = true

    const request = () =>
      api.generateNodeNarrative(ctx.characterId, ctx.nodeId, ctx.kind, {
        model: ctx.model,
        force: options?.force,
      })

    try {
      const first = await request()
      if (isCachedResponse(first)) {
        openWithArtifact(first.artifact, ctx, dialogTitle)
        if (!options?.force) ElMessage.success('已加载缓存')
        return true
      }

      const initialJob = first as Job
      const ok = await jobRunner.run({
        title: '叙事生成',
        submitLabel: `AI 正在撰写${kindLabel}`,
        submit: () => Promise.resolve(initialJob),
        resubmit: async () => {
          const res = await api.generateNodeNarrative(ctx.characterId, ctx.nodeId, ctx.kind, {
            model: ctx.model,
            force: true,
          })
          if (isCachedResponse(res)) {
            openWithArtifact(res.artifact, ctx, dialogTitle)
            throw new Error('已加载缓存')
          }
          return res
        },
        onProgress: (j) => {
          if (j.status === 'running') {
            const stage = j.stage_text ? `${j.stage_text} · ` : ''
            jobRunner.statusText.value = `${stage}${j.progress}% · ${modelDisplayLabel(j.model || ctx.model)}`
          }
        },
        afterSuccess: (job) => {
          content.value = parseJobContent(job.result)
          ElMessage.success(`${kindLabel}已生成`)
        },
      })
      return ok
    } catch (e: unknown) {
      if (!jobRunner.failed.value) {
        ElMessage.error(e instanceof Error ? e.message : '生成失败')
        if (!content.value) visible.value = false
      }
      return false
    } finally {
      running.value = false
    }
  }

  async function openNodeNarrative(
    characterId: string,
    node: LifeNode,
    kind: Exclude<NarrativeKind, 'light_novel'>,
    model: AIModelId = DEFAULT_AI_MODEL
  ) {
    await generateNodeNarrative({
      type: 'node',
      characterId,
      nodeId: node.id,
      versionId: node.version_id,
      kind,
      model,
      nodeLabel: `${node.year} 年 · ${node.title}`,
    })
  }

  async function regenerate() {
    const ctx = context.value
    if (!ctx) return
    await generateNodeNarrative(ctx, { force: true })
  }

  async function retryNarrativeJob() {
    await jobRunner.retry()
  }

  return {
    visible,
    title,
    content,
    running,
    progress: jobRunner.progress,
    statusText: jobRunner.statusText,
    failed: jobRunner.failed,
    errorText: jobRunner.errorText,
    retrying: jobRunner.retrying,
    jobPanelVisible: jobRunner.visible,
    context,
    close,
    generateNodeNarrative,
    openNodeNarrative,
    regenerate,
    retryNarrativeJob,
    clearJobState: jobRunner.clearState,
  }
}
