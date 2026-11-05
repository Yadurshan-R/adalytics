// Tests for toolList: every tool on the toolbar exists in the library and has
// a name and an icon, and each tool is in one group only.
import { describe, it, expect } from 'vitest'
import { findOverlaySpec } from 'lightweight-charts-drawing'
import { GROUPS, LABELS, DRAWING_KINDS, kindsOf, groupOf } from './toolList.js'
import { ICONS } from './toolIcons.js'

describe('toolList', () => {
  it('lists only tools the library has', () => {
    for (const kind of DRAWING_KINDS) expect(findOverlaySpec(kind), kind).toBeTruthy()
  })

  it('gives every tool a name and an icon', () => {
    for (const kind of GROUPS.flatMap(kindsOf)) {
      expect(LABELS[kind], kind).toBeTruthy()
      expect(ICONS[kind]?.icon, kind).toBeTruthy()
    }
  })

  it('puts each tool in one group, found by groupOf', () => {
    const all = GROUPS.flatMap(kindsOf)
    expect(new Set(all).size).toBe(all.length)
    for (const g of GROUPS) for (const kind of kindsOf(g)) expect(groupOf(kind)).toBe(g)
    expect(groupOf('xabcd-pattern')).toBeUndefined()
  })

  it('has the 16 drawing tools chosen for a token chart', () => {
    expect(DRAWING_KINDS).toHaveLength(16)
    expect(DRAWING_KINDS).toContain('long-position')
    expect(DRAWING_KINDS).not.toContain('cursor')
  })
})
