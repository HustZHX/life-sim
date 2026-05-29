<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { api, type Profile, type TimelineConfig } from '@/api/client'
import { DEFAULT_TARGET_NODE_COUNT, effectiveDeathYear } from '@/utils/timelineDensity'
import { useModelStore } from '@/stores/model'
import { useTimelineGenerate } from '@/composables/useTimelineGenerate'
import { useAwaitProgress } from '@/composables/useAwaitProgress'
import CandidateCards from '@/components/CandidateCards.vue'
import ProfileEditor from '@/components/ProfileEditor.vue'
import TimelineConfigPanel from '@/components/TimelineConfigPanel.vue'
import JobProgress from '@/components/JobProgress.vue'

const router = useRouter()
const step = ref(1)
const query = ref('')
const characterId = ref('')
const candidates = ref<import('@/api/client').ResolveCandidate[]>([])
const profile = ref<Profile | null>(null)
const loading = ref(false)
const modelStore = useModelStore()
const timelineConfig = ref<TimelineConfig>({ target_node_count: DEFAULT_TARGET_NODE_COUNT, start_year: 0, end_year: 0 })

const { jobRunning, progress, statusText, failed, errorText, retrying, retry, clearJobState, confirmAndGenerate } = useTimelineGenerate()
const profileJob = useAwaitProgress('搜寻人物资料')
const resolveJob = useAwaitProgress('查找候选人')

const progressVisible = computed(
  () => jobRunning.value || failed.value || profileJob.visible.value || resolveJob.visible.value
)
const progressValue = computed(() => {
  if (jobRunning.value) return progress.value
  if (profileJob.visible.value) return profileJob.progress.value
  if (resolveJob.visible.value) return resolveJob.progress.value
  return 0
})
const progressStatus = computed(() => {
  if (jobRunning.value) return statusText.value
  if (profileJob.visible.value) return profileJob.statusText.value
  if (resolveJob.visible.value) return resolveJob.statusText.value
  return ''
})
const progressTitle = computed(() => {
  if (jobRunning.value) return '时间轴生成中'
  if (profileJob.visible.value) return profileJob.title.value
  if (resolveJob.visible.value) return resolveJob.title.value
  return '处理中'
})

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
          const ch = await api.createCharacter('famous')
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

async function onSelect(index: number) {
  loading.value = true
  try {
    const p = await profileJob.run(
      () => api.generateProfile(characterId.value, { model: modelStore.aiModel }),
      {
        title: '搜寻人物资料',
        startText: '正在确认人物身份…',
        onMid: () => api.confirm(characterId.value, index).then(() => undefined),
        midText: 'AI 正在搜寻史料、整理人物档案…',
      }
    )
    profile.value = p
    timelineConfig.value = {
      target_node_count: DEFAULT_TARGET_NODE_COUNT,
      start_year: p.birth_year,
      end_year: effectiveDeathYear(p),
    }
    step.value = 3
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '确认失败')
  } finally {
    loading.value = false
  }
}

function onProfileUpdate(p: Profile) {
  profile.value = p
  timelineConfig.value.end_year = effectiveDeathYear(p)
}

async function onGenerateClick() {
  try {
    const result = await confirmAndGenerate(characterId.value, modelStore.aiModel, {
      displayName: profile.value?.display_name,
      config: timelineConfig.value,
      background: true,
    })
    if (result.ok) {
      router.push(`/characters/${characterId.value}`)
    }
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '无法启动生成')
  }
}
</script>

<template>
  <div class="page-card">
    <JobProgress
      :visible="progressVisible"
      :progress="progressValue"
      :status-text="progressStatus"
      :title="progressTitle"
      :failed="failed"
      :error-text="errorText"
      :retrying="retrying"
      @retry="retry()"
      @dismiss="clearJobState()"
    />

    <el-steps :active="step - 1" finish-status="success" align-center style="margin-bottom: 24px">
      <el-step title="输入姓名" />
      <el-step title="确认身份" />
      <el-step title="预览档案" />
    </el-steps>

    <div v-if="step === 1">
      <h2 class="step-title">输入历史名人姓名</h2>
      <el-input
        v-model="query"
        placeholder="例如：李白、拿破仑"
        size="large"
        :disabled="loading"
        @keyup.enter="start"
      />
      <el-button type="primary" size="large" :loading="loading" style="margin-top: 16px" @click="start">
        AI 查找候选人
      </el-button>
    </div>

    <div v-if="step === 2">
      <h2 class="step-title">请确认是哪位人物</h2>
      <CandidateCards :candidates="candidates" :disabled="loading" @select="onSelect" />
    </div>

    <div v-if="step === 3 && profile">
      <h2 class="step-title">人物档案预览</h2>
      <ProfileEditor
        :profile="profile"
        :character-id="characterId"
        @update:profile="onProfileUpdate"
      />
      <TimelineConfigPanel
        v-model="timelineConfig"
        :profile="profile"
        :disabled="jobRunning"
      />
      <el-button
        type="primary"
        size="large"
        :disabled="jobRunning"
        style="margin-top: 8px"
        @click="onGenerateClick"
      >
        生成人生时间轴
      </el-button>
    </div>
  </div>
</template>

