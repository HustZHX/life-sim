import type { LayoutLane } from '@/utils/forkTimelineLayout'

export interface LaneColorTheme {
  border: string
  bg: string
  headerBg: string
  headerText: string
  accent: string
  changeBg: string
  changeText: string
}

const TRUNK_THEME: LaneColorTheme = {
  border: '#b3d8ff',
  bg: '#ffffff',
  headerBg: '#ecf5ff',
  headerText: '#409eff',
  accent: '#409eff',
  changeBg: '#fdf6ec',
  changeText: '#b88230',
}

const ORIGINAL_THEME: LaneColorTheme = {
  border: '#dcdfe6',
  bg: '#fafafa',
  headerBg: '#f4f4f5',
  headerText: '#909399',
  accent: '#909399',
  changeBg: '#fef0f0',
  changeText: '#c45656',
}

const BRANCH_PALETTE: LaneColorTheme[] = [
  {
    border: '#b3e19d',
    bg: '#ffffff',
    headerBg: '#f0f9eb',
    headerText: '#67c23a',
    accent: '#67c23a',
    changeBg: '#fdf6ec',
    changeText: '#b88230',
  },
  {
    border: '#d3adf7',
    bg: '#ffffff',
    headerBg: '#f9f0ff',
    headerText: '#9254de',
    accent: '#9254de',
    changeBg: '#fff7e6',
    changeText: '#d48806',
  },
  {
    border: '#ffbb96',
    bg: '#ffffff',
    headerBg: '#fff2e8',
    headerText: '#fa541c',
    accent: '#fa541c',
    changeBg: '#fff1f0',
    changeText: '#cf1322',
  },
  {
    border: '#87e8de',
    bg: '#ffffff',
    headerBg: '#e6fffb',
    headerText: '#13c2c2',
    accent: '#13c2c2',
    changeBg: '#f6ffed',
    changeText: '#389e0d',
  },
  {
    border: '#ffadd2',
    bg: '#ffffff',
    headerBg: '#fff0f6',
    headerText: '#eb2f96',
    accent: '#eb2f96',
    changeBg: '#f9f0ff',
    changeText: '#722ed1',
  },
]

const ACTIVE_BRANCH_THEME: LaneColorTheme = {
  border: '#b3e19d',
  bg: '#ffffff',
  headerBg: '#f0f9eb',
  headerText: '#67c23a',
  accent: '#67c23a',
  changeBg: '#fdf6ec',
  changeText: '#b88230',
}

/** 非当前分支按列顺序依次取色（跳过当前分支用的绿色） */
const INACTIVE_BRANCH_PALETTE: LaneColorTheme[] = BRANCH_PALETTE.filter(
  (t) => t.accent !== ACTIVE_BRANCH_THEME.accent
)

function isActiveBranchLane(
  lane: LayoutLane,
  activeVersionId?: string,
  displayActiveBranchId?: string
): boolean {
  if (lane.isActive) return true
  const branchKey = displayActiveBranchId || activeVersionId
  return !!branchKey && lane.versionId === branchKey
}

/** 为各 lane 分配配色：当前分支固定绿色，其余分支按布局列顺序依次取色。 */
export function buildLaneColorMap(
  lanes: LayoutLane[],
  activeVersionId?: string,
  displayActiveBranchId?: string
): Map<string, LaneColorTheme> {
  const map = new Map<string, LaneColorTheme>()
  let inactiveBranchIdx = 0

  for (const lane of lanes) {
    if (lane.kind === 'trunk') {
      map.set(lane.id, TRUNK_THEME)
    } else if (lane.kind === 'original') {
      map.set(lane.id, ORIGINAL_THEME)
    } else if (lane.kind === 'branch') {
      if (isActiveBranchLane(lane, activeVersionId, displayActiveBranchId)) {
        map.set(lane.id, ACTIVE_BRANCH_THEME)
      } else {
        const palette =
          INACTIVE_BRANCH_PALETTE.length > 0 ? INACTIVE_BRANCH_PALETTE : BRANCH_PALETTE
        map.set(lane.id, palette[inactiveBranchIdx % palette.length])
        inactiveBranchIdx++
      }
    }
  }
  return map
}

export function laneThemeVars(theme: LaneColorTheme): Record<string, string> {
  return {
    '--lane-border': theme.border,
    '--lane-bg': theme.bg,
    '--lane-accent': theme.accent,
    '--lane-change-bg': theme.changeBg,
    '--lane-change-text': theme.changeText,
  }
}

export function headerThemeStyle(theme: LaneColorTheme): Record<string, string> {
  return {
    background: theme.headerBg,
    color: theme.headerText,
    borderColor: theme.border,
  }
}
