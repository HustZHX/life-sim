<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { EditPen } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import {
  api,
  type AIModelId,
  type LifeNode,
  type Profile,
  type TimelineConfig,
} from '@/api/client'
import type { EntityHighlightContext } from '@/utils/textEntities'
import { entityContextFromNode } from '@/utils/textEntities'
import ModelSelector from '@/components/ModelSelector.vue'
import HighlightedText from '@/components/HighlightedText.vue'
import TimelineConfigPanel from '@/components/TimelineConfigPanel.vue'
import NodeSceneMeta from '@/components/NodeSceneMeta.vue'
import { DEFAULT_TARGET_NODE_COUNT } from '@/utils/timelineDensity'
import { DEFAULT_AI_MODEL, DEFAULT_CASCADE_MODEL } from '@/constants/models'
import type { NarrativeKind } from '@/api/client'
import { fieldLabel } from '@/constants/fieldLabels'
import JobProgress from '@/components/JobProgress.vue'

const props = defineProps<{
  characterId: string
  node: LifeNode | null
  profile?: Profile | null
  loading?: boolean
  regenProgress?: number
  regenStatus?: string
  regenVisible?: boolean
}>()

const emit = defineEmits<{
  save: [
    payload: {
      mode: string
      model: AIModelId
      patch: Record<string, string>
      target_node_count?: number
    },
  ]
  narrative: [kind: Exclude<NarrativeKind, 'light_novel'>, model: AIModelId]
}>()

type UiMode = 'read' | 'edit'

const uiMode = ref<UiMode>('read')
const form = ref({
  title: '',
  events: '',
  thoughts: '',
  personality_snapshot: '',
})
const model = ref<AIModelId>(DEFAULT_AI_MODEL)
const cascadeModel = ref<AIModelId>(DEFAULT_CASCADE_MODEL)
const regenConfig = ref<TimelineConfig>({
  target_node_count: DEFAULT_TARGET_NODE_COUNT,
  start_year: 0,
  end_year: 0,
})

const eventsRegenerating = ref(false)

const entityContext = computed<EntityHighlightContext>(() =>
  entityContextFromNode(props.node, props.profile ?? null)
)

function syncFormFromNode(n: LifeNode) {
  form.value = {
    title: n.title,
    events: n.events,
    thoughts: n.thoughts,
    personality_snapshot: n.personality_snapshot,
  }
}

watch(
  () => props.node,
  (n) => {
    if (n) {
      syncFormFromNode(n)
      uiMode.value = 'read'
    }
  },
  { immediate: true }
)

watch(
  () => props.profile,
  (p) => {
    if (p) {
      regenConfig.value.start_year = p.birth_year
      regenConfig.value.end_year = p.death_year
    }
  },
  { immediate: true }
)

function buildPatch() {
  return { ...form.value }
}

function enterEdit() {
  if (props.node) syncFormFromNode(props.node)
  uiMode.value = 'edit'
}

function cancelEdit() {
  if (props.node) syncFormFromNode(props.node)
  uiMode.value = 'read'
}

function openNarrative(kind: Exclude<NarrativeKind, 'light_novel'>) {
  emit('narrative', kind, model.value)
}

async function regenerateEventsFromTitle() {
  if (!props.node || !props.characterId) return
  const title = (uiMode.value === 'edit' ? form.value.title : props.node.title).trim()
  if (!title) {
    ElMessage.warning('请先填写或确认节点标题')
    return
  }
  eventsRegenerating.value = true
  try {
    const res = await api.regenerateNodeEvents(props.characterId, props.node.id, {
      title,
      model: model.value,
    })
    form.value.events = res.events
    if (res.thoughts) form.value.thoughts = res.thoughts
    if (res.personality_snapshot) form.value.personality_snapshot = res.personality_snapshot
    uiMode.value = 'edit'
    ElMessage.success('已根据标题重新生成本节点经历、内心与性格')
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '生成经历失败')
  } finally {
    eventsRegenerating.value = false
  }
}

function syncCurrentInner() {
  emit('save', { mode: 'inner_current', model: model.value, patch: buildPatch() })
}

function deduceRemainingLife() {
  if (!form.value.events.trim()) {
    ElMessage.warning('请先填写经历后再推演后续')
    return
  }
  emit('save', {
    mode: 'full_cascade',
    model: cascadeModel.value,
    patch: buildPatch(),
    target_node_count: regenConfig.value.target_node_count,
  })
}
</script>

