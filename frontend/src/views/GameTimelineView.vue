<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import {
  api,
  type AIModelId,
  type GameNodeChoiceDisplay,
  type LifeNode,
  type Profile,
  type Timeline,
  type WorldLine,
} from '@/api/client'
import { useLayoutStore } from '@/stores/layout'
import { useModelStore } from '@/stores/model'
import { useChronicleJobs, type ChronicleJobEntry } from '@/composables/useChronicleJobs'
import NarrativeDialog from '@/components/NarrativeDialog.vue'
import LightNovelJobList from '@/components/LightNovelJobList.vue'
import TimelineAxis from '@/components/TimelineAxis.vue'
import WorldLinePanel from '@/components/WorldLinePanel.vue'
import ProfilePanel from '@/components/ProfilePanel.vue'
import GameChoicePanel from '@/components/GameChoicePanel.vue'
import NodeEditor from '@/components/NodeEditor.vue'
import CharacterDialogueDialog from '@/components/CharacterDialogueDialog.vue'
import JobProgress from '@/components/JobProgress.vue'
import { useJobRunner } from '@/composables/useJobRunner'

const route = useRoute()
const router = useRouter()
const charId = route.params.id as string
const layout = useLayoutStore()

const nodes = ref<LifeNode[]>([])
const gameChoices = ref<GameNodeChoiceDisplay[]>([])
const worldLine = ref<WorldLine | null>(null)
const profile = ref<Profile | null>(null)
const characterMode = ref<'famous' | 'random'>('random')
const currentTimeline = ref<Timeline | null>(null)
const selectedNode = ref<LifeNode | null>(null)
const pageLoading = ref(false)
const showProfile = ref(false)
const detailPanelOpen = ref(false)
const mobileTab = ref<'nodes' | 'detail' | 'sidebar'>('nodes')
const choicePanelOpen = ref(false)
const choiceSimulating = ref(false)
const jobBusy = ref(false)
const worldLineRefreshing = ref(false)
const dialogueRef = ref<InstanceType<typeof CharacterDialogueDialog> | null>(null)
const profileJobRunner = useJobRunner('刷新档案')

const currentVersionId = computed(() => nodes.value[0]?.version_id ?? '')

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
const modelStore = useModelStore()
const chronicleFrom = ref(0)
const chronicleTo = ref(0)
const chronicleVisible = ref(false)
const chronicleTitle = ref('')
const chronicleContent = ref('')
const chronicleViewContext = ref<{
  versionId: string
  fromSequence: number
  toSequence: number
  model: AIModelId
} | null>(null)

const nodeSequenceOptions = computed(() =>
  nodes.value.map((n) => ({
    value: n.sequence,
    label: `${n.sequence}. ${n.year} 年 · ${n.title}`,
  }))
)

const lastNode = computed(() => {
  const list = nodes.value
  if (!list.length) return null
  return list.reduce((a, b) => (a.sequence >= b.sequence ? a : b))
})

const appendDisabled = computed(() => pageLoading.value || jobBusy.value || !lastNode.value)
const rollbackEnabled = computed(() => !pageLoading.value && !jobBusy.value && nodes.value.length > 0)

const showChoiceForNode = computed(
  () =>
    choicePanelOpen.value &&
    !!lastNode.value &&
    !!selectedNode.value &&
    selectedNode.value.id === lastNode.value.id
)

const selectedIndex = computed(() => {
  if (!selectedNode.value) return -1
  return nodes.value.findIndex((n) => n.id === selectedNode.value?.id)
})
const hasPrevNode = computed(() => selectedIndex.value > 0)
const hasNextNode = computed(() => selectedIndex.value >= 0 && selectedIndex.value < nodes.value.length - 1)
const mobileDetailNavBusy = computed(() => pageLoading.value || jobBusy.value)

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
  choicePanelOpen.value = false
  scrollDetailToTop()
}

function mobileNextNode() {
  const idx = selectedIndex.value
  if (idx < 0 || idx >= nodes.value.length - 1) return
  selectedNode.value = nodes.value[idx + 1]
  choicePanelOpen.value = false
  scrollDetailToTop()
}

async function mobileBackToList() {
  mobileTab.value = 'nodes'
  await scrollToSelectedNode()
}

