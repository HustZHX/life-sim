<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { ArrowDown, ArrowRight } from '@element-plus/icons-vue'
import type { BranchNode, BranchOverviewEntry } from '@/api/client'
import { useForkTimelineLayout } from '@/composables/useForkTimelineLayout'
import { resolveGraphDensity, useElementWidth } from '@/composables/useElementWidth'
import type { ForkPoint, LayoutCell, LayoutLane } from '@/utils/forkTimelineLayout'
import { buildLaneColorMap, headerThemeStyle, type LaneColorTheme } from '@/utils/branchLaneColors'
import ForkTimelineNodeCard from '@/components/ForkTimelineNodeCard.vue'

const props = withDefaults(
  defineProps<{
    branches: BranchOverviewEntry[]
    branchRoots: BranchNode[]
    activeVersionId: string
    displayActiveBranchId?: string
    mode: 'focused' | 'expanded'
    /** 详情面板收起时拉宽节点列，占满可用宽度 */
    wide?: boolean
    loading?: boolean
    selectedVersionId?: string
    selectedSequence?: number
    forkSequenceOverrides?: Record<string, number>
    /** 编辑后滚到该分叉点 */
    scrollToForkSequence?: number | null
  }>(),
  {
    selectedVersionId: '',
    selectedSequence: -1,
    scrollToForkSequence: null,
    wide: false,
    displayActiveBranchId: '',
  }
)

const emit = defineEmits<{
  select: [payload: { versionId: string; sequence: number }]
  activate: [versionId: string]
  showDiff: [fork: ForkPoint]
}>()

const collapsedLaneIds = ref<Set<string>>(new Set())

const { layout } = useForkTimelineLayout({
  branches: () => props.branches,
  branchRoots: () => props.branchRoots,
  activeVersionId: () => props.activeVersionId,
  displayActiveBranchId: () => props.displayActiveBranchId || props.activeVersionId,
  mode: () => props.mode,
  forkSequenceOverrides: () => props.forkSequenceOverrides,
})

const visibleLanes = computed(() => layout.value.lanes)

function laneColumnStyle(lane: LayoutLane, density: string, wide: boolean): string {
  if (!isLaneCollapsed(lane.id)) {
    const minCol =
      density === 'dense' ? 108 : density === 'compact' ? 140 : wide ? 200 : 168
    const w = containerWidth.value
    const count = Math.max(visibleLanes.value.filter((l) => !isLaneCollapsed(l.id)).length, 1)
    const gap = density === 'dense' ? 6 : density === 'compact' ? 8 : 12
    if (w > 0) {
      const collapsedCount = visibleLanes.value.filter((l) => isLaneCollapsed(l.id)).length
      const collapsedWidth = collapsedCount * 36
      const usable = Math.max(w - gap * (visibleLanes.value.length - 1) - collapsedWidth, count * 96)
      const fitMin = Math.max(96, Math.min(minCol, Math.floor(usable / count)))
      return `minmax(${fitMin}px, 1fr)`
    }
    return `minmax(${minCol}px, 1fr)`
  }
  return '36px'
}

watch(
  () => [props.mode, props.branchRoots.map((b) => b.id).join(',')],
  () => {
    collapsedLaneIds.value = new Set()
  }
)

function canCollapseLane(lane: LayoutLane): boolean {
  return lane.kind !== 'trunk'
}

function descendantLaneIds(laneId: string): string[] {
  const match = laneId.match(/^(branch|original)_(.+)$/)
  if (!match) return []
  const versionId = match[2]
  const ids: string[] = []

  const collectDescendants = (node: BranchNode) => {
    ids.push(`branch_${node.id}`, `original_${node.id}`)
    for (const child of node.children ?? []) collectDescendants(child)
  }

  const walk = (nodes: BranchNode[]): boolean => {
    for (const node of nodes) {
      if (node.id === versionId) {
        for (const child of node.children ?? []) collectDescendants(child)
        return true
      }
      if (node.children?.length && walk(node.children)) return true
    }
    return false
  }

  walk(props.branchRoots)
  return ids
}

