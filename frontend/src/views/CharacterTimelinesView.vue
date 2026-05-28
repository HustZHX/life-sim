<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { api, type Character, type Profile, type Timeline } from '@/api/client'
import { formatDateTime, modeLabel, statusLabel } from '@/constants/characterLabels'
import { useLayoutStore } from '@/stores/layout'
import {
  isTimelineGenerating,
  isTimelineReady,
  timelineGenerationLabel,
  timelineJobLabel,
  versionChangeLabel,
} from '@/constants/timelineStatus'

const route = useRoute()
const router = useRouter()
const charId = route.params.id as string
const layout = useLayoutStore()

const character = ref<Character | null>(null)
const profile = ref<Profile | null>(null)
const timelines = ref<Timeline[]>([])
const loading = ref(false)
let pollTimer: ReturnType<typeof setInterval> | null = null

const activeJobCount = computed(
  () => timelines.value.filter((tl) => isTimelineGenerating(tl)).length
)

async function load(silent = false) {
  if (!silent) loading.value = true
  try {
    const [ch, prof, tlRes] = await Promise.all([
      api.getCharacter(charId),
      api.getProfile(charId).catch(() => null),
      api.listTimelines(charId).catch(() => ({ timelines: [] as Timeline[] })),
    ])
    character.value = ch
    profile.value = prof
    timelines.value = Array.isArray(tlRes.timelines) ? tlRes.timelines : []
    if (activeJobCount.value > 0) {
      startPollIfNeeded()
    } else {
      stopPoll()
    }
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '加载失败')
    if (!silent) router.push('/characters')
  } finally {
    if (!silent) loading.value = false
  }
}

function startPollIfNeeded() {
  if (pollTimer) return
  pollTimer = setInterval(() => {
    if (activeJobCount.value > 0) {
      void load(true)
    } else {
      stopPoll()
    }
  }, 3000)
}

function stopPoll() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

function openTimeline(tl: Timeline) {
  if (isTimelineGenerating(tl)) {
    ElMessage.info('该时间轴正在生成中，请稍候')
    return
  }
  if (tl.generation_status === 'failed') {
    return
  }
  if (!isTimelineReady(tl)) {
    ElMessage.warning('时间轴尚未就绪')
    return
  }
  router.push({ path: `/timeline/${charId}`, query: { timeline: tl.id } })
}

async function retryTimelineGeneration(tl: Timeline, event?: Event) {
  event?.stopPropagation()
  if (!tl.active_job?.id) {
    ElMessage.warning('无法重试，请重新创建时间轴')
    return
  }
  try {
    await api.retryJob(tl.active_job.id)
    ElMessage.success('已重新提交生成')
    await load(true)
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '重试失败')
  }
}

function createTimeline() {
  router.push(`/continue/${charId}`)
}

function editProfile() {
  router.push(`/continue/${charId}`)
}

function jobStatusText(tl: Timeline): string {
  if (tl.generation_status === 'failed') {
    return tl.generation_error || '生成失败'
  }
  const job = tl.active_job
  if (job) {
    const label = timelineJobLabel(job.type)
    if (job.status === 'pending') return `${label} · 排队中`
    const stage = job.stage_text ? `${job.stage_text} · ` : ''
    return `${label} · ${stage}${job.progress}%`
  }
  if (isTimelineGenerating(tl)) return '生成时间轴 · 进行中'
  return timelineGenerationLabel(tl.generation_status)
}

function statusTagType(tl: Timeline): 'success' | 'danger' | 'warning' | 'info' {
  if (tl.generation_status === 'failed') return 'danger'
  if (isTimelineGenerating(tl)) return 'warning'
  return 'success'
}

onMounted(() => load())

onUnmounted(stopPoll)
</script>