async function load() {
  pageLoading.value = true
  try {
    const ch = await api.getCharacter(charId)
    characterMode.value = ch.mode === 'famous' ? 'famous' : 'random'
    if (ch.play_style && ch.play_style !== 'game') {
      router.replace(`/timeline/${charId}`)
      return
    }
    const tlRes = await api.listTimelines(charId)
    const timelines = tlRes.timelines ?? []
    if (!timelines.length) {
      router.replace(characterMode.value === 'famous' ? '/game/famous' : '/game/random')
      return
    }
    const timelineId = (route.query.timeline as string) || ch.current_timeline_id || timelines[0].id
    const [tl, prof] = await Promise.all([
      api.getTimeline(charId, { timelineId }),
      api.getProfile(charId),
    ])
    nodes.value = tl.nodes
    gameChoices.value = tl.game_choices ?? []
    worldLine.value = tl.timeline.world_line ?? null
    currentTimeline.value = tl.timeline
    profile.value = prof
    const prevId = selectedNode.value?.id
    if (prevId) {
      selectedNode.value = tl.nodes.find((n) => n.id === prevId) ?? lastNodeFrom(tl.nodes)
    } else {
      selectedNode.value = tl.nodes.length ? lastNodeFrom(tl.nodes) : null
    }
    choicePanelOpen.value = false
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    pageLoading.value = false
  }
}

function lastNodeFrom(list: LifeNode[]) {
  return list.reduce((a, b) => (a.sequence >= b.sequence ? a : b))
}

function onSelect(node: LifeNode) {
  selectedNode.value = node
  if (node.id !== lastNode.value?.id) {
    choicePanelOpen.value = false
  }
  if (layout.isMobile) {
    mobileTab.value = 'detail'
    detailPanelOpen.value = false
    return
  }
  detailPanelOpen.value = true
}

function closeDetailPanel() {
  detailPanelOpen.value = false
  choicePanelOpen.value = false
}

function onAppend() {
  if (!lastNode.value || appendDisabled.value) return
  selectedNode.value = lastNode.value
  choicePanelOpen.value = true
  if (layout.isMobile) {
    mobileTab.value = 'detail'
  } else {
    detailPanelOpen.value = true
  }
}

function onChoiceSimulating(busy: boolean) {
  choiceSimulating.value = busy
  jobBusy.value = busy
}

async function onChoiceDone() {
  choiceSimulating.value = false
  jobBusy.value = false
  await load()
  choicePanelOpen.value = false
  if (lastNode.value) {
    selectedNode.value = lastNode.value
    if (!layout.isMobile) detailPanelOpen.value = true
  }
  ElMessage.success('已推演下一节点')
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
      `将回退到「${node.year} 年 · ${node.title}」，并丢弃其后 ${afterCount} 个节点；档案与世界线将一并恢复。`,
      '回退至此',
      { confirmButtonText: '确认回退', cancelButtonText: '取消', type: 'warning' }
    )
  } catch {
    return
  }
  pageLoading.value = true
  try {
    await api.rollbackToNode(charId, node.id)
    await load()
    ElMessage.success(`已回退至节点 #${node.sequence}`)
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '回退失败')
  } finally {
    pageLoading.value = false
  }
}

function onOpenDialogue() {
  if (!selectedNode.value) return
  dialogueRef.value?.open(
    charId,
    selectedNode.value,
    selectedNode.value.version_id,
    currentVersionId.value
  )
}

function onDialogueNodeUpdated(payload: { versionId: string; node: LifeNode }) {
  const idx = nodes.value.findIndex((n) => n.sequence === payload.node.sequence)
  if (idx >= 0) {
    nodes.value[idx] = payload.node
    selectedNode.value = payload.node
  }
}

async function refreshProfileFromNodes() {
  if (!currentTimeline.value?.id || !nodes.value.length) {
    ElMessage.warning('请先有人生节点')
    return
  }
  jobBusy.value = true
  const ok = await profileJobRunner.run({
    title: '刷新档案经历',
    submitLabel: '正在归纳人生经历…',
    submit: () =>
      api.gameRefreshProfile(charId, {
        model: modelStore.aiModel,
        timeline_id: currentTimeline.value!.id,
      }),
    resubmit: () =>
      api.gameRefreshProfile(charId, {
        model: modelStore.aiModel,
        timeline_id: currentTimeline.value!.id,
      }),
    afterSuccess: async () => {
      profile.value = await api.getProfile(charId)
      ElMessage.success('档案经历已更新为人生总述')
    },
  })
  jobBusy.value = false
  if (!ok && !profileJobRunner.failed.value) {
    ElMessage.error('刷新档案失败')
  }
}

