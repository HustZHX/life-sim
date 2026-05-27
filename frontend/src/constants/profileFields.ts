import type { Profile } from '@/api/client'

export type ProfileFieldKey = keyof Omit<Profile, 'character_id' | 'template_source' | 'raw_json'>

export interface ProfileFieldDef {
  key: ProfileFieldKey
  label: string
  type: 'text' | 'number' | 'textarea'
  span?: number
  randomizable?: boolean
}

export const PROFILE_FIELDS: ProfileFieldDef[] = [
  { key: 'display_name', label: '姓名', type: 'text' },
  { key: 'birth_year', label: '出生年', type: 'number' },
  { key: 'death_year', label: '卒年', type: 'number' },
  { key: 'era', label: '时代', type: 'text' },
  { key: 'nationality', label: '国籍', type: 'text' },
  { key: 'birth_place', label: '出生地', type: 'text' },
  { key: 'occupation', label: '职业', type: 'text' },
  { key: 'social_class', label: '阶层', type: 'text' },
  { key: 'education', label: '教育', type: 'text' },
  { key: 'personality_initial', label: '性格', type: 'textarea', span: 2 },
  { key: 'beliefs_motto', label: '信仰格言', type: 'textarea', span: 2 },
  { key: 'beliefs_politics', label: '信仰/政治', type: 'textarea', span: 2 },
  { key: 'era_background', label: '时代背景', type: 'textarea', span: 2 },
  { key: 'experiences', label: '经历', type: 'textarea', span: 2 },
  { key: 'family_relations', label: '关系人', type: 'textarea', span: 2 },
  { key: 'appearance', label: '外貌', type: 'textarea', span: 2 },
  { key: 'major_works', label: '主要成就', type: 'textarea', span: 2 },
  { key: 'literature', label: '文献', type: 'textarea', span: 2 },
  { key: 'controversies', label: '争议', type: 'textarea', span: 2 },
  { key: 'death_cause', label: '死因', type: 'text', span: 2 },
  { key: 'sources_note', label: '说明', type: 'textarea', span: 2, randomizable: false },
]

export function cloneProfile(p: Profile): Profile {
  return { ...p }
}
