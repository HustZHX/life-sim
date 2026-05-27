<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ArrowDown, ArrowUp } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { api, type Profile, type TimelineConfig, type TimelineRecommendation } from '@/api/client'
import type { AIModelId } from '@/constants/models'
import { DEFAULT_AI_MODEL, DEFAULT_NARRATIVE_DENSITY } from '@/constants/models'
import {
  DEFAULT_TARGET_NODE_COUNT,
  MAX_TARGET_NODE_COUNT,
  MIN_TARGET_NODE_COUNT,
  computeStepYears,
  effectiveDeathYear,
  formatRegenSummary,
  formatTimelineSummary,
} from '@/utils/timelineDensity'

const props = defineProps<{
  profile: Profile
  model?: AIModelId
  disabled?: boolean
  /** 仅展示节点规模（用于重算后续） */
  densityOnly?: boolean
  /** 重算后续时锚点年份，用于计算剩余跨度 */
  anchorYear?: number
  /** 推演余生时展示用卒年上界（档案或预览），不参与节点数计算 */
  overrideEndYear?: number
}>()

const config = defineModel<TimelineConfig>({ required: true })

const loadingRecs = ref(false)
const recommendations = ref<TimelineRecommendation[]>([])
const selectedRecIndex = ref<number | null>(null)
const expanded = ref(false)

const yearRange = computed(() => ({
  min: props.profile.birth_year,
  max: effectiveDeathYear(props.profile),
}))

const spanStartYear = computed(() => {
  if (props.densityOnly && props.anchorYear != null) {
    return props.anchorYear
  }
  return config.value.start_year || props.profile.birth_year
})

const spanEndYear = computed(() => {
  if (props.overrideEndYear != null && props.overrideEndYear > 0) {
    return props.overrideEndYear
  }
  const end = config.value.end_year
  if (end > props.profile.birth_year) return end
  return effectiveDeathYear(props.profile)
})

const spanYears = computed(() => Math.max(0, spanEndYear.value - spanStartYear.value))

const computedStepYears = computed(() =>
  computeStepYears(spanYears.value, config.value.target_node_count)
)

const configSummary = computed(() => {
  if (props.densityOnly && props.anchorYear != null) {
    return formatRegenSummary(
      props.anchorYear,
      spanEndYear.value,
      config.value.target_node_count
    )
  }
  return formatTimelineSummary(
    config.value.start_year,
    config.value.end_year,
    config.value.target_node_count
  )
})

watch(
  () => props.profile,
  (p) => {
    if (!config.value.start_year) config.value.start_year = p.birth_year
    if (!config.value.end_year || config.value.end_year <= p.birth_year) {
      config.value.end_year = effectiveDeathYear(p)
    }
    if (!config.value.target_node_count) config.value.target_node_count = DEFAULT_TARGET_NODE_COUNT
    if (!config.value.narrative_density) config.value.narrative_density = DEFAULT_NARRATIVE_DENSITY
  },
  { immediate: true }
)

watch(computedStepYears, (step) => {
  config.value.step_years = step
})

async function fetchRecommendations() {
  loadingRecs.value = true
  selectedRecIndex.value = null
  try {
    const res = await api.recommendTimeline(props.profile.character_id, props.model || DEFAULT_AI_MODEL)
    recommendations.value = res.recommendations
  } catch (e: unknown) {
    recommendations.value = []
    ElMessage.error(e instanceof Error ? e.message : '获取推荐失败')
  } finally {
    loadingRecs.value = false
  }
}

function applyRecommendation(rec: TimelineRecommendation, index: number) {
  selectedRecIndex.value = index
  config.value = {
    ...config.value,
    target_node_count: rec.target_node_count,
    start_year: rec.start_year,
    end_year: rec.end_year,
    step_years: rec.step_years,
  }
}

function toggleExpanded() {
  if (props.disabled) return
  expanded.value = !expanded.value
}
</script>

