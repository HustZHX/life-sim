import type { LifeNode, Profile } from '@/api/client'
import { fieldLabel } from '@/constants/fieldLabels'
import { sceneDisplayItems } from '@/utils/nodeScene'

function line(label: string, value: string | undefined | null): string {
  const v = value?.trim()
  if (!v) return ''
  return `${label}：${v}`
}

function formatTraitChanges(node: LifeNode): string {
  if (!node.trait_changes?.length) return ''
  const rows = node.trait_changes.map((t) => {
    const label = fieldLabel(t.field, 'trait')
    const change = `${t.before} → ${t.after}`.trim()
    const reason = t.reason?.trim()
    return reason ? `  · ${label}：${change}（${reason}）` : `  · ${label}：${change}`
  })
  return ['性格/思想变更：', ...rows].join('\n')
}

function formatScene(node: LifeNode): string {
  const items = sceneDisplayItems(node.scene)
  if (!items.length) return ''
  return `场景：${items.map((i) => `${i.label} ${i.value}`).join(' · ')}`
}

function formatNode(node: LifeNode, index: number): string {
  const header = `【节点 ${index + 1}】${node.year} 年 · ${node.age} 岁 · ${node.title}`
  const parts = [
    header,
    line('经历', node.events),
    line('内心想法', node.thoughts),
    line('性格快照', node.personality_snapshot),
    formatTraitChanges(node),
    formatScene(node),
  ].filter(Boolean)
  return parts.join('\n')
}

/** 将人物档案与时间轴节点格式化为可复制的纯文本 */
export function formatTimelineExport(profile: Profile | null | undefined, nodes: LifeNode[]): string {
  const sorted = [...nodes].sort((a, b) => a.sequence - b.sequence)
  const lines: string[] = []

  if (profile) {
    const span =
      profile.birth_year && profile.death_year
        ? `${profile.birth_year}—${profile.death_year} 年`
        : ''
    lines.push(`【人物】${profile.display_name}${span ? `（${span}）` : ''}`)
    if (profile.era?.trim()) lines.push(line('时代', profile.era))
    if (profile.personality_initial?.trim()) lines.push(line('初始性格', profile.personality_initial))
    if (profile.beliefs_motto?.trim()) lines.push(line('信念', profile.beliefs_motto))
    if (profile.experiences?.trim()) lines.push(line('经历摘要', profile.experiences))
    lines.push('')
  }

  if (!sorted.length) {
    lines.push('（暂无时间轴节点）')
    return lines.join('\n').trim()
  }

  lines.push(`【人生时间轴】共 ${sorted.length} 个节点`)
  lines.push('')

  sorted.forEach((node, i) => {
    if (i > 0) lines.push('')
    lines.push('---')
    lines.push(formatNode(node, i))
  })

  lines.push('')
  lines.push('---')
  lines.push('（由 Life-Sim 导出 · AI 生成内容仅供参考）')

  return lines.join('\n').trim()
}

export async function copyTextToClipboard(text: string): Promise<void> {
  if (navigator.clipboard?.writeText) {
    await navigator.clipboard.writeText(text)
    return
  }
  const ta = document.createElement('textarea')
  ta.value = text
  ta.style.position = 'fixed'
  ta.style.left = '-9999px'
  document.body.appendChild(ta)
  ta.select()
  document.execCommand('copy')
  document.body.removeChild(ta)
}
