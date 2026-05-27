import { onUnmounted, ref, watch, type Ref } from 'vue'

/** 监听元素宽度变化，用于容器内响应式布局。 */
export function useElementWidth(target: Ref<HTMLElement | null | undefined>) {
  const width = ref(0)
  let observer: ResizeObserver | null = null

  const stop = watch(
    target,
    (el) => {
      observer?.disconnect()
      observer = null
      if (!el) {
        width.value = 0
        return
      }
      observer = new ResizeObserver((entries) => {
        const entry = entries[0]
        if (entry) width.value = entry.contentRect.width
      })
      observer.observe(el)
      width.value = el.getBoundingClientRect().width
    },
    { immediate: true }
  )

  onUnmounted(() => {
    stop()
    observer?.disconnect()
  })

  return { width }
}

export type GraphLayoutDensity = 'comfortable' | 'compact' | 'dense'

export function resolveGraphDensity(
  containerWidth: number,
  laneCount: number
): GraphLayoutDensity {
  if (containerWidth <= 0 || laneCount <= 0) return 'comfortable'
  const perLane = containerWidth / laneCount
  if (perLane < 220) return 'dense'
  if (perLane < 340) return 'compact'
  return 'comfortable'
}
