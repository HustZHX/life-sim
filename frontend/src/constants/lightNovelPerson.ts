export type LightNovelPerson = 'first' | 'second' | 'third'

export const LIGHT_NOVEL_PERSON_OPTIONS: { value: LightNovelPerson; label: string }[] = [
  { value: 'first', label: '第一人称（我）' },
  { value: 'second', label: '第二人称（你）' },
  { value: 'third', label: '第三人称（他/她）' },
]

export function lightNovelPersonLabel(person?: string): string {
  return LIGHT_NOVEL_PERSON_OPTIONS.find((o) => o.value === person)?.label ?? '第一人称（我）'
}
