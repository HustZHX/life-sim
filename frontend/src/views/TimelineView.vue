<script setup lang="ts">
import { nextTick, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { api, pollJob } from '@/api/client'
import type { AIModelId, LifeNode, NodeFieldChange, TimelineVersion } from '@/api/client'
import { modelDisplayLabel } from '@/constants/models'
import TimelineAxis from '@/components/TimelineAxis.vue'
import NodeEditor from '@/components/NodeEditor.vue'
import ChangeDiffPanel from '@/components/ChangeDiffPanel.vue'
import VersionHistory from '@/components/VersionHistory.vue'
import ProfilePanel from '@/components/ProfilePanel.vue'
import type { Profile } from '@/api/client'

const route = useRoute()
const router = useRouter()
const charId = route.params.id as string

const nodes = ref<LifeNode[]>([])
const profile = ref<Profile | null>(null)
const selectedNode = ref<LifeNode | null>(null)
const versions = ref<TimelineVersion[]>([])
const currentVersionId = ref('')
const changes = ref<NodeFieldChange[]>([])
const diffLoading = ref(false)
const diffVersionId = ref('')
const diffPanelRef = ref<HTMLElement | null>(null)
const pageLoading = ref(false)
const showProfile = ref(false)

const regenVisible = ref(false)
const regenProgress = ref(0)
const regenStatus = ref('')

async function load() {
  pageLoading.value = true
  try {
    const [tl, prof, vers] = await Promise.all([
      api.getTimeline(charId),
      api.getProfile(charId).catch(() => null),
      api.listVersions(charId),
    ])
    nodes.value = tl.nodes
    currentVersionId.value = tl.version.id
    profile.value = prof
    versions.value = vers.versions
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    pageLoading.value = false
  }
}

function onSelect(node: LifeNode) {
  selectedNode.value = node
}

async function onSave(payload: {
  mode: string
  model: AIModelId
  patch: Record<string, string>
  target_node_count?: number
}) {
  if (!selectedNode.value) return
  regenVisible.value = true
  regenProgress.value = 5
  regenStatus.value = '正在提交重算任务…'
  changes.value = []
  try {
    const job = await api.patchNode(charId, selectedNode.value.id, {
      title: payload.patch.title,
      events: payload.patch.events,
      thoughts: payload.patch.thoughts,
      personality_snapshot: payload.patch.personality_snapshot,
      mode: payload.mode as 'full_cascade' | 'inner_current' | 'inner_subsequent',
      model: payload.model,
      target_node_count: payload.target_node_count,
    })
    const done = await pollJob(job.id, (j) => {
      regenProgress.value = Math.max(j.progress, 5)
      regenStatus.value =
        j.status === 'running'
          ? `重算中… ${j.progress}%（${modelDisplayLabel(j.model || payload.model)}）`
          : regenStatus.value
    })
    if (done.status === 'failed') throw new Error(done.error || '重算失败')
    regenProgress.value = 100
    regenStatus.value = '重算完成'
    if (done.result) {
      try {
        const diff = JSON.parse(done.result)
        changes.value = diff.changes ?? []
      } catch {
        /* ignore */
      }
    }
    await load()
    const updated = nodes.value.find((n: LifeNode) => n.sequence === selectedNode.value?.sequence)
    if (updated) selectedNode.value = updated
    const msg =
      payload.mode === 'full_cascade'
        ? '已重算寿命并全新生成后续时间轴'
        : '想法与性格已根据经历更新'
    ElMessage.success(msg)
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '保存失败')
  } finally {
    setTimeout(() => {
      regenVisible.value = false
      regenProgress.value = 0
      regenStatus.value = ''
    }, 1500)
  }
}

async function onDiff(v: TimelineVersion) {
  diffLoading.value = true
  diffVersionId.value = v.id
  changes.value = []
  try {
    const diff = await api.getVersionDiff(charId, v.id)
    changes.value = diff.changes ?? []
    await nextTick()
    diffPanelRef.value?.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
    if (!v.parent_version_id) {
      ElMessage.info('这是首个版本，没有上一版可对比')
    } else if (!changes.value.length) {
      ElMessage.info('该版本与上一版相比无字段变更')
    } else {
      ElMessage.success(`已加载 ${changes.value.length} 处变更`)
    }
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '获取 diff 失败')
  } finally {
    diffLoading.value = false
  }
}

