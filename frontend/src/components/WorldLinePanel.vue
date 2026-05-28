<script setup lang="ts">
import { computed, ref, toRaw, watch } from 'vue'
import { ElMessage } from 'element-plus'
import type { AIModelId, LifeNode, WorldLine, WorldLineEvent } from '@/api/client'
import { api } from '@/api/client'
import ModelSelector from '@/components/ModelSelector.vue'
import { DEFAULT_CASCADE_MODEL } from '@/constants/models'

const props = defineProps<{
  characterId: string
  timelineId: string
  worldLine: WorldLine | null | undefined
  nodes: LifeNode[]
  loading?: boolean
  refreshing?: boolean
}>()

const emit = defineEmits<{
  updated: [worldLine: WorldLine]
  applyJob: [jobId: string]
  refresh: []
}>()

const editing = ref(false)
const saving = ref(false)
const model = ref<AIModelId>(DEFAULT_CASCADE_MODEL)
const draft = ref<WorldLine | null>(null)

function cloneWorldLine<T>(v: T): T {
  // 防御性拷贝：
  // - structuredClone 在部分浏览器/环境会对 Vue Proxy 抛 DataCloneError
  // - iOS Safari 13 可能没有 structuredClone
  // 因 worldLine 是纯 JSON 数据，这里统一兜底为 JSON clone。
  const raw = toRaw(v) as T
  const sc = (globalThis as unknown as { structuredClone?: (x: T) => T }).structuredClone
  if (typeof sc === 'function') {
    try {
      return sc(raw)
    } catch {
      // fall through
    }
  }
  return JSON.parse(JSON.stringify(raw)) as T
}

watch(
  () => props.worldLine,
  (wl) => {
    if (!editing.value) {
      draft.value = wl ? cloneWorldLine(wl) : null
    }
  },
  { immediate: true, deep: true }
)

const hasEvents = computed(() => (draft.value?.events?.length ?? 0) > 0)

const sortedEvents = computed(() => {
  const events = draft.value?.events ?? []
  return events
    .map((ev, originalIndex) => ({ ev, originalIndex }))
    .sort((a, b) => a.ev.year - b.ev.year)
})

const yearRange = computed(() => {
  if (!draft.value?.start_year && !draft.value?.end_year && !hasEvents.value) return ''
  const start = draft.value?.start_year ?? sortedEvents.value[0]?.ev.year
  const end = draft.value?.end_year ?? sortedEvents.value[sortedEvents.value.length - 1]?.ev.year
  if (start && end) return `${start} — ${end}`
  return ''
})

function startEdit() {
  if (!props.worldLine) return
  draft.value = cloneWorldLine(props.worldLine)
  editing.value = true
}

function cancelEdit() {
  draft.value = props.worldLine ? cloneWorldLine(props.worldLine) : null
  editing.value = false
}

function updateEvent(index: number, patch: Partial<WorldLineEvent>) {
  if (!draft.value) return
  draft.value.events[index] = { ...draft.value.events[index], ...patch }
}

async function save(applyToNodes: boolean) {
  if (!draft.value || !props.timelineId) return
  saving.value = true
  try {
    const res = await api.updateWorldLine(props.characterId, props.timelineId, {
      world_line: draft.value,
      model: model.value,
      apply_to_nodes: applyToNodes,
    })
    emit('updated', res.world_line)
    draft.value = cloneWorldLine(res.world_line)
    editing.value = false
    if (res.job?.id) {
      emit('applyJob', res.job.id)
      ElMessage.success('世界线已保存，正在同步人生节点…')
    } else {
      ElMessage.success('世界线已保存')
    }
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '保存失败')
  } finally {
    saving.value = false
  }
}

function nodeLabel(seq?: number | null) {
  if (seq == null || seq < 0) return ''
  return `节点 #${seq + 1}`
}
</script>

