import type { YearMonth } from '@/api/types'
import { formatYearMonth } from '@/lib/format'

export function periodFrom(year: number, month: number): YearMonth {
  return { year, month }
}

export function formatPeriod(year: number, month: number) {
  return formatYearMonth(periodFrom(year, month))
}
