<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { api, pollJob } from '@/api/client'
import type { AIModelId, BranchNode, BranchOverviewEntry, LifeNode, Timeline } from '@/api/client'
import { DEFAULT_AI_MODEL, modelDisplayLabel } from '@/constants/models'
import TimelineAxis from '@/components/TimelineAxis.vue'
import NodeEditor from '@/components/NodeEditor.vue'
import BranchFlowchart from '@/components/BranchFlowchart.vue'
import BranchOverview from '@/components/BranchOverview.vue'
import ProfilePanel from '@/components/ProfilePanel.vue'
import NarrativeDialog from '@/components/NarrativeDialog.vue'
import ModelSelector from '@/components/ModelSelector.vue'
import type { Profile } from '@/api/client'
import { copyTextToClipboard, formatTimelineExport } from '@/utils/exportTimelineText'
import { useNarrativeGenerate } from '@/composables/useNarrativeGenerate'
import { DEFAULT_TARGET_NODE_COUNT } from '@/utils/timelineDensity'

const route = useRoute()
const router = useRouter()
const charId = route.params.id as string

const nodes = ref<LifeNode[]>([])
const profile = ref<Profile | null>(null)
const timelines = ref<Timeline[]>([])
const currentTimeline = ref<Timeline | null>(null)
const selectedTimelineId = ref('')
const selectedNode = ref<LifeNode | null>(null)
const currentVersionId = ref('')
const branchRoots = ref<BranchNode[]>([])
const branchLoading = ref(false)
const activatingBranchId = ref('')
const pageLoading = ref(false)
const showProfile = ref(false)
const detailPanelOpen = ref(false)
const viewMode = ref<'timeline' | 'overview'>('timeline')
const overviewBranches = ref<BranchOverviewEntry[]>([])
const overviewLoading = ref(false)

const regenVisible = ref(false)
const regenProgress = ref(0)
const regenStatus = ref('')

const exportVisible = ref(false)
const exportText = ref('')
const exportCopying = ref(false)

const narrative = useNarrativeGenerate()
const {
  visible: narrativeVisible,
  title: narrativeTitle,
  content: narrativeContent,
  running: narrativeRunning,
  progress: narrativeProgress,
  statusText: narrativeStatusText,
  close: closeNarrative,
  regenerate: regenerateNarrative,
  generateLightNovel,
  openNodeNarrative,
} = narrative
const lightNovelPickerVisible = ref(false)
const lightNovelModel = ref<AIModelId>(DEFAULT_AI_MODEL)
const lightNovelFrom = ref(0)
const lightNovelTo = ref(0)

const narrativeChangeVisible = ref(false)
const narrativeChangeInstruction = ref('')
const narrativeChangeModel = ref<AIModelId>(DEFAULT_AI_MODEL)
const narrativeChangeTargetNodes = ref(DEFAULT_TARGET_NODE_COUNT)

const nodeSequenceOptions = computed(() =>
  nodes.value.map((n) => ({
    value: n.sequence,
    label: `${n.sequence}. ${n.year} 年 · ${n.title}`,
  }))
)

function openLightNovelPicker() {
  if (!nodes.value.length || !currentVersionId.value) {
    ElMessage.warning('暂无节点可生成轻小说')
    return
  }
  const seqs = nodes.value.map((n) => n.sequence)
  lightNovelFrom.value = Math.min(...seqs)
  lightNovelTo.value = Math.max(...seqs)
  lightNovelPickerVisible.value = true
}

async function confirmLightNovel() {
  if (lightNovelFrom.value > lightNovelTo.value) {
    ElMessage.warning('起始节点不能晚于结束节点')
    return
  }
  lightNovelPickerVisible.value = false
  await generateLightNovel({
    type: 'light_novel',
    characterId: charId,
    versionId: currentVersionId.value,
    fromSequence: lightNovelFrom.value,
    toSequence: lightNovelTo.value,
    model: lightNovelModel.value,
  })
}