<template>
  <div v-if="node" class="node-detail">
    <div class="detail-head">
      <div>
        <h3>节点详情 · {{ node.year }} 年</h3>
        <p class="sub-meta">第 {{ node.sequence + 1 }} 节点 · {{ node.age }} 岁</p>
      </div>
      <div v-if="uiMode === 'read'" class="head-actions">
        <el-button type="primary" :icon="EditPen" @click="enterEdit">编辑</el-button>
      </div>
      <div v-else class="head-actions">
        <el-button @click="cancelEdit">取消</el-button>
      </div>
    </div>

    <NodeSceneMeta :scene="node.scene" />

    <!-- 阅读模式 -->
    <div v-if="uiMode === 'read'" class="read-panel">
      <div class="read-section">
        <div class="section-label">标题</div>
        <div class="section-body title-body">
          <HighlightedText :text="node.title" :context="entityContext" tag="span" />
        </div>
      </div>

      <div class="read-section">
        <div class="section-label">经历</div>
        <div class="section-body prose">
          <HighlightedText :text="node.events" :context="entityContext" tag="span" />
        </div>
      </div>

      <div v-if="node.thoughts" class="read-section">
        <div class="section-label">内心想法</div>
        <div class="section-body prose muted">
          <HighlightedText :text="node.thoughts" :context="entityContext" tag="span" />
        </div>
      </div>

      <div v-if="node.personality_snapshot" class="read-section">
        <div class="section-label">性格快照</div>
        <div class="section-body prose muted">
          <HighlightedText
            :text="node.personality_snapshot"
            :context="entityContext"
            tag="span"
          />
        </div>
      </div>

      <div v-if="node.trait_changes?.length" class="read-section">
        <div class="section-label">性格/思想变更</div>
        <el-card v-for="(t, i) in node.trait_changes" :key="i" size="small" class="trait-card">
          <strong>{{ fieldLabel(t.field, 'trait') }}</strong>：
          <HighlightedText :text="`${t.before} → ${t.after}`" :context="entityContext" tag="span" />
          <p class="reason">
            <HighlightedText :text="t.reason" :context="entityContext" tag="span" />
          </p>
        </el-card>
      </div>

      <ModelSelector v-model="model" class="read-model" />

      <div class="read-toolbar">
        <el-button
          type="primary"
          plain
          :loading="eventsRegenerating"
          :disabled="!!loading"
          @click="regenerateEventsFromTitle"
        >
          根据标题重新生成本节点
        </el-button>
      </div>

      <div class="narrative-box">
        <div class="narrative-head">多视角叙事</div>
        <p class="narrative-hint">按需生成，结果会缓存；可在弹窗中重新生成。</p>
        <div class="narrative-actions">
          <el-button plain @click="openNarrative('diary')">查看日记</el-button>
          <el-button plain @click="openNarrative('letter')">查看书信</el-button>
          <el-button plain @click="openNarrative('archive')">查看档案</el-button>
        </div>
      </div>
    </div>

    <!-- 编辑模式 -->
    <el-form v-else label-position="top" class="edit-panel">
      <el-form-item label="标题">
        <el-input v-model="form.title" />
      </el-form-item>

      <div class="regen-events-row">
        <el-button
          type="primary"
          plain
          :loading="eventsRegenerating"
          :disabled="!!loading"
          @click="regenerateEventsFromTitle"
        >
          根据标题重新生成本节点
        </el-button>
        <span class="regen-hint">保留标题，由 AI 重写经历、内心想法与性格快照</span>
      </div>

      <el-form-item label="经历">
        <el-input v-model="form.events" type="textarea" :rows="6" />
      </el-form-item>

      <div class="derived-box">
        <div class="derived-head">
          <span class="derived-title">根据经历推演</span>
        </div>

        <el-form-item label="内心想法">
          <el-input v-model="form.thoughts" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="性格快照">
          <el-input v-model="form.personality_snapshot" type="textarea" :rows="2" />
        </el-form-item>

        <div v-if="node.trait_changes?.length" class="traits-inline">
          <div class="trait-label">性格/思想变更（只读，重算后更新）</div>
          <el-card v-for="(t, i) in node.trait_changes" :key="i" size="small" class="trait-card">
            <strong>{{ fieldLabel(t.field, 'trait') }}</strong>：
            <HighlightedText :text="`${t.before} → ${t.after}`" :context="entityContext" tag="span" />
          </el-card>
        </div>
      </div>

      <ModelSelector v-model="model" />

      <div class="narrative-box">
        <div class="narrative-head">多视角叙事</div>
        <div class="narrative-actions">
          <el-button plain @click="openNarrative('diary')">查看日记</el-button>
          <el-button plain @click="openNarrative('letter')">查看书信</el-button>
          <el-button plain @click="openNarrative('archive')">查看档案</el-button>
        </div>
      </div>

      <div class="actions">
        <el-button type="primary" :loading="loading" @click="syncCurrentInner">
          根据经历更新本节点想法/性格
        </el-button>
      </div>

      <div v-if="profile" class="cascade-box">
        <div class="cascade-head">推演后续</div>
        <p class="cascade-hint">
          从锚点起新增若干节点（如 M=1 仅生成下一个人生阶段）。会创建新分支并自动切换；分支内寿命随推演演变，与档案原卒年无关。默认 Pro 模型。
        </p>
        <ModelSelector v-model="cascadeModel" />
        <TimelineConfigPanel
          v-model="regenConfig"
          :profile="profile"
          :model="cascadeModel"
          :anchor-year="node.year"
          density-only
          :disabled="loading"
        />
        <el-button type="warning" :loading="loading" @click="deduceRemainingLife">
          推演后续
        </el-button>
      </div>

      <el-alert
        title="编辑后可「更新本节点」或「推演后续」；每次保存会创建新分支。"
        type="info"
        :closable="false"
        show-icon
      />
    </el-form>

    <JobProgress
      :visible="!!regenVisible"
      :progress="regenProgress ?? 0"
      :status-text="regenStatus ?? ''"
      title="任务进度"
    />
  </div>
  <el-empty v-else description="点击左侧节点打开详情" />
