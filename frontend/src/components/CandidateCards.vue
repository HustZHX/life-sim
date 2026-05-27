<script setup lang="ts">
import type { ResolveCandidate } from '@/api/client'

defineProps<{
  candidates: ResolveCandidate[]
  disabled?: boolean
}>()

const emit = defineEmits<{
  select: [index: number]
}>()
</script>

<template>
  <div class="candidates">
    <el-card
      v-for="(c, i) in candidates"
      :key="i"
      class="candidate"
      :class="{ disabled: disabled }"
      shadow="hover"
      @click="!disabled && emit('select', i)"
    >
      <div class="head">
        <strong>{{ c.name }}</strong>
        <el-tag size="small" type="info">{{ c.nationality }}</el-tag>
        <el-tag size="small">{{ Math.round(c.confidence * 100) }}%</el-tag>
      </div>
      <p class="years" v-if="c.birth_year || c.death_year">
        {{ c.birth_year || '?' }} — {{ c.death_year || '?' }}
      </p>
      <p class="summary">{{ c.summary }}</p>
    </el-card>
  </div>
</template>

<style scoped>
.candidates {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.candidate {
  cursor: pointer;
}
.candidate.disabled {
  cursor: not-allowed;
  opacity: 0.6;
  pointer-events: none;
}
.head {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.years {
  color: #888;
  font-size: 0.85rem;
  margin: 4px 0;
}
.summary {
  margin: 0;
  color: #444;
}
</style>
