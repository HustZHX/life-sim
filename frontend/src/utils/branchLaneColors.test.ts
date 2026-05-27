import { describe, expect, it } from 'vitest'
import { buildLaneColorMap } from './branchLaneColors'
import type { LayoutLane } from './forkTimelineLayout'

describe('buildLaneColorMap', () => {
  it('assigns distinct colors to branch lanes', () => {
    const lanes: LayoutLane[] = [
      { id: 'trunk', kind: 'trunk', label: '主线', versionId: 'v1', isActive: true },
      { id: 'original_v1', kind: 'original', label: '原本', versionId: 'v1', isActive: false },
      { id: 'branch_v2', kind: 'branch', label: '分支A', versionId: 'v2', isActive: false },
      { id: 'branch_v3', kind: 'branch', label: '分支B', versionId: 'v3', isActive: true },
    ]
    const map = buildLaneColorMap(lanes, 'v3')
    expect(map.get('trunk')?.accent).toBe('#409eff')
    expect(map.get('branch_v3')?.accent).toBe('#67c23a')
    expect(map.get('branch_v2')?.accent).not.toBe('#67c23a')
    expect(map.get('branch_v2')?.accent).not.toBe(map.get('branch_v3')?.accent)
  })

  it('colors inactive branches by layout order not version id sort', () => {
    const lanes: LayoutLane[] = [
      { id: 'branch_v-z', kind: 'branch', label: '旧分支', versionId: 'v-z', isActive: false },
      { id: 'branch_v-a', kind: 'branch', label: '新分支', versionId: 'v-a', isActive: true },
    ]
    const map = buildLaneColorMap(lanes, 'v-a')
    expect(map.get('branch_v-a')?.accent).toBe('#67c23a')
    expect(map.get('branch_v-z')?.accent).toBe('#9254de')
  })
})
