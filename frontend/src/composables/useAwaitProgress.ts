import { ref } from 'vue'

export interface AwaitProgressOptions {
  title?: string
  startText: string
  /** 中间步骤（如 confirm），完成后更新文案 */
  midText?: string
  onMid?: () => Promise<void>
  /** 进度条上限（等待主任务时） */
  cap?: number
}

export function useAwaitProgress(defaultTitle = '处理中') {
  const visible = ref(false)
  const progress = ref(0)
  const statusText = ref('')
  const title = ref(defaultTitle)
  let timer: ReturnType<typeof setInterval> | null = null

  function startTick(from = 8, cap = 88) {
    progress.value = from
    timer = setInterval(() => {
      if (progress.value < cap) {
        progress.value = Math.min(cap, Math.round(progress.value + Math.random() * 4 + 2))
      }
    }, 450)
  }

  function stopTick(final = 100) {
    if (timer) {
      clearInterval(timer)
      timer = null
    }
    progress.value = final
  }

  async function run<T>(task: () => Promise<T>, options: AwaitProgressOptions): Promise<T> {
    visible.value = true
    title.value = options.title ?? defaultTitle
    statusText.value = options.startText
    startTick(5, options.cap ?? 88)
    try {
      if (options.onMid) {
        await options.onMid()
        progress.value = Math.max(progress.value, 32)
        if (options.midText) statusText.value = options.midText
      }
      const result = await task()
      stopTick(100)
      statusText.value = '完成'
      return result
    } catch (e) {
      stopTick(progress.value)
      throw e
    } finally {
      setTimeout(() => {
        visible.value = false
        progress.value = 0
        statusText.value = ''
      }, 700)
    }
  }

  return { visible, progress, statusText, title, run }
}
