/** 时间轴节点 diff 字段 */
export const NODE_FIELD_LABELS: Record<string, string> = {
  title: '标题',
  events: '经历',
  thoughts: '想法',
  personality_snapshot: '性格快照',
  trait_changes: '性格/思想变更',
}

/** 性格/思想变更 trait_changes.field（含 AI 可能输出的英文键） */
export const TRAIT_FIELD_LABELS: Record<string, string> = {
  occupation: '职业',
  beliefs: '信念',
  personality: '性格',
  identity: '身份',
  status: '地位',
  social_class: '阶层',
  nationality: '国籍',
  worldview: '世界观',
  values: '价值观',
  ambition: '志向',
  temperament: '气质',
  relationship: '人际关系',
  title: '头衔',
  role: '角色',
  mindset: '心态',
}

function lookup(map: Record<string, string>, field: string): string | undefined {
  return map[field] ?? map[field.toLowerCase()]
}

/** 节点 diff 或 trait 变更字段 → 中文标签；已是中文则原样返回 */
export function fieldLabel(field: string, kind: 'node' | 'trait' = 'node'): string {
  const map = kind === 'trait' ? TRAIT_FIELD_LABELS : NODE_FIELD_LABELS
  return lookup(map, field) ?? field
}
