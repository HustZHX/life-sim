<script setup lang="ts">
import { computed, ref } from 'vue'
import { Loading, Plus } from '@element-plus/icons-vue'
import type { GameNodeChoiceDisplay, LifeNode, Profile, WorldLine, WorldLineEvent } from '@/api/client'
import { fieldLabel } from '@/constants/fieldLabels'
import { entityContextFromNode } from '@/utils/textEntities'
import { isLivingProfile } from '@/utils/timelineDensity'
import HighlightedText from '@/components/HighlightedText.vue'
import NodeSceneMeta from '@/components/NodeSceneMeta.vue'

const props = defineProps<{
  nodes: LifeNode[]
  worldLine?: WorldLine | null
  selectedId?: string
  profile?: Profile | null
  /** 继续推演进行中或时间轴生成中 */
  appendDisabled?: boolean
  /** 是否允许回退至此（当前激活分支且未在任务中） */
  rollbackEnabled?: boolean
  /** 游戏模式：末节点始终显示「+」，不依赖 isLivingProfile */
  alwaysShowAppend?: boolean
  /** 末节点按钮文案 */
  appendButtonLabel?: string
  /** 游戏模式：已选抉择（显示在锚定节点与下一节点之间） */
  gameChoices?: GameNodeChoiceDisplay[]
  /** 时间轴末尾：世界线/抉择推演进行中 */
  tailSimulating?: boolean
  tailSimulatingLabel?: string
}>()

const emit = defineEmits<{
  select: [node: LifeNode]
  append: []
  rollback: [node: LifeNode]
}>()

const showAppendSlot = computed(() => {
  if (props.nodes.length === 0) return false
  if (props.alwaysShowAppend) return true
  return !!props.profile && isLivingProfile(props.profile)
})

const appendLabel = computed(() => props.appendButtonLabel || '继续推演')
const worldSimLabel = computed(() => props.tailSimulatingLabel || '世界模拟中')

const expandedIds = ref<Set<string>>(new Set())
const expandedWorldIds = ref<Set<string>>(new Set())
const expandedChoiceIds = ref<Set<string>>(new Set())

const hasGameChoices = computed(() => (props.gameChoices?.length ?? 0) > 0)

const showFullContent = (nodeId: string) => expandedIds.value.has(nodeId)
const showWorldFull = (id: string) => expandedWorldIds.value.has(id)

function choiceExpandKey(nodeId: string) {
  return `choice-${nodeId}`
}

const showChoiceFull = (nodeId: string) => expandedChoiceIds.value.has(choiceExpandKey(nodeId))

function contextFor(node: LifeNode) {
  return entityContextFromNode(node, props.profile ?? null)
}

function toggleExpand(nodeId: string) {
  const next = new Set(expandedIds.value)
  if (next.has(nodeId)) next.delete(nodeId)
  else next.add(nodeId)
  expandedIds.value = next
}

