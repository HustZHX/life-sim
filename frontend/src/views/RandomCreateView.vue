<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { api, type Profile, type TimelineConfig } from '@/api/client'
import { DEFAULT_TARGET_NODE_COUNT, effectiveDeathYear, isLivingProfile } from '@/utils/timelineDensity'
import { useTimelineGenerate } from '@/composables/useTimelineGenerate'
import { useModelStore } from '@/stores/model'
import { DEFAULT_RANDOM_NAME } from '@/utils/randomNames'
import ProfileEditor from '@/components/ProfileEditor.vue'
import TimelineConfigPanel from '@/components/TimelineConfigPanel.vue'
import JobProgress from '@/components/JobProgress.vue'

const router = useRouter()
const step = ref(1)
const characterId = ref('')
const profile = ref<Profile | null>(null)
const loading = ref(false)
const nameLoading = ref(false)

const displayName = ref('')
const nameCandidates = ref<string[]>([])
const background = ref('')
const introduction = ref('')
const modelStore = useModelStore()
const timelineConfig = ref<TimelineConfig>({ target_node_count: DEFAULT_TARGET_NODE_COUNT, start_year: 0, end_year: 0 })

const { jobRunning, progress, statusText, failed, errorText, retrying, retry, clearJobState, confirmAndGenerate } = useTimelineGenerate()

async function suggestNames() {
  nameLoading.value = true
  nameCandidates.value = []
  try {
    const res = await api.suggestNames({
      model: modelStore.aiModel,
      background: background.value.trim(),
      introduction: introduction.value.trim(),
    })
    nameCandidates.value = res.names
    if (res.names.length) {
      displayName.value = res.names[0]
    }
    ElMessage.success(`已生成 ${res.names.length} 个姓名备选`)
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '姓名推荐失败')
  } finally {
    nameLoading.value = false
  }
}

function pickName(name: string) {
  displayName.value = name
}

async function generateProfile() {
  loading.value = true
  try {
    const ch = await api.createCharacter('random')
    characterId.value = ch.id
    const name = displayName.value.trim() || DEFAULT_RANDOM_NAME
    profile.value = await api.generateProfile(ch.id, {
      model: modelStore.aiModel,
      display_name: name,
      background: background.value.trim(),
      introduction: introduction.value.trim(),
    })
    timelineConfig.value = {
      target_node_count: DEFAULT_TARGET_NODE_COUNT,
      start_year: isLivingProfile(profile.value) ? 0 : profile.value.birth_year,
      end_year: isLivingProfile(profile.value) ? 0 : effectiveDeathYear(profile.value),
    }
    step.value = 2
    ElMessage.success('人物档案已生成，可编辑或单独随机各字段')
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '生成失败')
  } finally {
    loading.value = false
  }
}

function onProfileUpdate(p: Profile) {
  profile.value = p
  timelineConfig.value = {
    ...timelineConfig.value,
    start_year: isLivingProfile(p) ? 0 : p.birth_year,
    end_year: isLivingProfile(p) ? 0 : effectiveDeathYear(p),
  }
}

async function onGenerateClick() {
  if (!characterId.value || !profile.value) return
  const result = await confirmAndGenerate(characterId.value, modelStore.aiModel, {
    displayName: profile.value.display_name,
    config: timelineConfig.value,
    profile: profile.value,
    background: true,
  })
  if (result.ok) {
    router.push(`/characters/${characterId.value}`)
  }
}
</script>

<template>
  <div class="page-card">
    <JobProgress
      :visible="jobRunning"
      :progress="progress"
      :status-text="statusText"
      title="时间轴生成中"
      :failed="failed"
      :error-text="errorText"
      :retrying="retrying"
      @retry="retry()"
      @dismiss="clearJobState()"
    />

    <el-steps :active="step - 1" finish-status="success" align-center style="margin-bottom: 24px">
      <el-step title="设定人物" />
      <el-step title="编辑档案" />
    </el-steps>

    <div v-if="step === 1">
      <h2 class="step-title">随机普通人</h2>
      <p class="hint">
        可先填写背景与介绍，再 AI 推荐姓名；也可留空直接随机 5 个备选。不填姓名则默认为「小明」。
      </p>

      <el-form label-width="100px" class="setup-form">
        <el-form-item label="背景设定">
          <el-input
            v-model="background"
            type="textarea"
            :rows="3"
            placeholder="可以是真实时代，如「1990 年代中国小城」；也可以是虚构世界，如「《哈利·波特》魔法界」「赛博朋克 2077」"
          />
        </el-form-item>
        <el-form-item label="一句话介绍">
          <el-input
            v-model="introduction"
            type="textarea"
            :rows="2"
            placeholder="用一句话描述你想创造的人物，如：内向但坚韧的程序员，从小在单亲家庭长大"
          />
        </el-form-item>
        <el-form-item label="姓名">
          <div class="name-block">
            <div class="name-row">
              <el-input
                v-model="displayName"
                :placeholder="DEFAULT_RANDOM_NAME"
                maxlength="20"
                show-word-limit
                style="max-width: 280px"
              />
              <el-button :loading="nameLoading" @click="suggestNames">AI 推荐姓名</el-button>
            </div>
            <div v-if="nameCandidates.length" class="name-candidates">
              <span class="candidates-label">备选（点击选用）：</span>
              <el-tag
                v-for="n in nameCandidates"
                :key="n"
                :type="displayName === n ? 'primary' : 'info'"
                class="name-tag"
                effect="plain"
                @click="pickName(n)"
              >
                {{ n }}
              </el-tag>
            </div>
          </div>
        </el-form-item>
      </el-form>

      <el-button type="primary" size="large" :loading="loading" style="margin-top: 16px" @click="generateProfile">
        生成人物档案
      </el-button>
    </div>

    <div v-if="step === 2 && profile">
      <h2 class="step-title">编辑档案 · {{ profile.display_name }}</h2>
      <p class="hint">所有字段可手动修改；点击字段右侧按钮可单独 AI 随机重生成。</p>

      <ProfileEditor
        :profile="profile"
        :character-id="characterId"
        @update:profile="onProfileUpdate"
      />

      <el-divider />

      <h3 class="sub-title">生成人生时间轴</h3>
      <el-alert
        v-if="isLivingProfile(profile)"
        title="人生未完结：仅选择推演节点数即可；经历与职业会在时间轴与后续推演中逐步填写。"
        type="info"
        :closable="false"
        show-icon
        style="margin-bottom: 12px"
      />
      <TimelineConfigPanel
        v-model="timelineConfig"
        :profile="profile"
        :disabled="jobRunning"
      />
      <el-button type="primary" size="large" :disabled="jobRunning" style="margin-top: 8px" @click="onGenerateClick">
        生成人生时间轴
      </el-button>
    </div>
  </div>
</template>

<style scoped>
.hint {
  color: #888;
  font-size: 0.9rem;
  margin: 0 0 16px;
}
.setup-form {
  max-width: 640px;
}
.name-block {
  width: 100%;
}
.name-row {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}
.name-candidates {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin-top: 10px;
}
.candidates-label {
  font-size: 0.85rem;
  color: #909399;
}
.name-tag {
  cursor: pointer;
}
.sub-title {
  margin: 0 0 12px;
  font-size: 1rem;
}
</style>
