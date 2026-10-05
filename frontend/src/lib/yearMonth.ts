import type { YearMonth } from '@/api/types'

export function currentYearMonth(date = new Date()): YearMonth {
  return { year: date.getFullYear(), month: date.getMonth() + 1 }
}

export function compareYearMonth(left: YearMonth, right: YearMonth) {
  return left.year === right.year ? left.month - right.month : left.year - right.year
}

export function addMonths(value: YearMonth, months: number): YearMonth {
  const date = yearMonthToDate(value)
  date.setMonth(date.getMonth() + months)
  return dateToYearMonth(date)
}

export function yearMonthToDate(value: YearMonth): Date {
  return new Date(value.year, value.month - 1, 1)
}

export function dateToYearMonth(date: Date): YearMonth {
  return { year: date.getFullYear(), month: date.getMonth() + 1 }
}

export function parseYearMonth(value: string): YearMonth | null {
  const match = /^(\d{4})-(\d{2})$/.exec(value)
  if (!match) return null
  const year = Number(match[1])
  const month = Number(match[2])
  if (month < 1 || month > 12) return null
  return { year, month }
}

export function toYearMonthInputValue(value: YearMonth): string {
  return `${value.year}-${String(value.month).padStart(2, '0')}`
}