function onNarrativeDialogClose(visible: boolean) {
  if (!visible) closeNarrative()
}

function openNarrativeChange() {
  if (!selectedTimelineId.value || !nodes.value.length) {
    ElMessage.warning('请先加载时间轴')
    return
  }
  narrativeChangeInstruction.value = ''
  narrativeChangeVisible.value = true
}

async function confirmNarrativeChange() {
  const instruction = narrativeChangeInstruction.value.trim()
  if (!instruction) {
    ElMessage.warning('请填写叙述变更内容')
    return
  }
  if (!selectedTimelineId.value) return

  narrativeChangeVisible.value = false
  regenVisible.value = true
  regenProgress.value = 5
  regenStatus.value = '正在提交叙述变更…'
  try {
    const job = await api.applyNarrativeChange(charId, {
      timeline_id: selectedTimelineId.value,
      instruction,
      model: narrativeChangeModel.value,
      target_node_count: narrativeChangeTargetNodes.value,
    })
    const done = await pollJob(job.id, (j) => {
      regenProgress.value = Math.max(j.progress, 5)
      if (j.status === 'running') {
        const stage = j.stage_text ? `${j.stage_text} · ` : '处理中… '
        regenStatus.value = `${stage}${j.progress}%（${modelDisplayLabel(j.model || narrativeChangeModel.value)}）`
      }
    })
    if (done.status === 'failed') throw new Error(done.error || '叙述变更失败')
    regenProgress.value = 100
    regenStatus.value = '变更完成'
    await load()
    ElMessage.success('已根据叙述创建新分支并推演后续节点')
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '叙述变更失败')
  } finally {
    setTimeout(() => {
      regenVisible.value = false
      regenProgress.value = 0
      regenStatus.value = ''
    }, 1500)
  }
}

function openExport() {
  if (!nodes.value.length) {
    ElMessage.warning('暂无节点可导出')
    return
  }
  exportText.value = formatTimelineExport(profile.value, nodes.value)
  exportVisible.value = true
}

async function copyExport() {
  if (!exportText.value) return
  exportCopying.value = true
  try {
    await copyTextToClipboard(exportText.value)
    ElMessage.success('已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败，请手动全选复制')
  } finally {
    exportCopying.value = false
  }
}

async function load() {
  pageLoading.value = true
  try {
    const tlRes = await api.listTimelines(charId).catch(() => ({ timelines: [] as Timeline[] }))
    timelines.value = tlRes.timelines
    if (!tlRes.timelines.length) {
      router.replace(`/characters/${charId}`)
      return
    }

    let timelineId = (route.query.timeline as string) || selectedTimelineId.value
    if (!timelineId && timelines.value.length) {
      timelineId = timelines.value[0].id
    }

    const [tl, prof, branches] = await Promise.all([
      api.getTimeline(charId, { timelineId }),
      api.getProfile(charId).catch(() => null),
      timelineId
        ? api.listBranches(charId, timelineId).catch(() => ({
            timeline_id: timelineId,
            active_version_id: '',
            roots: [] as BranchNode[],
          }))
        : Promise.resolve({ timeline_id: '', active_version_id: '', roots: [] as BranchNode[] }),
    ])
    nodes.value = tl.nodes
    currentTimeline.value = tl.timeline
    selectedTimelineId.value = tl.timeline.id
    currentVersionId.value = tl.version.id
    profile.value = prof
    branchRoots.value = branches.roots
    selectedNode.value = null
    detailPanelOpen.value = false
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    pageLoading.value = false
  }
}

function onTimelineChange(timelineId: string) {
  router.replace({ path: `/timeline/${charId}`, query: { timeline: timelineId } })
}

watch(
  () => route.query.timeline,
  (tid) => {
    const next = typeof tid === 'string' ? tid : ''
    if (next && next !== selectedTimelineId.value) {
      selectedTimelineId.value = next
      load()
    }
  }
)

