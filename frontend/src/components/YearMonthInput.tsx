import { MonthPickerInput } from '@mantine/dates'
import type { ReactNode } from 'react'

import type { YearMonth } from '@/api/types'
import { dateToYearMonth, parseYearMonth, yearMonthToDate } from '@/lib/yearMonth'

type MonthPickerValue = Date | string | null

type YearMonthInputProps = {
  value: YearMonth | null
  onChange: (value: YearMonth | null) => void
  label?: ReactNode
  placeholder?: string
  error?: ReactNode
  required?: boolean
  disabled?: boolean
  minDate?: Date
  maxDate?: Date
}

function toYearMonth(value: MonthPickerValue): YearMonth | null {
  if (!value) return null
  if (value instanceof Date) return dateToYearMonth(value)
  return parseYearMonth(value.slice(0, 7))
}

export function YearMonthInput({ value, onChange, ...props }: YearMonthInputProps) {
  return (
    <MonthPickerInput
      value={value ? yearMonthToDate(value) : null}
      onChange={(date) => onChange(toYearMonth(date))}
      valueFormat="MMMM YYYY"
      clearable
      {...props}
    />
  )
}
