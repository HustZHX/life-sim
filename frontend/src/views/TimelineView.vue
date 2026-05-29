<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { api, type SavedChronicleMeta, type SavedLightNovelMeta } from '@/api/client'
import type { AIModelId, BranchNode, LifeNode, Timeline, WorldLine } from '@/api/client'
import { useLayoutStore } from '@/stores/layout'
import { DEFAULT_AI_MODEL, modelDisplayLabel } from '@/constants/models'
import { useModelStore } from '@/stores/model'
import TimelineAxis from '@/components/TimelineAxis.vue'
import NodeEditor from '@/components/NodeEditor.vue'
import BranchFlowchart from '@/components/BranchFlowchart.vue'
import ProfilePanel from '@/components/ProfilePanel.vue'
import NarrativeDialog from '@/components/NarrativeDialog.vue'
import CharacterDialogueDialog from '@/components/CharacterDialogueDialog.vue'
import LightNovelJobList, { type NarrativeRangeJobEntry } from '@/components/LightNovelJobList.vue'
import JobProgress from '@/components/JobProgress.vue'
import type { Profile } from '@/api/client'
import { copyTextToClipboard, formatTimelineExport } from '@/utils/exportTimelineText'
import { useNarrativeGenerate } from '@/composables/useNarrativeGenerate'
import { useChronicleJobs, type ChronicleJobEntry } from '@/composables/useChronicleJobs'
import { useLightNovelJobs } from '@/composables/useLightNovelJobs'
import { useJobRunner } from '@/composables/useJobRunner'
import { DEFAULT_TARGET_NODE_COUNT, isLivingProfile } from '@/utils/timelineDensity'
import {
  LIGHT_NOVEL_PERSON_OPTIONS,
  lightNovelPersonLabel,
  type LightNovelPerson,
} from '@/constants/lightNovelPerson'
import { timelineJobLabel, versionChangeLabel, isTimelineGenerating } from '@/constants/timelineStatus'

const route = useRoute()
const router = useRouter()
const charId = route.params.id as string
const layout = useLayoutStore()

const nodes = ref<LifeNode[]>([])
const worldLine = ref<WorldLine | null>(null)
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
const mobileTab = ref<'nodes' | 'detail' | 'branch'>('nodes')
const mobileDetailNavBusy = computed(() => pageLoading.value || jobBusy.value)

const selectedIndex = computed(() => {
  if (!selectedNode.value) return -1
  return nodes.value.findIndex((n) => n.id === selectedNode.value?.id)
})
const hasPrevNode = computed(() => selectedIndex.value > 0)
const hasNextNode = computed(() => selectedIndex.value >= 0 && selectedIndex.value < nodes.value.length - 1)

const worldLineIntroSnippet = computed(() => {
  const wl = worldLine.value
  if (!wl) return ''
  const parts = [wl.historical_trend, wl.era_summary, wl.daily_life_context]
    .filter((s): s is string => typeof s === 'string' && !!s.trim())
    .map((s) => s.trim())
  const text = parts.join(' ')
  if (!text) return ''
  return text.length > 32 ? `${text.slice(0, 32)}…` : text
})

async function scrollToSelectedNode() {
  await nextTick()
  const id = selectedNode.value?.id
  if (!id) return
  const el = document.getElementById(`node-${id}`)
  if (el && typeof el.scrollIntoView === 'function') {
    el.scrollIntoView({ block: 'center' })
  }
}

function scrollDetailToTop() {
  // 详情页切换节点时，回到顶部便于从头阅读
  try {
    window.scrollTo({ top: 0, behavior: 'smooth' })
  } catch {
    window.scrollTo(0, 0)
  }
}

function mobilePrevNode() {
  const idx = selectedIndex.value
  if (idx <= 0) return
  selectedNode.value = nodes.value[idx - 1]
  scrollDetailToTop()
}

function mobileNextNode() {
  const idx = selectedIndex.value
  if (idx < 0 || idx >= nodes.value.length - 1) return
  selectedNode.value = nodes.value[idx + 1]
  scrollDetailToTop()
}

async function mobileBackToList() {
  mobileTab.value = 'nodes'
  await scrollToSelectedNode()
}

const dialogueRef = ref<InstanceType<typeof CharacterDialogueDialog> | null>(null)

const jobRunner = useJobRunner('任务进度')
const {
  visible: jobVisible,
  progress: jobProgress,
  statusText: jobStatusText,
  title: jobTitle,
  failed: jobFailed,
  errorText: jobErrorText,
  retrying: jobRetrying,
  retry: retryJobRunner,
  clearState: clearJobState,
} = jobRunner
const jobBusy = computed(() => jobVisible.value && !jobFailed.value)

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
  jobPanelVisible: narrativeJobVisible,
  failed: narrativeFailed,
  errorText: narrativeErrorText,
  retrying: narrativeRetrying,
  close: closeNarrative,
  regenerate: regenerateNodeNarrative,
  openNodeNarrative,
  retryNarrativeJob,
  clearJobState: clearNarrativeJobState,
} = narrative

const lightNovelJobs = useLightNovelJobs(charId)
const {
  jobs: lightNovelJobList,
  listVisible: lightNovelListVisible,
  loading: lightNovelListLoading,
  activeCount: lightNovelActiveCount,
  refreshList: refreshLightNovelJobs,
  submit: submitLightNovelJob,
  retryJob: retryLightNovelJob,
} = lightNovelJobs

const modelStore = useModelStore()

const lightNovelPickerVisible = ref(false)
const lightNovelSubmitting = ref(false)
const lightNovelPerson = ref<LightNovelPerson>('first')
const lightNovelFrom = ref(0)
const lightNovelTo = ref(0)
const lightNovelViewContext = ref<{
  versionId: string
  fromSequence: number
  toSequence: number
  person: LightNovelPerson
  model: AIModelId
} | null>(null)
const savedNovelsVisible = ref(false)
const savedNovelsLoading = ref(false)
const savedNovels = ref<SavedLightNovelMeta[]>([])
const savedNovelsBranch = ref('')

