import { DatesProvider } from '@mantine/dates'
import { useQueryClient } from '@tanstack/react-query'
import { Fragment, useCallback, useMemo, useState } from 'react'
import type { ReactNode } from 'react'

import { LanguageContext } from '@/i18n/context'
import { getLanguage, setLanguage as setCurrentLanguage } from '@/i18n/language'
import type { Language } from '@/i18n/language'

// LanguageProvider garda o idioma da web. Ao cambialo descarta os datos xa
// cargados (os textos que envía o servidor, como os tipos de activo ou os
// títulos das gráficas, veñen no idioma de cada petición) e volve montar a
// interface para que todo, tamén os importes e as datas, saia no novo idioma.
export function LanguageProvider({ children }: { children: ReactNode }) {
  const queryClient = useQueryClient()
  const [language, setLanguageState] = useState<Language>(getLanguage)

  const changeLanguage = useCallback(
    (next: Language) => {
      if (next === getLanguage()) return
      setCurrentLanguage(next)
      void queryClient.cancelQueries()
      queryClient.removeQueries()
      setLanguageState(next)
    },
    [queryClient],
  )

  const value = useMemo(() => ({ language, setLanguage: changeLanguage }), [language, changeLanguage])

  return (
    <LanguageContext.Provider value={value}>
      <DatesProvider settings={{ locale: language, firstDayOfWeek: 1, weekendDays: [0, 6] }}>
        <Fragment key={language}>{children}</Fragment>
      </DatesProvider>
    </LanguageContext.Provider>
  )
}
