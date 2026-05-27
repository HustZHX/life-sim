<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  api,
  type AIModelId,
  type LifeNode,
  type LifespanPreviewResponse,
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
      confirmed_death_year?: number
      confirmed_death_cause?: string
      lifespan_reasoning?: string
    },
  ]
  narrative: [kind: Exclude<NarrativeKind, 'light_novel'>, model: AIModelId]
}>()

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

const previewLoading = ref(false)
const lifespanPreview = ref<LifespanPreviewResponse | null>(null)
const lifespanConfirmed = ref(false)
const confirmedDeathYear = ref(0)
const confirmedDeathCause = ref('')
const confirmedLifespanReasoning = ref('')

const entityContext = computed<EntityHighlightContext>(() =>
  entityContextFromNode(props.node, props.profile ?? null)
)

function resetLifespanFlow() {
  lifespanPreview.value = null
  lifespanConfirmed.value = false
  confirmedDeathYear.value = 0
  confirmedDeathCause.value = ''
  confirmedLifespanReasoning.value = ''
}

watch(
  () => props.node,
  (n) => {
    if (n) {
      form.value = {
        title: n.title,
        events: n.events,
        thoughts: n.thoughts,
        personality_snapshot: n.personality_snapshot,
      }
      resetLifespanFlow()
    }
  },
  { immediate: true }
)

watch(
  () => [form.value.title, form.value.events] as const,
  () => {
    if (lifespanPreview.value || lifespanConfirmed.value) {
      resetLifespanFlow()
    }
  }
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

function openNarrative(kind: Exclude<NarrativeKind, 'light_novel'>) {
  emit('narrative', kind, model.value)
}

function syncCurrentInner() {
  emit('save', { mode: 'inner_current', model: model.value, patch: buildPatch() })
}

function syncSubsequentInner() {
  emit('save', { mode: 'inner_subsequent', model: model.value, patch: buildPatch() })
}

async function previewLifespan() {
  if (!props.node || !props.characterId) return
  if (!form.value.events.trim()) {
    ElMessage.warning('请先填写经历后再预览寿命')
    return
  }
  previewLoading.value = true
  resetLifespanFlow()
  try {
    lifespanPreview.value = await api.previewLifespan(props.characterId, props.node.id, {
      ...buildPatch(),
      model: model.value,
    })
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '寿命预览失败')
  } finally {
    previewLoading.value = false
  }
}

function confirmLifespan() {
  if (!lifespanPreview.value) return
  const p = lifespanPreview.value.preview
  if (p.death_year <= lifespanPreview.value.anchor_year) {
    ElMessage.error('预览卒年须大于锚点年份')
    return
  }
  confirmedDeathYear.value = p.death_year
  confirmedDeathCause.value = p.death_cause || ''
  confirmedLifespanReasoning.value = p.reasoning || ''
  lifespanConfirmed.value = true
  regenConfig.value.end_year = p.death_year
}

function regenerateAllSubsequent() {
  if (!lifespanConfirmed.value || confirmedDeathYear.value <= 0) {
    ElMessage.warning('请先预览并确认寿命')
    return
  }
  emit('save', {
    mode: 'full_cascade',
    model: cascadeModel.value,
    patch: buildPatch(),
    target_node_count: regenConfig.value.target_node_count,
    confirmed_death_year: confirmedDeathYear.value,
    confirmed_death_cause: confirmedDeathCause.value,
    lifespan_reasoning: confirmedLifespanReasoning.value,
  })
}
</script>

<template>
  <div v-if="node" class="editor">
    <h3>编辑节点 · {{ node.year }} 年</h3>
    <NodeSceneMeta :scene="node.scene" />

    <el-form label-position="top">
      <el-form-item label="标题">
        <el-input v-model="form.title" />
      </el-form-item>

      <el-form-item label="经历">
        <el-input v-model="form.events" type="textarea" :rows="4" />
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
          <div class="trait-label">性格/思想变更</div>
          <el-card v-for="(t, i) in node.trait_changes" :key="i" size="small" class="trait-card">
            <strong>{{ fieldLabel(t.field, 'trait') }}</strong>：
            <HighlightedText :text="`${t.before} → ${t.after}`" :context="entityContext" tag="span" />
            <p class="reason">
              <HighlightedText :text="t.reason" :context="entityContext" tag="span" />
            </p>
          </el-card>
        </div>
      </div>

      <ModelSelector v-model="model" />

      <div class="narrative-box">
        <div class="narrative-head">多视角叙事</div>
        <p class="narrative-hint">按需生成，结果会缓存；可在弹窗中重新生成。</p>
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
        <el-button type="primary" plain :loading="loading" @click="syncSubsequentInner">
          保留后续经历，重算内心与性格
        </el-button>
      </div>

      <div class="cascade-box">
        <div class="cascade-head">完全重算后续（两步）</div>
        <p class="cascade-hint">第一步根据新经历预览寿命；确认后再配置节点密度并生成。默认使用 Pro 模型。</p>

        <el-button
          type="warning"
          plain
          :loading="previewLoading"
          :disabled="loading"
          @click="previewLifespan"
        >
          第一步：预览寿命
        </el-button>

        <el-card v-if="lifespanPreview && !lifespanConfirmed" class="preview-card" shadow="never">
          <div class="preview-row">
            <span class="preview-label">档案卒年</span>
            <span>{{ lifespanPreview.current_death_year }} 年</span>
          </div>
          <div class="preview-row highlight">
            <span class="preview-label">预览卒年</span>
            <span>{{ lifespanPreview.preview.death_year }} 年</span>
          </div>
          <div v-if="lifespanPreview.preview.death_cause" class="preview-row">
            <span class="preview-label">死因</span>
            <span>{{ lifespanPreview.preview.death_cause }}</span>
          </div>
          <p class="preview-reason">{{ lifespanPreview.preview.reasoning }}</p>
          <el-button type="primary" size="small" @click="confirmLifespan">确认此寿命</el-button>
        </el-card>

        <el-alert
          v-if="lifespanConfirmed"
          :title="`已确认卒年：${confirmedDeathYear} 年（锚点 ${node.year} 年）`"
          type="success"
          :closable="false"
          show-icon
          class="confirmed-alert"
        />

        <template v-if="lifespanConfirmed && profile">
          <ModelSelector v-model="cascadeModel" />
          <TimelineConfigPanel
            v-model="regenConfig"
            :profile="profile"
            :model="cascadeModel"
            :anchor-year="node.year"
            :override-end-year="confirmedDeathYear"
            density-only
            :disabled="loading"
          />
          <el-button type="warning" :loading="loading" @click="regenerateAllSubsequent">
            第二步：生成后续时间轴
          </el-button>
        </template>
      </div>

      <el-alert
        title="三种重算粒度：① 仅本节点内心 ② 保留后续经历只重算内心 ③ 完全重算须先预览确认寿命再生成后续节点。"
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
  <el-empty v-else description="点击左侧节点查看或编辑" />
</template>

<style scoped>
.editor h3 {
  margin-top: 0;
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