async function refreshWorldLine() {
  if (!currentTimeline.value?.id || !nodes.value.length) {
    ElMessage.warning('当前时间轴尚无节点')
    return
  }
  worldLineRefreshing.value = true
  try {
    await api.refreshWorldLine(charId, currentTimeline.value.id, { model: modelStore.aiModel })
    await load()
    ElMessage.success('世界线已更新')
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '刷新世界线失败')
  } finally {
    worldLineRefreshing.value = false
  }
}

watch(
  () => mobileTab.value,
  (tab) => {
    if (!layout.isMobile) return
    if (tab === 'nodes') void scrollToSelectedNode()
  }
)

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
  if (chronicleFrom.value > chronicleTo.value || !currentVersionId.value) return
  chronicleSubmitting.value = true
  try {
    const ok = await submitChronicleJob({
      versionId: currentVersionId.value,
      fromSequence: chronicleFrom.value,
      toSequence: chronicleTo.value,
      model: modelStore.aiModel,
      onCached: (artifact) => {
        chroniclePickerVisible.value = false
        chronicleTitle.value = `史书 · 节点 ${chronicleFrom.value}–${chronicleTo.value}`
        chronicleContent.value = artifact.content
        chronicleVisible.value = true
      },
    })
    if (ok) chroniclePickerVisible.value = false
  } finally {
    chronicleSubmitting.value = false
  }
}

function viewChronicleJob(entry: ChronicleJobEntry) {
  if (!entry.content) return
  chronicleTitle.value = `史书 · 节点 ${entry.fromSequence}–${entry.toSequence}`
  chronicleContent.value = entry.content
  chronicleVisible.value = true
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
  chronicleVisible.value = false
  await submitChronicleJob({
    versionId: ctx.versionId,
    fromSequence: ctx.fromSequence,
    toSequence: ctx.toSequence,
    model: ctx.model,
    force: true,
    onCached: (artifact) => {
      chronicleTitle.value = `史书 · 节点 ${ctx.fromSequence}–${ctx.toSequence}`
      chronicleContent.value = artifact.content
      chronicleVisible.value = true
    },
  })
}

onMounted(async () => {
  await load()
  await refreshChronicleJobs()
  chronicleJobs.startBackgroundPoll()
})
</script>

