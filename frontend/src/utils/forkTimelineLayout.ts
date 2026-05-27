import type { BranchNode, BranchOverviewEntry, BranchOverviewNodeLite, TraitChange } from '@/api/client'

export type LayoutMode = 'focused' | 'expanded'

export type LaneKind = 'trunk' | 'original' | 'branch'

export interface LayoutLane {
  id: string
  kind: LaneKind
  versionId?: string
  label: string
  isActive: boolean
}

export interface LayoutCell {
  laneId: string
  versionId: string
  sequence: number
  year: number
  title: string
  traitChanges?: TraitChange[]
  isForkRow: boolean
  isForkPoint: boolean
  isActivePath: boolean
}

export interface ForkPoint {
  sequence: number
  year: number
  parentVersionId: string
  childVersionId: string
  childLabel: string
  /** 是否为「修改并推演后续」产生的分叉 */
  createsBranch: boolean
}

export interface LayoutEdge {
  fromLaneId: string
  toLaneId: string
  sequence: number
  kind: 'fork_horizontal' | 'vertical'
}

export interface ForkTimelineLayout {
  lanes: LayoutLane[]
  cells: LayoutCell[]
  forkPoints: ForkPoint[]
  edges: LayoutEdge[]
  rowSequences: number[]
}

export interface ForkTimelineLayoutInput {
  branches: BranchOverviewEntry[]
  branchRoots: BranchNode[]
  activeVersionId: string
  /** 折叠后分支树上用于定位活跃链的版本 id */
  displayActiveBranchId?: string
  mode: LayoutMode
  /** sequence -> fork for versions missing fork_sequence */
  forkSequenceOverrides?: Record<string, number>
}

export function resolveForkSequence(
  entry: BranchOverviewEntry,
  overrides?: Record<string, number>
): number {
  if (overrides?.[entry.version_id] != null) {
    return overrides[entry.version_id]
  }
  if (entry.fork_sequence != null && entry.fork_sequence >= 0) {
    return entry.fork_sequence
  }
  return 0
}

function branchMap(branches: BranchOverviewEntry[]): Map<string, BranchOverviewEntry> {
  return new Map(branches.map((b) => [b.version_id, b]))
}

function flattenTree(roots: BranchNode[]): BranchNode[] {
  const out: BranchNode[] = []
  const walk = (n: BranchNode) => {
    out.push(n)
    for (const c of n.children ?? []) walk(c)
  }
  for (const r of roots) walk(r)
  return out
}

function childrenOf(flat: BranchNode[]): Map<string, BranchNode[]> {
  const m = new Map<string, BranchNode[]>()
  for (const n of flat) {
    if (!n.parent_id) continue
    const list = m.get(n.parent_id) ?? []
    list.push(n)
    m.set(n.parent_id, list)
  }
  for (const [, list] of m) {
    list.sort((a, b) => a.created_at.localeCompare(b.created_at))
  }
  return m
}

/** 从激活版本沿 parent_id 回溯到根（使用折叠后的分支树 id） */
export function getActiveChain(
  activeVersionId: string,
  flat: BranchNode[],
  roots: BranchNode[],
  displayActiveBranchId?: string
): string[] {
  const treeActiveId = displayActiveBranchId || activeVersionId
  const byId = new Map(flat.map((n) => [n.id, n]))
  const rootIds = new Set(roots.map((r) => r.id))

  const chain: string[] = []
  let cur: string | undefined = treeActiveId
  const seen = new Set<string>()
  while (cur && !seen.has(cur)) {
    seen.add(cur)
    chain.unshift(cur)
    const node = byId.get(cur)
    const pid = node?.parent_id
    if (!pid || !byId.has(pid)) break
    cur = pid
  }

  if (chain.length === 0 && roots.length) {
    const activeRoot = roots.find((r) => r.id === treeActiveId)
    if (activeRoot) return [activeRoot.id]
    return [roots[0].id]
  }

  if (chain.length > 0 && !rootIds.has(chain[0])) {
    const first = byId.get(chain[0])
    if (first?.parent_id && byId.has(first.parent_id)) {
      // parent missing from flat — keep chain as-is
    }
  }

  return chain
}

function childCreatesBranch(
  childId: string,
  map: Map<string, BranchOverviewEntry>
): boolean {
  return map.get(childId)?.creates_branch ?? false
}

