import { describe, expect, it } from 'vitest'
import {
  buildForkTimelineLayout,
  cellAt,
  getActiveChain,
  resolveForkSequence,
} from './forkTimelineLayout'
import type { BranchNode, BranchOverviewEntry } from '@/api/client'

const rootEntry: BranchOverviewEntry = {
  version_id: 'v-root',
  label: '主枝',
  is_active: false,
  nodes: [
    { sequence: 0, year: 1990, title: '出生' },
    { sequence: 1, year: 2000, title: '转折' },
    { sequence: 2, year: 2010, title: '原后续A' },
    { sequence: 3, year: 2020, title: '原后续B' },
  ],
}

const childEntry: BranchOverviewEntry = {
  version_id: 'v-child',
  label: '推演分支',
  is_active: true,
  creates_branch: true,
  fork_sequence: 1,
  nodes: [
    { sequence: 0, year: 1990, title: '出生' },
    { sequence: 1, year: 2000, title: '转折(改)' },
    { sequence: 2, year: 2015, title: '新后续A' },
    { sequence: 3, year: 2025, title: '新后续B' },
  ],
}

const roots: BranchNode[] = [
  {
    id: 'v-root',
    label: '主枝',
    node_count: 4,
    is_active: false,
    created_at: '2020-01-01T00:00:00Z',
    children: [
      {
        id: 'v-child',
        parent_id: 'v-root',
        label: '推演分支',
        fork_sequence: 1,
        node_count: 4,
        is_active: true,
        created_at: '2020-01-02T00:00:00Z',
      },
    ],
  },
]

describe('resolveForkSequence', () => {
  it('uses fork_sequence when set', () => {
    expect(resolveForkSequence(childEntry)).toBe(1)
  })

  it('uses override when missing', () => {
    const noFork = { ...childEntry, fork_sequence: undefined }
    expect(resolveForkSequence(noFork, { 'v-child': 2 })).toBe(2)
  })
})

describe('getActiveChain', () => {
  it('returns root to active', () => {
    const flat = [
      roots[0],
      ...(roots[0].children ?? []),
    ]
    expect(getActiveChain('v-child', flat, roots)).toEqual(['v-root', 'v-child'])
  })
})

describe('buildForkTimelineLayout focused', () => {
  it('splits trunk, original tail, and branch at fork', () => {
    const layout = buildForkTimelineLayout({
      branches: [rootEntry, childEntry],
      branchRoots: roots,
      activeVersionId: 'v-child',
      mode: 'focused',
    })

    expect(layout.lanes.map((l) => l.id)).toContain('trunk')
    expect(layout.lanes.map((l) => l.id)).toContain('original_v-root')
    expect(layout.lanes.map((l) => l.id)).toContain('branch_v-child')

    expect(cellAt(layout, 'trunk', 0)?.title).toBe('出生')
    expect(cellAt(layout, 'trunk', 1)?.isForkPoint).toBe(true)
    expect(cellAt(layout, 'original_v-root', 2)?.title).toBe('原后续A')
    expect(cellAt(layout, 'branch_v-child', 2)?.title).toBe('新后续A')
    expect(layout.forkPoints).toHaveLength(1)
    expect(layout.forkPoints[0].createsBranch).toBe(true)
    expect(layout.forkPoints[0].year).toBe(2000)
  })
})

describe('buildForkTimelineLayout expanded', () => {
  it('includes all branch lanes', () => {
    const layout = buildForkTimelineLayout({
      branches: [rootEntry, childEntry],
      branchRoots: roots,
      activeVersionId: 'v-child',
      mode: 'expanded',
    })

    expect(layout.lanes.some((l) => l.id === 'branch_v-child')).toBe(true)
    expect(layout.lanes.some((l) => l.id === 'original_v-root')).toBe(true)
    expect(cellAt(layout, 'branch_v-child', 2)?.title).toBe('新后续A')
  })

  it('does not duplicate original tail for multiple siblings at same fork', () => {
    const zhugeRoot: BranchOverviewEntry = {
      version_id: 'v-root',
      label: '主枝',
      is_active: false,
      nodes: [
        { sequence: 0, year: 221, title: '刘备东征劝谏' },
        { sequence: 1, year: 228, title: '第一次北伐' },
        { sequence: 2, year: 234, title: '病逝五丈原' },
      ],
    }
    const branchA: BranchOverviewEntry = {
      version_id: 'v-a',
      label: '北伐胜利',
      is_active: true,
      creates_branch: true,
      fork_sequence: 1,
      nodes: [
        { sequence: 0, year: 221, title: '刘备东征劝谏' },
        { sequence: 1, year: 228, title: '北伐胜利' },
        { sequence: 2, year: 234, title: '病逝五丈原' },
      ],
    }
    const branchB: BranchOverviewEntry = {
      version_id: 'v-b',
      label: '乘胜抚定陇右',
      is_active: false,
      creates_branch: true,
      fork_sequence: 1,
      nodes: [
        { sequence: 0, year: 221, title: '刘备东征劝谏' },
        { sequence: 1, year: 228, title: '乘胜抚定陇右' },
        { sequence: 2, year: 230, title: '再出祁山' },
        { sequence: 3, year: 234, title: '病逝五丈原' },
      ],
    }
    const multiRoots: BranchNode[] = [
      {
        id: 'v-root',
        label: '主枝',
        node_count: 3,
        is_active: false,
        created_at: '2020-01-01T00:00:00Z',
        children: [
          {
            id: 'v-a',
            parent_id: 'v-root',
            label: '北伐胜利',
            fork_sequence: 1,
            node_count: 3,
            is_active: true,
            created_at: '2020-01-02T00:00:00Z',
          },
          {
            id: 'v-b',
            parent_id: 'v-root',
            label: '乘胜抚定陇右',
            fork_sequence: 1,
            node_count: 4,
            is_active: false,
            created_at: '2020-01-03T00:00:00Z',
          },
        ],
      },
    ]

    const layout = buildForkTimelineLayout({
      branches: [zhugeRoot, branchA, branchB],
      branchRoots: multiRoots,
      activeVersionId: 'v-a',
      mode: 'expanded',
    })

    const originalTail = layout.cells.filter(
      (c) => c.laneId === 'original_v-root' && c.sequence === 2
    )
    expect(originalTail).toHaveLength(1)
    expect(originalTail[0].title).toBe('病逝五丈原')

    expect(cellAt(layout, 'branch_v-a', 1)?.title).toBe('北伐胜利')
    expect(cellAt(layout, 'branch_v-b', 1)?.title).toBe('乘胜抚定陇右')
    expect(cellAt(layout, 'trunk', 1)?.title).toBe('第一次北伐')
  })
})
