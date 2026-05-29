<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { api, type GameEraOption, type Profile } from '@/api/client'
import { DEFAULT_AI_MODEL, type AIModelId } from '@/constants/models'
import { useAwaitProgress } from '@/composables/useAwaitProgress'
import { useGameCreate } from '@/composables/useGameCreate'
import CandidateCards from '@/components/CandidateCards.vue'
import EraPeriodCards from '@/components/EraPeriodCards.vue'
import ProfilePanel from '@/components/ProfilePanel.vue'
import ModelSelector from '@/components/ModelSelector.vue'
import JobProgress from '@/components/JobProgress.vue'

const router = useRouter()
const step = ref(1)
const query = ref('')
const characterId = ref('')
const candidates = ref<import('@/api/client').ResolveCandidate[]>([])
const eraDescription = ref('')
const eraOptions = ref<GameEraOption[]>([])
const selectedEra = ref<GameEraOption | null>(null)
const profile = ref<Profile | null>(null)
const loading = ref(false)
const eraLoading = ref(false)
const profileModel = ref<AIModelId>(DEFAULT_AI_MODEL)
const startModel = ref<AIModelId>(DEFAULT_AI_MODEL)

const resolveJob = useAwaitProgress('查找候选人')
const profileJob = useAwaitProgress('生成档案')
const { jobRunner, starting, startGameTimeline } = useGameCreate()

async function start() {
  if (!query.value.trim()) {
    ElMessage.warning('请输入名人姓名')
    return
  }
  loading.value = true
  try {
    let cid = ''
    const res = await resolveJob.run(
      async () => {
        const r = await api.resolve(cid, query.value.trim())
        return r
      },
      {
        title: '查找候选人',
        startText: '正在创建角色…',
        onMid: async () => {
          const ch = await api.createGameCharacter('famous')
          cid = ch.id
          characterId.value = ch.id
        },
        midText: 'AI 正在检索历史人物…',
      }
    )
    candidates.value = res.candidates
    step.value = 2
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '消歧失败')
  } finally {
    loading.value = false
  }
}

async function loadSavedEraState() {
  if (!characterId.value) return
  try {
    const ch = await api.getCharacter(characterId.value)
    const cfg = ch.game_config
    if (cfg?.era_description) eraDescription.value = cfg.era_description
    if (cfg?.era_options?.length) eraOptions.value = cfg.era_options
    if (cfg?.selected_era && cfg?.start_year_hint) {
      const match = cfg.era_options?.find((o) => o.label === cfg.selected_era)
      if (match) selectedEra.value = match
    }
  } catch {
    /* 忽略加载失败 */
  }
}

async function onSelect(index: number) {
  loading.value = true
  try {
    await api.confirm(characterId.value, index)
    step.value = 3
    await loadSavedEraState()
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '确认失败')
  } finally {
    loading.value = false
  }
}

async function generateEraOptions(regenerate = false) {
  eraLoading.value = true
  const previousOptions = [...eraOptions.value]
  try {
    const res = await api.gameEraOptions(characterId.value, {
      model: profileModel.value,
      era_description: eraDescription.value.trim(),
      regenerate,
    })
    eraOptions.value = res.options
    if (!res.options.length) {
      if (previousOptions.length) eraOptions.value = previousOptions
      ElMessage.warning('未生成时期选项')
    }
  } catch (e: unknown) {
    if (previousOptions.length) eraOptions.value = previousOptions
    ElMessage.error(e instanceof Error ? e.message : '生成失败')
  } finally {
    eraLoading.value = false
  }
}

function onEraSelect(opt: GameEraOption) {
  selectedEra.value = opt
}

async function confirmEraAndProfile() {
  if (!selectedEra.value) {
    ElMessage.warning('请选择一个时期')
    return
  }
  loading.value = true
  try {
    await api.updateGameConfig(characterId.value, {
      era_description: eraDescription.value.trim(),
      era_options: eraOptions.value.length ? eraOptions.value : undefined,
      selected_era: selectedEra.value.label,
      start_year_hint: selectedEra.value.start_year,
    })
    const p = await profileJob.run(
      () => api.gameGenerateProfile(characterId.value, { model: profileModel.value }),
      { title: '生成阶段档案', startText: 'AI 正在整理该时期的人物状态…' }
    )
    profile.value = p
    step.value = 4
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '档案生成失败')
  } finally {
    loading.value = false
  }
}

