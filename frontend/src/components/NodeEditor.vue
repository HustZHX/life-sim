<script setup lang="ts">

import { computed, ref, watch } from 'vue'

import type { AIModelId, LifeNode, Profile, TimelineConfig } from '@/api/client'

import type { EntityHighlightContext } from '@/utils/textEntities'

import { entityContextFromNode } from '@/utils/textEntities'

import ModelSelector from '@/components/ModelSelector.vue'

import HighlightedText from '@/components/HighlightedText.vue'

import TimelineConfigPanel from '@/components/TimelineConfigPanel.vue'

import NodeSceneMeta from '@/components/NodeSceneMeta.vue'

import { DEFAULT_TARGET_NODE_COUNT } from '@/utils/timelineDensity'

import { DEFAULT_AI_MODEL } from '@/constants/models'

import { fieldLabel } from '@/constants/fieldLabels'

import JobProgress from '@/components/JobProgress.vue'



const props = defineProps<{

  node: LifeNode | null

  profile?: Profile | null

  loading?: boolean

  regenProgress?: number

  regenStatus?: string

  regenVisible?: boolean

}>()



const emit = defineEmits<{

  save: [

    payload: {

      mode: string

      model: AIModelId

      patch: Record<string, string>

      target_node_count?: number

    },

  ]

}>()



const form = ref({

  title: '',

  events: '',

  thoughts: '',

  personality_snapshot: '',

})

const model = ref<AIModelId>(DEFAULT_AI_MODEL)

const regenConfig = ref<TimelineConfig>({ target_node_count: DEFAULT_TARGET_NODE_COUNT, start_year: 0, end_year: 0 })



const entityContext = computed<EntityHighlightContext>(() =>

  entityContextFromNode(props.node, props.profile ?? null)

)



watch(

  () => props.node,

  (n) => {

    if (n) {

      form.value = {

        title: n.title,

        events: n.events,

        thoughts: n.thoughts,

        personality_snapshot: n.personality_snapshot,

      }

    }

  },

  { immediate: true }

)



watch(

  () => props.profile,

  (p) => {

    if (p) {

      regenConfig.value.start_year = p.birth_year

      regenConfig.value.end_year = p.death_year

    }

  },

  { immediate: true }

)



function buildPatch() {

  return { ...form.value }

}



/** 根据当前经历，仅重算本节点的想法、性格与变更 */

function syncCurrentInner() {

  emit('save', { mode: 'inner_current', model: model.value, patch: buildPatch() })

}



/** 完全重算：保存本节点后，由 AI 推演后续所有人生节点 */

function regenerateAllSubsequent() {

  emit('save', {

    mode: 'full_cascade',

    model: model.value,

    patch: buildPatch(),

    target_node_count: regenConfig.value.target_node_count,

  })

}

</script>



<template>

  <div v-if="node" class="editor">

    <h3>编辑节点 · {{ node.year }} 年</h3>
    <NodeSceneMeta :scene="node.scene" />

    <el-form label-position="top">

      <el-form-item label="标题">

        <el-input v-model="form.title" />

      </el-form-item>



      <el-form-item label="经历">

        <el-input v-model="form.events" type="textarea" :rows="4" />

      </el-form-item>



      <div class="derived-box">

        <div class="derived-head">

          <span class="derived-title">根据经历推演</span>

        </div>



        <el-form-item label="内心想法">

          <el-input v-model="form.thoughts" type="textarea" :rows="3" />

        </el-form-item>

        <el-form-item label="性格快照">

          <el-input v-model="form.personality_snapshot" type="textarea" :rows="2" />

        </el-form-item>



        <div v-if="node.trait_changes?.length" class="traits-inline">

          <div class="trait-label">性格/思想变更</div>

          <el-card v-for="(t, i) in node.trait_changes" :key="i" size="small" class="trait-card">

            <strong>{{ fieldLabel(t.field, 'trait') }}</strong>：

            <HighlightedText :text="`${t.before} → ${t.after}`" :context="entityContext" tag="span" />

            <p class="reason">

              <HighlightedText :text="t.reason" :context="entityContext" tag="span" />

            </p>

          </el-card>

        </div>

      </div>



      <ModelSelector v-model="model" />



      <TimelineConfigPanel
        v-if="profile"
        v-model="regenConfig"
        :profile="profile"
        :model="model"
        :anchor-year="node.year"
        density-only
        :disabled="loading"
      />



      <div class="actions">

        <el-button type="primary" :loading="loading" @click="syncCurrentInner">

          根据经历更新本节点想法/性格

        </el-button>

        <el-button type="warning" plain :loading="loading" @click="regenerateAllSubsequent">

          完全重算后续（重算寿命 + 全新时间轴）

        </el-button>

      </div>



      <el-alert

        title="「更新本节点」只改当前节点的想法与性格。「完全重算」会先根据新经历重算人物寿命，再按目标节点数（默认约 20 个）全新生成后续所有节点（时间、经历、性格均不沿用旧轴）。"

        type="info"

        :closable="false"

        show-icon

      />

    </el-form>



    <JobProgress

      :visible="!!regenVisible"

      :progress="regenProgress ?? 0"

      :status-text="regenStatus ?? ''"

      title="任务进度"

    />

  </div>

  <el-empty v-else description="点击左侧节点查看或编辑" />

</template>



<style scoped>

.editor h3 {

  margin-top: 0;

}

.derived-box {

  border: 2px solid #dcdfe6;

  border-radius: 8px;

  padding: 12px 14px 4px;

  margin-bottom: 8px;

  background: #fafbfc;

}

.derived-head {

  margin-bottom: 8px;

}

.derived-title {
  font-weight: 600;
  color: #303133;
  font-size: 0.9rem;
}
.traits-inline {

  margin-bottom: 8px;

}

.trait-label {

  font-size: 0.85rem;

  color: #606266;

  margin-bottom: 6px;

}

.trait-card {

  margin-bottom: 6px;

}

.reason {

  margin: 4px 0 0;

  color: #888;

  font-size: 0.85rem;

}

.actions {

  display: flex;

  flex-direction: column;

  gap: 10px;

}

</style>

