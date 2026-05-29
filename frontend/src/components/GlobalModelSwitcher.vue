<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ArrowDown, ArrowUp } from '@element-plus/icons-vue'
import { storeToRefs } from 'pinia'
import { useModelStore } from '@/stores/model'
import { isAIModelId } from '@/constants/models'

const modelStore = useModelStore()
const { aiModel, models, loading } = storeToRefs(modelStore)
const expanded = ref(false)

const current = computed(() => models.value.find((m) => m.id === aiModel.value))

onMounted(() => {
  void modelStore.loadModels()
})

function toggle() {
  expanded.value = !expanded.value
}

function select(id: string) {
  if (!isAIModelId(id)) return
  modelStore.setModel(id)
  expanded.value = false
}
</script>

<template>
  <div v-loading="loading" class="global-model-switcher">
    <button
      type="button"
      class="global-model-trigger"
      :class="{ open: expanded }"
      :title="'当前模型：' + (current?.label || modelStore.currentLabel)"
      @click="toggle"
    >
      <span class="trigger-prefix">模型</span>
      <span class="trigger-name">{{ current?.label || modelStore.currentLabel }}</span>
      <el-icon class="trigger-icon"><ArrowUp v-if="expanded" /><ArrowDown v-else /></el-icon>
    </button>
    <Transition name="global-model-panel">
      <div v-if="expanded" class="global-model-panel">
        <button
          v-for="m in models"
          :key="m.id"
          type="button"
          class="model-option"
          :class="{ active: m.id === aiModel }"
          @click="select(m.id)"
        >
          <div class="model-option-head">
            <strong>{{ m.label }}</strong>
            <el-tag size="small" type="info">{{ m.speed }}</el-tag>
          </div>
          <p class="model-option-desc">{{ m.description }}</p>
          <p class="model-option-est">{{ m.est_time }}</p>
        </button>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.global-model-switcher {
  position: relative;
}

.global-model-trigger {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  border: 1px solid rgba(232, 213, 183, 0.45);
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.08);
  color: #e8d5b7;
  cursor: pointer;
  font-size: 0.85rem;
  transition: background 0.15s, border-color 0.15s;
}

.global-model-trigger:hover,
.global-model-trigger.open {
  background: rgba(255, 255, 255, 0.14);
  border-color: rgba(232, 213, 183, 0.75);
  color: #fff;
}

.trigger-prefix {
  opacity: 0.75;
  font-size: 0.8rem;
}

.trigger-name {
  font-weight: 600;
  max-width: 140px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.trigger-icon {
  font-size: 0.75rem;
  opacity: 0.85;
}

.global-model-panel {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  z-index: 2000;
  width: min(320px, calc(100vw - 32px));
  padding: 10px;
  border-radius: 8px;
  background: #fff;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.18);
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.model-option {
  text-align: left;
  padding: 10px 12px;
  border: 1px solid #dcdfe6;
  border-radius: 6px;
  background: #fafbfc;
  cursor: pointer;
  transition: border-color 0.15s, background 0.15s;
}

.model-option:hover {
  border-color: #409eff;
  background: #f5f9ff;
}

.model-option.active {
  border-color: #409eff;
  background: #ecf5ff;
  box-shadow: 0 0 0 1px #409eff;
}

.model-option-head {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.model-option-desc {
  margin: 6px 0 4px;
  font-size: 0.8rem;
  color: #606266;
  line-height: 1.4;
}

.model-option-est {
  margin: 0;
  font-size: 0.75rem;
  color: #409eff;
}

.global-model-panel-enter-active,
.global-model-panel-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}

.global-model-panel-enter-from,
.global-model-panel-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

.app-shell--mobile .trigger-prefix {
  display: none;
}

.app-shell--mobile .trigger-name {
  max-width: 88px;
}
</style>