function onSelect(node: LifeNode) {
  selectedNode.value = node
  detailPanelOpen.value = true
}

function closeDetailPanel() {
  detailPanelOpen.value = false
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
  regenStatus.value =
    payload.mode === 'full_cascade' ? '正在提交推演后续…' : '正在更新本节点…'
  try {
    const job = await api.patchNode(charId, selectedNode.value.id, {
      title: payload.patch.title,
      events: payload.patch.events,
      thoughts: payload.patch.thoughts,
      personality_snapshot: payload.patch.personality_snapshot,
      mode: payload.mode as 'full_cascade' | 'inner_current',
      model: payload.model,
      target_node_count: payload.target_node_count,
    })
    const done = await pollJob(job.id, (j) => {
      regenProgress.value = Math.max(j.progress, 5)
      if (j.status === 'running') {
        const stage = j.stage_text ? `${j.stage_text} · ` : ''
        const label = payload.mode === 'full_cascade' ? '推演后续' : '更新中'
        regenStatus.value = `${stage}${label} ${j.progress}%（${modelDisplayLabel(j.model || payload.model)}）`
      }
    })
    if (done.status === 'failed') throw new Error(done.error || '任务失败')
    regenProgress.value = 100
    regenStatus.value = payload.mode === 'full_cascade' ? '推演后续完成' : '更新完成'
    await load()
    const updated = nodes.value.find((n: LifeNode) => n.sequence === selectedNode.value?.sequence)
    if (updated) selectedNode.value = updated
    const msg =
      payload.mode === 'full_cascade'
        ? '已创建新分支并推演后续节点'
        : '已创建新分支并更新本节点内心/性格'
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

async function loadOverview() {
  if (!selectedTimelineId.value) return
  overviewLoading.value = true
  try {
    const res = await api.getBranchOverview(charId, selectedTimelineId.value)
    overviewBranches.value = res.branches
    currentVersionId.value = res.active_version_id
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '加载总览失败')
  } finally {
    overviewLoading.value = false
  }
}

async function enterOverview() {
  if (!selectedTimelineId.value) {
    ElMessage.warning('请先选择时间轴')
    return
  }
  viewMode.value = 'overview'
  detailPanelOpen.value = false
  await loadOverview()
}

function exitOverview() {
  viewMode.value = 'timeline'
}

async function onActivateBranch(versionId: string, options?: { returnToTimeline?: boolean }) {
  if (!selectedTimelineId.value) return
  activatingBranchId.value = versionId
  pageLoading.value = true
  try {
    await api.activateBranch(charId, selectedTimelineId.value, versionId)
    await load()
    if (options?.returnToTimeline !== false) {
      viewMode.value = 'timeline'
    } else if (viewMode.value === 'overview') {
      await loadOverview()
    }
    ElMessage.success('已切换分支')
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '切换分支失败')
  } finally {
    activatingBranchId.value = ''
    pageLoading.value = false
  }
}

async function onOverviewActivate(versionId: string) {
  await onActivateBranch(versionId, { returnToTimeline: true })
}

onMounted(load)
</script>

