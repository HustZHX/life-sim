export const DEFAULT_TARGET_NODE_COUNT = 20
export const MIN_TARGET_NODE_COUNT = 8
export const MAX_TARGET_NODE_COUNT = 40

/** 档案未标注卒年或仍在世时，用当前年份作为时间轴上界 */
export function effectiveDeathYear(profile: { birth_year: number; death_year: number }): number {
  if (profile.death_year > profile.birth_year) {
    return profile.death_year
  }
  const y = new Date().getFullYear()
  if (y > profile.birth_year) return y
  return profile.birth_year > 0 ? profile.birth_year + 1 : y
}

export function isLivingProfile(profile: { birth_year: number; death_year: number }): boolean {
  return profile.death_year <= 0 || profile.death_year <= profile.birth_year
}

export function formatProfileLifeSpan(profile: { birth_year: number; death_year: number }): string {
  if (isLivingProfile(profile)) {
    return `${profile.birth_year} — 至今`
  }
  return `${profile.birth_year} — ${profile.death_year}`
}

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
  _anchorYear: number,
  _deathYear: number,
  targetNodeCount: number
): string {
  return `后续约 ${targetNodeCount} 个节点（与推演卒年无机械对应）`
}

const CHUNK_MIN_NODES = 15
const CHUNK_MIN_SPAN = 60

/** 与后端 splitTimelineRange 规则一致的段数估算 */
export function estimateTimelineChunks(targetNodeCount: number, spanYears: number): number {
  if (targetNodeCount <= CHUNK_MIN_NODES && spanYears <= CHUNK_MIN_SPAN) {
    return 1
  }
  let numChunks = 2
  if (targetNodeCount > 24 || spanYears > 120) numChunks = 3
  if (targetNodeCount > 36 || spanYears > 200) numChunks = 4
  return numChunks
}
