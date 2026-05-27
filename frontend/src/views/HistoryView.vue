<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { api, type HistoryItem } from '@/api/client'

const router = useRouter()
const items = ref<HistoryItem[]>([])
const loading = ref(false)
const filterMode = ref('')

const modeLabel: Record<string, string> = {
  famous: '历史名人',
  random: '随机人物',
}

const statusLabel: Record<string, string> = {
  draft: '草稿',
  resolved: '待确认',
  confirmed: '已确认',
  profile_ready: '档案已生成',
  timeline_ready: '时间轴已生成',
}

async function load() {
  loading.value = true
  try {
    const res = await api.listHistory()
    items.value = res.items
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

function filteredItems() {
  if (!filterMode.value) return items.value
  return items.value.filter((i) => i.mode === filterMode.value)
}

function openRecord(row: HistoryItem) {
  if (row.status === 'timeline_ready') {
    router.push(`/timeline/${row.id}`)
    return
  }
  if (row.status === 'profile_ready') {
    router.push(`/continue/${row.id}`)
    return
  }
  ElMessage.info('该记录尚未生成完整档案，请重新创建')
}

function formatTime(t: string) {
  return new Date(t).toLocaleString('zh-CN')
}

onMounted(load)
</script>

<template>
  <div class="page-card">
    <div class="head">
      <h2>生成历史</h2>
      <el-button @click="router.push('/')">返回首页</el-button>
    </div>
    <p class="hint">点击记录可跳转到对应的人生时间轴或继续生成</p>

    <div class="toolbar">
      <el-radio-group v-model="filterMode">
        <el-radio-button value="">全部</el-radio-button>
        <el-radio-button value="famous">名人</el-radio-button>
        <el-radio-button value="random">随机</el-radio-button>
      </el-radio-group>
      <el-button :loading="loading" @click="load">刷新</el-button>
    </div>

    <el-table
      v-loading="loading"
      :data="filteredItems()"
      stripe
      style="width: 100%; margin-top: 16px"
      empty-text="暂无生成记录，去首页创建一个人物吧"
      @row-click="openRecord"
    >
      <el-table-column label="人物" min-width="140">
        <template #default="{ row }">
          <strong>{{ row.display_name }}</strong>
        </template>
      </el-table-column>
      <el-table-column label="类型" width="100">
        <template #default="{ row }">
          <el-tag size="small" :type="row.mode === 'famous' ? 'primary' : 'success'">
            {{ modeLabel[row.mode] || row.mode }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="时代" width="100" prop="era" />
      <el-table-column label="生卒" width="120">
        <template #default="{ row }">
          <span v-if="row.birth_year || row.death_year">
            {{ row.birth_year || '?' }} — {{ row.death_year || '?' }}
          </span>
          <span v-else>—</span>
        </template>
      </el-table-column>
      <el-table-column label="节点数" width="80" align="center">
        <template #default="{ row }">
          {{ row.node_count > 0 ? row.node_count : '—' }}
        </template>
      </el-table-column>
      <el-table-column label="状态" width="130">
        <template #default="{ row }">
          <el-tag
            size="small"
            :type="row.status === 'timeline_ready' ? 'success' : row.status === 'profile_ready' ? 'warning' : 'info'"
          >
            {{ statusLabel[row.status] || row.status }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="更新时间" width="180">
        <template #default="{ row }">
          {{ formatTime(row.updated_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="100" fixed="right">
        <template #default="{ row }">
          <el-button
            v-if="row.status === 'timeline_ready' || row.status === 'profile_ready'"
            type="primary"
            link
            @click.stop="openRecord(row)"
          >
            打开
          </el-button>
          <span v-else class="muted">—</span>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<style scoped>
.head {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.head h2 {
  margin: 0;
}
.hint {
  color: #888;
  font-size: 0.9rem;
  margin: 8px 0 0;
}
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 16px;
}
:deep(.el-table__row) {
  cursor: pointer;
}
.muted {
  color: #ccc;
}
</style>
