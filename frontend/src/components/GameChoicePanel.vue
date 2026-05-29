<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { api, type GameChoiceOption } from '@/api/client'
import { useJobRunner } from '@/composables/useJobRunner'
import { useModelStore } from '@/stores/model'
import JobProgress from '@/components/JobProgress.vue'

const props = defineProps<{
  charId: string
  nodeId: string
  isFamous?: boolean
  visible: boolean
}>()

const emit = defineEmits<{
  done: []
  simulating: [busy: boolean]
}>()

const options = ref<GameChoiceOption[]>([])
const loading = ref(false)
/** 后台预生成尚未完成（GET regenerate=false 缓存未命中） */
const prefetchWaiting = ref(false)
let prefetchPollTimer: ReturnType<typeof setInterval> | null = null

const PREFETCH_POLL_MS = 2500

/** 仅首次请求选项时全屏 loading；后台预生成不阻塞自定义推演 */
const optionsFetching = computed(() => loading.value && !prefetchWaiting.value)
const customText = ref('')
const alreadyChosen = ref(false)
const chosenLabel = ref('')
const lockedChoiceId = ref<string | null>(null)
const lockedUseCustom = ref(false)
/** 已确认抉择（含推演进行中）：选项置灰不可再选 */
const choiceLocked = ref(false)
const modelStore = useModelStore()
const jobRunner = useJobRunner('抉择推演')

const optionsDisabled = computed(
  () =>
    choiceLocked.value ||
    alreadyChosen.value ||
    optionsFetching.value ||
    prefetchWaiting.value ||
    jobRunner.visible.value
)

/** 待确认：点击选项仅高亮，需再点「确认选择」才推演 */
const pendingChoiceId = ref<string | null>(null)
const pendingUseCustom = ref(false)

const canConfirm = computed(() => {
  if (choiceLocked.value || optionsFetching.value || jobRunner.visible.value) return false
  const custom = customText.value.trim()
  if (pendingUseCustom.value && custom) return true
  if (custom && !pendingChoiceId.value) return true
  return !!pendingChoiceId.value
})

function applyLockedFromResponse(res: { already_chosen?: boolean; chosen_id?: string; custom_text?: string }) {
  if (!res.already_chosen) {
    choiceLocked.value = false
    lockedChoiceId.value = null
    lockedUseCustom.value = false
    return
  }
  choiceLocked.value = true
  lockedChoiceId.value = res.chosen_id || null
  lockedUseCustom.value = !!res.custom_text?.trim()
  const picked = res.chosen_id ? options.value.find((o) => o.id === res.chosen_id) : undefined
  chosenLabel.value = res.custom_text?.trim() || picked?.label || '已抉择'
}

function selectOption(opt: GameChoiceOption) {
  if (optionsDisabled.value) return
  pendingChoiceId.value = opt.id
  pendingUseCustom.value = false
}

function selectCustom() {
  if (choiceLocked.value || optionsFetching.value || jobRunner.visible.value) return
  if (!customText.value.trim()) {
    ElMessage.warning('请先输入自定义抉择')
    return
  }
  pendingUseCustom.value = true
  pendingChoiceId.value = null
}

function clearPending() {
  pendingChoiceId.value = null
  pendingUseCustom.value = false
}

function isPrefetchPendingError(e: unknown): boolean {
  const msg = e instanceof Error ? e.message : String(e)
  return msg.includes('尚未生成') || msg.includes('请稍后')
}

function stopPrefetchPoll() {
  if (prefetchPollTimer) {
    clearInterval(prefetchPollTimer)
    prefetchPollTimer = null
  }
}

function startPrefetchPoll() {
  if (prefetchPollTimer || choiceLocked.value) return
  prefetchPollTimer = setInterval(() => {
    if (!props.visible || choiceLocked.value || options.value.length || loading.value) {
      if (options.value.length) stopPrefetchPoll()
      return
    }
    void loadChoices(false)
  }, PREFETCH_POLL_MS)
}

function resetPrefetchState() {
  prefetchWaiting.value = false
  stopPrefetchPoll()
}