<template>
  <div class="page-card" v-loading="loading">
    <div class="head">
      <div>
        <el-button link class="back-link" @click="router.push('/characters')">← 人物列表</el-button>
        <h2>{{ profile?.display_name || character?.display_name || '人物' }}</h2>
        <div v-if="character" class="meta">
          <el-tag size="small" :type="character.mode === 'famous' ? 'primary' : 'success'">
            {{ modeLabel[character.mode] || character.mode }}
          </el-tag>
          <el-tag v-if="character.status" size="small" type="info">
            {{ statusLabel[character.status] || character.status }}
          </el-tag>
          <span v-if="profile?.era" class="meta-text">{{ profile.era }}</span>
          <span v-if="profile?.birth_year || profile?.death_year" class="meta-text">
            {{ profile?.birth_year || '?' }} — {{ profile?.death_year || '?' }}
          </span>
        </div>
      </div>
      <div class="head-actions">
        <el-button @click="editProfile">编辑档案</el-button>
        <el-button type="primary" @click="createTimeline">新建时间轴</el-button>
      </div>
    </div>

    <el-alert
      v-if="activeJobCount"
      :title="`${activeJobCount} 条时间轴正在后台生成，进度将自动刷新`"
      type="warning"
      :closable="false"
      show-icon
      class="active-alert"
    />

    <el-alert
      v-if="profile"
      :title="profile.personality_initial || '暂无性格摘要'"
      type="info"
      :closable="false"
      show-icon
      class="profile-alert"
    />

    <section class="timeline-section">
      <div class="section-head">
        <h3>时间轴列表（{{ timelines.length }}）</h3>
        <el-button :loading="loading" @click="load()">刷新</el-button>
      </div>

      <el-empty v-if="!timelines.length && !loading" description="该人物尚无时间轴">
        <el-button type="primary" @click="createTimeline">生成第一条时间轴</el-button>
      </el-empty>

      <div v-else-if="layout.isMobile" class="mobile-cards">
        <button
          v-for="row in timelines"
          :key="row.id"
          type="button"
          class="mobile-card"
          @click="openTimeline(row)"
        >
          <div class="card-head">
            <div class="card-title">{{ row.title || '未命名时间轴' }}</div>
            <el-tag size="small" :type="statusTagType(row)" effect="plain">
              {{ timelineGenerationLabel(row.generation_status) }}
            </el-tag>
          </div>

          <div v-if="isTimelineGenerating(row)" class="job-cell card-job">
            <el-progress
              :percentage="Math.max(row.active_job?.progress ?? 8, 5)"
              :stroke-width="6"
              striped
              striped-flow
            />
            <span class="job-status">{{ jobStatusText(row) }}</span>
          </div>
          <p v-if="row.generation_status === 'failed' && row.generation_error" class="fail-hint">
            {{ row.generation_error }}
          </p>

          <div class="card-meta">
            <span class="meta-item">节点 {{ row.node_count ?? '—' }}</span>
            <span class="meta-item">版本 {{ versionChangeLabel(row.version_count) }}</span>
          </div>

          <div class="card-foot">
            <span class="time-text">{{ formatDateTime(row.updated_at) }}</span>
            <div class="card-actions">
              <el-button
                v-if="row.generation_status === 'failed'"
                type="warning"
                link
                @click.stop="retryTimelineGeneration(row, $event)"
              >
                重试
              </el-button>
              <el-button type="primary" link :disabled="!isTimelineReady(row)" @click.stop="openTimeline(row)">
                查看
              </el-button>
            </div>
          </div>
        </button>
      </div>

      <el-table v-else :data="timelines" stripe style="width: 100%" @row-click="openTimeline">
        <el-table-column label="名称" min-width="200" prop="title" />
        <el-table-column label="节点数" width="80" align="center">
          <template #default="{ row }">
            {{ row.node_count ?? '—' }}
          </template>
        </el-table-column>
        <el-table-column label="版本数" width="130" align="center">
          <template #default="{ row }">
            {{ versionChangeLabel(row.version_count) }}
          </template>
        </el-table-column>
        <el-table-column label="生成状态" min-width="220">
          <template #default="{ row }">
            <div v-if="isTimelineGenerating(row)" class="job-cell">
              <el-progress
                :percentage="Math.max(row.active_job?.progress ?? 8, 5)"
                :stroke-width="6"
                striped
                striped-flow
              />
              <span class="job-status">{{ jobStatusText(row) }}</span>
            </div>
            <el-tag v-else size="small" :type="statusTagType(row)" effect="plain">
              {{ timelineGenerationLabel(row.generation_status) }}
            </el-tag>
            <p v-if="row.generation_status === 'failed' && row.generation_error" class="fail-hint">
              {{ row.generation_error }}
            </p>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="更新时间" width="160">
          <template #default="{ row }">
            {{ formatDateTime(row.updated_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.generation_status === 'failed'"
              type="warning"
              link
              @click.stop="retryTimelineGeneration(row, $event)"
            >
              重试
            </el-button>
            <el-button type="primary" link :disabled="!isTimelineReady(row)" @click.stop="openTimeline(row)">
              查看
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </section>
  </div>
</template>

<style scoped>
.head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 20px;
}
.back-link {
  padding: 0;
  margin-bottom: 4px;
  font-size: 0.85rem;
}
.head h2 {
  margin: 0;
}
.meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
}
.meta-text {
  font-size: 0.85rem;
  color: #606266;
}
.head-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}
.active-alert {
  margin-bottom: 12px;
}
.profile-alert {
  margin-bottom: 20px;
}
.timeline-section h3 {
  margin: 0;
  font-size: 1rem;
}
.section-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}
.job-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.job-status {
  font-size: 0.78rem;
  color: #606266;
  line-height: 1.3;
}
.fail-hint {
  margin: 4px 0 0;
  font-size: 0.75rem;
  color: #f56c6c;
  line-height: 1.3;
}
:deep(.el-table__row) {
  cursor: pointer;
}

.mobile-cards {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.mobile-card {
  width: 100%;
  text-align: left;
  border: 1px solid #ebeef5;
  border-radius: 12px;
  background: #fff;
  padding: 14px 14px 12px;
  box-shadow: 0 1px 6px rgba(0, 0, 0, 0.04);
}
.card-head {
  display: flex;
  gap: 10px;
  align-items: center;
  justify-content: space-between;
}
.card-title {
  font-size: 1.02rem;
  font-weight: 700;
  color: #16213e;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.card-meta {
  margin-top: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  color: #606266;
  font-size: 0.9rem;
}
.meta-item {
  padding: 2px 8px;
  border-radius: 999px;
  background: #f5f7fa;
}
.card-job {
  margin-top: 10px;
}
.card-foot {
  margin-top: 10px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}
.time-text {
  color: #909399;
  font-size: 0.85rem;
  white-space: nowrap;
}
.card-actions {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-shrink: 0;
}
</style>
