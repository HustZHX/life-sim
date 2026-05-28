<script setup lang="ts">
import { computed } from 'vue'
import { ElMessage } from 'element-plus'
import { useLayoutStore } from '@/stores/layout'

const layout = useLayoutStore()

const label = computed(() => (layout.mode === 'mobile' ? '手机' : '桌面'))

function toggle() {
  layout.toggleMode()
  ElMessage.success(`已切换到${layout.mode === 'mobile' ? '手机' : '桌面'}布局`)
}
</script>

<template>
  <el-button size="small" text @click="toggle">布局：{{ label }}</el-button>
</template>