async function onRollback(v: TimelineVersion) {
  pageLoading.value = true
  try {
    await api.rollback(charId, v.id)
    await load()
    changes.value = []
    diffVersionId.value = ''
    ElMessage.success('已回溯到选定版本')
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '回溯失败')
  } finally {
    pageLoading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="timeline-page">
    <div class="toolbar">
      <div class="toolbar-left">
        <h2>{{ profile?.display_name || '人生时间轴' }}</h2>
        <div class="name-legend">
          <span class="legend-item"><mark class="name-protagonist">主角</mark></span>
          <span class="legend-item"><mark class="name-person">他人</mark></span>
          <span class="legend-item"><mark class="name-place">地名</mark></span>
        </div>
      </div>
      <div class="toolbar-actions">
        <el-button @click="router.push('/history')">生成历史</el-button>
        <el-button @click="showProfile = !showProfile">
          {{ showProfile ? '隐藏' : '查看' }}档案
        </el-button>
      </div>
    </div>

    <el-collapse-transition>
      <div v-if="showProfile && profile" class="page-card profile-block">
        <ProfilePanel :profile="profile" />
      </div>
    </el-collapse-transition>

    <div class="timeline-layout">
      <section class="col-timeline">
        <div v-loading="pageLoading" class="page-card timeline-scroll">
          <TimelineAxis
            :nodes="nodes"
            :selected-id="selectedNode?.id"
            :profile="profile"
            @select="onSelect"
          />
        </div>
      </section>

      <aside class="col-dock col-dock-editor">
        <div class="dock-panel page-card">
          <NodeEditor
            :node="selectedNode"
            :profile="profile"
            :loading="regenVisible"
            :regen-visible="regenVisible"
            :regen-progress="regenProgress"
            :regen-status="regenStatus"
            @save="onSave"
          />
        </div>
      </aside>

      <aside class="col-dock col-dock-side">
        <div class="dock-panel page-card">
          <VersionHistory
            :versions="versions"
            :current-version-id="currentVersionId"
            :loading="pageLoading"
            :diff-loading="diffLoading"
            :diff-version-id="diffVersionId"
            @diff="onDiff"
            @rollback="onRollback"
          />
          <el-divider />
          <div ref="diffPanelRef">
            <ChangeDiffPanel :changes="changes" :loading="diffLoading" />
          </div>
        </div>
      </aside>
    </div>
  </div>
</template>

<style scoped>
/* 时间轴页需要更宽的三栏布局 */
.timeline-page {
  padding-bottom: 24px;
  max-width: 1500px;
  margin: 0 auto;
}
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.toolbar h2 {
  margin: 0;
}
.toolbar-left {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.name-legend {
  display: flex;
  gap: 12px;
  font-size: 0.8rem;
}
.name-legend mark {
  background: none;
  padding: 0 4px;
  font-size: 0.8rem;
}
.name-legend .name-protagonist {
  font-weight: 700;
  color: #b45309;
  background: linear-gradient(180deg, rgba(251, 191, 36, 0.35) 0%, rgba(251, 191, 36, 0.12) 100%);
  border-bottom: 2px solid #f59e0b;
  border-radius: 3px;
}
.name-legend .name-person {
  font-weight: 600;
  color: #1d4ed8;
  background: rgba(59, 130, 246, 0.12);
  border-radius: 3px;
}
.name-legend .name-place {
  font-weight: 600;
  color: #047857;
  background: rgba(16, 185, 129, 0.14);
  border-bottom: 1px dashed #10b981;
  border-radius: 3px;
}
.legend-item {
  color: #888;
}
.toolbar-actions {
  display: flex;
  gap: 8px;
}
.profile-block {
  margin-bottom: 16px;
}

/*
 * 关键：列拉伸到与左侧时间轴同高，sticky 子元素才能在页面滚动时“钉”在视口
 */
.timeline-layout {
  display: grid;
  grid-template-columns: minmax(280px, 1fr) 400px 320px;
  gap: 16px;
  align-items: stretch;
}

.col-timeline {
  min-width: 0;
}

.col-dock {
  min-width: 0;
  /* grid align-items:stretch 使本列与时间轴列同高，sticky 才能生效 */
}

.dock-panel {
  position: -webkit-sticky;
  position: sticky;
  top: 16px;
  max-height: calc(100vh - 32px);
  overflow-y: auto;
  overflow-x: hidden;
  z-index: 5;
}

@media (max-width: 1280px) {
  .timeline-layout {
    grid-template-columns: minmax(0, 1fr) 360px;
  }
  .col-dock-side {
    grid-column: 1 / -1;
  }
  .col-dock-side .dock-panel {
    max-height: 50vh;
  }
}

@media (max-width: 900px) {
  .timeline-layout {
    grid-template-columns: 1fr;
  }
  .col-dock .dock-panel {
    position: static;
    max-height: none;
  }
}
</style>
