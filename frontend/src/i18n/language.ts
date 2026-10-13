import dayjs from 'dayjs'
import 'dayjs/locale/gl'

import { en } from '@/i18n/en'

export type Language = 'gl' | 'en'

// Cada idioma co seu nome no propio idioma: estes nomes non se traducen.
export const languages: ReadonlyArray<{ value: Language; label: string }> = [
  { value: 'gl', label: 'Galego' },
  { value: 'en', label: 'English' },
]

export const LANGUAGE_STORAGE_KEY = 'invest-tracker-lang'

// parseLanguage recoñece un idioma a partir dunha etiqueta como "gl", "en-US"
// ou "en_GB".
export function parseLanguage(value: string | null | undefined): Language | null {
  const tag = value?.trim().toLowerCase() ?? ''
  if (tag.startsWith('en')) return 'en'
  if (tag.startsWith('gl')) return 'gl'
  return null
}

// detectLanguage devolve o idioma escollido na web e, se aínda non se escolleu
// ningún, inglés se o navegador está en inglés e galego en calquera outro caso.
export function detectLanguage(): Language {
  try {
    const saved = parseLanguage(window.localStorage.getItem(LANGUAGE_STORAGE_KEY))
    if (saved) return saved
  } catch {
    // Sen localStorage (p.ex. bloqueado polo navegador) úsase o do navegador.
  }
  return parseLanguage(navigator.languages?.[0] ?? navigator.language) === 'en' ? 'en' : 'gl'
}

let current: Language = detectLanguage()
apply(current)

function apply(language: Language) {
  document.documentElement.lang = language
  dayjs.locale(language)
}

// getLanguage devolve o idioma actual da web.
export function getLanguage(): Language {
  return current
}

// setLanguage cambia o idioma e lémbrao para as próximas visitas. Para que a
// interface se actualice úsase useLanguage().setLanguage, que ademais volve
// pedir os datos ao servidor no novo idioma.
export function setLanguage(language: Language) {
  current = language
  try {
    window.localStorage.setItem(LANGUAGE_STORAGE_KEY, language)
  } catch {
    // Se non se pode gardar, o idioma dura ata recargar a páxina.
  }
  apply(language)
}

export type TranslationParams = Record<string, string | number>

// translate traduce un texto en galego ao idioma indicado e substitúe os
// parámetros {nome} polos seus valores.
export function translate(language: Language, key: string, params?: TranslationParams): string {
  const text = language === 'en' ? (en[key] ?? key) : key
  if (!params) return text
  return text.replace(/\{(\w+)\}/g, (match, name: string) => (name in params ? String(params[name]) : match))
}

// t traduce un texto en galego ao idioma actual. Os textos escríbense en galego
// e úsanse como clave; as traducións inglesas están en src/i18n/en/.
export function t(key: string, params?: TranslationParams): string {
  return translate(current, key, params)
}

// tKey marca un texto para traducilo despois con t() (p.ex. nas listas
// definidas fóra dos compoñentes). Non traduce nada: devolve o propio texto.
export function tKey<const T extends string>(key: T): T {
  return key
}
