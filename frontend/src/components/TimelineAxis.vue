<script setup lang="ts">
import { computed, ref } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import type { LifeNode, Profile } from '@/api/client'
import { fieldLabel } from '@/constants/fieldLabels'
import { entityContextFromNode } from '@/utils/textEntities'
import { isLivingProfile } from '@/utils/timelineDensity'
import HighlightedText from '@/components/HighlightedText.vue'
import NodeSceneMeta from '@/components/NodeSceneMeta.vue'

const props = defineProps<{
  nodes: LifeNode[]
  selectedId?: string
  profile?: Profile | null
  /** 继续推演进行中或时间轴生成中 */
  appendDisabled?: boolean
  /** 是否允许回退至此（当前激活分支且未在任务中） */
  rollbackEnabled?: boolean
}>()

const emit = defineEmits<{
  select: [node: LifeNode]
  append: []
  rollback: [node: LifeNode]
}>()

const showAppendSlot = computed(
  () => !!props.profile && isLivingProfile(props.profile) && props.nodes.length > 0
)

const expandedIds = ref<Set<string>>(new Set())

const showFullContent = (nodeId: string) => expandedIds.value.has(nodeId)

function contextFor(node: LifeNode) {
  return entityContextFromNode(node, props.profile ?? null)
}

function toggleExpand(nodeId: string) {
  const next = new Set(expandedIds.value)
  if (next.has(nodeId)) {
    next.delete(nodeId)
  } else {
    next.add(nodeId)
  }
  expandedIds.value = next
}

function expandAll() {
  expandedIds.value = new Set(props.nodes.map((n) => n.id))
}

function collapseAll() {
  expandedIds.value = new Set()
}

const maxSequence = computed(() => {
  if (!props.nodes.length) return 0
  return props.nodes.reduce((max, n) => (n.sequence > max ? n.sequence : max), 0)
})

function canRollbackTo(node: LifeNode) {
  return !!props.rollbackEnabled && node.sequence < maxSequence.value
}
</script>

<template>
  <div class="timeline-axis">
    <div v-if="nodes.length" class="axis-toolbar">
      <el-button size="small" text type="primary" @click="expandAll">全部展开</el-button>
      <el-button size="small" text @click="collapseAll">全部收起</el-button>
    </div>

    <el-timeline>
      <el-timeline-item
        v-for="node in nodes"
        :key="node.id"
        :timestamp="`${node.year} 年 · ${node.age} 岁`"
        placement="top"
        :type="node.trait_changes?.length ? 'warning' : 'primary'"
        :hollow="selectedId !== node.id"
      >
        <el-card
          class="node-card"
          :class="{ active: selectedId === node.id, expanded: showFullContent(node.id) }"
          shadow="hover"
          @click="emit('select', node)"
        >
          <div class="card-head">
            <h4>
              <HighlightedText :text="node.title" :context="contextFor(node)" tag="span" />
            </h4>
            <div class="card-actions">
              <el-button
                v-if="canRollbackTo(node)"
                size="small"
                text
                type="danger"
                class="rollback-btn"
                @click.stop="emit('rollback', node)"
              >
                回退至此
              </el-button>
              <el-button
                size="small"
                text
                type="primary"
                class="expand-btn"
                @click.stop="toggleExpand(node.id)"
              >
                {{ showFullContent(node.id) ? '收起' : '展开' }}
              </el-button>
            </div>
          </div>

          <template v-if="showFullContent(node.id)">
            <NodeSceneMeta :scene="node.scene" />
            <div class="read-block">
              <div class="read-label">经历</div>
              <p class="read-text">
                <HighlightedText :text="node.events" :context="contextFor(node)" tag="span" />
              </p>
            </div>
            <div v-if="node.thoughts" class="read-block muted">
              <div class="read-label">内心</div>
              <p class="read-text">
                <HighlightedText :text="node.thoughts" :context="contextFor(node)" tag="span" />
              </p>
            </div>
            <div v-if="node.personality_snapshot" class="read-block muted">
              <div class="read-label">性格</div>
              <p class="read-text">
                <HighlightedText
                  :text="node.personality_snapshot"
                  :context="contextFor(node)"
                  tag="span"
                />
              </p>
            </div>
            <div v-if="node.trait_changes?.length" class="trait-section">
              <div class="read-label">性格/思想变更</div>
              <ul class="trait-list">
                <li v-for="(t, i) in node.trait_changes" :key="i" class="trait-item">
                  <strong>{{ fieldLabel(t.field, 'trait') }}</strong>：
                  <HighlightedText
                    :text="`${t.before} → ${t.after}`"
                    :context="contextFor(node)"
                    tag="span"
                  />
                  <p v-if="t.reason" class="trait-reason">
                    <HighlightedText :text="t.reason" :context="contextFor(node)" tag="span" />
                  </p>
                </li>
              </ul>
            </div>
          </template>
          <ul v-else-if="node.trait_changes?.length" class="trait-compact">
            <li v-for="(t, i) in node.trait_changes" :key="i">
              <span class="trait-field">{{ fieldLabel(t.field, 'trait') }}</span>：
              <HighlightedText
                :text="`${t.before} → ${t.after}`"
                :context="contextFor(node)"
                tag="span"
              />
            </li>
          </ul>
        </el-card>
      </el-timeline-item>

      <el-timeline-item v-if="showAppendSlot" placement="top" type="info" hollow>
        <button
          type="button"
          class="append-next-btn"
          :disabled="appendDisabled"
          :title="appendDisabled ? '请等待当前任务完成' : '继续推演下一个节点'"
          @click="emit('append')"
        >
          <el-icon :size="22"><Plus /></el-icon>
          <span>继续推演</span>
        </button>
      </el-timeline-item>
    </el-timeline>
  </div>