function isLaneCollapsed(laneId: string): boolean {
  return collapsedLaneIds.value.has(laneId)
}

function toggleLaneCollapse(lane: LayoutLane) {
  const next = new Set(collapsedLaneIds.value)
  if (next.has(lane.id)) {
    next.delete(lane.id)
    for (const id of descendantLaneIds(lane.id)) next.delete(id)
  } else {
    next.add(lane.id)
    for (const id of descendantLaneIds(lane.id)) next.add(id)
  }
  collapsedLaneIds.value = next
}

function collapseInactiveLanes() {
  const next = new Set<string>()
  for (const lane of layout.value.lanes) {
    if (canCollapseLane(lane) && !lane.isActive) next.add(lane.id)
  }
  collapsedLaneIds.value = next
}

function expandAllLanes() {
  collapsedLaneIds.value = new Set()
}

const laneColors = computed(() =>
  buildLaneColorMap(
    layout.value.lanes,
    props.activeVersionId,
    props.displayActiveBranchId || props.activeVersionId
  )
)

function laneTheme(laneId: string): LaneColorTheme | undefined {
  return laneColors.value.get(laneId)
}

const graphRef = ref<HTMLElement | null>(null)
const { width: containerWidth } = useElementWidth(graphRef)

const laneCount = computed(() => Math.max(visibleLanes.value.length, 1))

const layoutDensity = computed(() =>
  resolveGraphDensity(containerWidth.value, laneCount.value)
)

const laneGridStyle = computed(() => {
  const lanes = visibleLanes.value
  const count = Math.max(lanes.length, 1)
  const density = layoutDensity.value
  const rowCount = layout.value.rowSequences.length
  const gap = density === 'dense' ? 6 : density === 'compact' ? 8 : 12

  const base: Record<string, string> = {
    display: 'grid',
    gap: `${gap}px`,
    width: '100%',
  }

  if (props.wide && count === 1 && !collapsedLaneIds.value.size) {
    return {
      ...base,
      gridTemplateColumns: '1fr',
      gridTemplateRows: `auto repeat(${rowCount}, auto)`,
    }
  }

  const columns = lanes.map((lane) => laneColumnStyle(lane, density, props.wide)).join(' ')
  return {
    ...base,
    gridTemplateColumns: columns,
    gridTemplateRows: `auto repeat(${rowCount}, auto)`,
  }
})

function gridPlacement(colIdx: number, rowIdx: number): Record<string, string> {
  return {
    gridColumn: String(colIdx + 1),
    gridRow: String(rowIdx + 1),
  }
}

const useWideCards = computed(
  () => props.wide && layout.value.lanes.length <= 2 && layoutDensity.value === 'comfortable'
)

const cardCompact = computed(() => !useWideCards.value || layoutDensity.value !== 'comfortable')
const cardDense = computed(() => layoutDensity.value === 'dense')

function cellsInLane(laneId: string, sequence: number): LayoutCell | undefined {
  return layout.value.cells.find((c) => c.laneId === laneId && c.sequence === sequence)
}

function laneClass(lane: LayoutLane): Record<string, boolean> {
  return {
    'lane-trunk': lane.kind === 'trunk',
    'lane-original': lane.kind === 'original',
    'lane-branch': lane.kind === 'branch',
    'lane-active': lane.isActive,
  }
}

function laneHeaderStyle(lane: LayoutLane): Record<string, string> {
  const theme = laneTheme(lane.id)
  if (!theme) return {}
  return {
    ...headerThemeStyle(theme),
    borderWidth: '1px',
    borderStyle: 'solid',
  }
}

function laneCellStyle(lane: LayoutLane): Record<string, string> | undefined {
  const theme = laneTheme(lane.id)
  if (!theme) return undefined
  if (lane.kind === 'original') {
    return { borderLeft: `2px dashed ${theme.accent}` }
  }
  if (lane.kind === 'branch') {
    return { borderLeft: `3px solid ${theme.accent}` }
  }
  return undefined
}

function forkAtSequence(seq: number): ForkPoint | undefined {
  return layout.value.forkPoints.find((f) => f.sequence === seq && f.createsBranch)
}

