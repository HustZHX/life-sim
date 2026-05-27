<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { api, type AIModelId, type Profile, type TimelineConfig } from '@/api/client'
import { DEFAULT_TARGET_NODE_COUNT } from '@/utils/timelineDensity'
import { DEFAULT_AI_MODEL } from '@/constants/models'
import { useTimelineGenerate } from '@/composables/useTimelineGenerate'
import ProfilePanel from '@/components/ProfilePanel.vue'
import ModelSelector from '@/components/ModelSelector.vue'
import TimelineConfigPanel from '@/components/TimelineConfigPanel.vue'
import JobProgress from '@/components/JobProgress.vue'

const route = useRoute()
const router = useRouter()
const charId = route.params.id as string

const profile = ref<Profile | null>(null)
const displayName = ref('')
const pageLoading = ref(false)
const aiModel = ref<AIModelId>(DEFAULT_AI_MODEL)
const timelineConfig = ref<TimelineConfig>({ target_node_count: DEFAULT_TARGET_NODE_COUNT, start_year: 0, end_year: 0 })
const { jobRunning, progress, statusText, confirmAndGenerate } = useTimelineGenerate()

async function load() {
  pageLoading.value = true
  try {
    const [ch, prof] = await Promise.all([
      api.getCharacter(charId),
      api.getProfile(charId),
    ])
    profile.value = prof
    displayName.value = ch.display_name || prof.display_name
    timelineConfig.value = {
      target_node_count: DEFAULT_TARGET_NODE_COUNT,
      start_year: prof.birth_year,
      end_year: prof.death_year,
    }
    if (ch.status === 'timeline_ready') {
      router.replace(`/timeline/${charId}`)
    }
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '加载失败')
    router.push('/history')
  } finally {
    pageLoading.value = false
  }
}

async function onGenerateClick() {
  const ok = await confirmAndGenerate(charId, aiModel.value, {
    displayName: displayName.value,
    config: timelineConfig.value,
  })
  if (ok) router.push(`/timeline/${charId}`)
}

onMounted(load)
</script>

<template>
  <div class="page-card" v-loading="pageLoading">
    <JobProgress
      :visible="jobRunning"
      :progress="progress"
      :status-text="statusText"
      title="时间轴生成中"
    />

    <div class="head">
      <h2>{{ displayName || '继续生成' }}</h2>
      <el-button @click="router.push('/history')">返回历史</el-button>
    </div>
    <el-alert
      title="选择阶段区间与目标节点数（默认约 20 个，随寿命自动适配间隔）。可点击 AI 推荐方案或手动调整后生成。"
      type="info"
      :closable="false"
      show-icon
      style="margin-bottom: 16px"
    />
    <ProfilePanel v-if="profile" :profile="profile" />
    <ModelSelector v-if="profile" v-model="aiModel" />
    <TimelineConfigPanel
      v-if="profile"
      v-model="timelineConfig"
      :profile="profile"
      :model="aiModel"
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
</template>

<style scoped>
.head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.head h2 {
  margin: 0;
}
</style>