</template>

<style scoped>
.timeline-axis {
  min-width: 0;
}

.axis-toolbar {
  display: flex;
  justify-content: flex-end;
  gap: 4px;
  margin-bottom: 8px;
}

.node-card {
  cursor: pointer;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.node-card.active {
  border-color: #409eff;
  box-shadow: 0 2px 12px rgba(64, 158, 255, 0.15);
}

.node-card.expanded {
  border-color: #79bbff;
}

.card-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 4px;
}

.node-card h4 {
  margin: 0;
  line-height: 1.45;
  flex: 1;
  min-width: 0;
}

.expand-btn {
  flex-shrink: 0;
  padding: 0 4px;
}

.card-actions {
  display: flex;
  align-items: center;
  gap: 2px;
  flex-shrink: 0;
}

.rollback-btn {
  padding: 0 4px;
}

.trait-compact {
  margin: 8px 0 0;
  padding: 8px 10px;
  list-style: none;
  background: #fffbf0;
  border-radius: 6px;
  border-left: 3px solid #e6a23c;
  font-size: 0.85rem;
  line-height: 1.5;
  color: #606266;
}

.trait-compact li + li {
  margin-top: 6px;
}

.trait-field {
  font-weight: 600;
  color: #303133;
}

.trait-section {
  margin-top: 10px;
}

.trait-list {
  margin: 6px 0 0;
  padding: 0;
  list-style: none;
}

.trait-item {
  margin: 0 0 8px;
  padding: 8px 10px;
  background: #fffbf0;
  border-radius: 6px;
  font-size: 0.88rem;
  line-height: 1.5;
  color: #606266;
}

.trait-item strong {
  color: #303133;
}

.trait-reason {
  margin: 4px 0 0;
  font-size: 0.82rem;
  color: #909399;
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

.append-next-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: 100%;
  min-height: 56px;
  padding: 12px 16px;
  border: 2px dashed #c0c4cc;
  border-radius: 8px;
  background: #fafbfc;
  color: #409eff;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: border-color 0.2s, background 0.2s, color 0.2s;
}

.append-next-btn:hover:not(:disabled) {
  border-color: #409eff;
  background: #ecf5ff;
  color: #337ecc;
}

.append-next-btn:disabled {
  cursor: not-allowed;
  opacity: 0.55;
  color: #909399;
}
</style>
