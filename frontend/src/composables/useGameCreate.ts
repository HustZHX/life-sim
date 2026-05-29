import { ref } from 'vue'
import { api, pollJob, type AIModelId, type Job } from '@/api/client'
import { DEFAULT_AI_MODEL } from '@/constants/models'
import { useJobRunner } from '@/composables/useJobRunner'

export function useGameCreate() {
  const jobRunner = useJobRunner('人生游戏')
  const starting = ref(false)

  async function startGameTimeline(
    characterId: string,
    model: AIModelId = DEFAULT_AI_MODEL
  ): Promise<{ ok: boolean; timelineId?: string }> {
    starting.value = true
    const ok = await jobRunner.run({
      title: '开启人生游戏',
      submitLabel: '正在生成首节点与世界线…',
      submit: () => api.gameStartTimeline(characterId, { model, title: '人生游戏' }),
      resubmit: () => api.gameStartTimeline(characterId, { model, title: '人生游戏' }),
      afterSuccess: async (job: Job) => {
        if (job.result) {
          try {
            const r = JSON.parse(job.result) as { timeline_id?: string }
            if (r.timeline_id) return
          } catch {
            /* ignore */
          }
        }
      },
    })
    starting.value = false
    if (!ok) return { ok: false }
    try {
      const ch = await api.getCharacter(characterId)
      return { ok: true, timelineId: ch.current_timeline_id }
    } catch {
      return { ok: true }
    }
  }

  return {
    jobRunner,
    starting,
    startGameTimeline,
  }
}

export async function pollGameJob(jobId: string, onProgress?: (j: Job) => void) {
  return pollJob(jobId, onProgress)
}
