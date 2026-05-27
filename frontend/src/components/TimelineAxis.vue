<script setup lang="ts">
import type { LifeNode, Profile } from '@/api/client'
import { entityContextFromNode } from '@/utils/textEntities'
import HighlightedText from '@/components/HighlightedText.vue'
import NodeSceneMeta from '@/components/NodeSceneMeta.vue'

const props = defineProps<{
  nodes: LifeNode[]
  selectedId?: string
  profile?: Profile | null
}>()

const emit = defineEmits<{
  select: [node: LifeNode]
}>()

function contextFor(node: LifeNode) {
  return entityContextFromNode(node, props.profile ?? null)
}
</script>

<template>
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
        :class="{ active: selectedId === node.id }"
        shadow="hover"
        @click="emit('select', node)"
      >
        <h4>
          <HighlightedText :text="node.title" :context="contextFor(node)" tag="span" />
        </h4>
        <NodeSceneMeta :scene="node.scene" compact />
        <p class="preview">
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
</template>

<style scoped>
.node-card {
  cursor: pointer;
}
.node-card.active {
  border-color: #409eff;
}
.node-card h4 {
  margin: 0 0 8px;
  line-height: 1.45;
}
.preview {
  margin: 0 0 8px;
  color: #666;
  font-size: 0.9rem;
  line-height: 1.55;
}
</style>