<template>
  <div v-if="densityOnly" class="timeline-config density-only">
      <div class="block-label">推演后续 · 节点数</div>

    <button
      type="button"
      class="config-trigger"
      :class="{ open: expanded, disabled: disabled }"
      :disabled="disabled"
      @click="toggleExpanded"
    >
      <span class="trigger-main">
        <span class="trigger-name">约 {{ config.target_node_count }} 个后续节点</span>
        <span v-if="!expanded" class="trigger-sub">{{ configSummary }}</span>
      </span>
      <span class="trigger-action">
        {{ expanded ? '收起' : '调整配置' }}
        <el-icon><ArrowUp v-if="expanded" /><ArrowDown v-else /></el-icon>
      </span>
    </button>

    <Transition name="config-panel">
      <div v-if="expanded" class="config-panel">
        <p class="density-hint">
          本次仅新增指定个数的后续节点；年份由 AI 按事件安排，与寿命/卒年无机械对应。
        </p>
        <div class="form-row slider-row">
          <span class="field-label">目标节点数（约）</span>
          <el-slider
            v-model="config.target_node_count"
            :min="MIN_TARGET_NODE_COUNT"
            :max="MAX_TARGET_NODE_COUNT"
            :step="1"
            show-stops
            :disabled="disabled"
          />
        </div>
        <p class="computed-hint">当前：{{ configSummary }}</p>
      </div>
    </Transition>
  </div>

  <div v-else class="timeline-config">
    <div class="section-head">
      <span class="label">时间轴配置</span>
      <span class="summary">{{ configSummary }}</span>
    </div>

    <el-button
      type="primary"
      plain
      :loading="loadingRecs"
      :disabled="disabled"
      @click="fetchRecommendations"
    >
      AI 生成推荐方案
    </el-button>

    <div v-if="recommendations.length" v-loading="loadingRecs" class="rec-list">
      <el-card
        v-for="(rec, i) in recommendations"
        :key="i"
        shadow="hover"
        class="rec-card"
        :class="{ selected: selectedRecIndex === i }"
        @click="applyRecommendation(rec, i)"
      >
        <div class="rec-head">
          <strong>{{ rec.label }}</strong>
          <el-tag size="small" type="info">约 {{ rec.target_node_count }} 个节点</el-tag>
        </div>
        <p class="rec-desc">{{ rec.description }}</p>
        <p class="rec-range">
          {{ rec.start_year }} — {{ rec.end_year }} 年
          <span v-if="rec.focus_phase"> · {{ rec.focus_phase }}</span>
        </p>
      </el-card>
    </div>

    <p class="density-hint">
      节点数随人物寿命自动适配（如千年寿命仍约 20 个节点）。AI 按关键事件安排年份，非机械间隔。
    </p>

    <div class="custom-form">
      <div class="form-row density-mode-row">
        <span class="field-label">叙事密度</span>
        <el-radio-group v-model="config.narrative_density" :disabled="disabled">
          <el-radio value="rich">细腻（骨架 + 扩写，推荐）</el-radio>
          <el-radio value="standard">标准（单次生成，更快）</el-radio>
        </el-radio-group>
        <p class="density-hint">
          细腻模式在史实严谨前提下加厚 events/thoughts；Flash / Pro 均可选，Pro 文笔更细。
        </p>
      </div>
      <div class="form-row era-events-row">
        <span class="field-label">时代背景与大事记</span>
        <el-input
          v-model="config.era_events"
          type="textarea"
          :rows="4"
          :disabled="disabled"
          placeholder="留空则 AI 自动整理：如三国平民会补充黄巾、官渡、赤壁等对生计的影响；架空/小说世界观可自由创作大事"
        />
        <p v-if="profile.era_background" class="density-hint profile-era">
          档案中的时代背景：{{ profile.era_background }}
        </p>
        <p class="density-hint">
          生成时间轴前会先整理本区间大事，再写入各人生节点；真实历史须符合史实，虚构世界观可自创设定。
        </p>
      </div>
      <div class="form-row instructions-row">
        <span class="field-label">备注与特殊要求</span>
        <el-input
          v-model="config.instructions"
          type="textarea"
          :rows="3"
          :disabled="disabled"
          placeholder="例如：托孤后穿越到 2025 年；或聚焦青年科举阶段；或按《三体》世界观展开…"
        />
        <p class="density-hint">AI 生成时间轴时会优先遵循此处说明（穿越、架空、特定阶段等）。</p>
      </div>
      <div class="form-row slider-row">
        <span class="field-label">目标节点数（约）</span>
        <el-slider
          v-model="config.target_node_count"
          :min="MIN_TARGET_NODE_COUNT"
          :max="MAX_TARGET_NODE_COUNT"
          :step="1"
          show-stops
          :disabled="disabled"
        />
      </div>
      <p class="computed-hint">跨度 {{ spanYears }} 年 → 参考间隔约 {{ computedStepYears }} 年</p>
      <div class="form-row years">
        <div>
          <span class="field-label">起始年份</span>
          <el-input-number
            v-model="config.start_year"
            :min="yearRange.min"
            :max="yearRange.max"
            :disabled="disabled"
          />
        </div>
        <div>
          <span class="field-label">结束年份</span>
          <el-input-number
            v-model="config.end_year"
            :min="config.start_year || yearRange.min"
            :max="yearRange.max"
            :disabled="disabled"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.timeline-config {
  margin: 16px 0;
  padding: 16px;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  background: #fafbfc;
  min-width: 0;
  overflow: hidden;
}
.timeline-config.density-only {
  margin: 0 0 16px;
  padding: 0;
  border: none;
  background: transparent;
}
.block-label {
  font-weight: 600;
  margin-bottom: 8px;
  color: #333;
}
.config-trigger {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  width: 100%;
  padding: 10px 14px;
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  background: #fafbfc;
  cursor: pointer;
  text-align: left;
  transition: border-color 0.2s, background 0.2s;
  min-width: 0;
}
.config-trigger:hover:not(:disabled),
.config-trigger.open {
  border-color: #409eff;
  background: #f5f9ff;
}
.config-trigger:disabled {
  cursor: not-allowed;
  opacity: 0.65;
}
.trigger-main {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
  flex: 1;
}
.trigger-name {
  font-weight: 600;
  color: #303133;
  word-break: break-word;
}
.trigger-sub {
  font-size: 0.78rem;
  color: #909399;
  line-height: 1.35;
  word-break: break-word;
}
.trigger-action {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
  font-size: 0.82rem;
  color: #409eff;
  white-space: nowrap;
}
.config-panel {
  margin-top: 10px;
  padding: 12px 14px;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  background: #fafbfc;
  min-width: 0;
  overflow: hidden;
}
.section-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}
.label {
  font-weight: 600;
  color: #303133;
}
.summary {
  font-size: 0.85rem;
  color: #409eff;
  text-align: right;
  word-break: break-word;
  line-height: 1.4;
}
.rec-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin: 14px 0;
}
.rec-card {
  cursor: pointer;
  transition: border-color 0.2s;
}
.rec-card.selected {
  border-color: #409eff;
  background: #f0f7ff;
}
.rec-head {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 6px;
}
.rec-desc {
  margin: 0 0 6px;
  font-size: 0.88rem;
  color: #606266;
  line-height: 1.45;
  word-break: break-word;
}
.rec-range {
  margin: 0;
  font-size: 0.8rem;
  color: #909399;
  word-break: break-word;
}
.density-hint,
.computed-hint {
  margin: 0 0 12px;
  font-size: 0.82rem;
  color: #909399;
  line-height: 1.5;
  word-break: break-word;
}
.computed-hint {
  margin-top: 4px;
  color: #409eff;
}
.custom-form {
  margin-top: 16px;
}
.form-row {
  margin-bottom: 12px;
  min-width: 0;
}
.form-row.years {
  display: flex;
  gap: 24px;
  flex-wrap: wrap;
}
.field-label {
  display: block;
  font-size: 0.85rem;
  color: #606266;
  margin-bottom: 6px;
}
.slider-row :deep(.el-slider) {
  box-sizing: border-box;
  width: calc(100% - 16px);
  max-width: 100%;
  margin: 8px 8px 12px;
}
.slider-row :deep(.el-slider__runway) {
  margin: 0;
}
.config-panel-enter-active,
.config-panel-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}
.config-panel-enter-from,
.config-panel-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}
</style>
