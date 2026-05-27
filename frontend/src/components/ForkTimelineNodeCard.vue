<script setup lang="ts">
import { computed } from 'vue'
import { WarningFilled } from '@element-plus/icons-vue'
import type { LayoutCell } from '@/utils/forkTimelineLayout'
import type { LaneColorTheme } from '@/utils/branchLaneColors'
import { laneThemeVars } from '@/utils/branchLaneColors'
import { fieldLabel } from '@/constants/fieldLabels'

const props = defineProps<{
  cell: LayoutCell
  selected?: boolean
  compact?: boolean
  /** 极窄容器：精简展示，便于总览 */
  dense?: boolean
  readOnlyHint?: boolean
  laneTheme?: LaneColorTheme
}>()

const emit = defineEmits<{
  select: []
}>()

const hasChanges = computed(() => (props.cell.traitChanges?.length ?? 0) > 0)

const displayTraits = computed(() => {
  const list = props.cell.traitChanges ?? []
  if (props.dense) return list.slice(0, 1)
  return list.slice(0, 2)
})

const extraTraitCount = computed(() => {
  const n = props.cell.traitChanges?.length ?? 0
  return n > 2 ? n - 2 : 0
})

const cardStyle = computed(() => {
  const theme = props.laneTheme
  if (!theme) return undefined
  return laneThemeVars(theme)
})

function traitSummary(field: string, after: string): string {
  const label = fieldLabel(field, 'trait')
  const text = after.trim()
  if (props.dense) return label
  if (!text) return label
  const maxLen = props.compact ? 8 : 12
  const short = text.length > maxLen ? `${text.slice(0, maxLen)}…` : text
  return `${label}→${short}`
}
</script>

<template>
  <div
    class="fork-node-card"
    :class="{
      active: selected,
      compact,
      dense,
      'fork-point': cell.isForkPoint,
      'fork-row': cell.isForkRow && !cell.isForkPoint,
      'inactive-path': !cell.isActivePath,
      readonly: readOnlyHint,
      'has-changes': hasChanges,
    }"
    :style="cardStyle"
    role="button"
    tabindex="0"
    @click="emit('select')"
    @keydown.enter="emit('select')"
  >
    <div class="card-body">
      <div class="card-main">
        <div class="card-year-row">
          <span class="card-year">{{ cell.year }} 年</span>
          <el-icon v-if="hasChanges" class="change-icon" title="本节点有性格/思想变更">
            <WarningFilled />
          </el-icon>
        </div>
        <div class="card-title">{{ cell.title || '（无标题）' }}</div>
        <div v-if="dense && hasChanges" class="dense-change-hint">
          {{ cell.traitChanges!.length }} 项变更
        </div>
        <span v-if="readOnlyHint" class="readonly-tag">非当前分支</span>
        <span v-if="cell.isForkPoint" class="fork-tag">分叉点</span>
      </div>

      <div v-if="hasChanges && !dense" class="card-changes">
        <span
          v-for="(t, i) in displayTraits"
          :key="`${t.field}-${i}`"
          class="change-chip"
          :title="`${fieldLabel(t.field, 'trait')}：${t.before} → ${t.after}`"
        >
          {{ traitSummary(t.field, t.after) }}
        </span>
        <span v-if="extraTraitCount > 0" class="change-more">+{{ extraTraitCount }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.fork-node-card {
  cursor: pointer;
  border: 1px solid var(--lane-border, #e4e7ed);
  border-radius: 8px;
  padding: 10px 12px;
  background: var(--lane-bg, #fff);
  transition: border-color 0.2s, box-shadow 0.2s;
  min-width: 0;
}

.fork-node-card:hover {
  border-color: var(--lane-accent, #a0cfff);
}

.fork-node-card.active {
  border-color: var(--lane-accent, #409eff);
  box-shadow: 0 2px 10px color-mix(in srgb, var(--lane-accent, #409eff) 25%, transparent);
}

.fork-node-card.has-changes {
  border-left: 3px solid var(--lane-change-text, #e6a23c);
}

.fork-node-card.fork-point {
  border-color: #e6a23c;
  background: #fdf6ec;
}

.fork-node-card.inactive-path {
  opacity: 0.88;
  border-style: dashed;
}

.fork-node-card.readonly .readonly-tag {
  display: inline-block;
}

.fork-node-card.dense {
  padding: 6px 8px;
}

.fork-node-card.dense .card-body {
  flex-direction: column;
  gap: 4px;
}

.fork-node-card.dense .card-title {
  font-size: 0.82rem;
  line-height: 1.35;
}

.fork-node-card.dense .card-year {
  font-size: 0.72rem;
}

.dense-change-hint {
  font-size: 0.68rem;
  color: var(--lane-change-text, #b88230);
  margin-top: 2px;
}

.fork-node-card.compact {
  padding: 8px 10px;
}

.card-body {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  min-width: 0;
}

.card-main {
  flex: 1;
  min-width: 0;
}

.card-year-row {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-bottom: 4px;
}

.card-year {
  font-size: 0.78rem;
  color: #909399;
}

.change-icon {
  font-size: 12px;
  color: var(--lane-change-text, #e6a23c);
}

.card-title {
  font-size: 0.92rem;
  font-weight: 600;
  color: #303133;
  line-height: 1.4;
  word-break: break-word;
}

.card-changes {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
  max-width: 42%;
}

.change-chip {
  display: inline-block;
  font-size: 0.68rem;
  line-height: 1.35;
  padding: 2px 6px;
  border-radius: 4px;
  background: var(--lane-change-bg, #fdf6ec);
  color: var(--lane-change-text, #b88230);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 100%;
}

.change-more {
  font-size: 0.65rem;
  color: #909399;
}

.fork-node-card:not(.compact) .card-title {
  font-size: 0.98rem;
}

.fork-node-card:not(.compact) .card-changes {
  max-width: 50%;
}

.fork-node-card:not(.compact) .change-chip {
  font-size: 0.72rem;
  padding: 3px 8px;
}

.fork-tag {
  display: inline-block;
  margin-top: 6px;
  font-size: 0.72rem;
  color: #e6a23c;
  font-weight: 600;
}

.readonly-tag {
  display: none;
  margin-top: 6px;
  font-size: 0.72rem;
  color: #909399;
}

@media (max-width: 900px) {
  .card-body {
    flex-direction: column;
  }
  .card-changes {
    flex-direction: row;
    flex-wrap: wrap;
    align-items: flex-start;
    max-width: 100%;
  }
}
</style>
