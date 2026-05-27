<script setup lang="ts">
import { computed } from 'vue'
import type { EntityHighlightContext } from '@/utils/textEntities'
import { segmentEntities } from '@/utils/textEntities'

const props = withDefaults(
  defineProps<{
    text: string
    context?: EntityHighlightContext
    tag?: keyof HTMLElementTagNameMap
  }>(),
  {
    context: () => ({ protagonist: [], persons: [], places: [] }),
    tag: 'span',
  }
)

const segments = computed(() => segmentEntities(props.text, props.context))
</script>

<template>
  <component :is="tag" class="highlighted-text">
    <template v-for="(seg, i) in segments" :key="i">
      <span v-if="seg.kind === 'text'">{{ seg.value }}</span>
      <mark v-else-if="seg.kind === 'protagonist'" class="name-protagonist" :title="seg.value">{{
        seg.value
      }}</mark>
      <mark v-else-if="seg.kind === 'person'" class="name-person" :title="seg.value">{{
        seg.value
      }}</mark>
      <mark v-else class="name-place" :title="seg.value">{{ seg.value }}</mark>
    </template>
  </component>
</template>

<style scoped>
.highlighted-text {
  line-height: inherit;
}

mark {
  background: none;
  padding: 0;
}

.name-protagonist {
  font-weight: 700;
  color: #b45309;
  background: linear-gradient(180deg, rgba(251, 191, 36, 0.35) 0%, rgba(251, 191, 36, 0.12) 100%);
  border-bottom: 2px solid #f59e0b;
  border-radius: 3px;
  padding: 0 2px;
  box-decoration-break: clone;
  -webkit-box-decoration-break: clone;
}

.name-person {
  font-weight: 600;
  color: #1d4ed8;
  background: rgba(59, 130, 246, 0.12);
  border-radius: 3px;
  padding: 0 2px;
  box-decoration-break: clone;
  -webkit-box-decoration-break: clone;
}

.name-place {
  font-weight: 600;
  color: #047857;
  background: rgba(16, 185, 129, 0.14);
  border-bottom: 1px dashed #10b981;
  border-radius: 3px;
  padding: 0 2px;
  box-decoration-break: clone;
  -webkit-box-decoration-break: clone;
}
</style>