function toggleWorld(id: string) {
  const next = new Set(expandedWorldIds.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  expandedWorldIds.value = next
}

function toggleChoiceExpand(nodeId: string) {
  const key = choiceExpandKey(nodeId)
  const next = new Set(expandedChoiceIds.value)
  if (next.has(key)) next.delete(key)
  else next.add(key)
  expandedChoiceIds.value = next
}

function expandAll() {
  expandedIds.value = new Set(props.nodes.map((n) => n.id))
  if (hasGameChoices.value) {
    expandedChoiceIds.value = new Set(
      props.nodes.filter((n) => choiceForNode(n)).map((n) => choiceExpandKey(n.id))
    )
  }
}

function collapseAll() {
  expandedIds.value = new Set()
  expandedChoiceIds.value = new Set()
}

type AxisItem =
  | { kind: 'world'; id: string; year: number; originalIndex: number; ev: WorldLineEvent }
  | { kind: 'node'; id: string; year: number; node: LifeNode }

const axisItems = computed<AxisItem[]>(() => {
  const out: AxisItem[] = []
  const events = props.worldLine?.events ?? []
  for (let i = 0; i < events.length; i++) {
    const ev = events[i]
    if (!ev || !ev.year) continue
    out.push({ kind: 'world', id: `w-${ev.year}-${i}`, year: ev.year, originalIndex: i, ev })
  }
  for (const node of props.nodes) {
    out.push({ kind: 'node', id: node.id, year: node.year, node })
  }
  out.sort((a, b) => {
    if (a.year !== b.year) return a.year - b.year
    if (a.kind !== b.kind) return a.kind === 'world' ? -1 : 1 // 同年：世界线在前
    if (a.kind === 'node' && b.kind === 'node') return a.node.sequence - b.node.sequence
    if (a.kind === 'world' && b.kind === 'world') return a.originalIndex - b.originalIndex
    return 0
  })
  return out
})

const maxSequence = computed(() => {
  if (!props.nodes.length) return 0
  return props.nodes.reduce((max, n) => (n.sequence > max ? n.sequence : max), 0)
})

function canRollbackTo(node: LifeNode) {
  return !!props.rollbackEnabled && node.sequence < maxSequence.value
}

function choiceForNode(node: LifeNode) {
  const list = props.gameChoices ?? []
  const byId = list.find((c) => c.node_id === node.id)
  if (byId?.display_text?.trim()) return byId
  const bySeq = list.find((c) => c.node_sequence === node.sequence)
  if (bySeq?.display_text?.trim()) return bySeq
  return null
}

function choiceSnippet(text: string, max = 40) {
  const t = text.trim()
  if (t.length <= max) return t
  return `${t.slice(0, max)}…`
}
</script>

<template>
  <div class="timeline-axis">
    <div v-if="nodes.length" class="axis-toolbar">
      <el-button size="small" text type="primary" @click="expandAll">全部展开</el-button>
      <el-button size="small" text @click="collapseAll">全部收起</el-button>
    </div>

    <div class="axis-list">
      <template v-for="item in axisItems" :key="item.id">
      <div class="axis-row">
        <div class="axis-col axis-col--world">
          <el-card
            v-if="item.kind === 'world'"
            class="world-card"
            :class="{ expanded: showWorldFull(item.id) }"
            shadow="hover"
            @click="toggleWorld(item.id)"
          >
            <div class="world-head">
              <h4 class="world-title">{{ item.ev.name }}</h4>
              <el-button
                size="small"
                text
                type="primary"
                class="world-expand-btn"
                @click.stop="toggleWorld(item.id)"
              >
                {{ showWorldFull(item.id) ? '收起' : '展开' }}
              </el-button>
            </div>
            <template v-if="showWorldFull(item.id)">
              <p v-if="item.ev.description" class="world-text">{{ item.ev.description }}</p>
              <p v-if="item.ev.impact" class="world-text muted">{{ item.ev.impact }}</p>
              <p v-if="item.ev.divergence_note" class="world-text warn">{{ item.ev.divergence_note }}</p>
              <span v-if="item.ev.caused_by_node_sequence != null" class="world-ref">
                来源节点 #{{ item.ev.caused_by_node_sequence + 1 }}
              </span>
            </template>
          </el-card>
        </div>

        <div class="axis-mid" aria-hidden="true">
          <div class="axis-dot" />
          <div class="axis-timestamp">
            {{ item.kind === 'node' ? `${item.node.year} 年 · ${item.node.age} 岁` : `${item.ev.year} 年` }}
          </div>
        </div>

        <div class="axis-col axis-col--node">
          <el-card
            v-if="item.kind === 'node'"
            :id="`node-${item.node.id}`"
            class="node-card"
            :class="{
              active: selectedId === item.node.id,
              expanded: showFullContent(item.node.id),
              warning: !!item.node.trait_changes?.length,
            }"
            shadow="hover"
            @click="emit('select', item.node)"
          >
            <div class="card-head">
              <h4>
                <HighlightedText :text="item.node.title" :context="contextFor(item.node)" tag="span" />
              </h4>
              <div class="card-actions">
                <el-button
                  v-if="canRollbackTo(item.node)"
                  size="small"
                  text
                  type="danger"
                  class="rollback-btn"
                  @click.stop="emit('rollback', item.node)"
                >
                  回退至此
                </el-button>
                <el-button
                  size="small"
                  text
                  type="primary"
                  class="expand-btn"
                  @click.stop="toggleExpand(item.node.id)"
                >
                  {{ showFullContent(item.node.id) ? '收起' : '展开' }}
                </el-button>
              </div>
            </div>

            <template v-if="showFullContent(item.node.id)">
              <NodeSceneMeta :scene="item.node.scene" />
              <div class="read-block">
                <div class="read-label">经历</div>
                <p class="read-text">
                  <HighlightedText :text="item.node.events" :context="contextFor(item.node)" tag="span" />
                </p>
              </div>
              <div v-if="item.node.thoughts" class="read-block muted">
                <div class="read-label">内心</div>
                <p class="read-text">
                  <HighlightedText :text="item.node.thoughts" :context="contextFor(item.node)" tag="span" />
                </p>
              </div>
              <div v-if="item.node.personality_snapshot" class="read-block muted">
                <div class="read-label">性格</div>
                <p class="read-text">
                  <HighlightedText
                    :text="item.node.personality_snapshot"
                    :context="contextFor(item.node)"
                    tag="span"
                  />
                </p>
              </div>
              <div v-if="item.node.trait_changes?.length" class="trait-section">
                <div class="read-label">性格/思想变更</div>
                <ul class="trait-list">
                  <li v-for="(t, i) in item.node.trait_changes" :key="i" class="trait-item">
                    <strong>{{ fieldLabel(t.field, 'trait') }}</strong>：
                    <HighlightedText :text="`${t.before} → ${t.after}`" :context="contextFor(item.node)" tag="span" />
                    <p v-if="t.reason" class="trait-reason">
                      <HighlightedText :text="t.reason" :context="contextFor(item.node)" tag="span" />
                    </p>
                  </li>
                </ul>
              </div>
            </template>
            <ul v-else-if="item.node.trait_changes?.length" class="trait-compact">
              <li v-for="(t, i) in item.node.trait_changes" :key="i">
                <span class="trait-field">{{ fieldLabel(t.field, 'trait') }}</span>：
                <HighlightedText :text="`${t.before} → ${t.after}`" :context="contextFor(item.node)" tag="span" />
              </li>
            </ul>
          </el-card>
        </div>
      </div>

      <div
        v-if="item.kind === 'node' && choiceForNode(item.node)"
        class="axis-row axis-row--choice"
      >
        <div class="axis-col axis-col--world" />
        <div class="axis-mid axis-mid--choice" aria-hidden="true">
          <div class="axis-dot axis-dot--choice" />
        </div>
        <div class="axis-col axis-col--node">
          <div
            class="game-choice-bridge"
            :class="{ expanded: showChoiceFull(item.node.id) }"
            role="note"
            @click="toggleChoiceExpand(item.node.id)"
          >
            <div class="game-choice-bridge__head">
              <span class="game-choice-bridge__label">人生抉择</span>
              <el-button
                size="small"
                text
                type="primary"
                class="game-choice-expand-btn"
                @click.stop="toggleChoiceExpand(item.node.id)"
              >
                {{ showChoiceFull(item.node.id) ? '收起' : '展开' }}
              </el-button>
            </div>
            <p v-if="showChoiceFull(item.node.id)" class="game-choice-bridge__text">
              {{ choiceForNode(item.node)?.display_text }}
            </p>
            <p v-else class="game-choice-bridge__snippet">
              {{ choiceSnippet(choiceForNode(item.node)?.display_text ?? '') }}
            </p>
          </div>
        </div>
      </div>
      </template>

      <div
        v-if="showAppendSlot"
        class="axis-row"
        :class="tailSimulating ? 'axis-row--world-sim' : 'axis-row--append'"
      >
        <div class="axis-col axis-col--world">
          <div v-if="tailSimulating" class="world-sim-loading" role="status" aria-live="polite">
            <el-icon class="world-sim-loading__icon is-loading" :size="20">
              <Loading />
            </el-icon>
            <div class="world-sim-loading__body">
              <span class="world-sim-loading__title">{{ worldSimLabel }}</span>
              <span class="world-sim-loading__sub">正在根据你的抉择推演世界线与下一人生阶段…</span>
            </div>
          </div>
        </div>
        <div class="axis-mid" aria-hidden="true">
          <div
            class="axis-dot"
            :class="tailSimulating ? 'axis-dot--simulating' : 'axis-dot--append'"
          />
        </div>
        <div class="axis-col axis-col--node">
          <div v-if="tailSimulating" class="node-sim-placeholder">
            <el-icon class="is-loading" :size="18"><Loading /></el-icon>
            <span>人生推演中</span>
          </div>
          <button
            v-else
            type="button"
            class="append-next-btn"
            :disabled="appendDisabled"
            :title="appendDisabled ? '请等待当前任务完成' : appendLabel"
            @click="emit('append')"
          >
            <el-icon :size="22"><Plus /></el-icon>
            <span>{{ appendLabel }}</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.timeline-axis {
  min-width: 0;
}

