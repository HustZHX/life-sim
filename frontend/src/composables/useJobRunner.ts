import { ref } from 'vue'
import { api, pollJob, type Job } from '@/api/client'

export function useJobRunner(defaultTitle = '任务进度') {
  const visible = ref(false)
  const progress = ref(0)
  const statusText = ref('')
  const title = ref(defaultTitle)
  const failed = ref(false)
  const errorText = ref('')
  const retrying = ref(false)

  let lastJobId = ''
  let resubmitFn: (() => Promise<Job>) | null = null
  let afterSuccess: ((job: Job) => Promise<void> | void) | null = null
  let onPollProgress: ((j: Job) => void) | null = null

  function clearState() {
    visible.value = false
    failed.value = false
    errorText.value = ''
    progress.value = 0
    statusText.value = ''
    retrying.value = false
    lastJobId = ''
    resubmitFn = null
    afterSuccess = null
    onPollProgress = null
  }

  async function pollUntilDone(jobId: string) {
    lastJobId = jobId
    const done = await pollJob(jobId, (j) => {
      progress.value = Math.max(j.progress, 5)
      onPollProgress?.(j)
    })
    if (done.status === 'failed') {
      throw new Error(done.error || '任务失败')
    }
    return done
  }

  async function run(options: {
    title?: string
    submitLabel?: string
    submit: () => Promise<Job>
    resubmit?: () => Promise<Job>
    onProgress?: (j: Job) => void
    afterSuccess?: (job: Job) => Promise<void> | void
  }): Promise<boolean> {
    title.value = options.title ?? defaultTitle
    visible.value = true
    failed.value = false
    errorText.value = ''
    progress.value = 5
    statusText.value = options.submitLabel ?? '正在提交…'
    resubmitFn = options.resubmit ?? options.submit
    afterSuccess = options.afterSuccess ?? null
    onPollProgress = options.onProgress ?? null

    try {
      const job = await options.submit()
      const done = await pollUntilDone(job.id)
      progress.value = 100
      statusText.value = '完成'
      await afterSuccess?.(done)
      setTimeout(clearState, 1500)
      return true
    } catch (e: unknown) {
      failed.value = true
      errorText.value = e instanceof Error ? e.message : '任务失败'
      statusText.value = '任务失败'
      return false
    }
  }

  async function retry(): Promise<boolean> {
    if (!lastJobId && !resubmitFn) return false
    retrying.value = true
    failed.value = false
    errorText.value = ''
    progress.value = 5
    statusText.value = '正在重试…'
    try {
      let job: Job
      if (lastJobId) {
        try {
          job = await api.retryJob(lastJobId)
        } catch {
          if (!resubmitFn) throw new Error('无法重试，请重新提交')
          job = await resubmitFn()
        }
      } else if (resubmitFn) {
        job = await resubmitFn()
      } else {
        return false
      }
      const done = await pollUntilDone(job.id)
      progress.value = 100
      statusText.value = '完成'
      await afterSuccess?.(done)
      setTimeout(clearState, 1500)
      return true
    } catch (e: unknown) {
      failed.value = true
      errorText.value = e instanceof Error ? e.message : '重试失败'
      statusText.value = '任务失败'
      return false
    } finally {
      retrying.value = false
    }
  }

  return {
    visible,
    progress,
    statusText,
    title,
    failed,
    errorText,
    retrying,
    run,
    retry,
    clearState,
  }
}
