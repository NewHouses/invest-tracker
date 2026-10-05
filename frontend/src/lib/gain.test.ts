import { describe, expect, it } from 'vitest'

import { gainColor } from '@/lib/gain'

describe('gainColor', () => {
  it('maps gains to CLI-like colors', () => {
    expect(gainColor(1)).toBe('green')
    expect(gainColor(-1)).toBe('red')
    expect(gainColor(0)).toBe('yellow')
  })
})