async function loadChoices(regenerate = false) {
  if (!props.nodeId) return
  if (loading.value && !regenerate) return
  loading.value = true
  if (regenerate) resetPrefetchState()
  clearPending()
  try {
    const res = await api.gameGetChoices(props.charId, props.nodeId, modelStore.aiModel, regenerate)
    options.value = res.options
    alreadyChosen.value = !!res.already_chosen
    applyLockedFromResponse(res)
    prefetchWaiting.value = false
    stopPrefetchPoll()
    if (regenerate && res.options.length) {
      ElMessage.success('已刷新抉择选项')
    }
  } catch (e: unknown) {
    if (!regenerate && isPrefetchPendingError(e)) {
      options.value = []
      prefetchWaiting.value = true
      startPrefetchPoll()
    } else {
      resetPrefetchState()
      ElMessage.error(e instanceof Error ? e.message : '加载抉择失败')
    }
  } finally {
    loading.value = false
  }
}

onUnmounted(() => {
  stopPrefetchPoll()
})

watch(
  () => [props.visible, props.nodeId] as const,
  ([vis, nid]) => {
    if (!vis) {
      resetPrefetchState()
      return
    }
    if (nid) {
      customText.value = ''
      clearPending()
      choiceLocked.value = false
      lockedChoiceId.value = null
      lockedUseCustom.value = false
      options.value = []
      resetPrefetchState()
      void loadChoices(false)
    }
  },
  { immediate: true }
)

async function confirmChoice() {
  if (alreadyChosen.value) {
    ElMessage.info('该节点已做出抉择，请回退后重试')
    return
  }
  const custom = customText.value.trim()
  const useCustom = pendingUseCustom.value || (!pendingChoiceId.value && !!custom)
  const choiceId = useCustom ? undefined : pendingChoiceId.value ?? undefined
  const customPayload = useCustom ? custom : ''
  if (!choiceId && !customPayload) {
    ElMessage.warning('请先点选一项抉择，或填写自定义抉择')
    return
  }
  choiceLocked.value = true
  alreadyChosen.value = true
  lockedChoiceId.value = choiceId ?? null
  lockedUseCustom.value = !!customPayload
  if (customPayload) {
    chosenLabel.value = customPayload
  } else {
    const picked = options.value.find((o) => o.id === choiceId)
    chosenLabel.value = picked?.label || '已抉择'
  }
  emit('simulating', true)
  const ok = await jobRunner.run({
    title: '抉择推演',
    submitLabel: '正在推演后果…',
    submit: () =>
      api.gameChoose(props.charId, props.nodeId, {
        model: modelStore.aiModel,
        choice_id: choiceId,
        custom_text: customPayload || undefined,
      }),
    resubmit: () =>
      api.gameChoose(props.charId, props.nodeId, {
        model: modelStore.aiModel,
        choice_id: choiceId,
        custom_text: customPayload || undefined,
      }),
    afterSuccess: () => {
      emit('done')
    },
  })
  emit('simulating', false)
  if (!ok && !jobRunner.failed.value) {
    ElMessage.error('提交失败')
    choiceLocked.value = false
    alreadyChosen.value = false
    lockedChoiceId.value = null
    lockedUseCustom.value = false
    void loadChoices(false)
  }
}
</script>

<template>
  <div v-if="visible" class="game-choice-panel">
    <JobProgress
      :visible="jobRunner.visible.value"
      :progress="jobRunner.progress.value"
      :status-text="jobRunner.statusText.value"
      :title="jobRunner.title.value"
      :failed="jobRunner.failed.value"
      :error-text="jobRunner.errorText.value"
      :retrying="jobRunner.retrying.value"
      @retry="jobRunner.retry()"
      @dismiss="jobRunner.clearState()"
    />

    <div class="panel-head">
      <h3>人生抉择</h3>
      <el-button
        v-if="!choiceLocked"
        text
        type="primary"
        :loading="loading"
        @click="loadChoices(true)"
      >
        刷新选项
      </el-button>
    </div>

    <p v-if="!choiceLocked && !optionsFetching" class="panel-hint">
      <template v-if="prefetchWaiting">
        预设选项正在后台生成，你可先填写自定义抉择并推演，无需等待。
      </template>
      <template v-else>
        以下为人生大事级分岔；点选或填写自定义后，再点「确认选择」开始推演。
      </template>
    </p>

    <div v-if="choiceLocked" class="chosen-hint">
      <template v-if="jobRunner.visible.value">已确认抉择，正在推演…</template>
      <template v-else>已选择：{{ chosenLabel }}</template>
    </div>

    <div v-else-if="optionsFetching" v-loading="true" class="loading-block" element-loading-text="正在加载选项…">
      <span class="loading-hint">正在加载选项…</span>
    </div>

    <p v-else-if="prefetchWaiting && !options.length" class="prefetch-banner">
      预设选项生成中… 可先使用下方自定义抉择。
    </p>

    <div v-if="options.length" class="options" :class="{ 'options--locked': choiceLocked }">
      <button
        v-for="opt in options"
        :key="opt.id"
        type="button"
        class="choice-card"
        :class="{
          historical: opt.is_historical && isFamous,
          'is-selected':
            !choiceLocked && pendingChoiceId === opt.id && !pendingUseCustom,
          'is-chosen': choiceLocked && !lockedUseCustom && lockedChoiceId === opt.id,
          'is-dimmed': choiceLocked && (lockedUseCustom || lockedChoiceId !== opt.id),
        }"
        :disabled="optionsDisabled"
        @click="selectOption(opt)"
      >
        <span v-if="opt.is_historical && isFamous" class="hist-badge">史实</span>
        <strong>{{ opt.label }}</strong>
        <p>{{ opt.description }}</p>
      </button>
    </div>

    <div v-if="!choiceLocked && !optionsFetching" class="custom-block">
      <el-input
        v-model="customText"
        type="textarea"
        :rows="2"
        placeholder="输入你自己的抉择（选项未生成时也可直接推演）…"
        :disabled="choiceLocked || jobRunner.visible.value"
        @input="clearPending"
      />
      <div class="custom-actions">
        <el-button
          size="small"
          :type="pendingUseCustom ? 'primary' : 'default'"
          plain
          @click="selectCustom"
        >
          选用自定义输入
        </el-button>
      </div>
    </div>

    <div v-if="!choiceLocked" class="confirm-row">
      <el-button type="primary" size="large" :disabled="!canConfirm" @click="confirmChoice">
        确认选择
      </el-button>
    </div>

  </div>
