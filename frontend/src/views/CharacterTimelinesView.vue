<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { api, type Character, type Profile, type Timeline } from '@/api/client'
import { formatDateTime, modeLabel, statusLabel } from '@/constants/characterLabels'

const route = useRoute()
const router = useRouter()
const charId = route.params.id as string

const character = ref<Character | null>(null)
const profile = ref<Profile | null>(null)
const timelines = ref<Timeline[]>([])
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    const [ch, prof, tlRes] = await Promise.all([
      api.getCharacter(charId),
      api.getProfile(charId).catch(() => null),
      api.listTimelines(charId).catch(() => ({ timelines: [] as Timeline[] })),
    ])
    character.value = ch
    profile.value = prof
    timelines.value = tlRes.timelines
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '加载失败')
    router.push('/characters')
  } finally {
    loading.value = false
  }
}

function openTimeline(tl: Timeline) {
  router.push({ path: `/timeline/${charId}`, query: { timeline: tl.id } })
}

function createTimeline() {
  router.push(`/continue/${charId}`)
}

function editProfile() {
  router.push(`/continue/${charId}`)
}

onMounted(load)
</script>

<template>
  <div class="page-card" v-loading="loading">
    <div class="head">
      <div>
        <el-button link class="back-link" @click="router.push('/characters')">← 人物列表</el-button>
        <h2>{{ profile?.display_name || character?.display_name || '人物' }}</h2>
        <div v-if="character" class="meta">
          <el-tag size="small" :type="character.mode === 'famous' ? 'primary' : 'success'">
            {{ modeLabel[character.mode] || character.mode }}
          </el-tag>
          <el-tag v-if="character.status" size="small" type="info">
            {{ statusLabel[character.status] || character.status }}
          </el-tag>
          <span v-if="profile?.era" class="meta-text">{{ profile.era }}</span>
          <span v-if="profile?.birth_year || profile?.death_year" class="meta-text">
            {{ profile?.birth_year || '?' }} — {{ profile?.death_year || '?' }}
          </span>
        </div>
      </div>
      <div class="head-actions">
        <el-button @click="editProfile">编辑档案</el-button>
        <el-button type="primary" @click="createTimeline">新建时间轴</el-button>
      </div>
    </div>

    <el-alert
      v-if="profile"
      :title="profile.personality_initial || '暂无性格摘要'"
      type="info"
      :closable="false"
      show-icon
      class="profile-alert"
    />

    <section class="timeline-section">
      <div class="section-head">
        <h3>时间轴列表（{{ timelines.length }}）</h3>
        <el-button :loading="loading" @click="load">刷新</el-button>
      </div>

      <el-empty v-if="!timelines.length && !loading" description="该人物尚无时间轴">
        <el-button type="primary" @click="createTimeline">生成第一条时间轴</el-button>
      </el-empty>

      <el-table
        v-else
        :data="timelines"
        stripe
        style="width: 100%"
        @row-click="openTimeline"
      >
        <el-table-column label="名称" min-width="220" prop="title" />
        <el-table-column label="节点数" width="90" align="center">
          <template #default="{ row }">
            {{ row.node_count ?? '—' }}
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="更新时间" width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.updated_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click.stop="openTimeline(row)">查看</el-button>
          </template>
        </el-table-column>
      </el-table>
    </section>
  </div>
</template>

<style scoped>
.head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 20px;
}
.back-link {
  padding: 0;
  margin-bottom: 4px;
  font-size: 0.85rem;
}
.head h2 {
  margin: 0;
}
.meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
}
.meta-text {
  font-size: 0.85rem;
  color: #606266;
}
.head-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}
.profile-alert {
  margin-bottom: 20px;
}
.timeline-section h3 {
  margin: 0;
  font-size: 1rem;
}
.section-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}
:deep(.el-table__row) {
  cursor: pointer;
}
</style>