<template>
  <div class="timeline-page game-timeline-page">
    <div class="toolbar">
      <div class="toolbar-left">
        <h2>{{ profile?.display_name || '人生游戏' }}</h2>
        <p v-if="currentTimeline" class="timeline-subtitle">{{ currentTimeline.title }}</p>
        <div class="toolbar-tags">
          <el-tag type="warning" size="small">游戏模式</el-tag>
        </div>
        <div class="name-legend">
          <span class="legend-item"><mark class="name-protagonist">主角</mark></span>
          <span class="legend-item"><mark class="name-person">他人</mark></span>
          <span class="legend-item"><mark class="name-place">地名</mark></span>
        </div>
      </div>
      <div class="toolbar-actions">
        <template v-if="layout.isMobile">
          <el-button @click="showProfile = !showProfile">{{ showProfile ? '隐藏' : '档案' }}</el-button>
          <el-dropdown trigger="click">
            <el-button>更多</el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="router.push('/characters')">人物列表</el-dropdown-item>
                <el-dropdown-item
                  @click="router.push(characterMode === 'famous' ? '/game/famous' : '/game/random')"
                >
                  新建游戏
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
        <template v-else>
          <el-button :disabled="!nodes.length || pageLoading" @click="openChroniclePicker">
            编纂史书
          </el-button>
          <el-badge :value="chronicleActiveCount" :hidden="!chronicleActiveCount" type="warning">
            <el-button :disabled="pageLoading" @click="chronicleListVisible = true">史书队列</el-button>
          </el-badge>
          <el-button @click="router.push('/characters')">人物列表</el-button>
          <el-button @click="router.push(characterMode === 'famous' ? '/game/famous' : '/game/random')">
            新建游戏
          </el-button>
          <el-button @click="showProfile = !showProfile">
            {{ showProfile ? '隐藏' : '查看' }}档案
          </el-button>
        </template>
      </div>
    </div>

    <el-collapse-transition>
      <div v-if="showProfile && profile" class="page-card profile-block">
        <div class="profile-section-head">
          <span class="profile-section-head__title">人物档案</span>
          <el-button
            size="small"
            type="primary"
            plain
            :loading="profileJobRunner.visible.value"
            :disabled="!nodes.length || pageLoading || jobBusy"
            @click="refreshProfileFromNodes"
          >
            刷新经历摘要
          </el-button>
        </div>
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
                          :loading="worldLineRefreshing"
                          :disabled="pageLoading || jobBusy"
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
                :game-choices="gameChoices"
                :tail-simulating="choiceSimulating"
                :selected-id="selectedNode?.id"
                :profile="profile"
                :append-disabled="appendDisabled"
                :rollback-enabled="rollbackEnabled"
                always-show-append
                append-button-label="人生抉择"
                @select="onSelect"
                @append="onAppend"
                @rollback="onRollbackToNode"
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
            <template v-else>
              <el-alert
                v-if="lastNode && selectedNode.id !== lastNode.id"
                type="info"
                :closable="false"
                show-icon
                class="game-choice-hint"
                title="人生抉择仅在时间轴最后一个节点进行"
                description="请选中末节点，或点击时间轴下方的「人生抉择」按钮。"
              />
              <NodeEditor
                :character-id="charId"
                :node="selectedNode"
                :profile="profile"
                read-only
                @dialogue="onOpenDialogue"
              />
              <GameChoicePanel
                v-if="lastNode"
                :char-id="charId"
                :node-id="lastNode.id"
                :is-famous="characterMode === 'famous'"
                :visible="showChoiceForNode"
                @simulating="onChoiceSimulating"
                @done="onChoiceDone"
              />
            </template>
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

        <el-tab-pane label="侧栏" name="sidebar">
          <div class="page-card sidebar-stack">
            <section v-if="profile" class="sidebar-section">
              <div class="profile-section-head">
                <h4 class="sidebar-section__title">人物档案</h4>
                <el-button
                  size="small"
                  type="primary"
                  plain
                  :loading="profileJobRunner.visible.value"
                  :disabled="!nodes.length || pageLoading || jobBusy"
                  @click="refreshProfileFromNodes"
                >
                  刷新经历摘要
                </el-button>
              </div>
              <ProfilePanel :profile="profile" />
            </section>
            <WorldLinePanel
              v-if="worldLine && currentTimeline"
              :character-id="charId"
              :timeline-id="currentTimeline.id"
              :world-line="worldLine"
              :nodes="nodes"
              :refreshing="worldLineRefreshing"
              @updated="(wl) => (worldLine = wl)"
              @refresh="load"
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
                        :loading="worldLineRefreshing"
                        :disabled="pageLoading || jobBusy"
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
              :game-choices="gameChoices"
              :tail-simulating="choiceSimulating"
              :selected-id="selectedNode?.id"
              :profile="profile"
              :append-disabled="appendDisabled"
              :rollback-enabled="rollbackEnabled"
              always-show-append
              append-button-label="人生抉择"
              @select="onSelect"
              @append="onAppend"
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
            <el-empty v-if="!selectedNode" description="请选择一个节点" />
            <template v-else>
              <el-alert
                v-if="lastNode && selectedNode.id !== lastNode.id"
                type="info"
                :closable="false"
                show-icon
                class="game-choice-hint"
                title="人生抉择仅在时间轴最后一个节点进行"
                description="请选中末节点，或点击时间轴下方的「人生抉择」按钮。"
              />
              <NodeEditor
                :character-id="charId"
                :node="selectedNode"
                :profile="profile"
                read-only
                @dialogue="onOpenDialogue"
              />
              <GameChoicePanel
                v-if="lastNode"
                :char-id="charId"
                :node-id="lastNode.id"
                :is-famous="characterMode === 'famous'"
                :visible="showChoiceForNode"
                @simulating="onChoiceSimulating"
                @done="onChoiceDone"
              />
            </template>
          </div>
        </div>
      </aside>

      <aside class="col-dock col-dock-side">
        <div class="dock-panel page-card sidebar-stack">
          <section v-if="profile" class="sidebar-section">
            <div class="profile-section-head">
              <h4 class="sidebar-section__title">人物档案</h4>
              <el-button
                size="small"
                type="primary"
                plain
                :loading="profileJobRunner.visible.value"
                :disabled="!nodes.length || pageLoading || jobBusy"
                @click="refreshProfileFromNodes"
              >
                刷新经历摘要
              </el-button>
            </div>
            <ProfilePanel :profile="profile" />
          </section>
          <WorldLinePanel
            v-if="worldLine && currentTimeline"
            :character-id="charId"
            :timeline-id="currentTimeline.id"
            :world-line="worldLine"
            :nodes="nodes"
            :refreshing="worldLineRefreshing"
            @updated="(wl) => (worldLine = wl)"
            @refresh="load"
          />
        </div>
      </aside>
    </div>

    <el-dialog v-model="chroniclePickerVisible" title="编纂史书" width="min(520px, 92vw)" destroy-on-close>
      <p class="export-hint">综合世界线与人生节点撰写史书，后台生成。</p>
      <el-form label-position="top">
        <el-form-item label="起始节点">
          <el-select v-model="chronicleFrom" style="width: 100%">
            <el-option
              v-for="opt in nodeSequenceOptions"
              :key="'g-ch-from-' + opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="结束节点">
          <el-select v-model="chronicleTo" style="width: 100%">
            <el-option
              v-for="opt in nodeSequenceOptions"
              :key="'g-ch-to-' + opt.value"
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

    <LightNovelJobList
      v-model="chronicleListVisible"
      :jobs="chronicleJobList"
      :loading="chronicleListLoading"
      drawer-title="史书编纂队列"
      hint="编纂任务在后台运行。"
      empty-description="暂无编纂任务"
      :show-person="false"
      @refresh="refreshChronicleJobs"
      @view="viewChronicleJob"
      @retry="retryChronicleJob"
    />

    <NarrativeDialog
      :model-value="chronicleVisible"
      :title="chronicleTitle"
      :content="chronicleContent"
      :loading="false"
      @update:model-value="
        (v) => {
          chronicleVisible = v
          if (!v) chronicleViewContext = null
        }
      "
      @regenerate="regenerateChronicle"
    />

    <JobProgress
      :visible="profileJobRunner.visible.value"
      :progress="profileJobRunner.progress.value"
      :status-text="profileJobRunner.statusText.value"
      :title="profileJobRunner.title.value"
      :failed="profileJobRunner.failed.value"
      :error-text="profileJobRunner.errorText.value"
      :retrying="profileJobRunner.retrying.value"
      @retry="profileJobRunner.retry()"
      @dismiss="profileJobRunner.clearState()"
    />

    <CharacterDialogueDialog
      ref="dialogueRef"
      :all-nodes="nodes"
      @node-updated="onDialogueNodeUpdated"
    />
  </div>