<template>
  <aside class="world-line-panel" v-loading="loading || saving">
    <header class="panel-head">
      <div>
        <h3>世界线</h3>
        <p v-if="yearRange" class="year-range">{{ yearRange }}</p>
      </div>
      <div v-if="!editing && nodes.length" class="head-actions">
        <el-button
          size="small"
          text
          type="primary"
          :loading="refreshing"
          :disabled="loading || saving"
          @click="emit('refresh')"
        >
          刷新世界线
        </el-button>
        <el-button v-if="draft && hasEvents" size="small" text type="primary" @click="startEdit">
          编辑
        </el-button>
      </div>
    </header>

    <div v-if="draft && hasEvents" class="panel-body">
      <div v-if="!editing" class="trend-block">
        <p v-if="draft.historical_trend" class="historical-trend">{{ draft.historical_trend }}</p>
        <p v-if="draft.era_summary" class="era-summary">{{ draft.era_summary }}</p>
        <p v-if="draft.daily_life_context" class="daily-context">{{ draft.daily_life_context }}</p>
      </div>

      <div v-else class="edit-meta">
        <el-form label-position="top" size="small">
          <el-form-item label="历史走向">
            <el-input v-model="draft.historical_trend" type="textarea" :rows="3" />
          </el-form-item>
          <el-form-item label="时代总述">
            <el-input v-model="draft.era_summary" type="textarea" :rows="2" />
          </el-form-item>
        </el-form>
        <ModelSelector v-model="model" />
      </div>

      <el-timeline class="world-events">
        <el-timeline-item
          v-for="item in sortedEvents"
          :key="`${item.ev.year}-${item.ev.name}-${item.originalIndex}`"
          :timestamp="`${item.ev.year}`"
          placement="top"
          :type="item.ev.divergence_note ? 'warning' : 'success'"
          hollow
        >
          <div class="event-card" :class="{ diverged: !!item.ev.divergence_note }">
            <template v-if="!editing">
              <h4>{{ item.ev.name }}</h4>
              <p v-if="item.ev.description" class="desc">{{ item.ev.description }}</p>
              <p v-if="item.ev.impact" class="impact">{{ item.ev.impact }}</p>
              <p v-if="item.ev.divergence_note" class="divergence">{{ item.ev.divergence_note }}</p>
              <span v-if="item.ev.caused_by_node_sequence != null" class="node-ref">
                {{ nodeLabel(item.ev.caused_by_node_sequence) }}
              </span>
            </template>
            <template v-else>
              <el-input
                :model-value="item.ev.name"
                placeholder="事件名称"
                size="small"
                class="ev-field"
                @update:model-value="(v: string) => updateEvent(item.originalIndex, { name: v })"
              />
              <el-input
                :model-value="item.ev.year"
                type="number"
                placeholder="年"
                size="small"
                class="ev-field ev-year"
                @update:model-value="(v: string | number) => updateEvent(item.originalIndex, { year: Number(v) || 0 })"
              />
              <el-input
                :model-value="item.ev.description"
                type="textarea"
                :rows="2"
                placeholder="描述"
                size="small"
                class="ev-field"
                @update:model-value="(v: string) => updateEvent(item.originalIndex, { description: v })"
              />
              <el-input
                :model-value="item.ev.divergence_note"
                type="textarea"
                :rows="2"
                placeholder="偏差说明"
                size="small"
                class="ev-field"
                @update:model-value="(v: string) => updateEvent(item.originalIndex, { divergence_note: v })"
              />
            </template>
          </div>
        </el-timeline-item>
      </el-timeline>
    </div>

    <div v-else class="panel-empty">
      <p class="empty-title">暂无世界线数据</p>
      <p class="empty-hint">新建时间轴或推演余生后会自动生成天下大事与历史走向。</p>
    </div>

    <footer v-if="editing" class="panel-foot">
      <el-button size="small" @click="cancelEdit">取消</el-button>
      <el-button size="small" type="primary" :loading="saving" @click="save(false)">保存</el-button>
      <el-button size="small" type="warning" :loading="saving" @click="save(true)">同步节点</el-button>
    </footer>
  </aside>
</template>

<style scoped>
.world-line-panel {
  display: flex;
  flex-direction: column;
  background: linear-gradient(180deg, #f8fafc 0%, #fff 48px);
  border-right: 1px solid #e4e7ed;
}
.panel-head {
  flex-shrink: 0;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 8px;
  padding: 14px 14px 10px;
  border-bottom: 1px solid #ebeef5;
  background: #f8fafc;
}
.panel-head h3 {
  margin: 0;
  font-size: 0.95rem;
  font-weight: 600;
  color: #303133;
}
.year-range {
  margin: 4px 0 0;
  font-size: 0.75rem;
  color: #909399;
  font-variant-numeric: tabular-nums;
}
.head-actions {
  flex-shrink: 0;
}
.panel-body {
  padding: 12px 10px 16px 14px;
}
.trend-block {
  margin-bottom: 14px;
  padding: 10px 12px;
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  font-size: 0.82rem;
  line-height: 1.55;
  color: #606266;
}
.historical-trend {
  margin: 0 0 8px;
  font-weight: 500;
  color: #303133;
}
.era-summary,
.daily-context {
  margin: 6px 0 0;
  color: #909399;
  font-size: 0.78rem;
}
.world-events {
  margin-top: 4px;
  padding-left: 2px;
}
.world-events :deep(.el-timeline-item__timestamp) {
  font-size: 0.72rem;
  font-weight: 600;
  color: #67c23a;
}
.world-events :deep(.el-timeline-item__wrapper) {
  padding-left: 20px;
}
.event-card {
  padding: 8px 10px;
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  border-left: 3px solid #67c23a;
}
.event-card.diverged {
  border-left-color: #e6a23c;
  background: #fffbf5;
}
.event-card h4 {
  margin: 0 0 4px;
  font-size: 0.85rem;
  line-height: 1.35;
}
.desc,
.impact {
  margin: 0;
  font-size: 0.78rem;
  line-height: 1.45;
  color: #606266;
}
.impact {
  margin-top: 4px;
  color: #909399;
}
.divergence {
  margin: 6px 0 0;
  font-size: 0.75rem;
  color: #e6a23c;
  line-height: 1.4;
}
.node-ref {
  display: inline-block;
  margin-top: 6px;
  padding: 1px 6px;
  font-size: 0.7rem;
  color: #909399;
  background: #f4f4f5;
  border-radius: 4px;
}
.ev-field {
  margin-bottom: 6px;
}
.ev-year {
  max-width: 88px;
}
.edit-meta {
  margin-bottom: 12px;
}
.panel-empty {
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 24px 16px;
  text-align: center;
}
.empty-title {
  margin: 0 0 8px;
  font-size: 0.88rem;
  color: #606266;
}
.empty-hint {
  margin: 0;
  font-size: 0.78rem;
  line-height: 1.5;
  color: #909399;
}
.panel-foot {
  flex-shrink: 0;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  padding: 10px 12px;
  border-top: 1px solid #ebeef5;
  background: #fff;
}
</style>