function branchLanesAtFork(seq: number): LayoutLane[] {
  const fps = layout.value.forkPoints.filter((f) => f.sequence === seq)
  const ids = new Set(fps.map((f) => `branch_${f.childVersionId}`))
  return layout.value.lanes.filter((l) => ids.has(l.id))
}

function onCellSelect(cell: LayoutCell) {
  emit('select', { versionId: cell.versionId, sequence: cell.sequence })
}

function isSelected(cell: LayoutCell): boolean {
  return (
    props.selectedVersionId === cell.versionId && props.selectedSequence === cell.sequence
  )
}

function isReadOnly(cell: LayoutCell): boolean {
  return cell.versionId !== props.activeVersionId
}

async function scrollToFork(seq: number | null | undefined) {
  if (seq == null || seq < 0) return
  await nextTick()
  const el = graphRef.value?.querySelector(`[data-fork-seq="${seq}"]`)
  el?.scrollIntoView({ behavior: 'smooth', block: 'center' })
}

watch(
  () => props.scrollToForkSequence,
  (seq) => {
    void scrollToFork(seq)
  },
  { immediate: true }
)

defineExpose({ scrollToFork })
</script>

<template>
  <div
    ref="graphRef"
    v-loading="loading"
    class="fork-timeline-graph"
    :class="{
      'fork-timeline-graph--wide': wide,
      [`fork-timeline-graph--${layoutDensity}`]: true,
    }"
  >
    <div v-if="!layout.lanes.length && !loading" class="graph-empty">
      <el-empty description="暂无分支数据" :image-size="72" />
    </div>

    <div v-else class="graph-shell">
      <div v-if="layout.lanes.length > 1" class="graph-toolbar">
        <el-button size="small" text type="primary" @click="collapseInactiveLanes">
          折叠非当前分支
        </el-button>
        <el-button size="small" text @click="expandAllLanes">全部展开</el-button>
        <span v-if="collapsedLaneIds.size" class="toolbar-hint">
          已折叠 {{ collapsedLaneIds.size }} 列
        </span>
      </div>

      <div ref="graphRef" class="graph-scroll" :style="laneGridStyle">
      <header
        v-for="(lane, colIdx) in visibleLanes"
        :key="`header-${lane.id}`"
        class="lane-header"
        :class="[laneClass(lane), { 'lane-header--collapsed': isLaneCollapsed(lane.id) }]"
        :style="{ ...laneHeaderStyle(lane), ...gridPlacement(colIdx, 0) }"
      >
        <button
          v-if="canCollapseLane(lane)"
          type="button"
          class="lane-collapse-btn"
          :title="isLaneCollapsed(lane.id) ? '展开此列' : '折叠此列'"
          @click.stop="toggleLaneCollapse(lane)"
        >
          <el-icon><ArrowRight v-if="isLaneCollapsed(lane.id)" /><ArrowDown v-else /></el-icon>
        </button>
        <span v-if="!isLaneCollapsed(lane.id)" class="lane-label">{{ lane.label }}</span>
        <span v-else class="lane-label-collapsed" :title="lane.label">{{ lane.label }}</span>
        <el-tag v-if="!isLaneCollapsed(lane.id) && lane.isActive" size="small" type="primary" effect="dark">当前</el-tag>
        <el-button
          v-else-if="!isLaneCollapsed(lane.id) && lane.kind === 'branch' && lane.versionId"
          size="small"
          text
          type="primary"
          @click.stop="emit('activate', lane.versionId!)"
        >
          切换到此分支
        </el-button>
      </header>

      <template v-for="(seq, rowIdx) in layout.rowSequences" :key="seq">
        <div
          v-for="(lane, colIdx) in visibleLanes"
          :key="`${lane.id}-${seq}`"
          class="lane-cell"
          :class="[laneClass(lane), { 'lane-cell--collapsed': isLaneCollapsed(lane.id) }]"
          :style="{ ...laneCellStyle(lane), ...gridPlacement(colIdx, rowIdx + 1) }"
          :data-fork-seq="
            colIdx === 0 && forkAtSequence(seq) ? seq : undefined
          "
        >
          <template v-if="!isLaneCollapsed(lane.id) && cellsInLane(lane.id, seq)">
            <ForkTimelineNodeCard
              :cell="cellsInLane(lane.id, seq)!"
              :selected="isSelected(cellsInLane(lane.id, seq)!)"
              :read-only-hint="isReadOnly(cellsInLane(lane.id, seq)!)"
              :lane-theme="laneTheme(lane.id)"
              :compact="cardCompact"
              :dense="cardDense"
              @select="onCellSelect(cellsInLane(lane.id, seq)!)"
            />
            <div
              v-if="cellsInLane(lane.id, seq)?.isForkPoint && forkAtSequence(seq)"
              class="fork-actions"
            >
              <span class="fork-connector-hint">→ 推演分叉</span>
              <el-button
                size="small"
                text
                type="warning"
                @click.stop="emit('showDiff', forkAtSequence(seq)!)"
              >
                查看变更
              </el-button>
            </div>
          </template>

          <template
            v-else-if="
              !isLaneCollapsed(lane.id) &&
              lane.kind === 'branch' &&
              branchLanesAtFork(seq).some((b) => b.id === lane.id)
            "
          >
            <div v-if="forkAtSequence(seq)" class="fork-connector">
              <span class="fork-line-h" />
            </div>
          </template>

          <div
            v-else-if="!isLaneCollapsed(lane.id) && lane.kind === 'original' && seq > (forkAtSequence(seq - 1)?.sequence ?? -1)"
            class="lane-gap"
          />
        </div>
      </template>
      </div>
    </div>
  </div>