function nodeAt(nodes: BranchOverviewNodeLite[], seq: number): BranchOverviewNodeLite | undefined {
  return nodes.find((n) => n.sequence === seq)
}

function pushCell(
  cells: LayoutCell[],
  lane: LayoutLane,
  node: BranchOverviewNodeLite,
  opts: { isForkRow: boolean; isForkPoint: boolean; isActivePath: boolean }
) {
  cells.push({
    laneId: lane.id,
    versionId: lane.versionId ?? '',
    sequence: node.sequence,
    year: node.year,
    title: node.title,
    traitChanges: node.trait_changes?.length ? node.trait_changes : undefined,
    isForkRow: opts.isForkRow,
    isForkPoint: opts.isForkPoint,
    isActivePath: opts.isActivePath,
  })
}

/** 同一 lane + sequence 只保留首个 cell，避免多 sibling 展开时重复写入 */
function dedupeCells(cells: LayoutCell[]): LayoutCell[] {
  const seen = new Map<string, LayoutCell>()
  for (const c of cells) {
    const key = `${c.laneId}:${c.sequence}`
    if (!seen.has(key)) seen.set(key, c)
  }
  return [...seen.values()]
}

function ensureLane(
  lanes: LayoutLane[],
  id: string,
  kind: LaneKind,
  label: string,
  versionId: string | undefined,
  isActive: boolean
): LayoutLane {
  let lane = lanes.find((l) => l.id === id)
  if (!lane) {
    lane = { id, kind, label, versionId, isActive }
    lanes.push(lane)
  }
  return lane
}

function buildSingleTrunk(
  entry: BranchOverviewEntry,
  activeVersionId: string
): ForkTimelineLayout {
  const lane: LayoutLane = {
    id: 'trunk',
    kind: 'trunk',
    label: entry.label || '时间轴',
    versionId: entry.version_id,
    isActive: entry.version_id === activeVersionId,
  }
  const cells: LayoutCell[] = entry.nodes.map((n) => ({
    laneId: 'trunk',
    versionId: entry.version_id,
    sequence: n.sequence,
    year: n.year,
    title: n.title,
    traitChanges: n.trait_changes?.length ? n.trait_changes : undefined,
    isForkRow: false,
    isForkPoint: false,
    isActivePath: true,
  }))
  const rowSequences = [...new Set(entry.nodes.map((n) => n.sequence))].sort((a, b) => a - b)
  return { lanes: [lane], cells, forkPoints: [], edges: [], rowSequences }
}

