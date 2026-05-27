import type { NodeScene } from '@/api/client'

export interface SceneDisplayItem {
  key: string
  label: string
  value: string
}

const SCENE_LABELS: Record<keyof NodeScene, string> = {
  datetime: '时日',
  time_of_day: '时辰',
  season: '季节',
  weather: '天气',
  scene: '场景',
}

/** 将有内容的 scene 字段转为展示列表，空则返回 [] */
export function sceneDisplayItems(scene?: NodeScene | null): SceneDisplayItem[] {
  if (!scene) return []
  const order: (keyof NodeScene)[] = ['datetime', 'time_of_day', 'season', 'weather', 'scene']
  const items: SceneDisplayItem[] = []
  for (const key of order) {
    const value = scene[key]?.trim()
    if (value) items.push({ key, label: SCENE_LABELS[key], value })
  }
  return items
}

export function hasScene(scene?: NodeScene | null): boolean {
  return sceneDisplayItems(scene).length > 0
}
