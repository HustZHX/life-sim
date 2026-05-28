<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { RefreshRight } from '@element-plus/icons-vue'
import { api, type Profile } from '@/api/client'
import type { AIModelId } from '@/constants/models'
import { DEFAULT_AI_MODEL } from '@/constants/models'
import { PROFILE_FIELDS, cloneProfile, type ProfileFieldKey } from '@/constants/profileFields'
import { formatProfileLifeSpan, isLivingProfile } from '@/utils/timelineDensity'

const props = withDefaults(
  defineProps<{
    profile: Profile
    characterId: string
    model?: AIModelId
    editable?: boolean
  }>(),
  { editable: true, model: DEFAULT_AI_MODEL }
)

const emit = defineEmits<{
  'update:profile': [profile: Profile]
}>()

const local = ref<Profile>(cloneProfile(props.profile))
const saving = ref(false)
const randomizingField = ref<string | null>(null)

watch(
  () => props.profile,
  (p) => {
    local.value = cloneProfile(p)
  },
  { deep: true }
)

function fieldValue(key: ProfileFieldKey): string | number {
  return local.value[key] as string | number
}

function setField(key: ProfileFieldKey, val: string | number | undefined) {
  if (key === 'birth_year' || key === 'death_year') {
    local.value[key] = typeof val === 'number' ? val : Number(val) || 0
    if (key === 'death_year' && isLivingProfile(local.value)) {
      local.value.experiences = ''
      local.value.occupation = ''
      local.value.death_cause = ''
    }
  } else {
    ;(local.value as unknown as Record<string, string>)[key] = String(val ?? '')
  }
}

function deathYearDisplay(): number | undefined {
  return isLivingProfile(local.value) ? undefined : local.value.death_year || undefined
}

function onDeathYearChange(v: number | undefined) {
  if (v == null || v <= 0) {
    local.value.death_year = 0
    local.value.experiences = ''
    local.value.occupation = ''
    local.value.death_cause = ''
    return
  }
  local.value.death_year = v
}

function markLivingProfile() {
  local.value.death_year = 0
  local.value.experiences = ''
  local.value.occupation = ''
  local.value.death_cause = ''
}

async function saveProfile() {
  if (!props.editable) return
  saving.value = true
  try {
    const updated = await api.updateProfile(props.characterId, local.value)
    local.value = cloneProfile(updated)
    emit('update:profile', updated)
    ElMessage.success('档案已保存')
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '保存失败')
  } finally {
    saving.value = false
  }
}

async function randomizeField(key: ProfileFieldKey) {
  if (!props.editable) return
  randomizingField.value = key
  try {
    const updated = await api.randomizeProfileField(props.characterId, key, props.model)
    local.value = cloneProfile(updated)
    emit('update:profile', updated)
    ElMessage.success('已重新随机')
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '随机失败')
  } finally {
    randomizingField.value = null
  }
}
</script>

<template>
  <div class="profile-editor">
    <div v-if="editable" class="editor-toolbar">
      <span class="life-span">生卒：{{ formatProfileLifeSpan(local) }}</span>
      <el-button type="primary" :loading="saving" @click="saveProfile">保存档案修改</el-button>
    </div>

    <el-descriptions v-else :title="local.display_name" :column="2" border>
      <el-descriptions-item label="生卒">{{ formatProfileLifeSpan(local) }}</el-descriptions-item>
      <el-descriptions-item v-for="f in PROFILE_FIELDS" :key="f.key" :label="f.label" :span="f.span ?? 1">
        {{ fieldValue(f.key) || '—' }}
      </el-descriptions-item>
    </el-descriptions>

    <el-form v-if="editable" label-width="96px" class="field-form">
      <el-form-item label="姓名">
        <div class="field-row">
          <el-input
            :model-value="String(local.display_name ?? '')"
            @update:model-value="(v: string) => setField('display_name', v)"
          />
          <el-button
            :icon="RefreshRight"
            :loading="randomizingField === 'display_name'"
            title="AI 随机此字段"
            @click="randomizeField('display_name')"
          />
        </div>
      </el-form-item>
      <el-form-item label="出生年">
        <div class="field-row">
          <el-input-number
            :model-value="local.birth_year || undefined"
            :controls="true"
            style="width: 160px"
            @update:model-value="(v: number | undefined) => setField('birth_year', v ?? 0)"
          />
        </div>
      </el-form-item>
      <el-form-item label="卒年">
        <div class="death-year-block">
          <div class="field-row">
            <el-input-number
              :model-value="deathYearDisplay()"
              :controls="true"
              :disabled="isLivingProfile(local)"
              placeholder="未完结"
              style="width: 160px"
              @update:model-value="onDeathYearChange"
            />
            <el-button v-if="isLivingProfile(local)" plain @click="local.death_year = local.birth_year + 60">
              填写卒年
            </el-button>
            <el-button v-else plain @click="markLivingProfile">人生未完结</el-button>
          </div>
          <p class="field-hint">留空表示人生未完结；此时经历、职业留待剧情推演填写。</p>
        </div>
      </el-form-item>
      <el-form-item v-for="f in PROFILE_FIELDS.filter((x) => x.key !== 'birth_year' && x.key !== 'death_year' && x.key !== 'display_name')" :key="f.key" :label="f.label">
        <div class="field-row">
          <el-input-number
            v-if="f.type === 'number'"
            :model-value="fieldValue(f.key) as number"
            :controls="true"
            style="width: 160px"
            @update:model-value="(v: number | undefined) => setField(f.key, v ?? 0)"
          />
          <el-input
            v-else-if="f.type === 'text'"
            :model-value="String(fieldValue(f.key) ?? '')"
            @update:model-value="(v: string) => setField(f.key, v)"
          />
          <el-input
            v-else
            type="textarea"
            :rows="3"
            :model-value="String(fieldValue(f.key) ?? '')"
            @update:model-value="(v: string) => setField(f.key, v)"
          />
          <el-button
            v-if="f.randomizable !== false"
            :icon="RefreshRight"
            :loading="randomizingField === f.key"
            title="AI 随机此字段"
            @click="randomizeField(f.key)"
          />
        </div>
      </el-form-item>
    </el-form>
  </div>
</template>

<style scoped>
.profile-editor {
  margin-top: 8px;
}
.editor-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  gap: 12px;
  flex-wrap: wrap;
}
.life-span {
  font-size: 0.9rem;
  color: #606266;
}
.field-form {
  max-width: 900px;
}
.field-row {
  display: flex;
  gap: 8px;
  align-items: flex-start;
  width: 100%;
}
.field-row :deep(.el-input),
.field-row :deep(.el-textarea) {
  flex: 1;
}
.death-year-block {
  width: 100%;
}
.field-hint {
  margin: 6px 0 0;
  font-size: 0.82rem;
  color: #909399;
  line-height: 1.45;
}
</style>