const chronicleJobs = useChronicleJobs(charId)
const {
  jobs: chronicleJobList,
  listVisible: chronicleListVisible,
  loading: chronicleListLoading,
  activeCount: chronicleActiveCount,
  refreshList: refreshChronicleJobs,
  submit: submitChronicleJob,
  retryJob: retryChronicleJob,
} = chronicleJobs

const chroniclePickerVisible = ref(false)
const chronicleSubmitting = ref(false)
const chronicleFrom = ref(0)
const chronicleTo = ref(0)
const chronicleViewContext = ref<{
  versionId: string
  fromSequence: number
  toSequence: number
  model: AIModelId
} | null>(null)
const savedChroniclesVisible = ref(false)
const savedChroniclesLoading = ref(false)
const savedChronicles = ref<SavedChronicleMeta[]>([])
const savedChroniclesBranch = ref('')

const narrativeChangeVisible = ref(false)
const narrativeChangeInstruction = ref('')
const narrativeChangeTargetNodes = ref(DEFAULT_TARGET_NODE_COUNT)

const appendNextVisible = ref(false)
const appendNextTitle = ref('')

const profileLiving = computed(() => !!profile.value && isLivingProfile(profile.value))

const lastTimelineNode = computed(() => {
  const list = nodes.value
  if (!list.length) return null
  return list.reduce((a, b) => (a.sequence >= b.sequence ? a : b))
})

const appendNextDisabled = computed(
  () =>
    pageLoading.value ||
    jobBusy.value ||
    !profileLiving.value ||
    !lastTimelineNode.value ||
    !!(
      currentTimelineMeta.value && isTimelineGenerating(currentTimelineMeta.value)
    )
)

const rollbackEnabled = computed(
  () =>
    !pageLoading.value &&
    !jobBusy.value &&
    nodes.value.length > 1 &&
    nodes.value.every((n) => n.version_id === currentVersionId.value) &&
    !(currentTimelineMeta.value && isTimelineGenerating(currentTimelineMeta.value))
)

const nodeSequenceOptions = computed(() =>
  nodes.value.map((n) => ({
    value: n.sequence,
    label: `${n.sequence}. ${n.year} 年 · ${n.title}`,
  }))
)

const selectedNodeReadOnly = computed(
  () => !!selectedNode.value && selectedNode.value.version_id !== currentVersionId.value
)

const currentTimelineMeta = computed(() =>
  timelines.value.find((tl) => tl.id === selectedTimelineId.value)
)

function currentJobStatusText(): string {
  const job = currentTimelineMeta.value?.active_job
  if (!job) return ''
  const label = timelineJobLabel(job.type)
  if (job.status === 'pending') return `${label} · 排队中`
  const stage = job.stage_text ? `${job.stage_text} · ` : ''
  return `${label} · ${stage}${job.progress}%`
}

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
  if (!currentVersionId.value) return

  lightNovelSubmitting.value = true
  try {
    const ok = await submitLightNovelJob({
      versionId: currentVersionId.value,
      fromSequence: lightNovelFrom.value,
      toSequence: lightNovelTo.value,
      person: lightNovelPerson.value,
      model: modelStore.aiModel,
      onCached: (artifact) => {
        lightNovelPickerVisible.value = false
        openLightNovelContent(
          artifact.content,
          lightNovelFrom.value,
          lightNovelTo.value,
          lightNovelPerson.value
        )
        lightNovelViewContext.value = {
          versionId: currentVersionId.value,
          fromSequence: lightNovelFrom.value,
          toSequence: lightNovelTo.value,
          person: lightNovelPerson.value,
          model: modelStore.aiModel,
        }
      },
    })
    if (ok) lightNovelPickerVisible.value = false
  } finally {
    lightNovelSubmitting.value = false
  }
}

function openLightNovelContent(
  text: string,
  fromSequence: number,
  toSequence: number,
  person: LightNovelPerson
) {
  narrativeTitle.value = `轻小说 · 节点 ${fromSequence}–${toSequence} · ${lightNovelPersonLabel(person)}`
  narrativeContent.value = text
  narrativeVisible.value = true
}

function openLightNovelJobList() {
  lightNovelListVisible.value = true
  void refreshLightNovelJobs()
}

function viewLightNovelJob(entry: NarrativeRangeJobEntry) {
  if (!entry.content) return
  openLightNovelContent(
    entry.content,
    entry.fromSequence,
    entry.toSequence,
    entry.person || 'first'
  )
  lightNovelViewContext.value = {
    versionId: currentVersionId.value,
    fromSequence: entry.fromSequence,
    toSequence: entry.toSequence,
    person: entry.person || 'first',
    model: entry.model,
  }
}

async function regenerateLightNovel() {
  const ctx = lightNovelViewContext.value
  if (!ctx) return
  narrativeVisible.value = false
  await submitLightNovelJob({
    versionId: ctx.versionId,
    fromSequence: ctx.fromSequence,
    toSequence: ctx.toSequence,
    person: ctx.person,
    model: ctx.model,
    force: true,
    onCached: (artifact) => {
      openLightNovelContent(
        artifact.content,
        ctx.fromSequence,
        ctx.toSequence,
        ctx.person
      )
    },
  })
}

async function openSavedNovels() {
  savedNovelsVisible.value = true
  savedNovelsLoading.value = true
  try {
    const res = await api.listSavedLightNovels({ character_id: charId })
    savedNovels.value = res.items
    savedNovelsBranch.value = res.branch || ''
  } catch (e: unknown) {
    savedNovels.value = []
    ElMessage.error(e instanceof Error ? e.message : '加载已保存轻小说失败')
  } finally {
    savedNovelsLoading.value = false
  }
}

async function viewSavedNovel(item: SavedLightNovelMeta) {
  savedNovelsLoading.value = true
  try {
    const full = await api.getSavedLightNovel(item.id)
    savedNovelsVisible.value = false
    const person = (full.person as LightNovelPerson) || 'first'
    openLightNovelContent(full.content, full.from_sequence, full.to_sequence, person)
    lightNovelViewContext.value = {
      versionId: full.version_id,
      fromSequence: full.from_sequence,
      toSequence: full.to_sequence,
      person,
      model: (full.model as AIModelId) || DEFAULT_AI_MODEL,
    }
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '读取轻小说失败')
  } finally {
    savedNovelsLoading.value = false
  }
}

