export const modeLabel: Record<string, string> = {
  famous: '历史名人',
  random: '随机人物',
}

export const statusLabel: Record<string, string> = {
  draft: '草稿',
  resolved: '待确认',
  confirmed: '已确认',
  profile_ready: '档案已生成',
  timeline_ready: '时间轴已生成',
}

export function formatDateTime(t: string): string {
  return new Date(t).toLocaleString('zh-CN')
}
