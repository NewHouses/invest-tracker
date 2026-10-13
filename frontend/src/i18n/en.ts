import { analysis } from '@/i18n/en/analysis'
import { core } from '@/i18n/en/core'
import { operations } from '@/i18n/en/operations'

// Traducións ao inglés por áreas da aplicación (texto en galego → inglés). Un
// mesmo texto pode aparecer en varias áreas, pero sempre coa mesma tradución:
// o test do catálogo compróbao.
export const enParts: Record<string, Record<string, string>> = { core, operations, analysis }

export const en: Record<string, string> = Object.assign({}, ...Object.values(enParts))