function buildFocusedLayout(
  chain: string[],
  map: Map<string, BranchOverviewEntry>,
  activeVersionId: string,
  overrides?: Record<string, number>
): ForkTimelineLayout {
  if (chain.length === 0) {
    return { lanes: [], cells: [], forkPoints: [], edges: [], rowSequences: [] }
  }
  if (chain.length === 1) {
    const entry = map.get(chain[0])
    if (!entry) return { lanes: [], cells: [], forkPoints: [], edges: [], rowSequences: [] }
    return buildSingleTrunk(entry, activeVersionId)
  }

  const lanes: LayoutLane[] = []
  const cells: LayoutCell[] = []
  const forkPoints: ForkPoint[] = []
  const edges: LayoutEdge[] = []
  const rowSet = new Set<number>()

  ensureLane(lanes, 'trunk', 'trunk', '共享主干', chain[0], chain[chain.length - 1] === chain[0])

  for (let i = 1; i < chain.length; i++) {
    const parentId = chain[i - 1]
    const childId = chain[i]
    const parent = map.get(parentId)
    const child = map.get(childId)
    if (!parent || !child) continue

    const forkSeq = resolveForkSequence(child, overrides)
    const forkNode = nodeAt(parent.nodes, forkSeq) ?? nodeAt(child.nodes, forkSeq)
    const forkYear = forkNode?.year ?? 0

    forkPoints.push({
      sequence: forkSeq,
      year: forkYear,
      parentVersionId: parentId,
      childVersionId: childId,
      childLabel: child.label,
      createsBranch: childCreatesBranch(childId, map),
    })

    const originalLaneId = `original_${parentId}`
    const branchLaneId = `branch_${childId}`
    const originalLane = ensureLane(
      lanes,
      originalLaneId,
      'original',
      '原本时间轴后续',
      parentId,
      false
    )
    const branchLane = ensureLane(
      lanes,
      branchLaneId,
      'branch',
      child.label || '新时间轴',
      childId,
      childId === activeVersionId
    )

    edges.push({
      fromLaneId: i === 1 ? 'trunk' : `branch_${parentId}`,
      toLaneId: branchLaneId,
      sequence: forkSeq,
      kind: 'fork_horizontal',
    })

    if (i === 1) {
      for (const n of parent.nodes) {
        if (n.sequence < forkSeq) {
          rowSet.add(n.sequence)
          pushCell(cells, lanes.find((l) => l.id === 'trunk')!, n, {
            isForkRow: false,
            isForkPoint: false,
            isActivePath: true,
          })
        }
      }
      const trunkFork = nodeAt(parent.nodes, forkSeq)
      if (trunkFork) {
        rowSet.add(forkSeq)
        pushCell(cells, lanes.find((l) => l.id === 'trunk')!, trunkFork, {
          isForkRow: true,
          isForkPoint: true,
          isActivePath: true,
        })
      }
    } else {
      const prevBranchId = `branch_${parentId}`
      const prevLane = lanes.find((l) => l.id === prevBranchId)!
      for (const n of parent.nodes) {
        if (n.sequence < forkSeq) {
          rowSet.add(n.sequence)
          pushCell(cells, prevLane, n, {
            isForkRow: false,
            isForkPoint: false,
            isActivePath: true,
          })
        }
      }
      const spineFork = nodeAt(parent.nodes, forkSeq)
      if (spineFork) {
        rowSet.add(forkSeq)
        pushCell(cells, prevLane, spineFork, {
          isForkRow: true,
          isForkPoint: true,
          isActivePath: true,
        })
      }
    }

    const childForkNode = nodeAt(child.nodes, forkSeq)
    if (childForkNode) {
      rowSet.add(forkSeq)
      pushCell(cells, branchLane, childForkNode, {
        isForkRow: true,
        isForkPoint: false,
        isActivePath: childId === activeVersionId,
      })
    }

    for (const n of parent.nodes) {
      if (n.sequence > forkSeq) {
        rowSet.add(n.sequence)
        pushCell(cells, originalLane, n, {
          isForkRow: false,
          isForkPoint: false,
          isActivePath: false,
        })
      }
    }

    for (const n of child.nodes) {
      if (n.sequence > forkSeq) {
        rowSet.add(n.sequence)
        pushCell(cells, branchLane, n, {
          isForkRow: false,
          isForkPoint: false,
          isActivePath: childId === activeVersionId,
        })
      }
    }
  }

  const rowSequences = [...rowSet].sort((a, b) => a - b)
  return { lanes, cells, forkPoints, edges, rowSequences }
}

function collectVisibleVersionIdsFocused(
  chain: string[],
  map: Map<string, BranchOverviewEntry>
): Set<string> {
  const s = new Set(chain)
  for (const id of chain) {
    const e = map.get(id)
    if (e) s.add(e.version_id)
  }
  return s
}

