<script setup lang="ts">
import { computed } from 'vue'
import type { NodeScene } from '@/api/client'
import { sceneDisplayItems } from '@/utils/nodeScene'

const props = withDefaults(
  defineProps<{
    scene?: NodeScene | null
    compact?: boolean
  }>(),
  { compact: false }
)

const items = computed(() => sceneDisplayItems(props.scene))
</script>

<template>
  <div v-if="items.length" class="node-scene" :class="{ compact }">
    <span v-for="item in items" :key="item.key" class="scene-chip">
      <span class="chip-label">{{ item.label }}</span>
      <span class="chip-value">{{ item.value }}</span>
    </span>
  </div>
</template>

<style scoped>
.node-scene {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 8px;
}
.node-scene.compact {
  margin-bottom: 6px;
}
.scene-chip {
  display: inline-flex;
  align-items: baseline;
  gap: 4px;
  max-width: 100%;
  padding: 2px 8px;
  border-radius: 4px;
  background: #f0f9ff;
  border: 1px solid #dbeafe;
  font-size: 0.78rem;
  line-height: 1.4;
  word-break: break-word;
}
.chip-label {
  color: #64748b;
  flex-shrink: 0;
}
.chip-value {
  color: #334155;
}
</style>
