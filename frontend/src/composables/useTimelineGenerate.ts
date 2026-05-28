import { ElMessage, ElMessageBox } from 'element-plus'
import { DEFAULT_TARGET_NODE_COUNT, computeStepYears, estimateTimelineChunks, isLivingProfile } from '@/utils/timelineDensity'
import { api, type AIModelId, type Job, type Profile, type TimelineConfig } from '@/api/client'
import { modelDisplayLabel, estimateTimelineDuration, narrativeDensityLabel, normalizeNarrativeDensity } from '@/constants/models'
import type { NarrativeDensity } from '@/constants/models'
import { useJobRunner } from '@/composables/useJobRunner'

function escapeHtml(text: string): string {
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}

function buildTimelineConfirmMessage(options: {
  displayName?: string
  config?: TimelineConfig
  modelLabel: string
  estTime: string
  density: NarrativeDensity
  living?: boolean
}): string {
  const name = options.displayName ? `「${escapeHtml(options.displayName)}」` : '该人物'
  const cfg = options.config
  const rows: Array<[string, string]> = []

  if (cfg?.title) {
    rows.push(['时间轴名称', escapeHtml(cfg.title)])
  }
  if (cfg?.instructions) {
    rows.push(['特殊要求', escapeHtml(cfg.instructions)])
  }
  if (options.living) {
    rows.push(['人生状态', '未完结（进行中）'])
    rows.push(['节点规模', `约 ${cfg?.target_node_count ?? DEFAULT_TARGET_NODE_COUNT} 个`])
  } else if (cfg?.start_year && cfg?.end_year) {
    const span = Math.max(0, cfg.end_year - cfg.start_year)
    rows.push(['时间区间', `${cfg.start_year} — ${cfg.end_year} 年`])
    rows.push(['寿命跨度', `${span} 年`])
    if (cfg.target_node_count) {
      rows.push(['节点规模', `约 ${cfg.target_node_count} 个`])
      const step = computeStepYears(span, cfg.target_node_count)
      rows.push(['参考间隔', `约 ${step} 年`])
      const chunks = estimateTimelineChunks(cfg.target_node_count, span)
      if (chunks > 1) {
        rows.push(['生成方式', `分 ${chunks} 段生成（进度更准确）`])
      }
    }
  } else {
    rows.push(['节点规模', `约 ${DEFAULT_TARGET_NODE_COUNT} 个（全生命周期）`])
  }

  rows.push(['AI 模型', escapeHtml(options.modelLabel)])
  rows.push(['叙事密度', escapeHtml(narrativeDensityLabel(options.density))])
  rows.push(['预计耗时', escapeHtml(options.estTime)])

  const rowHtml = rows
    .map(
      ([label, value]) =>
        `<div class="tl-confirm-row"><span class="tl-confirm-label">${label}</span><span class="tl-confirm-value">${value}</span></div>`
    )
    .join('')

  return `
    <p class="tl-confirm-lead">将为 <strong>${name}</strong> 生成一条新的人生时间轴</p>
    <div class="tl-confirm-panel">${rowHtml}</div>
    <p class="tl-confirm-hint">任务将在后台运行，提交后可返回时间轴列表查看进度。</p>
  `
}

function parseTimelineIdFromJob(job: Job): string | undefined {
  if (job.request_json) {
    try {
      const req = JSON.parse(job.request_json) as { timeline_id?: string }
      if (req.timeline_id) return req.timeline_id
    } catch {
      /* ignore */
    }
  }
  if (job.result) {
    try {
      const res = JSON.parse(job.result) as { timeline_id?: string }
      if (res.timeline_id) return res.timeline_id
    } catch {
      /* ignore */
    }
  }
  return undefined
}

export interface TimelineGenerateResult {
  ok: boolean
  timelineId?: string
  jobId?: string
  background?: boolean
}

export function useTimelineGenerate() {
  const jobRunner = useJobRunner('生成时间轴')

  async function confirmAndGenerate(
    characterId: string,
    aiModel: AIModelId,
    options?: {
      displayName?: string
      config?: TimelineConfig
      background?: boolean
      profile?: Profile
    }
  ): Promise<TimelineGenerateResult> {
    const modelLabel = modelDisplayLabel(aiModel)
    const cfg = options?.config
    const living = options?.profile ? isLivingProfile(options.profile) : false
    const density = normalizeNarrativeDensity(cfg?.narrative_density)
    const targetNodes = cfg?.target_node_count ?? DEFAULT_TARGET_NODE_COUNT
    const estTime = estimateTimelineDuration(aiModel, density, targetNodes)

    try {
      await ElMessageBox.confirm(
        buildTimelineConfirmMessage({
          displayName: options?.displayName,
          config: cfg,
          modelLabel,
          estTime,
          density,
          living,
        }),
        '确认生成时间轴',
        {
          confirmButtonText: '开始生成',
          cancelButtonText: '取消',
          type: 'info',
          dangerouslyUseHTMLString: true,
          customClass: 'timeline-confirm-box',
        }
      )
    } catch {
      return { ok: false }
    }

    const submit = () => api.generateTimeline(characterId, aiModel, cfg)

    if (options?.background) {
      try {
        const job = await submit()
        const timelineId = parseTimelineIdFromJob(job)
        ElMessage.success('已提交后台生成，可在时间轴列表查看进度')
        return { ok: true, timelineId, jobId: job.id, background: true }
      } catch (e: unknown) {
        ElMessage.error(e instanceof Error ? e.message : '提交失败')
        return { ok: false }
      }
    }

    let timelineId: string | undefined
    const ok = await jobRunner.run({
      title: '生成时间轴',
      submitLabel: 'AI 正在生成人生时间轴',
      submit: async () => {
        const job = await submit()
        timelineId = parseTimelineIdFromJob(job)
        return job
      },
      resubmit: async () => {
        const job = await submit()
        timelineId = parseTimelineIdFromJob(job)
        return job
      },
      onProgress: (j) => {
        if (j.status === 'running') {
          const stage = j.stage_text ? `${j.stage_text} · ` : ''
          jobRunner.statusText.value = `${stage}${j.progress}% · ${modelDisplayLabel(j.model || aiModel)}`
        }
      },
      afterSuccess: () => {
        ElMessage.success('人生时间轴已生成')
      },
    })

    return {
      ok,
      timelineId,
    }
  }

  return {
    jobRunning: jobRunner.visible,
    progress: jobRunner.progress,
    statusText: jobRunner.statusText,
    failed: jobRunner.failed,
    errorText: jobRunner.errorText,
    retrying: jobRunner.retrying,
    retry: jobRunner.retry,
    clearJobState: jobRunner.clearState,
    confirmAndGenerate,
  }
}
