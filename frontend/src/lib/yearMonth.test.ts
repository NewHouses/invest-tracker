import { describe, expect, it } from 'vitest'

import {
  addMonths,
  compareYearMonth,
  currentYearMonth,
  dateToYearMonth,
  parseYearMonth,
  toYearMonthInputValue,
  yearMonthToDate,
} from '@/lib/yearMonth'

describe('yearMonth helpers', () => {
  it('returns the current month from a date', () => {
    expect(currentYearMonth(new Date(2026, 2, 15))).toEqual({ year: 2026, month: 3 })
  })

  it('compares and adds months', () => {
    expect(compareYearMonth({ year: 2026, month: 3 }, { year: 2026, month: 4 })).toBeLessThan(0)
    expect(addMonths({ year: 2026, month: 12 }, 2)).toEqual({ year: 2027, month: 2 })
  })

  it('converts to and from dates and inputs', () => {
    const date = yearMonthToDate({ year: 2026, month: 3 })
    expect(dateToYearMonth(date)).toEqual({ year: 2026, month: 3 })
    expect(parseYearMonth('2026-03')).toEqual({ year: 2026, month: 3 })
    expect(parseYearMonth('2026-13')).toBeNull()
    expect(toYearMonthInputValue({ year: 2026, month: 3 })).toBe('2026-03')
  })
})