</template>

<style scoped>
.fork-timeline-graph {
  min-height: 200px;
  width: 100%;
  container-type: inline-size;
}

.graph-shell {
  width: 100%;
}

.graph-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
  flex-wrap: wrap;
}

.toolbar-hint {
  font-size: 0.78rem;
  color: #909399;
}

.lane-collapse-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: transparent;
  padding: 0;
  margin: 0;
  cursor: pointer;
  color: #606266;
  flex-shrink: 0;
}

.lane-collapse-btn:hover {
  color: #409eff;
}

.lane-header--collapsed {
  padding: 6px 4px;
  writing-mode: vertical-rl;
  text-orientation: mixed;
  justify-content: flex-start;
  min-height: 72px;
}

.lane-label-collapsed {
  font-size: 0.72rem;
  line-height: 1.2;
  max-height: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.lane-cell--collapsed {
  min-height: 0;
  padding: 0;
  border: none !important;
}

.lane-header .lane-label {
  flex: 1;
  min-width: 0;
}

.fork-timeline-graph--wide {
  width: 100%;
}

.fork-timeline-graph--wide .graph-scroll {
  width: 100%;
  overflow-x: auto;
}

.fork-timeline-graph--dense .lane-cell {
  margin-bottom: 0;
}

.fork-timeline-graph--dense .lane-header {
  padding: 6px 8px;
  font-size: 0.78rem;
}

.graph-scroll {
  overflow-x: auto;
  padding-bottom: 8px;
  width: 100%;
  min-width: 0;
}

.lane-header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  border-radius: 8px;
  font-size: 0.85rem;
  font-weight: 600;
  min-width: 0;
}

.lane-cell {
  position: relative;
  min-height: 48px;
  min-width: 0;
  align-self: start;
}

.lane-cell.lane-original {
  padding-left: 4px;
}

.fork-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 6px;
  flex-wrap: wrap;
}

.fork-connector-hint {
  font-size: 0.78rem;
  color: #e6a23c;
}

.fork-connector {
  display: flex;
  align-items: center;
  min-height: 40px;
  padding-left: 8px;
}

.fork-line-h {
  display: block;
  width: 48px;
  height: 2px;
  background: #e6a23c;
  position: relative;
}

.fork-line-h::after {
  content: '';
  position: absolute;
  right: -4px;
  top: -4px;
  border: 5px solid transparent;
  border-left-color: #e6a23c;
}

.lane-gap {
  min-height: 8px;
}

.graph-empty {
  padding: 24px;
}
</style>
