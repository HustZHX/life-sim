export type TimelineJobType =
  | 'timeline_generate'
  | 'timeline_regenerate'
  | 'node_inner_current'
  | 'node_inner_subsequent'
  | 'timeline_narrative_change'
  | 'world_line_sync'
  | 'world_line_refresh'

export interface TimelineActiveJob {
  id: string
  type: TimelineJobType | string
  status: string
  progress: number
  stage_text?: string
}

const TIMELINE_JOB_LABELS: Record<string, string> = {
  timeline_generate: '生成时间轴',
  timeline_regenerate: '推演后续',
  node_inner_current: '更新节点',
  node_inner_subsequent: '推演内心',
  timeline_narrative_change: '叙述变更',
  world_line_sync: '世界线同步',
  world_line_refresh: '刷新世界线',
}

export function timelineJobLabel(type: string): string {
  return TIMELINE_JOB_LABELS[type] || type
}

export type TimelineGenerationStatus = 'ready' | 'generating' | 'failed'

export function timelineGenerationLabel(status?: TimelineGenerationStatus): string {
  switch (status) {
    case 'generating':
      return '生成中'
    case 'failed':
      return '生成失败'
    case 'ready':
    default:
      return '就绪'
  }
}

export function isTimelineGenerating(tl: {
  generation_status?: TimelineGenerationStatus
  active_job?: { status?: string }
  current_version_id?: string
}): boolean {
  if (tl.generation_status === 'generating') return true
  if (tl.active_job?.status === 'pending' || tl.active_job?.status === 'running') return true
  return !tl.current_version_id && tl.generation_status !== 'failed' && tl.generation_status !== 'ready'
}

export function isTimelineReady(tl: {
  generation_status?: TimelineGenerationStatus
  current_version_id?: string
}): boolean {
  return tl.generation_status === 'ready' || (!!tl.current_version_id && tl.generation_status !== 'generating')
}

export function versionChangeLabel(count?: number): string {
  const n = count ?? 0
  if (n <= 0) return '—'
  if (n === 1) return '1（初始）'
  return `${n}（+${n - 1} 次变更）`
}
