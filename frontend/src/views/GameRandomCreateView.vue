<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { api, type Profile } from '@/api/client'
import { DEFAULT_AI_MODEL, type AIModelId } from '@/constants/models'
import { useAwaitProgress } from '@/composables/useAwaitProgress'
import { useGameCreate } from '@/composables/useGameCreate'
import ProfilePanel from '@/components/ProfilePanel.vue'
import ModelSelector from '@/components/ModelSelector.vue'
import JobProgress from '@/components/JobProgress.vue'

const router = useRouter()
const step = ref(1)
const characterId = ref('')
const displayName = ref('')
const birthBackground = ref('')
const profile = ref<Profile | null>(null)
const loading = ref(false)
const profileModel = ref<AIModelId>(DEFAULT_AI_MODEL)
const startModel = ref<AIModelId>(DEFAULT_AI_MODEL)

const profileJob = useAwaitProgress('生成出生设定')
const { jobRunner, starting, startGameTimeline } = useGameCreate()

async function start() {
  if (!birthBackground.value.trim()) {
    ElMessage.warning('请描述出生背景')
    return
  }
  loading.value = true
  try {
    const ch = await api.createGameCharacter('random')
    characterId.value = ch.id
    await api.updateGameConfig(ch.id, { birth_background: birthBackground.value.trim() })
    const p = await profileJob.run(
      () =>
        api.gameGenerateProfile(ch.id, {
          model: profileModel.value,
          display_name: displayName.value.trim() || undefined,
          birth_background: birthBackground.value.trim(),
        }),
      { title: '生成出生设定', startText: 'AI 正在构建时代与家庭…' }
    )
    profile.value = p
    step.value = 2
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '创建失败')
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
      :visible="profileJob.visible.value || jobRunner.visible.value"
      :progress="jobRunner.visible.value ? jobRunner.progress.value : profileJob.progress.value"
      :status-text="jobRunner.visible.value ? jobRunner.statusText.value : profileJob.statusText.value"
      :title="jobRunner.visible.value ? jobRunner.title.value : profileJob.title.value"
      :failed="jobRunner.failed.value"
      :error-text="jobRunner.errorText.value"
      :retrying="jobRunner.retrying.value"
      @retry="jobRunner.retry()"
      @dismiss="jobRunner.clearState()"
    />

    <el-steps :active="step - 1" finish-status="success" align-center style="margin-bottom: 24px">
      <el-step title="出生背景" />
      <el-step title="开始游戏" />
    </el-steps>

    <div v-if="step === 1">
      <h2 class="step-title">设定出生背景</h2>
      <el-input v-model="displayName" placeholder="姓名（可选，默认小明）" size="large" style="margin-bottom: 12px" />
      <el-input
        v-model="birthBackground"
        type="textarea"
        :rows="4"
        placeholder="时代、地区、家庭阶级与父母情况等，例如：1990 年代江南小城，父亲工人母亲教师…"
        :disabled="loading"
      />
      <ModelSelector v-model="profileModel" style="margin-top: 16px" />
      <el-button type="primary" size="large" :loading="loading" style="margin-top: 16px" @click="start">
        生成出生设定
      </el-button>
    </div>

    <div v-if="step === 2 && profile">
      <h2 class="step-title">出生设定预览</h2>
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
</style>