function buildExpandedLayout(
  roots: BranchNode[],
  map: Map<string, BranchOverviewEntry>,
  activeVersionId: string,
  overrides?: Record<string, number>
): ForkTimelineLayout {
  const flat = flattenTree(roots)
  const childMap = childrenOf(flat)
  const lanes: LayoutLane[] = []
  const cells: LayoutCell[] = []
  const forkPoints: ForkPoint[] = []
  const edges: LayoutEdge[] = []
  const rowSet = new Set<number>()

  /** spineStartSeq：该泳道从哪个 sequence 开始展示（含） */
  function walk(
    versionId: string,
    spineLaneId: string,
    isRoot: boolean,
    spineStartSeq = 0
  ) {
    const entry = map.get(versionId)
    if (!entry) return

    const children = childMap.get(versionId) ?? []
    if (!children.length) {
      const lane = ensureLane(
        lanes,
        spineLaneId,
        isRoot ? 'trunk' : 'branch',
        entry.label,
        versionId,
        versionId === activeVersionId
      )
      for (const n of entry.nodes) {
        if (n.sequence < spineStartSeq) continue
        rowSet.add(n.sequence)
        pushCell(cells, lane, n, {
          isForkRow: n.sequence === spineStartSeq && !isRoot,
          isForkPoint: false,
          isActivePath: versionId === activeVersionId,
        })
      }
      return
    }

    const minForkAmongChildren = Math.min(
      ...children.map((c) => {
        const e = map.get(c.id)
        return e ? resolveForkSequence(e, overrides) : 0
      })
    )

    const spineLane = ensureLane(
      lanes,
      spineLaneId,
      isRoot ? 'trunk' : 'branch',
      isRoot ? '共享主干' : entry.label,
      isRoot ? versionId : versionId,
      !isRoot && versionId === activeVersionId
    )

    for (const n of entry.nodes) {
      if (n.sequence < spineStartSeq || n.sequence >= minForkAmongChildren) continue
      rowSet.add(n.sequence)
      pushCell(cells, spineLane, n, {
        isForkRow: false,
        isForkPoint: false,
        isActivePath: isRoot ? false : versionId === activeVersionId,
      })
    }

    const forkNode = nodeAt(entry.nodes, minForkAmongChildren)
    if (forkNode) {
      rowSet.add(minForkAmongChildren)
      pushCell(cells, spineLane, forkNode, {
        isForkRow: true,
        isForkPoint: true,
        isActivePath: isRoot ? false : versionId === activeVersionId,
      })
    }

    const originalLaneId = `original_${versionId}`
    ensureLane(lanes, originalLaneId, 'original', '原本时间轴后续', versionId, false)

    for (const child of children) {
      const childEntry = map.get(child.id)
      if (!childEntry) continue
      const forkSeq = resolveForkSequence(childEntry, overrides)
      const forkNodeMeta = nodeAt(entry.nodes, forkSeq) ?? nodeAt(childEntry.nodes, forkSeq)
      forkPoints.push({
        sequence: forkSeq,
        year: forkNodeMeta?.year ?? 0,
        parentVersionId: versionId,
        childVersionId: child.id,
        childLabel: childEntry.label,
        createsBranch: childCreatesBranch(child.id, map),
      })

      const branchLaneId = `branch_${child.id}`
      ensureLane(
        lanes,
        branchLaneId,
        'branch',
        childEntry.label,
        child.id,
        child.id === activeVersionId
      )

      edges.push({
        fromLaneId: isRoot ? 'trunk' : spineLaneId,
        toLaneId: branchLaneId,
        sequence: forkSeq,
        kind: 'fork_horizontal',
      })

      walk(child.id, branchLaneId, false, forkSeq)
    }

    const originalLane = lanes.find((l) => l.id === originalLaneId)!
    for (const n of entry.nodes) {
      if (n.sequence <= minForkAmongChildren) continue
      rowSet.add(n.sequence)
      pushCell(cells, originalLane, n, {
        isForkRow: false,
        isForkPoint: false,
        isActivePath: false,
      })
    }
  }

  for (const root of roots) {
    if (map.has(root.id)) walk(root.id, 'trunk', true, 0)
  }

  const rowSequences = [...rowSet].sort((a, b) => a - b)
  return { lanes, cells: dedupeCells(cells), forkPoints, edges, rowSequences }
}

export function buildForkTimelineLayout(input: ForkTimelineLayoutInput): ForkTimelineLayout {
  const map = branchMap(input.branches)
  if (!input.branches.length) {
    return { lanes: [], cells: [], forkPoints: [], edges: [], rowSequences: [] }
  }

  const flat = flattenTree(input.branchRoots)

  if (input.mode === 'focused') {
    const chain = getActiveChain(
      input.activeVersionId,
      flat,
      input.branchRoots,
      input.displayActiveBranchId
    )
    const visible = collectVisibleVersionIdsFocused(chain, map)
    const filteredBranches = input.branches.filter((b) => visible.has(b.version_id))
    const filteredMap = branchMap(filteredBranches.length ? filteredBranches : input.branches)
    return buildFocusedLayout(chain, filteredMap, input.activeVersionId, input.forkSequenceOverrides)
  }

  return buildExpandedLayout(
    input.branchRoots,
    map,
    input.activeVersionId,
    input.forkSequenceOverrides
  )
}

/** 单元测试用：按 lane + sequence 取 cell */
export function cellAt(
  layout: ForkTimelineLayout,
  laneId: string,
  sequence: number
): LayoutCell | undefined {
  return layout.cells.find((c) => c.laneId === laneId && c.sequence === sequence)
}
