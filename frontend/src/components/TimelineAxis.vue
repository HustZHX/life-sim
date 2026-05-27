<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { LifeNode, Profile } from '@/api/client'
import { entityContextFromNode } from '@/utils/textEntities'
import HighlightedText from '@/components/HighlightedText.vue'
import NodeSceneMeta from '@/components/NodeSceneMeta.vue'

const props = defineProps<{
  nodes: LifeNode[]
  selectedId?: string
  profile?: Profile | null
  /** 详情面板收起时：宽屏阅读模式，默认展示全文 */
  readingFocus?: boolean
}>()

const emit = defineEmits<{
  select: [node: LifeNode]
}>()

const expandedIds = ref<Set<string>>(new Set())

const showFullContent = (nodeId: string) =>
  props.readingFocus || expandedIds.value.has(nodeId)

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

watch(
  () => props.readingFocus,
  (focus) => {
    if (focus) {
      expandedIds.value = new Set()
    }
  }
)

const axisClass = computed(() => ({
  'timeline-axis--focus': !!props.readingFocus,
}))
</script>

<template>
  <div class="timeline-axis" :class="axisClass">
    <div v-if="nodes.length && !readingFocus" class="axis-toolbar">
      <el-button size="small" text type="primary" @click="expandAll">全部展开</el-button>
      <el-button size="small" text @click="collapseAll">全部收起</el-button>
    </div>
    <p v-else-if="nodes.length && readingFocus" class="focus-hint">
      阅读模式：点击节点查看详情与编辑；全文已展开便于通读。
    </p>

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
          :class="{
            active: selectedId === node.id,
            expanded: showFullContent(node.id),
            'focus-card': readingFocus,
          }"
          shadow="hover"
          @click="emit('select', node)"
        >
          <div class="card-head">
            <h4>
              <HighlightedText :text="node.title" :context="contextFor(node)" tag="span" />
            </h4>
            <el-button
              v-if="!readingFocus"
              size="small"
              text
              type="primary"
              class="expand-btn"
              @click.stop="toggleExpand(node.id)"
            >
              {{ showFullContent(node.id) ? '收起' : '展开阅读' }}
            </el-button>
            <span v-else class="open-hint">点击查看详情 →</span>
          </div>

          <NodeSceneMeta :scene="node.scene" :compact="!readingFocus" />

          <template v-if="showFullContent(node.id)">
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
          </template>
          <p v-else class="preview">
            <HighlightedText
              :text="node.events.slice(0, 120) + (node.events.length > 120 ? '…' : '')"
              :context="contextFor(node)"
              tag="span"
            />
          </p>

          <el-tag v-if="node.trait_changes?.length" size="small" type="warning">性格/思想变更</el-tag>
        </el-card>
      </el-timeline-item>
    </el-timeline>
  </div>
</template>

<style scoped>
.timeline-axis {
  min-width: 0;
}

.timeline-axis--focus {
  max-width: 720px;
  margin: 0 auto;
}

.focus-hint {
  margin: 0 0 16px;
  padding: 10px 14px;
  font-size: 0.85rem;
  color: #606266;
  background: #f0f7ff;
  border-radius: 8px;
  border-left: 3px solid #409eff;
  line-height: 1.5;
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

.node-card.focus-card {
  padding: 2px 0;
}

.node-card.focus-card:hover {
  border-color: #a0cfff;
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

.timeline-axis--focus .node-card h4 {
  font-size: 1.05rem;
}

.expand-btn {
  flex-shrink: 0;
  padding: 0 4px;
}

.open-hint {
  flex-shrink: 0;
  font-size: 0.78rem;
  color: #909399;
  white-space: nowrap;
}

.preview {
  margin: 0 0 8px;
  color: #666;
  font-size: 0.9rem;
  line-height: 1.55;
}

.read-block {
  margin: 10px 0;
  padding: 10px 12px;
  background: #f8fafc;
  border-radius: 6px;
  border-left: 3px solid #409eff;
}

.timeline-axis--focus .read-block {
  padding: 12px 14px;
  margin: 12px 0;
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

.timeline-axis--focus .read-text {
  font-size: 0.96rem;
  line-height: 1.75;
}

.timeline-axis--focus :deep(.el-timeline-item__timestamp) {
  font-size: 0.88rem;
}
</style>
