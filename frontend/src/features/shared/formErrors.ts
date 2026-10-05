import { ApiError } from '@/api/client'

export function fieldErrorMap(error: unknown, aliases: Record<string, string> = {}) {
  if (!(error instanceof ApiError) || !error.fields) return null
  const mapped: Record<string, string> = {}
  for (const [key, value] of Object.entries(error.fields)) {
    mapped[aliases[key] ?? key] = value
  }
  return mapped
}

export function errorMessage(error: unknown) {
  return error instanceof ApiError ? error.message : 'Produciuse un erro inesperado.'
}

export function parseNumber(value: number | string | null | undefined) {
  if (typeof value === 'number') return value
  if (value === null || value === undefined || value === '') return Number.NaN
  return Number(String(value).replace(',', '.'))
}
