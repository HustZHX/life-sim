import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { DEFAULT_TARGET_NODE_COUNT } from '@/utils/timelineDensity'
import { api, pollJob, type AIModelId, type TimelineConfig } from '@/api/client'
import { modelDisplayLabel } from '@/constants/models'

export function useTimelineGenerate() {
  const jobRunning = ref(false)
  const progress = ref(0)
  const statusText = ref('')

  async function confirmAndGenerate(
    characterId: string,
    aiModel: AIModelId,
    options?: { displayName?: string; config?: TimelineConfig }
  ): Promise<boolean> {
    const modelLabel = modelDisplayLabel(aiModel)
    const name = options?.displayName ? `「${options.displayName}」` : '该人物'
    const estTime = aiModel === 'pro' ? '1～3 分钟' : '约 30 秒～1 分钟'
    const cfg = options?.config
    const rangeText = cfg
      ? formatTimelineSummary(cfg.start_year, cfg.end_year, cfg.target_node_count)
      : '全生命周期（默认约 20 个节点）'

    try {
      await ElMessageBox.confirm(
        `将为${name}生成时间轴。\n\n区间：${rangeText}\n模型：${modelLabel}\n预计耗时：${estTime}\n\n生成过程中可继续浏览档案，进度显示在左上角。`,
        '确认生成时间轴',
        {
          confirmButtonText: '开始生成',
          cancelButtonText: '取消',
          type: 'info',
        }
      )
    } catch {
      return false
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
      return true
    } catch (e: unknown) {
      ElMessage.error(e instanceof Error ? e.message : '生成失败')
      return false
    } finally {
      jobRunning.value = false
      statusText.value = ''
      progress.value = 0
    }
  }

  return { jobRunning, progress, statusText, confirmAndGenerate }
}