</template>

<style scoped>
.game-choice-panel {
  margin-top: 16px;
  padding: 16px;
  border: 1px dashed #dcdfe6;
  border-radius: 8px;
  background: #fafafa;
}
.panel-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}
.panel-head h3 {
  margin: 0;
  font-size: 1rem;
}
.panel-hint {
  margin: 0 0 12px;
  font-size: 0.85rem;
  color: #909399;
  line-height: 1.5;
}
.options {
  display: grid;
  gap: 10px;
}
.choice-card {
  position: relative;
  text-align: left;
  padding: 12px 14px;
  border: 2px solid #dcdfe6;
  border-radius: 8px;
  background: #fff;
  cursor: pointer;
  transition: border-color 0.15s, background 0.15s, box-shadow 0.15s;
}
.choice-card:hover {
  border-color: #409eff;
}
.choice-card.is-selected {
  border-color: #409eff;
  background: #ecf5ff;
  box-shadow: 0 0 0 1px #409eff;
}
.choice-card.historical {
  border-color: #e6a23c;
  background: #fdf6ec;
}
.choice-card.historical.is-selected {
  border-color: #e6a23c;
  background: #faecd8;
  box-shadow: 0 0 0 2px rgba(230, 162, 60, 0.45);
}
.options--locked .choice-card {
  cursor: default;
}
.choice-card.is-chosen {
  border-color: #a8abb2;
  background: #f4f4f5;
  box-shadow: none;
  opacity: 1;
}
.choice-card.is-dimmed {
  opacity: 0.45;
  filter: grayscale(0.15);
}
.choice-card:disabled {
  cursor: not-allowed;
}
.choice-card.is-dimmed:hover,
.choice-card.is-chosen:hover {
  border-color: #dcdfe6;
}
.choice-card.historical.is-chosen {
  border-color: #c0c4cc;
  background: #f4f4f5;
}
.hist-badge {
  position: absolute;
  top: 8px;
  right: 10px;
  font-size: 0.75rem;
  color: #e6a23c;
  font-weight: 600;
}
.choice-card strong {
  display: block;
  margin-bottom: 4px;
}
.choice-card p {
  margin: 0;
  font-size: 0.85rem;
  color: #606266;
}
.custom-block {
  margin-top: 12px;
}
.custom-actions {
  margin-top: 8px;
}
.confirm-row {
  margin-top: 16px;
  display: flex;
  justify-content: center;
}
.chosen-hint,
.loading-hint,
.prefetch-banner {
  color: #909399;
  font-size: 0.9rem;
}
.prefetch-banner {
  margin: 0 0 12px;
  padding: 8px 12px;
  background: #f4f4f5;
  border-radius: 6px;
  line-height: 1.5;
}
.loading-block {
  min-height: 72px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 8px;
}
.loading-block .loading-hint {
  visibility: hidden;
}
</style>
