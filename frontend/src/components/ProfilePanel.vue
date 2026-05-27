<script setup lang="ts">
import type { Profile } from '@/api/client'
import { formatProfileLifeSpan } from '@/utils/timelineDensity'

defineProps<{
  profile: Profile
}>()
</script>

<template>
  <el-descriptions :title="profile.display_name" :column="2" border>
    <el-descriptions-item label="生卒">
      {{ formatProfileLifeSpan(profile) }}
    </el-descriptions-item>
    <el-descriptions-item label="时代">{{ profile.era }}</el-descriptions-item>
    <el-descriptions-item label="国籍">{{ profile.nationality }}</el-descriptions-item>
    <el-descriptions-item label="出生地">{{ profile.birth_place }}</el-descriptions-item>
    <el-descriptions-item label="职业">{{ profile.occupation }}</el-descriptions-item>
    <el-descriptions-item label="阶层">{{ profile.social_class }}</el-descriptions-item>
    <el-descriptions-item label="性格" :span="2">{{ profile.personality_initial }}</el-descriptions-item>
    <el-descriptions-item label="信仰格言" :span="2">{{ profile.beliefs_motto }}</el-descriptions-item>
    <el-descriptions-item label="时代背景" :span="2">{{ profile.era_background }}</el-descriptions-item>
    <el-descriptions-item label="经历" :span="2">{{ profile.experiences }}</el-descriptions-item>
    <el-descriptions-item label="主要成就" :span="2">{{ profile.major_works }}</el-descriptions-item>
    <el-descriptions-item label="文献" :span="2" v-if="profile.literature">{{ profile.literature }}</el-descriptions-item>
    <el-descriptions-item label="关系人" :span="2">{{ profile.family_relations }}</el-descriptions-item>
    <el-descriptions-item label="外貌" :span="2" v-if="profile.appearance">{{ profile.appearance }}</el-descriptions-item>
    <el-descriptions-item label="模板来源" v-if="profile.template_source">{{ profile.template_source }}</el-descriptions-item>
    <el-descriptions-item label="说明" :span="2" v-if="profile.sources_note">
      <el-text type="warning">{{ profile.sources_note }}</el-text>
    </el-descriptions-item>
  </el-descriptions>
</template>