</template>

<style scoped>
.export-hint {
  margin: 0 0 12px;
  font-size: 0.85rem;
  color: #909399;
  line-height: 1.5;
}

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
  flex-wrap: wrap;
  gap: 12px;
}

.toolbar h2 {
  margin: 0;
}

.toolbar-left {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.toolbar-tags {
  display: flex;
  gap: 8px;
  align-items: center;
}

.timeline-subtitle {
  margin: 0;
  font-size: 0.85rem;
  color: #909399;
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

.profile-block {
  margin-bottom: 16px;
}

.profile-section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 10px;
}

.profile-section-head__title,
.profile-section-head .sidebar-section__title {
  margin: 0;
  font-size: 0.95rem;
  font-weight: 600;
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

.timeline-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 400px) minmax(0, 300px);
  gap: 16px;
  align-items: stretch;
  transition: grid-template-columns 0.25s ease;
  width: 100%;
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

.sidebar-stack {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.sidebar-section__title {
  margin: 0 0 12px;
  font-size: 0.9rem;
  font-weight: 600;
  color: #303133;
}

.game-choice-hint {
  margin-bottom: 12px;
}

.game-node-detail__title {
  margin: 0 0 12px;
  font-size: 1.05rem;
  line-height: 1.4;
  color: #303133;
}

.read-block {
  margin: 10px 0;
  padding: 10px 12px;
  background: #f8fafc;
  border-radius: 6px;
  border-left: 3px solid #409eff;
}

.read-block.muted {
  border-left-color: #c0c4cc;
  background: #fafafa;
}

.read-label {
  font-size: 0.75rem;
  font-weight: 600;
  color: #909399;
  margin-bottom: 6px;
  letter-spacing: 0.04em;
}

.read-text {
  margin: 0;
  color: #303133;
  font-size: 0.92rem;
  line-height: 1.65;
  white-space: pre-wrap;
  word-break: break-word;
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

@media (max-width: 900px) {
  .timeline-layout,
  .timeline-layout.detail-collapsed {
    grid-template-columns: 1fr;
    gap: 12px;
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
