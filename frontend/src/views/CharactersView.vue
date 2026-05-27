<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { api, type HistoryItem } from '@/api/client'
import { formatDateTime, modeLabel, statusLabel } from '@/constants/characterLabels'

const router = useRouter()
const items = ref<HistoryItem[]>([])
const loading = ref(false)
const filterMode = ref('')
const keyword = ref('')

const filteredItems = computed(() => {
  let list = items.value
  if (filterMode.value) {
    list = list.filter((i) => i.mode === filterMode.value)
  }
  const q = keyword.value.trim().toLowerCase()
  if (q) {
    list = list.filter(
      (i) =>
        i.display_name.toLowerCase().includes(q) ||
        (i.era && i.era.toLowerCase().includes(q)) ||
        (i.resolve_query && i.resolve_query.toLowerCase().includes(q))
    )
  }
  return list
})

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

function openCharacter(row: HistoryItem) {
  if (row.status === 'profile_ready' || row.status === 'timeline_ready') {
    router.push(`/characters/${row.id}`)
    return
  }
  if (row.status === 'confirmed' || row.status === 'resolved') {
    router.push(row.mode === 'famous' ? `/famous?id=${row.id}` : `/random?id=${row.id}`)
    return
  }
  ElMessage.info('该人物尚未完成档案，请重新创建或继续编辑')
}

onMounted(load)
</script>

<template>
  <div class="page-card">
    <div class="head">
      <div>
        <h2>人物列表</h2>
        <p class="hint">一个人物可拥有多条独立时间轴，点击进入查看该人物的所有时间轴</p>
      </div>
      <div class="head-actions">
        <el-button @click="router.push('/')">新建人物</el-button>
      </div>
    </div>

    <div class="toolbar">
      <el-radio-group v-model="filterMode">
        <el-radio-button value="">全部</el-radio-button>
        <el-radio-button value="famous">名人</el-radio-button>
        <el-radio-button value="random">随机</el-radio-button>
      </el-radio-group>
      <div class="toolbar-right">
        <el-input
          v-model="keyword"
          placeholder="搜索姓名、时代…"
          clearable
          style="width: 220px"
        />
        <el-button :loading="loading" @click="load">刷新</el-button>
      </div>
    </div>

    <el-table
      v-loading="loading"
      :data="filteredItems"
      stripe
      style="width: 100%; margin-top: 16px"
      empty-text="暂无人物，去首页创建一位吧"
      @row-click="openCharacter"
    >
      <el-table-column label="人物" min-width="140">
        <template #default="{ row }">
          <strong>{{ row.display_name || '未命名' }}</strong>
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
      <el-table-column label="时间轴" width="80" align="center">
        <template #default="{ row }">
          {{ row.timeline_count && row.timeline_count > 0 ? row.timeline_count : '—' }}
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
            :type="
              row.status === 'timeline_ready'
                ? 'success'
                : row.status === 'profile_ready'
                  ? 'warning'
                  : 'info'
            "
          >
            {{ statusLabel[row.status] || row.status }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="更新时间" width="180">
        <template #default="{ row }">
          {{ formatDateTime(row.updated_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="100" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" link @click.stop="openCharacter(row)">打开</el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<style scoped>
.head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
}
.head h2 {
  margin: 0;
}
.hint {
  color: #888;
  font-size: 0.9rem;
  margin: 8px 0 0;
}
.head-actions {
  flex-shrink: 0;
}
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 16px;
}
.toolbar-right {
  display: flex;
  gap: 8px;
  align-items: center;
}
:deep(.el-table__row) {
  cursor: pointer;
}
</style>