async function onStartGame() {
  const result = await startGameTimeline(characterId.value, startModel.value)
  if (result.ok) {
    router.push(`/game/timeline/${characterId.value}`)
  }
}
</script>

<template>
  <div class="page-card">
    <JobProgress
      :visible="resolveJob.visible.value || profileJob.visible.value || jobRunner.visible.value"
      :progress="jobRunner.visible.value ? jobRunner.progress.value : resolveJob.visible.value ? resolveJob.progress.value : profileJob.progress.value"
      :status-text="jobRunner.visible.value ? jobRunner.statusText.value : resolveJob.visible.value ? resolveJob.statusText.value : profileJob.statusText.value"
      :title="jobRunner.visible.value ? jobRunner.title.value : resolveJob.visible.value ? resolveJob.title.value : profileJob.title.value"
      :failed="jobRunner.failed.value"
      :error-text="jobRunner.errorText.value"
      :retrying="jobRunner.retrying.value"
      @retry="jobRunner.retry()"
      @dismiss="jobRunner.clearState()"
    />

    <el-steps :active="step - 1" finish-status="success" align-center style="margin-bottom: 24px">
      <el-step title="输入姓名" />
      <el-step title="确认身份" />
      <el-step title="选择时期" />
      <el-step title="开始游戏" />
    </el-steps>

    <div v-if="step === 1">
      <h2 class="step-title">输入历史名人姓名</h2>
      <el-input v-model="query" placeholder="例如：诸葛亮、李白" size="large" :disabled="loading" @keyup.enter="start" />
      <ModelSelector v-model="profileModel" style="margin-top: 16px" />
      <el-button type="primary" size="large" :loading="loading" style="margin-top: 16px" @click="start">查找候选人</el-button>
    </div>

    <div v-if="step === 2">
      <h2 class="step-title">请确认是哪位人物</h2>
      <CandidateCards :candidates="candidates" :disabled="loading" @select="onSelect" />
    </div>

    <div v-if="step === 3">
      <h2 class="step-title">从什么时期开始？</h2>
      <div class="era-toolbar">
        <el-input
          v-model="eraDescription"
          type="textarea"
          :rows="3"
          placeholder="描述你想从他人生的哪个阶段开始玩，例如：隆中对之前、安史之乱后…"
          :disabled="loading"
        />
        <div class="era-actions">
          <el-button type="primary" :loading="eraLoading" :disabled="loading" @click="generateEraOptions(false)">
            生成时期选项
          </el-button>
          <el-button
            v-if="eraOptions.length"
            :loading="eraLoading"
            :disabled="loading"
            @click="generateEraOptions(true)"
          >
            重新生成
          </el-button>
        </div>
      </div>
      <EraPeriodCards
        v-if="eraOptions.length"
        :options="eraOptions"
        :selected-label="selectedEra?.label"
        :disabled="loading"
        style="margin-top: 16px"
        @select="onEraSelect"
      />
      <el-button
        type="primary"
        size="large"
        :disabled="!selectedEra || loading"
        style="margin-top: 20px"
        @click="confirmEraAndProfile"
      >
        确认时期并生成档案
      </el-button>
    </div>

    <div v-if="step === 4 && profile">
      <h2 class="step-title">档案预览 · {{ selectedEra?.label }}</h2>
      <ProfilePanel :profile="profile" />
      <ModelSelector v-model="startModel" style="margin-top: 16px" />
      <el-button type="primary" size="large" :loading="starting" style="margin-top: 16px" @click="onStartGame">
        开始人生游戏
      </el-button>
    </div>
  </div>
</template>

<style scoped>
.step-title {
  margin-bottom: 16px;
}
.era-toolbar {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.era-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
@media (min-width: 640px) {
  .era-toolbar {
    flex-direction: row;
    align-items: flex-start;
  }
  .era-actions {
    flex-shrink: 0;
    flex-direction: column;
  }
}
</style>