<template>
  <div class="timeline-page">
    <div class="toolbar">
      <div class="toolbar-left">
        <h2>{{ profile?.display_name || '人生时间轴' }}</h2>
        <p v-if="currentTimeline" class="timeline-subtitle">{{ currentTimeline.title }}</p>
        <div class="name-legend">
          <span class="legend-item"><mark class="name-protagonist">主角</mark></span>
          <span class="legend-item"><mark class="name-person">他人</mark></span>
          <span class="legend-item"><mark class="name-place">地名</mark></span>
        </div>
      </div>
      <div class="toolbar-actions">
        <el-select
          v-if="timelines.length"
          :model-value="selectedTimelineId"
          placeholder="选择时间轴"
          style="width: min(280px, 40vw)"
          @change="onTimelineChange"
        >
          <el-option
            v-for="tl in timelines"
            :key="tl.id"
            :label="`${tl.title}（${tl.node_count ?? 0} 节点）`"
            :value="tl.id"
          />
        </el-select>
        <el-button @click="router.push(`/characters/${charId}`)">时间轴列表</el-button>
        <el-button @click="router.push(`/continue/${charId}`)">新建时间轴</el-button>
        <el-button
          :type="viewMode === 'overview' ? 'primary' : 'default'"
          :disabled="!selectedTimelineId || pageLoading"
          @click="viewMode === 'overview' ? exitOverview() : enterOverview()"
        >
          {{ viewMode === 'overview' ? '返回时间轴' : '总览视图' }}
        </el-button>
        <el-button :disabled="!nodes.length || pageLoading" @click="openExport">导出文字</el-button>
        <el-button :disabled="!nodes.length || pageLoading" @click="openNarrativeChange">
          叙述变更
        </el-button>
        <el-button :disabled="!nodes.length || pageLoading" @click="openLightNovelPicker">
          生成轻小说
        </el-button>
        <el-button @click="router.push('/characters')">人物列表</el-button>
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

    <div v-if="viewMode === 'overview'" class="page-card overview-wrap">
      <BranchOverview
        :branches="overviewBranches"
        :loading="overviewLoading || pageLoading"
        :activating-id="activatingBranchId"
        @activate="onOverviewActivate"
        @back="exitOverview"
      />
    </div>

    <div v-else class="timeline-layout" :class="{ 'detail-collapsed': !detailPanelOpen }">
      <section class="col-timeline">
        <div v-loading="pageLoading" class="page-card timeline-scroll">
          <TimelineAxis
            :nodes="nodes"
            :selected-id="selectedNode?.id"
            :profile="profile"
            :reading-focus="!detailPanelOpen"
            @select="onSelect"
          />
        </div>
      </section>

      <aside v-show="detailPanelOpen" class="col-dock col-dock-editor">
        <div class="dock-panel page-card">
          <div class="dock-panel-head">
            <el-button text type="primary" @click="closeDetailPanel">← 收起详情</el-button>
          </div>
          <NodeEditor
            :character-id="charId"
            :node="selectedNode"
            :profile="profile"
            :loading="regenVisible"
            :regen-visible="regenVisible"
            :regen-progress="regenProgress"
            :regen-status="regenStatus"
            @save="onSave"
            @narrative="(kind, model) => selectedNode && openNodeNarrative(charId, selectedNode, kind, model)"
          />
        </div>
      </aside>

      <aside class="col-dock col-dock-side">
        <div class="dock-panel page-card">
          <BranchFlowchart
            :roots="branchRoots"
            :active-version-id="currentVersionId"
            :loading="branchLoading || pageLoading"
            :activating-id="activatingBranchId"
            @activate="onActivateBranch"
          />
        </div>
      </aside>
    </div>

    <el-dialog
      v-model="narrativeChangeVisible"
      title="叙述变更"
      width="min(560px, 92vw)"
      destroy-on-close
    >
      <p class="export-hint">
        用一句话描述希望如何变更时间轴，AI 会自动定位相关节点、修改或插入，并重算后续人生。无需手动选中节点。
      </p>
      <el-form label-position="top">
        <el-form-item label="变更叙述" required>
          <el-input
            v-model="narrativeChangeInstruction"
            type="textarea"
            :rows="4"
            placeholder="例如：诸葛亮病逝于白帝城托孤后一天"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="后续节点规模">
          <el-input-number v-model="narrativeChangeTargetNodes" :min="5" :max="50" />
        </el-form-item>
        <ModelSelector v-model="narrativeChangeModel" />
      </el-form>
      <template #footer>
        <el-button @click="narrativeChangeVisible = false">取消</el-button>
        <el-button type="primary" :disabled="regenVisible" @click="confirmNarrativeChange">
          应用并重算后续
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="lightNovelPickerVisible"
      title="生成轻小说"
      width="min(520px, 92vw)"
      destroy-on-close
    >
      <p class="export-hint">选择人生节点区间，AI 将以第一人称撰写连贯的轻小说章节（超过 5 个节点会自动分章）。</p>
      <el-form label-position="top">
        <el-form-item label="起始节点">
          <el-select v-model="lightNovelFrom" style="width: 100%">
            <el-option
              v-for="opt in nodeSequenceOptions"
              :key="'from-' + opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="结束节点">
          <el-select v-model="lightNovelTo" style="width: 100%">
            <el-option
              v-for="opt in nodeSequenceOptions"
              :key="'to-' + opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </el-select>
        </el-form-item>
        <ModelSelector v-model="lightNovelModel" />
      </el-form>
      <template #footer>
        <el-button @click="lightNovelPickerVisible = false">取消</el-button>
        <el-button type="primary" :loading="narrativeRunning" @click="confirmLightNovel">
          开始生成
        </el-button>
      </template>
    </el-dialog>

    <NarrativeDialog
      :model-value="narrativeVisible"
      :title="narrativeTitle"
      :content="narrativeContent"
      :loading="narrativeRunning"
      :progress="narrativeProgress"
      :status-text="narrativeStatusText"
      @update:model-value="onNarrativeDialogClose"
      @regenerate="regenerateNarrative"
    />

    <el-dialog
      v-model="exportVisible"
      title="导出人生时间轴"
      width="min(720px, 92vw)"
      class="export-dialog"
      destroy-on-close
    >
      <p class="export-hint">以下为当前版本全部节点的纯文本，可复制到笔记或文档中。</p>
      <el-input v-model="exportText" type="textarea" :rows="18" readonly class="export-textarea" />
      <template #footer>
        <el-button @click="exportVisible = false">关闭</el-button>
        <el-button type="primary" :loading="exportCopying" @click="copyExport">复制全部</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
