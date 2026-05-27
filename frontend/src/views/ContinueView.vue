<script setup lang="ts">

import { onMounted, ref } from 'vue'

import { useRoute, useRouter } from 'vue-router'

import { ElMessage } from 'element-plus'

import { api, type AIModelId, type Profile, type TimelineConfig } from '@/api/client'

import {

  DEFAULT_TARGET_NODE_COUNT,

  effectiveDeathYear,

  isLivingProfile,

} from '@/utils/timelineDensity'

import { DEFAULT_AI_MODEL } from '@/constants/models'

import { useTimelineGenerate } from '@/composables/useTimelineGenerate'

import ProfileEditor from '@/components/ProfileEditor.vue'

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
const profileModel = ref<AIModelId>(DEFAULT_AI_MODEL)

const timelineTitle = ref('')

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

      end_year: effectiveDeathYear(prof),

    }

  } catch (e: unknown) {

    ElMessage.error(e instanceof Error ? e.message : '加载失败')

    router.push('/characters')

  } finally {

    pageLoading.value = false

  }

}



function onProfileUpdate(p: Profile) {
  profile.value = p
  timelineConfig.value.end_year = effectiveDeathYear(p)
}

async function onGenerateClick() {
  const result = await confirmAndGenerate(charId, aiModel.value, {
    displayName: displayName.value,
    config: {
      ...timelineConfig.value,
      title: timelineTitle.value.trim() || undefined,
    },
    background: true,
  })
  if (result.ok) {
    router.push(`/characters/${charId}`)
  } else {
    await load()
  }
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

      <el-button @click="router.push(`/characters/${charId}`)">返回时间轴列表</el-button>

    </div>

    <h3>新建时间轴</h3>

    <el-alert

      v-if="profile && isLivingProfile(profile)"

      title="该人物档案未标注卒年（仍在世），时间轴默认生成至当前年份，可在下方调整结束年份。"

      type="warning"

      :closable="false"

      show-icon

      style="margin-bottom: 16px"

    />

    <el-alert

      title="每次生成都会创建一条独立时间轴。选择阶段区间与目标节点数，可点击 AI 推荐方案或手动调整后生成。"

      type="info"

      :closable="false"

      show-icon

      style="margin-bottom: 16px"

    />

    <p v-if="profile" class="profile-edit-hint">档案字段可编辑；随机单字段时使用下方档案 AI 模型。</p>
    <ModelSelector v-if="profile" v-model="profileModel" label="档案 AI 模型" />
    <ProfileEditor
      v-if="profile"
      :profile="profile"
      :character-id="charId"
      :model="profileModel"
      @update:profile="onProfileUpdate"
    />

    <el-form v-if="profile" label-width="100px" style="margin-top: 12px">

      <el-form-item label="时间轴名称">

        <el-input

          v-model="timelineTitle"

          placeholder="可选，如「青年期」「架空人生 A」"

          maxlength="60"

          show-word-limit

          :disabled="jobRunning"

        />

      </el-form-item>

    </el-form>

    <ModelSelector v-if="profile" v-model="aiModel" label="时间轴 AI 模型" />

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

      生成时间轴

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

h3 {

  margin: 0 0 12px;

  font-size: 1rem;

}

.profile-edit-hint {
  margin: 0 0 8px;
  font-size: 0.85rem;
  color: #909399;
}

</style>

