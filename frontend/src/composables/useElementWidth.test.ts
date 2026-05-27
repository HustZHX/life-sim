import { describe, expect, it } from 'vitest'
import { resolveGraphDensity } from '@/composables/useElementWidth'

describe('resolveGraphDensity', () => {
  it('returns dense when per-lane width is small', () => {
    expect(resolveGraphDensity(400, 2)).toBe('dense')
  })

  it('returns compact in middle range', () => {
    expect(resolveGraphDensity(600, 2)).toBe('compact')
  })

  it('returns comfortable when wide enough', () => {
    expect(resolveGraphDensity(900, 2)).toBe('comfortable')
  })
})
