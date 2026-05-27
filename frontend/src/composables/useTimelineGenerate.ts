import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { DEFAULT_TARGET_NODE_COUNT, computeStepYears } from '@/utils/timelineDensity'
import { api, pollJob, type AIModelId, type TimelineConfig } from '@/api/client'
import { modelDisplayLabel } from '@/constants/models'

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
  if (cfg?.start_year && cfg?.end_year) {
    const span = Math.max(0, cfg.end_year - cfg.start_year)
    rows.push(['时间区间', `${cfg.start_year} — ${cfg.end_year} 年`])
    rows.push(['寿命跨度', `${span} 年`])
    if (cfg.target_node_count) {
      rows.push(['节点规模', `约 ${cfg.target_node_count} 个`])
      const step = computeStepYears(span, cfg.target_node_count)
      rows.push(['参考间隔', `约 ${step} 年`])
    }
  } else {
    rows.push(['节点规模', `约 ${DEFAULT_TARGET_NODE_COUNT} 个（全生命周期）`])
  }

  rows.push(['AI 模型', escapeHtml(options.modelLabel)])
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
    <p class="tl-confirm-hint">生成过程中可继续浏览档案，进度显示在左上角。</p>
  `
}

export interface TimelineGenerateResult {
  ok: boolean
  timelineId?: string
}

export function useTimelineGenerate() {
  const jobRunning = ref(false)
  const progress = ref(0)
  const statusText = ref('')

  async function confirmAndGenerate(
    characterId: string,
    aiModel: AIModelId,
    options?: { displayName?: string; config?: TimelineConfig }
  ): Promise<TimelineGenerateResult> {
    const modelLabel = modelDisplayLabel(aiModel)
    const estTime = aiModel === 'pro' ? '1～3 分钟' : '约 30 秒～1 分钟'
    const cfg = options?.config

    try {
      await ElMessageBox.confirm(
        buildTimelineConfirmMessage({
          displayName: options?.displayName,
          config: cfg,
          modelLabel,
          estTime,
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

    jobRunning.value = true
    progress.value = 5
    statusText.value = '正在提交任务…'

    try {
      const job = await api.generateTimeline(characterId, aiModel, cfg)
      statusText.value = 'AI 正在生成人生时间轴'
      const done = await pollJob(job.id, (j) => {
        progress.value = Math.max(j.progress, 5)
        if (j.status === 'running') {
          statusText.value = `生成中 ${j.progress}% · ${modelDisplayLabel(j.model || aiModel)}`
        }
      })
      if (done.status === 'failed') throw new Error(done.error || '生成失败')
      progress.value = 100
      statusText.value = '生成完成'
      ElMessage.success('人生时间轴已生成')
      let timelineId: string | undefined
      if (done.result) {
        try {
          const parsed = JSON.parse(done.result) as { timeline_id?: string }
          timelineId = parsed.timeline_id
        } catch {
          /* ignore */
        }
      }
      return { ok: true, timelineId }
    } catch (e: unknown) {
      ElMessage.error(e instanceof Error ? e.message : '生成失败')
      return { ok: false }
    } finally {
      jobRunning.value = false
      statusText.value = ''
      progress.value = 0
    }
  }

  return { jobRunning, progress, statusText, confirmAndGenerate }
}