function formatSavedNovelTime(iso?: string): string {
  if (!iso) return ''
  const d = new Date(iso)
  return Number.isNaN(d.getTime()) ? iso : d.toLocaleString()
}

function openChroniclePicker() {
  if (!nodes.value.length || !currentVersionId.value) {
    ElMessage.warning('暂无节点可编纂史书')
    return
  }
  const seqs = nodes.value.map((n) => n.sequence)
  chronicleFrom.value = Math.min(...seqs)
  chronicleTo.value = Math.max(...seqs)
  chroniclePickerVisible.value = true
}

async function confirmChronicle() {
  if (chronicleFrom.value > chronicleTo.value) {
    ElMessage.warning('起始节点不能晚于结束节点')
    return
  }
  if (!currentVersionId.value) return

  chronicleSubmitting.value = true
  try {
    const ok = await submitChronicleJob({
      versionId: currentVersionId.value,
      fromSequence: chronicleFrom.value,
      toSequence: chronicleTo.value,
      model: modelStore.aiModel,
      onCached: (artifact) => {
        chroniclePickerVisible.value = false
        openChronicleContent(artifact.content, chronicleFrom.value, chronicleTo.value)
        chronicleViewContext.value = {
          versionId: currentVersionId.value,
          fromSequence: chronicleFrom.value,
          toSequence: chronicleTo.value,
          model: modelStore.aiModel,
        }
      },
    })
    if (ok) chroniclePickerVisible.value = false
  } finally {
    chronicleSubmitting.value = false
  }
}

function openChronicleContent(text: string, fromSequence: number, toSequence: number) {
  narrativeTitle.value = `史书 · 节点 ${fromSequence}–${toSequence}`
  narrativeContent.value = text
  narrativeVisible.value = true
}

function openChronicleJobList() {
  chronicleListVisible.value = true
  void refreshChronicleJobs()
}

function viewChronicleJob(entry: ChronicleJobEntry) {
  if (!entry.content) return
  openChronicleContent(entry.content, entry.fromSequence, entry.toSequence)
  chronicleViewContext.value = {
    versionId: currentVersionId.value,
    fromSequence: entry.fromSequence,
    toSequence: entry.toSequence,
    model: entry.model,
  }
}

async function regenerateChronicle() {
  const ctx = chronicleViewContext.value
  if (!ctx) return
  narrativeVisible.value = false
  await submitChronicleJob({
    versionId: ctx.versionId,
    fromSequence: ctx.fromSequence,
    toSequence: ctx.toSequence,
    model: ctx.model,
    force: true,
    onCached: (artifact) => {
      openChronicleContent(artifact.content, ctx.fromSequence, ctx.toSequence)
    },
  })
}

async function openSavedChronicles() {
  savedChroniclesVisible.value = true
  savedChroniclesLoading.value = true
  try {
    const res = await api.listSavedChronicles({ character_id: charId })
    savedChronicles.value = res.items
    savedChroniclesBranch.value = res.branch || ''
  } catch (e: unknown) {
    savedChronicles.value = []
    ElMessage.error(e instanceof Error ? e.message : '加载已保存史书失败')
  } finally {
    savedChroniclesLoading.value = false
  }
}

async function viewSavedChronicle(item: SavedChronicleMeta) {
  savedChroniclesLoading.value = true
  try {
    const full = await api.getSavedChronicle(item.id)
    savedChroniclesVisible.value = false
    openChronicleContent(full.content, full.from_sequence, full.to_sequence)
    chronicleViewContext.value = {
      versionId: full.version_id,
      fromSequence: full.from_sequence,
      toSequence: full.to_sequence,
      model: (full.model as AIModelId) || DEFAULT_AI_MODEL,
    }
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '读取史书失败')
  } finally {
    savedChroniclesLoading.value = false
  }
}

function onNarrativeDialogClose(visible: boolean) {
  if (!visible) {
    closeNarrative()
    lightNovelViewContext.value = null
    chronicleViewContext.value = null
  }
}

