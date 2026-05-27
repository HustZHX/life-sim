<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { api, type Profile, type TimelineConfig } from '@/api/client'
import { DEFAULT_TARGET_NODE_COUNT } from '@/utils/timelineDensity'
import { DEFAULT_AI_MODEL, type AIModelId } from '@/constants/models'
import { useTimelineGenerate } from '@/composables/useTimelineGenerate'
import ProfilePanel from '@/components/ProfilePanel.vue'
import ModelSelector from '@/components/ModelSelector.vue'
import TimelineConfigPanel from '@/components/TimelineConfigPanel.vue'
import JobProgress from '@/components/JobProgress.vue'

const router = useRouter()
const characterId = ref('')
const profile = ref<Profile | null>(null)
const loading = ref(false)
const aiModel = ref<AIModelId>(DEFAULT_AI_MODEL)
const timelineConfig = ref<TimelineConfig>({ target_node_count: DEFAULT_TARGET_NODE_COUNT, start_year: 0, end_year: 0 })
const useTemplate = ref(false)
const famousQuery = ref('')
const { jobRunning, progress, statusText, confirmAndGenerate } = useTimelineGenerate()

async function generateProfile() {
  loading.value = true
  try {
    const ch = await api.createCharacter('random')
    characterId.value = ch.id
    profile.value = await api.generateProfile(ch.id)
    if (useTemplate.value && famousQuery.value.trim()) {
      profile.value = await api.importTemplate(ch.id, famousQuery.value.trim())
    }
    timelineConfig.value = {
      target_node_count: DEFAULT_TARGET_NODE_COUNT,
      start_year: profile.value.birth_year,
      end_year: profile.value.death_year,
    }
    ElMessage.success('随机人物档案已生成')
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '生成失败')
  } finally {
    loading.value = false
  }
}

async function onGenerateClick() {
  if (!characterId.value) return
  const ok = await confirmAndGenerate(characterId.value, aiModel.value, {
    displayName: profile.value?.display_name,
    config: timelineConfig.value,
  })
  if (ok) router.push(`/timeline/${characterId.value}`)
}
</script>

<template>
  <div class="page-card">
    <JobProgress
      :visible="jobRunning"
      :progress="progress"
      :status-text="statusText"
      title="时间轴生成中"
    />

    <h2 class="step-title">随机普通人</h2>
    <p>AI 将随机生成姓名、时代与人生背景。你也可以导入某位名人的性格与时代气质。</p>

    <el-checkbox v-model="useTemplate" style="margin: 16px 0">沿用名人的性格与时代背景</el-checkbox>
    <el-input
      v-if="useTemplate"
      v-model="famousQuery"
      placeholder="输入名人姓名，如：苏轼"
      style="margin-bottom: 16px; max-width: 400px"
    />

    <div class="actions">
      <el-button type="primary" :loading="loading && !profile" @click="generateProfile">
        随机生成档案
      </el-button>
      <el-button v-if="profile" type="success" :disabled="jobRunning" @click="onGenerateClick">
        生成人生时间轴
      </el-button>
    </div>

    <ProfilePanel v-if="profile" :profile="profile" style="margin-top: 24px" />
    <ModelSelector v-if="profile" v-model="aiModel" />
    <TimelineConfigPanel
      v-if="profile"
      v-model="timelineConfig"
      :profile="profile"
      :model="aiModel"
      :disabled="jobRunning"
    />
  </div>
</template>

<style scoped>
.actions {
  display: flex;
  gap: 12px;
}
</style>
