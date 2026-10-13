import { ApiError } from '@/api/client'
import { getLanguage, t } from '@/i18n'

export function fieldErrorMap(error: unknown, aliases: Record<string, string> = {}) {
  if (!(error instanceof ApiError) || !error.fields) return null
  const mapped: Record<string, string> = {}
  for (const [key, value] of Object.entries(error.fields)) {
    mapped[aliases[key] ?? key] = value
  }
  return mapped
}

export function errorMessage(error: unknown) {
  return error instanceof ApiError ? error.message : t('Produciuse un erro inesperado.')
}

// parseNumber le o valor dun NumberInput. Normalmente xa é un número; se é un
// texto (p.ex. a medio escribir) lese cos separadores do idioma actual.
export function parseNumber(value: number | string | null | undefined) {
  if (typeof value === 'number') return value
  if (value === null || value === undefined || value === '') return Number.NaN
  const text = String(value).trim()
  return Number(getLanguage() === 'en' ? text.replaceAll(',', '') : text.replace(',', '.'))
}