/* 时间轴页需要更宽的三栏布局 */
.timeline-page {
  padding-bottom: 24px;
  max-width: 1500px;
  margin: 0 auto;
}
.overview-wrap {
  padding: 20px;
  min-height: 480px;
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
.timeline-subtitle {
  margin: 4px 0 0;
  font-size: 0.85rem;
  color: #909399;
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
  flex-wrap: wrap;
}
.export-hint {
  margin: 0 0 12px;
  font-size: 0.85rem;
  color: #909399;
}
.export-textarea :deep(textarea) {
  font-family: 'Consolas', 'PingFang SC', 'Microsoft YaHei', monospace;
  font-size: 0.85rem;
  line-height: 1.55;
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
  transition: grid-template-columns 0.25s ease;
}

.timeline-layout.detail-collapsed {
  grid-template-columns: minmax(0, 1fr) 300px;
  max-width: 1100px;
  margin: 0 auto;
}

.timeline-layout.detail-collapsed .col-timeline .timeline-scroll {
  padding: 20px 28px 28px;
}

.col-timeline {
  min-width: 0;
}

.dock-panel-head {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  margin: -4px 0 4px;
  padding-bottom: 4px;
  border-bottom: 1px solid #ebeef5;
  position: sticky;
  top: 0;
  background: #fff;
  z-index: 2;
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
  .timeline-layout.detail-collapsed {
    grid-template-columns: minmax(0, 1fr) 280px;
    max-width: none;
  }
  .col-dock-side {
    grid-column: 1 / -1;
  }
  .col-dock-side .dock-panel {
    max-height: 50vh;
  }
}

@media (max-width: 900px) {
  .timeline-layout,
  .timeline-layout.detail-collapsed {
    grid-template-columns: 1fr;
    max-width: none;
  }
  .col-dock-editor {
    order: 2;
  }
  .col-dock-side {
    order: 3;
  }
  .col-dock .dock-panel {
    position: static;
    max-height: none;
  }
}
</style>
