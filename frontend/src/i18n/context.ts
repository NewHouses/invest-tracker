import { createContext, useContext } from 'react'

import { getLanguage } from '@/i18n/language'
import type { Language } from '@/i18n/language'

export type LanguageContextValue = {
  language: Language
  setLanguage: (language: Language) => void
}

export const LanguageContext = createContext<LanguageContextValue>({
  language: getLanguage(),
  setLanguage: () => undefined,
})

// useLanguage devolve o idioma actual e a función para cambialo.
export function useLanguage() {
  return useContext(LanguageContext)
}