</template>

<style scoped>
.node-detail h3 {
  margin: 0;
  font-size: 1.05rem;
}
.detail-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 12px;
}
.sub-meta {
  margin: 4px 0 0;
  font-size: 0.82rem;
  color: #909399;
}
.head-actions {
  flex-shrink: 0;
}
.read-panel {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.read-section {
  margin-bottom: 12px;
}
.section-label {
  font-size: 0.78rem;
  font-weight: 600;
  color: #909399;
  margin-bottom: 6px;
}
.section-body {
  line-height: 1.65;
  color: #303133;
  font-size: 0.92rem;
}
.section-body.prose {
  padding: 10px 12px;
  background: #f8fafc;
  border-radius: 6px;
  white-space: pre-wrap;
  word-break: break-word;
}
.section-body.prose.muted {
  background: #fafafa;
  color: #606266;
}
.title-body {
  font-weight: 600;
  font-size: 1rem;
}
.read-toolbar {
  margin: 8px 0 12px;
}
.read-model {
  margin-bottom: 8px;
}
.regen-events-row {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 6px;
  margin-bottom: 12px;
}
.regen-hint {
  font-size: 0.8rem;
  color: #909399;
}
.derived-box {
  border: 2px solid #dcdfe6;
  border-radius: 8px;
  padding: 12px 14px 4px;
  margin-bottom: 8px;
  background: #fafbfc;
}
.derived-head {
  margin-bottom: 8px;
}
.derived-title {
  font-weight: 600;
  color: #303133;
  font-size: 0.9rem;
}
.traits-inline {
  margin-bottom: 8px;
}
.trait-label {
  font-size: 0.85rem;
  color: #606266;
  margin-bottom: 6px;
}
.trait-card {
  margin-bottom: 6px;
}
.reason {
  margin: 4px 0 0;
  color: #888;
  font-size: 0.85rem;
}
.actions {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 16px;
}
.narrative-box {
  margin: 12px 0 16px;
  padding: 12px 14px;
  border: 1px solid #d9ecff;
  border-radius: 8px;
  background: #f5faff;
}
.narrative-head {
  font-weight: 600;
  color: #409eff;
  margin-bottom: 4px;
}
.narrative-hint {
  margin: 0 0 10px;
  font-size: 0.82rem;
  color: #909399;
  line-height: 1.45;
}
.narrative-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.cascade-box {
  margin: 16px 0;
  padding: 14px;
  border: 1px solid #f0c78a;
  border-radius: 8px;
  background: #fffbf5;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.cascade-head {
  font-weight: 600;
  color: #e6a23c;
}
.cascade-hint {
  margin: 0;
  font-size: 0.85rem;
  color: #909399;
  line-height: 1.5;
}
.preview-card {
  background: #fff;
}
.preview-row {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  font-size: 0.9rem;
  margin-bottom: 8px;
}
.preview-row.highlight {
  color: #e6a23c;
  font-weight: 600;
}
.preview-label {
  color: #909399;
  flex-shrink: 0;
}
.preview-reason {
  margin: 0 0 12px;
  font-size: 0.85rem;
  color: #606266;
  line-height: 1.5;
}
.confirmed-alert {
  margin: 0;
}
</style>