.axis-toolbar {
  display: flex;
  justify-content: flex-end;
  gap: 4px;
  margin-bottom: 8px;
}

.node-card {
  cursor: pointer;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.node-card.active {
  border-color: #409eff;
  box-shadow: 0 2px 12px rgba(64, 158, 255, 0.15);
}

.node-card.expanded {
  border-color: #79bbff;
}

.card-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 4px;
}

.node-card h4 {
  margin: 0;
  line-height: 1.45;
  flex: 1;
  min-width: 0;
}

.expand-btn {
  flex-shrink: 0;
  padding: 0 4px;
}

.card-actions {
  display: flex;
  align-items: center;
  gap: 2px;
  flex-shrink: 0;
}

.rollback-btn {
  padding: 0 4px;
}

.axis-list {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.axis-row {
  display: grid;
  grid-template-columns: minmax(180px, 0.9fr) 56px minmax(360px, 1.4fr);
  column-gap: 12px;
  align-items: stretch;
  padding-bottom: 16px;
}

.axis-row:last-child {
  padding-bottom: 0;
}

.axis-col {
  min-width: 0;
}

.axis-col--world {
  grid-column: 1;
}

.axis-mid {
  grid-column: 2;
}

.axis-col--node {
  grid-column: 3;
}

.axis-mid {
  position: relative;
  min-height: 24px;
  display: flex;
  flex-direction: column;
  align-items: center;
  align-self: stretch;
}

.axis-mid::before {
  content: '';
  position: absolute;
  top: 0;
  bottom: -16px;
  left: 50%;
  transform: translateX(-50%);
  width: 2px;
  background: #e5e7eb;
  z-index: 0;
}

.axis-row:first-child .axis-mid::before {
  top: 10px;
}

.axis-row:last-child .axis-mid::before,
.axis-row--append .axis-mid::before {
  bottom: 0;
}

.axis-dot {
  position: relative;
  z-index: 1;
  width: 12px;
  height: 12px;
  border-radius: 999px;
  background: #ffffff;
  border: 2px solid #67c23a;
  margin-top: 4px;
  flex-shrink: 0;
}

.axis-dot--append {
  border-color: #909399;
}

.axis-row--choice {
  padding-bottom: 4px;
}

.axis-mid--choice::before {
  top: 0;
  bottom: 0;
}

.axis-dot--choice {
  width: 10px;
  height: 10px;
  border-color: #409eff;
  background: #ecf5ff;
}

.game-choice-bridge {
  margin: 0 0 8px;
  padding: 10px 14px;
  border-radius: 8px;
  border: 1px dashed #b3d8ff;
  background: linear-gradient(135deg, #f5faff 0%, #ecf5ff 100%);
  cursor: pointer;
  transition: border-color 0.15s, box-shadow 0.15s;
}

.game-choice-bridge:hover {
  border-color: #79bbff;
}

.game-choice-bridge.expanded {
  border-color: #409eff;
  box-shadow: 0 0 0 1px rgba(64, 158, 255, 0.2);
}

.game-choice-bridge__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 4px;
}

.game-choice-bridge__label {
  font-size: 0.72rem;
  font-weight: 600;
  letter-spacing: 0.04em;
  color: #409eff;
  text-transform: uppercase;
}

.game-choice-expand-btn {
  flex-shrink: 0;
  padding: 0 4px;
}

.game-choice-bridge__text {
  margin: 0;
  font-size: 0.88rem;
  line-height: 1.55;
  color: #303133;
  white-space: pre-wrap;
  word-break: break-word;
}

.game-choice-bridge__snippet {
  margin: 0;
  font-size: 0.85rem;
  line-height: 1.45;
  color: #606266;
}

.axis-timestamp {
  position: relative;
  z-index: 1;
  margin-top: 6px;
  padding: 2px 6px;
  font-size: 0.78rem;
  line-height: 1.3;
  color: #909399;
  background: #ffffff;
  border-radius: 999px;
  white-space: nowrap;
}

.node-card.warning {
  border-color: #e6a23c;
}

.world-card {
  cursor: pointer;
  background: #fafafa;
  border: 1px solid #ebeef5;
  border-left: 2px solid rgba(103, 194, 58, 0.65);
}

.world-card.expanded {
  border-left-color: rgba(230, 162, 60, 0.85);
}

.world-head {
  display: flex;
  gap: 8px;
  align-items: flex-start;
  justify-content: space-between;
}

.world-title {
  margin: 0;
  font-size: 0.86rem;
  line-height: 1.28;
  font-weight: 650;
  color: #374151;
}

.world-expand-btn {
  padding: 0 4px;
  flex-shrink: 0;
}

.world-text {
  margin: 6px 0 0;
  font-size: 0.79rem;
  line-height: 1.5;
  color: #6b7280;
  white-space: pre-wrap;
}

.world-text.muted {
  color: #9ca3af;
}

.world-text.warn {
  color: #e6a23c;
}

.world-ref {
  display: inline-block;
  margin-top: 6px;
  padding: 1px 6px;
  font-size: 0.7rem;
  color: #9ca3af;
  background: rgba(244, 244, 245, 0.7);
  border-radius: 4px;
}

.axis-col--node :deep(.el-card__body),
.axis-col--world :deep(.el-card__body) {
  min-width: 0;
}

.axis-col--world :deep(.el-card__body) {
  padding: 10px 12px;
}

@media (max-width: 960px) {
  .axis-row {
    grid-template-columns: 24px 1fr;
    column-gap: 10px;
    padding-bottom: 12px;
  }

  .axis-mid::before {
    bottom: -12px;
  }

  .axis-col--world,
  .axis-col--node {
    grid-column: 2;
  }

  .axis-col--world {
    margin-bottom: 10px;
    margin-left: 6px;
    max-width: 92%;
  }

  .axis-mid {
    grid-column: 1;
    grid-row: 1 / span 2;
    align-items: flex-start;
    padding-top: 2px;
  }

  .axis-mid::before {
    top: 0;
    bottom: -12px;
    left: 12px;
    transform: none;
  }

  .axis-row:first-child .axis-mid::before {
    top: 8px;
  }

  .axis-row:last-child .axis-mid::before,
  .axis-row--append .axis-mid::before {
    bottom: 0;
  }

  .axis-dot {
    margin-top: 0;
    transform: translateX(6px);
  }

  .axis-timestamp {
    margin-top: 8px;
    transform: translateX(6px);
    white-space: normal;
    max-width: 100%;
  }
}

.trait-compact {
  margin: 8px 0 0;
  padding: 8px 10px;
  list-style: none;
  background: #fffbf0;
  border-radius: 6px;
  border-left: 3px solid #e6a23c;
  font-size: 0.85rem;
  line-height: 1.5;
  color: #606266;
}

.trait-compact li + li {
  margin-top: 6px;
}

.trait-field {
  font-weight: 600;
  color: #303133;
}

.trait-section {
  margin-top: 10px;
}

.trait-list {
  margin: 6px 0 0;
  padding: 0;
  list-style: none;
}

.trait-item {
  margin: 0 0 8px;
  padding: 8px 10px;
  background: #fffbf0;
  border-radius: 6px;
  font-size: 0.88rem;
  line-height: 1.5;
  color: #606266;
}

.trait-item strong {
  color: #303133;
}

.trait-reason {
  margin: 4px 0 0;
  font-size: 0.82rem;
  color: #909399;
}

.read-block {
  margin: 10px 0;
  padding: 10px 12px;
  background: #f8fafc;
  border-radius: 6px;
  border-left: 3px solid #409eff;
}

.read-block.muted {
  border-left-color: #c0c4cc;
  background: #fafafa;
}

.read-label {
  font-size: 0.75rem;
  font-weight: 600;
  color: #909399;
  margin-bottom: 6px;
  letter-spacing: 0.04em;
}

.read-text {
  margin: 0;
  color: #303133;
  font-size: 0.92rem;
  line-height: 1.65;
  white-space: pre-wrap;
  word-break: break-word;
}

.append-next-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: 100%;
  min-height: 56px;
  padding: 12px 16px;
  border: 2px dashed #c0c4cc;
  border-radius: 8px;
  background: #fafbfc;
  color: #409eff;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: border-color 0.2s, background 0.2s, color 0.2s;
}

