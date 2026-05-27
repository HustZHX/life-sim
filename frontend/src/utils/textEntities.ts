import type { LifeNode, Profile } from '@/api/client'

export interface NodeEntities {
  protagonist?: string[]
  persons?: string[]
  places?: string[]
}

export type HighlightKind = 'text' | 'protagonist' | 'person' | 'place'

export interface EntityHighlightContext {
  protagonist: string[]
  persons: string[]
  places: string[]
}

export interface TextSegment {
  kind: HighlightKind
  value: string
}

function sortByLength(items: string[]) {
  return [...items].sort((a, b) => b.length - a.length)
}

/** 从 AI 标注的节点实体构建高亮上下文；旧节点无 entities 时仅标主角名 */
export function entityContextFromNode(
  node: LifeNode | null,
  profile: Profile | null
): EntityHighlightContext {
  const fallbackProtagonist = profile?.display_name?.trim()
    ? [profile.display_name.trim()]
    : []

  if (node?.entities) {
    const protagonist =
      node.entities.protagonist?.length ? node.entities.protagonist : fallbackProtagonist
    return {
      protagonist: sortByLength(protagonist),
      persons: sortByLength(node.entities.persons || []),
      places: sortByLength(node.entities.places || []),
    }
  }

  return {
    protagonist: fallbackProtagonist,
    persons: [],
    places: [],
  }
}

interface Match {
  start: number
  end: number
  kind: HighlightKind
  priority: number
}

function findTermMatches(
  text: string,
  term: string,
  kind: HighlightKind,
  priority: number
): Match[] {
  const out: Match[] = []
  if (!term) return out
  let from = 0
  while (from <= text.length - term.length) {
    const idx = text.indexOf(term, from)
    if (idx === -1) break
    out.push({ start: idx, end: idx + term.length, kind, priority })
    from = idx + term.length
  }
  return out
}

function resolveOverlaps(matches: Match[]): Match[] {
  if (!matches.length) return []
  const sorted = [...matches].sort((a, b) => {
    if (a.start !== b.start) return a.start - b.start
    if (a.priority !== b.priority) return b.priority - a.priority
    return b.end - b.start - (a.end - a.start)
  })
  const picked: Match[] = []
  for (const m of sorted) {
    if (picked.some((p) => m.start < p.end && m.end > p.start)) continue
    picked.push(m)
  }
  return picked.sort((a, b) => a.start - b.start)
}

/** 按 AI 返回的实体列表切分文本 */
export function segmentEntities(text: string, ctx: EntityHighlightContext): TextSegment[] {
  if (!text) return [{ kind: 'text', value: '' }]

  const matches: Match[] = []
  for (const name of ctx.protagonist) {
    matches.push(...findTermMatches(text, name, 'protagonist', 4))
  }
  for (const name of ctx.persons) {
    matches.push(...findTermMatches(text, name, 'person', 3))
  }
  for (const name of ctx.places) {
    matches.push(...findTermMatches(text, name, 'place', 2))
  }

  const resolved = resolveOverlaps(matches)
  if (!resolved.length) return [{ kind: 'text', value: text }]

  const segments: TextSegment[] = []
  let cursor = 0
  for (const hit of resolved) {
    if (hit.start > cursor) {
      segments.push({ kind: 'text', value: text.slice(cursor, hit.start) })
    }
    segments.push({ kind: hit.kind, value: text.slice(hit.start, hit.end) })
    cursor = hit.end
  }
  if (cursor < text.length) {
    segments.push({ kind: 'text', value: text.slice(cursor) })
  }
  return segments
}
