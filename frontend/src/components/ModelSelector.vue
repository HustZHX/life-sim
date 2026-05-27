<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ArrowDown, ArrowUp } from '@element-plus/icons-vue'
import { api, type AIModel } from '@/api/client'
import { DEFAULT_AI_MODEL, isAIModelId, normalizeAIModel, type AIModelId } from '@/constants/models'

const props = withDefaults(
  defineProps<{
    label?: string
  }>(),
  {}
)
const model = defineModel<AIModelId>({ default: DEFAULT_AI_MODEL })
const models = ref<AIModel[]>([])
const loading = ref(false)
const expanded = ref(false)

const current = computed(() => models.value.find((m) => m.id === model.value))

onMounted(async () => {
  loading.value = true
  try {
    const res = await api.listModels()
    models.value = res.models.filter((m) => isAIModelId(m.id))
    model.value = normalizeAIModel(model.value)
  } finally {
    loading.value = false
  }
})

function toggle() {
  expanded.value = !expanded.value
}

function onSelect() {
  expanded.value = false
}
</script>

<template>
  <div v-loading="loading" class="model-selector">
    <div class="label">{{ props.label || 'AI 模型' }}</div>

    <button type="button" class="model-trigger" :class="{ open: expanded }" @click="toggle">
      <span class="trigger-main">
        <span class="trigger-name">{{ current?.label || 'V4 Flash (推荐)' }}</span>
        <span v-if="current?.est_time && !expanded" class="trigger-est">{{ current.est_time }}</span>
      </span>
      <span class="trigger-action">
        {{ expanded ? '收起' : '选择模型' }}
        <el-icon><ArrowUp v-if="expanded" /><ArrowDown v-else /></el-icon>
      </span>
    </button>

    <Transition name="model-panel">
      <el-radio-group v-if="expanded" v-model="model" class="model-group" @change="onSelect">
        <el-radio
          v-for="m in models"
          :key="m.id"
          :value="m.id"
          border
          class="model-item"
        >
          <div class="model-head">
            <strong>{{ m.label }}</strong>
            <el-tag size="small" type="info">{{ m.speed }}</el-tag>
            <el-tag size="small">{{ m.quality }}</el-tag>
          </div>
          <p class="desc">{{ m.description }}</p>
          <p class="est">{{ m.est_time }}</p>
        </el-radio>
      </el-radio-group>
    </Transition>
  </div>
</template>

<style scoped>
.model-selector {
  margin-bottom: 16px;
}

.label {
  font-weight: 600;
  margin-bottom: 8px;
  color: #333;
}

.model-trigger {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  width: 100%;
  padding: 10px 14px;
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  background: #fafbfc;
  cursor: pointer;
  text-align: left;
  transition: border-color 0.2s, background 0.2s;
}

.model-trigger:hover,
.model-trigger.open {
  border-color: #409eff;
  background: #f5f9ff;
}

.trigger-main {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.trigger-name {
  font-weight: 600;
  color: #303133;
}

.trigger-est {
  font-size: 0.8rem;
  color: #909399;
}

.trigger-action {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
  font-size: 0.85rem;
  color: #409eff;
}

.model-group {
  display: flex;
  flex-direction: column;
  gap: 10px;
  width: 100%;
  margin-top: 10px;
}

.model-item {
  width: 100%;
  height: auto;
  padding: 12px 16px;
  margin-right: 0;
  align-items: flex-start;
}

.model-head {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.desc {
  margin: 6px 0 4px;
  font-size: 0.85rem;
  color: #666;
  line-height: 1.4;
  white-space: normal;
}

.est {
  margin: 0;
  font-size: 0.8rem;
  color: #409eff;
}

:deep(.el-radio__label) {
  width: 100%;
}

.model-panel-enter-active,
.model-panel-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.model-panel-enter-from,
.model-panel-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}
</style>