async function onNarrativeRegenerate() {
  if (lightNovelViewContext.value) {
    await regenerateLightNovel()
  } else if (chronicleViewContext.value) {
    await regenerateChronicle()
  } else {
    await regenerateNodeNarrative()
  }
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
  const body = {
    timeline_id: selectedTimelineId.value,
    instruction,
    model: modelStore.aiModel,
    target_node_count: narrativeChangeTargetNodes.value,
  }
  const ok = await jobRunner.run({
    title: '叙述变更',
    submitLabel: '正在提交叙述变更…',
    submit: () => api.applyNarrativeChange(charId, body),
    resubmit: () => api.applyNarrativeChange(charId, body),
    onProgress: (j) => {
      if (j.status === 'running') {
        const stage = j.stage_text ? `${j.stage_text} · ` : '处理中… '
        jobRunner.statusText.value = `${stage}${j.progress}%（${modelDisplayLabel(j.model || modelStore.aiModel)}）`
      }
    },
    afterSuccess: async () => {
      await load()
      ElMessage.success('已根据叙述创建新分支并推演后续节点')
    },
  })
  if (!ok && !jobFailed.value) {
    ElMessage.error('叙述变更失败')
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

async function refreshWorldLine() {
  if (!selectedTimelineId.value || !nodes.value.length) {
    ElMessage.warning('当前时间轴尚无节点')
    return
  }
  const ok = await jobRunner.run({
    title: '刷新世界线',
    submitLabel: '正在提交刷新任务…',
    submit: () =>
      api.refreshWorldLine(charId, selectedTimelineId.value, { model: modelStore.aiModel }),
    resubmit: () =>
      api.refreshWorldLine(charId, selectedTimelineId.value, { model: modelStore.aiModel }),
    onProgress: (j) => {
      if (j.status === 'running' && j.stage_text) {
        jobRunner.statusText.value = j.stage_text
      }
    },
    afterSuccess: async () => {
      await load()
      ElMessage.success('世界线已更新')
    },
  })
  if (!ok && !jobFailed.value) {
    ElMessage.error('刷新世界线失败')
  }
}

async function load() {
  pageLoading.value = true
  try {
    const tlRes = await api.listTimelines(charId).catch(() => ({ timelines: [] as Timeline[] }))
    timelines.value = Array.isArray(tlRes.timelines) ? tlRes.timelines : []
    if (!timelines.value.length) {
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
    worldLine.value = tl.timeline.world_line ?? null
    const listMeta = timelines.value.find((t) => t.id === tl.timeline.id)
    currentTimeline.value = listMeta
      ? { ...tl.timeline, version_count: listMeta.version_count, active_job: listMeta.active_job }
      : tl.timeline
    selectedTimelineId.value = tl.timeline.id
    currentVersionId.value = tl.version.id
    profile.value = prof
    branchRoots.value = branches.roots
    selectedNode.value = null
    detailPanelOpen.value = false
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '加载失败'
    if (msg.includes('正在生成') || msg.includes('尚无版本')) {
      ElMessage.warning('该时间轴正在生成中，请返回列表等待完成')
      router.replace(`/characters/${charId}`)
      return
    }
    ElMessage.error(msg)
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

watch(
  () => mobileTab.value,
  (tab) => {
    if (!layout.isMobile) return
    if (tab === 'nodes') void scrollToSelectedNode()
  }
)

function onSelect(node: LifeNode) {
  selectedNode.value = node
  if (layout.isMobile) {
    mobileTab.value = 'detail'
    detailPanelOpen.value = false
    // 进入详情不强制跳到页面顶部：保持当前位置
    return
  }
  detailPanelOpen.value = true
}

function closeDetailPanel() {
  detailPanelOpen.value = false
}

function onOpenDialogue(_model: AIModelId) {
  if (!selectedNode.value) return
  dialogueRef.value?.open(
    charId,
    selectedNode.value,
    selectedNode.value.version_id,
    currentVersionId.value
  )
}

function onOpenDialogueHistory() {
  dialogueRef.value?.openHistory(charId, currentVersionId.value, currentVersionId.value)
}

async function onDialogueNodeUpdated(payload: { versionId: string; node: LifeNode }) {
  currentVersionId.value = payload.versionId
  if (!selectedTimelineId.value) return
  try {
    const tl = await api.getTimeline(charId, {
      timelineId: selectedTimelineId.value,
      version: payload.versionId,
    })
    nodes.value = tl.nodes
    selectedNode.value = payload.node
    const branches = await api.listBranches(charId, selectedTimelineId.value)
    branchRoots.value = branches.roots
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '刷新时间轴失败')
  }
}

async function onSave(payload: {
  mode: string
  model: AIModelId
  patch: Record<string, string>
  target_node_count?: number
}) {
  if (!selectedNode.value) return
  const nodeId = selectedNode.value.id
  const patchBody = {
    title: payload.patch.title,
    events: payload.patch.events,
    thoughts: payload.patch.thoughts,
    personality_snapshot: payload.patch.personality_snapshot,
    mode: payload.mode as 'full_cascade' | 'inner_current',
    model: payload.model,
    target_node_count: payload.target_node_count,
  }
  const ok = await jobRunner.run({
    title: '任务进度',
    submitLabel: payload.mode === 'full_cascade' ? '正在提交推演后续…' : '正在更新本节点…',
    submit: () => api.patchNode(charId, nodeId, patchBody),
    resubmit: () => api.patchNode(charId, nodeId, patchBody),
    onProgress: (j) => {
      if (j.status === 'running') {
        const stage = j.stage_text ? `${j.stage_text} · ` : ''
        const label = payload.mode === 'full_cascade' ? '推演后续' : '更新中'
        jobRunner.statusText.value = `${stage}${label} ${j.progress}%（${modelDisplayLabel(j.model || payload.model)}）`
      }
    },
    afterSuccess: async () => {
      await load()
      const updated = nodes.value.find((n: LifeNode) => n.sequence === selectedNode.value?.sequence)
      if (updated) {
        selectedNode.value = updated
      }
      const msg =
        payload.mode === 'full_cascade'
          ? '已创建新分支并推演后续节点'
          : '已更新本节点（未创建新分支；推演后续才会分叉）'
      ElMessage.success(msg)
    },
  })
  if (!ok && !jobFailed.value) {
    ElMessage.error('保存失败')
  }
}

function openAppendNext() {
  if (appendNextDisabled.value) return
  appendNextTitle.value = ''
  appendNextVisible.value = true
}

async function confirmAppendNext(randomTitle: boolean) {
  const anchor = lastTimelineNode.value
  if (!anchor || appendNextDisabled.value) return
  const titleHint = randomTitle ? '' : appendNextTitle.value.trim()
  appendNextVisible.value = false
  const patchBody = {
    title: anchor.title,
    events: anchor.events,
    thoughts: anchor.thoughts ?? '',
    personality_snapshot: anchor.personality_snapshot ?? '',
    mode: 'append_next' as const,
    model: modelStore.aiModel,
    target_node_count: 1,
    next_node_title: titleHint,
  }
  const ok = await jobRunner.run({
    title: '继续推演',
    submitLabel: '正在继续推演…',
    submit: () => api.patchNode(charId, anchor.id, patchBody),
    resubmit: () => api.patchNode(charId, anchor.id, patchBody),
    onProgress: (j) => {
      if (j.status === 'running') {
        const stage = j.stage_text ? `${j.stage_text} · ` : ''
        jobRunner.statusText.value = `${stage}继续推演 ${j.progress}%（${modelDisplayLabel(j.model || modelStore.aiModel)}）`
      }
    },
    afterSuccess: async () => {
      await load()
      const newLast = lastTimelineNode.value
      if (newLast) {
        selectedNode.value = newLast
        detailPanelOpen.value = true
      }
      ElMessage.success('已追加 1 个节点并创建新分支')
    },
  })
  if (!ok && !jobFailed.value) {
    ElMessage.error('继续推演失败')
  }
}

async function onActivateBranch(versionId: string) {
  if (!selectedTimelineId.value) return
  activatingBranchId.value = versionId
  pageLoading.value = true
  try {
    await api.activateBranch(charId, selectedTimelineId.value, versionId)
    await load()
    ElMessage.success('已切换分支')
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '切换分支失败')
  } finally {
    activatingBranchId.value = ''
    pageLoading.value = false
  }
}

async function onRollbackToNode(node: LifeNode) {
  if (!rollbackEnabled.value) return
  const afterCount = nodes.value.filter((n) => n.sequence > node.sequence).length
  if (afterCount <= 0) {
    ElMessage.warning('该节点已是时间轴末尾')
    return
  }
  try {
    await ElMessageBox.confirm(
      `将回退到「${node.year} 年 · ${node.title}」，并丢弃其后 ${afterCount} 个节点。此操作会创建新分支，原时间轴仍可在分支图中查看。`,
      '回退至此',
      {
        confirmButtonText: '确认回退',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
  } catch {
    return
  }

  pageLoading.value = true
  try {
    await api.rollbackToNode(charId, node.id)
    await load()
    const updated = nodes.value.find((n) => n.sequence === node.sequence)
    if (updated) {
      selectedNode.value = updated
      detailPanelOpen.value = true
    }
    ElMessage.success(`已回退至节点 #${node.sequence}`)
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '回退失败')
  } finally {
    pageLoading.value = false
  }
}

onMounted(async () => {
  await load()
  await refreshLightNovelJobs()
  lightNovelJobs.startBackgroundPoll()
  await refreshChronicleJobs()
  chronicleJobs.startBackgroundPoll()
})
</script>

<template>
  <div class="timeline-page">
    <div class="toolbar">
      <div class="toolbar-left">
        <h2>{{ profile?.display_name || '人生时间轴' }}</h2>
        <p v-if="currentTimeline" class="timeline-subtitle">{{ currentTimeline.title }}</p>
        <p v-if="currentTimelineMeta?.version_count" class="timeline-meta">
          版本：{{ versionChangeLabel(currentTimelineMeta.version_count) }}
        </p>
        <div v-if="currentTimelineMeta && isTimelineGenerating(currentTimelineMeta)" class="timeline-job-banner">
          <el-progress
            :percentage="Math.max(currentTimelineMeta.active_job?.progress ?? 8, 5)"
            :stroke-width="6"
            striped
            striped-flow
          />
          <span class="timeline-job-text">{{ currentJobStatusText() }}</span>
        </div>
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
        <template v-if="layout.isMobile">
          <el-button :disabled="!nodes.length || pageLoading" @click="openExport">导出</el-button>
          <el-button :disabled="pageLoading" @click="openSavedNovels">轻小说</el-button>
          <el-button @click="showProfile = !showProfile">{{ showProfile ? '隐藏' : '档案' }}</el-button>
          <el-dropdown trigger="click">
            <el-button>更多</el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="router.push(`/characters/${charId}`)">时间轴列表</el-dropdown-item>
                <el-dropdown-item @click="router.push('/characters')">人物列表</el-dropdown-item>
                <el-dropdown-item :disabled="pageLoading" @click="onOpenDialogueHistory">历史对话</el-dropdown-item>
                <el-dropdown-item :disabled="pageLoading" @click="openLightNovelJobList">
                  生成队列{{ lightNovelActiveCount ? `（${lightNovelActiveCount}）` : '' }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
        <template v-else>
          <el-button @click="router.push(`/characters/${charId}`)">时间轴列表</el-button>
          <el-button @click="router.push(`/continue/${charId}`)">新建时间轴</el-button>
          <el-button :disabled="!nodes.length || pageLoading" @click="openExport">导出文字</el-button>
          <el-button :disabled="pageLoading" @click="onOpenDialogueHistory">历史对话</el-button>
          <el-button :disabled="!nodes.length || pageLoading" @click="openNarrativeChange">
            叙述变更
          </el-button>
          <el-button :disabled="!nodes.length || pageLoading" @click="openLightNovelPicker">
            生成轻小说
          </el-button>
          <el-button :disabled="!nodes.length || pageLoading" @click="openChroniclePicker">
            编纂史书
          </el-button>
          <el-badge :value="lightNovelActiveCount" :hidden="!lightNovelActiveCount" type="warning">
            <el-button :disabled="pageLoading" @click="openLightNovelJobList">轻小说队列</el-button>
          </el-badge>
          <el-badge :value="chronicleActiveCount" :hidden="!chronicleActiveCount" type="warning">
            <el-button :disabled="pageLoading" @click="openChronicleJobList">史书队列</el-button>
          </el-badge>
          <el-button :disabled="pageLoading" @click="openSavedNovels">已保存轻小说</el-button>
          <el-button :disabled="pageLoading" @click="openSavedChronicles">已保存史书</el-button>
          <el-button @click="router.push('/characters')">人物列表</el-button>
          <el-button @click="showProfile = !showProfile">
            {{ showProfile ? '隐藏' : '查看' }}档案
          </el-button>
        </template>
      </div>
    </div>

    <el-collapse-transition>
      <div v-if="showProfile && profile" class="page-card profile-block">
        <ProfilePanel :profile="profile" />
      </div>
    </el-collapse-transition>

    <div v-if="layout.isMobile" class="mobile-tabs">
      <el-tabs v-model="mobileTab" stretch>
        <el-tab-pane label="节点" name="nodes">
          <div v-loading="pageLoading" class="page-card timeline-dual">
            <div class="track-life">
              <header class="track-life-head">
                <h3>人生节点</h3>
                <span v-if="nodes.length" class="track-life-meta">{{ nodes.length }} 个节点</span>
              </header>
              <div v-if="worldLine" class="worldline-intro">
                <el-collapse accordion>
                  <el-collapse-item name="intro">
                    <template #title>
                      <div class="worldline-intro-title">
                        <span>世界线简介</span>
                        <span v-if="worldLineIntroSnippet" class="worldline-intro-snippet">
                          {{ worldLineIntroSnippet }}
                        </span>
                        <el-button
                          size="small"
                          type="primary"
                          :disabled="jobBusy || pageLoading"
                          @click.stop="refreshWorldLine"
                        >
                          刷新世界线
                        </el-button>
                      </div>
                    </template>
                    <div class="worldline-intro-body">
                      <p v-if="worldLine.historical_trend" class="worldline-intro-p">
                        <strong>历史趋势：</strong>{{ worldLine.historical_trend }}
                      </p>
                      <p v-if="worldLine.era_summary" class="worldline-intro-p">
                        <strong>时代概览：</strong>{{ worldLine.era_summary }}
                      </p>
                      <p v-if="worldLine.daily_life_context" class="worldline-intro-p">
                        <strong>日常背景：</strong>{{ worldLine.daily_life_context }}
                      </p>
                    </div>
                  </el-collapse-item>
                </el-collapse>
              </div>
              <TimelineAxis
                :nodes="nodes"
                :world-line="worldLine"
                :selected-id="selectedNode?.id"
                :profile="profile"
                :append-disabled="true"
                :rollback-enabled="false"
                @select="onSelect"
              />
            </div>
          </div>
        </el-tab-pane>

        <el-tab-pane label="详情" name="detail">
          <div class="page-card mobile-detail-card">
            <div v-if="selectedNode" class="mobile-detail-topbar">
              <el-button size="small" @click="mobileBackToList">返回列表</el-button>
            </div>
            <el-empty v-if="!selectedNode" description="请选择一个节点查看详情" />
            <NodeEditor
              v-else
              :character-id="charId"
              :node="selectedNode"
              :profile="profile"
              :read-only="true"
              :loading="jobBusy"
              @narrative="(kind, model) => selectedNode && openNodeNarrative(charId, selectedNode, kind, model)"
              @dialogue="onOpenDialogue"
            />
            <button
              v-if="selectedNode"
              type="button"
              class="mobile-detail-arrow mobile-detail-arrow--left"
              :disabled="mobileDetailNavBusy || !hasPrevNode"
              aria-label="上一个节点"
              @click="mobilePrevNode"
            >
              ←
            </button>
            <button
              v-if="selectedNode"
              type="button"
              class="mobile-detail-arrow mobile-detail-arrow--right"
              :disabled="mobileDetailNavBusy || !hasNextNode"
              aria-label="下一个节点"
              @click="mobileNextNode"
            >
              →
            </button>
          </div>
        </el-tab-pane>

        <el-tab-pane label="分支" name="branch">
          <div class="page-card">
            <BranchFlowchart
              :roots="branchRoots"
              :active-version-id="currentVersionId"
              :loading="branchLoading || pageLoading"
              :activating-id="activatingBranchId"
              @activate="onActivateBranch"
            />
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>

    <div v-else class="timeline-layout" :class="{ 'detail-collapsed': !detailPanelOpen }">
      <section class="col-timeline">
        <div v-loading="pageLoading" class="page-card timeline-dual">
          <div class="track-life">
            <header class="track-life-head">
              <h3>人生节点</h3>
              <span v-if="nodes.length" class="track-life-meta">{{ nodes.length }} 个节点</span>
            </header>
            <div v-if="worldLine" class="worldline-intro">
              <el-collapse accordion>
                <el-collapse-item name="intro">
                  <template #title>
                    <div class="worldline-intro-title">
                      <span>世界线简介</span>
                      <span v-if="worldLineIntroSnippet" class="worldline-intro-snippet">
                        {{ worldLineIntroSnippet }}
                      </span>
                      <el-button
                        size="small"
                        type="primary"
                        :disabled="jobBusy || pageLoading"
                        @click.stop="refreshWorldLine"
                      >
                        刷新世界线
                      </el-button>
                    </div>
                  </template>
                  <div class="worldline-intro-body">
                    <p v-if="worldLine.historical_trend" class="worldline-intro-p">
                      <strong>历史趋势：</strong>{{ worldLine.historical_trend }}
                    </p>
                    <p v-if="worldLine.era_summary" class="worldline-intro-p">
                      <strong>时代概览：</strong>{{ worldLine.era_summary }}
                    </p>
                    <p v-if="worldLine.daily_life_context" class="worldline-intro-p">
                      <strong>日常背景：</strong>{{ worldLine.daily_life_context }}
                    </p>
                  </div>
                </el-collapse-item>
              </el-collapse>
            </div>
            <TimelineAxis
              :nodes="nodes"
              :world-line="worldLine"
              :selected-id="selectedNode?.id"
              :profile="profile"
              :append-disabled="appendNextDisabled"
              :rollback-enabled="rollbackEnabled"
              @select="onSelect"
              @append="openAppendNext"
              @rollback="onRollbackToNode"
            />
          </div>
        </div>
      </section>

      <aside v-show="detailPanelOpen" class="col-dock col-dock-editor">
        <div class="dock-panel dock-panel--editor page-card">
          <div class="dock-panel-toolbar">
            <el-button size="small" :icon="ArrowLeft" @click="closeDetailPanel">收起详情</el-button>
          </div>
          <div class="dock-panel-body">
            <NodeEditor
              :character-id="charId"
              :node="selectedNode"
              :profile="profile"
              :read-only="selectedNodeReadOnly"
              :loading="jobBusy"
              @save="onSave"
              @narrative="(kind, model) => selectedNode && openNodeNarrative(charId, selectedNode, kind, model)"
              @dialogue="onOpenDialogue"
            />
          </div>
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
      v-model="appendNextVisible"
      title="继续推演"
      width="min(480px, 92vw)"
      destroy-on-close
    >
      <p class="export-hint">
        在时间轴末尾追加 1 个节点。可填写下一节点标题，或留空由 AI 随机拟定标题与经历。
      </p>
      <el-form label-position="top">
        <el-form-item label="下一节点标题（可选）">
          <el-input
            v-model="appendNextTitle"
            maxlength="80"
            show-word-limit
            placeholder="留空则随机生成标题"
            clearable
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="appendNextVisible = false">取消</el-button>
        <el-button :disabled="jobBusy" @click="confirmAppendNext(true)">
          随机生成
        </el-button>
        <el-button type="primary" :disabled="jobBusy" @click="confirmAppendNext(false)">
          开始推演
        </el-button>
      </template>
    </el-dialog>

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
          <el-input-number v-model="narrativeChangeTargetNodes" :min="1" :max="25" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="narrativeChangeVisible = false">取消</el-button>
        <el-button type="primary" :disabled="jobBusy" @click="confirmNarrativeChange">
          应用并重算后续
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="savedNovelsVisible"
      :title="savedNovelsBranch ? `已保存轻小说（分支 ${savedNovelsBranch}）` : '已保存轻小说（当前 Git 分支）'"
      width="min(640px, 92vw)"
      destroy-on-close
    >
      <p class="export-hint">
        每次生成会新增一条记录（不覆盖旧文件）。文件保存在项目 data/light-novels/&lt;分支名&gt;/ 下。
      </p>
      <div v-loading="savedNovelsLoading">
        <el-empty v-if="!savedNovels.length && !savedNovelsLoading" description="当前分支下暂无已保存轻小说" />
        <el-table v-else :data="savedNovels" size="small" stripe @row-click="viewSavedNovel">
          <el-table-column label="节点区间" width="110">
            <template #default="{ row }">{{ row.from_sequence }}–{{ row.to_sequence }}</template>
          </el-table-column>
          <el-table-column label="人称" width="120">
            <template #default="{ row }">{{ lightNovelPersonLabel(row.person) }}</template>
          </el-table-column>
          <el-table-column prop="display_name" label="人物" min-width="100" show-overflow-tooltip />
          <el-table-column label="保存时间" min-width="150">
            <template #default="{ row }">{{ formatSavedNovelTime(row.created_at) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="72" align="center">
            <template #default="{ row }">
              <el-button link type="primary" @click.stop="viewSavedNovel(row)">查看</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
      <template #footer>
        <el-button @click="savedNovelsVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="lightNovelPickerVisible"
      title="生成轻小说"
      width="min(520px, 92vw)"
      destroy-on-close
    >
      <p class="export-hint">选择人生节点区间与叙述人称，任务将在后台生成（超过 5 个节点会自动分章）。</p>
      <el-form label-position="top">
        <el-form-item label="叙述人称">
          <el-radio-group v-model="lightNovelPerson">
            <el-radio
              v-for="opt in LIGHT_NOVEL_PERSON_OPTIONS"
              :key="opt.value"
              :value="opt.value"
            >
              {{ opt.label }}
            </el-radio>
          </el-radio-group>
        </el-form-item>
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
      </el-form>
      <template #footer>
        <el-button @click="lightNovelPickerVisible = false">取消</el-button>
        <el-button type="primary" :loading="lightNovelSubmitting" @click="confirmLightNovel">
          提交后台生成
        </el-button>
      </template>
    </el-dialog>

    <LightNovelJobList
      v-model="lightNovelListVisible"
      :jobs="lightNovelJobList"
      :loading="lightNovelListLoading"
      @refresh="refreshLightNovelJobs"
      @view="viewLightNovelJob"
      @retry="retryLightNovelJob"
    />

    <LightNovelJobList
      v-model="chronicleListVisible"
      :jobs="chronicleJobList"
      :loading="chronicleListLoading"
      drawer-title="史书编纂队列"
      hint="编纂任务在后台运行，可关闭此面板继续编辑。超过 5 个节点会自动分篇。"
      empty-description="暂无编纂任务，点击「编纂史书」提交"
      :show-person="false"
      @refresh="refreshChronicleJobs"
      @view="viewChronicleJob"
      @retry="retryChronicleJob"
    />

    <el-dialog
      v-model="chroniclePickerVisible"
      title="编纂史书"
      width="min(520px, 92vw)"
      destroy-on-close
    >
      <p class="export-hint">
        综合世界线、人生节点与人物档案撰写史书篇章；任务在后台生成（超过 5 个节点自动分篇）。
      </p>
      <el-form label-position="top">
        <el-form-item label="起始节点">
          <el-select v-model="chronicleFrom" style="width: 100%">
            <el-option
              v-for="opt in nodeSequenceOptions"
              :key="'ch-from-' + opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="结束节点">
          <el-select v-model="chronicleTo" style="width: 100%">
            <el-option
              v-for="opt in nodeSequenceOptions"
              :key="'ch-to-' + opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="chroniclePickerVisible = false">取消</el-button>
        <el-button type="primary" :loading="chronicleSubmitting" @click="confirmChronicle">
          提交后台编纂
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="savedChroniclesVisible"
      :title="savedChroniclesBranch ? `已保存史书（分支 ${savedChroniclesBranch}）` : '已保存史书（当前 Git 分支）'"
      width="min(640px, 92vw)"
      destroy-on-close
    >
      <p class="export-hint">
        每次编纂会新增一条记录。文件保存在项目 data/chronicles/&lt;分支名&gt;/ 下。
      </p>
      <div v-loading="savedChroniclesLoading">
        <el-empty v-if="!savedChronicles.length && !savedChroniclesLoading" description="当前分支下暂无已保存史书" />
        <el-table v-else :data="savedChronicles" size="small" stripe @row-click="viewSavedChronicle">
          <el-table-column label="节点区间" width="110">
            <template #default="{ row }">{{ row.from_sequence }}–{{ row.to_sequence }}</template>
          </el-table-column>
          <el-table-column prop="display_name" label="人物" min-width="100" show-overflow-tooltip />
          <el-table-column label="保存时间" min-width="150">
            <template #default="{ row }">{{ formatSavedNovelTime(row.created_at) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="72" align="center">
            <template #default="{ row }">
              <el-button link type="primary" @click.stop="viewSavedChronicle(row)">查看</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
      <template #footer>
        <el-button @click="savedChroniclesVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <CharacterDialogueDialog
      ref="dialogueRef"
      :all-nodes="nodes"
      @node-updated="onDialogueNodeUpdated"
    />

    <NarrativeDialog
      :model-value="narrativeVisible"
      :title="narrativeTitle"
      :content="narrativeContent"
      :loading="(narrativeRunning || narrativeJobVisible) && !lightNovelViewContext && !chronicleViewContext"
      :progress="narrativeProgress"
      :status-text="narrativeStatusText"
      :failed="narrativeFailed"
      :error-text="narrativeErrorText"
      :retrying="narrativeRetrying"
      @update:model-value="onNarrativeDialogClose"
      @regenerate="onNarrativeRegenerate"
      @retry="retryNarrativeJob"
      @dismiss-job="clearNarrativeJobState"
    />

    <JobProgress
      :visible="jobVisible"
      :progress="jobProgress"
      :status-text="jobStatusText"
      :title="jobTitle"
      :failed="jobFailed"
      :error-text="jobErrorText"
      :retrying="jobRetrying"
      @retry="retryJobRunner()"
      @dismiss="clearJobState()"
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
  max-width: 1680px;
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
.timeline-subtitle {
  margin: 4px 0 0;
  font-size: 0.85rem;
  color: #909399;
}
.timeline-meta {
  margin: 2px 0 0;
  font-size: 0.82rem;
  color: #606266;
}
.timeline-job-banner {
  margin-top: 8px;
  max-width: 320px;
}
.timeline-job-text {
  display: block;
  margin-top: 4px;
  font-size: 0.78rem;
  color: #e6a23c;
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

.mobile-detail-card {
  position: relative;
  padding-bottom: calc(var(--layout-bottom-nav-h) + 12px);
}

.mobile-detail-topbar {
  position: sticky;
  top: 0;
  z-index: 20;
  display: flex;
  justify-content: flex-start;
  padding-bottom: 10px;
  margin-bottom: 10px;
  background: #fff;
  border-bottom: 1px solid #ebeef5;
}

.mobile-detail-arrow {
  position: fixed;
  top: 50%;
  transform: translateY(-50%);
  z-index: 40;
  width: 46px;
  height: 46px;
  border-radius: 999px;
  border: 1px solid rgba(0, 0, 0, 0.08);
  background: rgba(255, 255, 255, 0.55);
  color: #111827;
  font-size: 22px;
  line-height: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  -webkit-tap-highlight-color: transparent;
  backdrop-filter: blur(6px);
}

.mobile-detail-arrow--left {
  left: 10px;
}

.mobile-detail-arrow--right {
  right: 10px;
}

.mobile-detail-arrow:disabled {
  opacity: 0.25;
  cursor: not-allowed;
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
  grid-template-columns: minmax(0, 1fr) minmax(0, 400px) minmax(0, 300px);
  gap: 16px;
  align-items: stretch;
  transition: grid-template-columns 0.25s ease;
  width: 100%;
}

.col-timeline .timeline-scroll {
  width: 100%;
  min-width: 0;
  container-type: inline-size;
}

.timeline-dual {
  padding: 0;
}

.track-life {
  min-width: 0;
  padding: 0 20px 24px;
}

.track-life-head {
  display: flex;
  align-items: baseline;
  gap: 10px;
  padding: 14px 0 12px;
  background: #fff;
  border-bottom: 1px solid #ebeef5;
  margin-bottom: 8px;
}

.track-life-head h3 {
  margin: 0;
  font-size: 0.95rem;
  font-weight: 600;
  color: #303133;
}

.track-life-meta {
  font-size: 0.78rem;
  color: #909399;
}

.timeline-layout.detail-collapsed {
  grid-template-columns: minmax(0, 1fr) minmax(220px, 280px);
  width: 100%;
  max-width: none;
}

.timeline-layout.detail-collapsed .col-timeline {
  min-width: 0;
}

.timeline-layout.detail-collapsed .track-life {
  padding: 0 20px 24px;
}

.timeline-layout.detail-collapsed .col-dock-side .dock-panel {
  max-height: calc(100vh - 32px);
}

.col-timeline {
  min-width: 0;
}

.dock-panel--editor {
  display: flex;
  flex-direction: column;
  padding: 0;
  overflow: hidden;
}

.dock-panel-toolbar {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  padding: 10px 16px;
  border-bottom: 1px solid #ebeef5;
  background: #fff;
  border-radius: 12px 12px 0 0;
  position: sticky;
  top: 0;
  z-index: 3;
}

.dock-panel-body {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 8px 20px 20px;
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

@media (max-width: 1400px) {
  .timeline-layout {
    grid-template-columns: minmax(0, 1fr) minmax(0, 340px) minmax(0, 260px);
    gap: 12px;
  }
}

@media (max-width: 1280px) {
  .timeline-layout {
    grid-template-columns: minmax(0, 1fr) minmax(0, 360px);
  }
  .timeline-layout.detail-collapsed {
    grid-template-columns: minmax(0, 1fr) minmax(180px, 240px);
    max-width: none;
  }
  .col-dock-side {
    grid-column: 1 / -1;
  }
  .col-dock-side .dock-panel {
    max-height: 40vh;
  }
  .track-life {
    padding: 0 16px 20px;
  }
}

@media (max-width: 960px) {
  .track-life {
    padding: 0 12px 20px;
  }
}

.worldline-intro {
  margin: 12px 0 10px;
}
.worldline-intro-title {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}
.worldline-intro-snippet {
  flex: 1;
  min-width: 0;
  color: #909399;
  font-size: 0.82rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.worldline-intro-body {
  padding: 6px 0 2px;
}
.worldline-intro-p {
  margin: 0 0 10px;
  color: #303133;
  line-height: 1.55;
  white-space: pre-wrap;
}

@media (max-width: 900px) {
  .timeline-layout,
  .timeline-layout.detail-collapsed {
    grid-template-columns: 1fr;
    max-width: none;
    gap: 12px;
  }
  .col-timeline .timeline-dual {
    min-height: auto;
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
  .dock-panel--editor {
    max-height: none;
  }
  .dock-panel-body {
    overflow: visible;
  }
}
</style>
