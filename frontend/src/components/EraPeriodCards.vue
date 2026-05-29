<script setup lang="ts">
import type { GameEraOption } from '@/api/client'

defineProps<{
  options: GameEraOption[]
  selectedLabel?: string
  disabled?: boolean
}>()

const emit = defineEmits<{
  select: [option: GameEraOption]
}>()
</script>

<template>
  <div class="era-cards">
    <button
      v-for="opt in options"
      :key="opt.label"
      type="button"
      class="era-card"
      :class="{ selected: selectedLabel === opt.label }"
      :disabled="disabled"
      @click="emit('select', opt)"
    >
      <div class="era-head">
        <strong>{{ opt.label }}</strong>
        <span class="year-range">{{ opt.year_range }}</span>
      </div>
      <p class="era-desc">{{ opt.description }}</p>
    </button>
  </div>
</template>

<style scoped>
.era-cards {
  display: grid;
  gap: 12px;
}
.era-card {
  text-align: left;
  padding: 14px 16px;
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  background: #fff;
  cursor: pointer;
  transition: border-color 0.2s, box-shadow 0.2s;
}
.era-card:hover:not(:disabled) {
  border-color: #409eff;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.15);
}
.era-card.selected {
  border-color: #409eff;
  background: #ecf5ff;
}
.era-card:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.era-head {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: 8px;
  margin-bottom: 6px;
}
.year-range {
  font-size: 0.85rem;
  color: #909399;
}
.era-desc {
  margin: 0;
  font-size: 0.9rem;
  color: #606266;
  line-height: 1.5;
}
</style>