.append-next-btn:hover:not(:disabled) {
  border-color: #409eff;
  background: #ecf5ff;
  color: #337ecc;
}

.append-next-btn:disabled {
  cursor: not-allowed;
  opacity: 0.55;
  color: #909399;
}

.axis-row--world-sim .axis-mid::before {
  bottom: 0;
}

.axis-dot--simulating {
  border-color: #67c23a;
  background: #f0f9eb;
  animation: axis-dot-pulse 1.2s ease-in-out infinite;
}

@keyframes axis-dot-pulse {
  0%,
  100% {
    box-shadow: 0 0 0 0 rgba(103, 194, 58, 0.35);
  }
  50% {
    box-shadow: 0 0 0 6px rgba(103, 194, 58, 0);
  }
}

.world-sim-loading {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  min-height: 56px;
  padding: 12px 14px;
  border-radius: 8px;
  border: 1px solid #e1f3d8;
  border-left: 3px solid #67c23a;
  background: linear-gradient(135deg, #f6ffed 0%, #fafafa 100%);
}

.world-sim-loading__icon {
  flex-shrink: 0;
  margin-top: 2px;
  color: #67c23a;
}

.world-sim-loading__body {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.world-sim-loading__title {
  font-size: 0.88rem;
  font-weight: 600;
  color: #529b2e;
}

.world-sim-loading__sub {
  font-size: 0.78rem;
  line-height: 1.45;
  color: #909399;
}

.node-sim-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: 100%;
  min-height: 56px;
  padding: 12px 16px;
  border: 2px dashed #c0c4cc;
  border-radius: 8px;
  background: #fafbfc;
  color: #909399;
  font-size: 0.88rem;
}
</style>
