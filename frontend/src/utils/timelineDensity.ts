export const DEFAULT_TARGET_NODE_COUNT = 20
export const MIN_TARGET_NODE_COUNT = 8
export const MAX_TARGET_NODE_COUNT = 40

/** 按区间跨度与目标节点数计算参考间隔（年） */
export function computeStepYears(spanYears: number, targetNodes: number): number {
  const target = targetNodes > 0 ? targetNodes : DEFAULT_TARGET_NODE_COUNT
  if (spanYears <= 0) return 1
  return Math.max(1, Math.floor(spanYears / target))
}

export function formatTimelineSummary(
  startYear: number,
  endYear: number,
  targetNodeCount: number
): string {
  const span = Math.max(0, endYear - startYear)
  const step = computeStepYears(span, targetNodeCount)
  return `${startYear}—${endYear} 年 · 约 ${targetNodeCount} 个节点 · 参考间隔 ${step} 年`
}

export function formatRegenSummary(
  anchorYear: number,
  deathYear: number,
  targetNodeCount: number
): string {
  const span = Math.max(0, deathYear - anchorYear)
  const step = computeStepYears(span, targetNodeCount)
  return `后续约 ${targetNodeCount} 个节点 · 跨度 ${span} 年 · 参考间隔 ${step} 年`
}
