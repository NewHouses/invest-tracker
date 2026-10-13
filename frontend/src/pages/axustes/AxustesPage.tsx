import { Select, Stack, useMantineColorScheme } from '@mantine/core'

import { PageHeader } from '@/components/PageHeader'
import { languages, parseLanguage, t, useLanguage } from '@/i18n'

export function AxustesPage() {
  const { colorScheme, setColorScheme } = useMantineColorScheme()
  const { language, setLanguage } = useLanguage()

  return (
    <Stack maw={520}>
      <PageHeader title={t('Axustes')} description={t('Preferencias da aplicación')} />
      <Select
        label={t('Idioma')}
        data={languages.map((option) => ({ value: option.value, label: option.label }))}
        value={language}
        allowDeselect={false}
        onChange={(value) => {
          const next = parseLanguage(value)
          if (next) setLanguage(next)
        }}
      />
      <Select
        label={t('Esquema de cor')}
        data={[
          { value: 'light', label: t('Claro') },
          { value: 'dark', label: t('Escuro') },
          { value: 'auto', label: t('Automático') },
        ]}
        value={colorScheme}
        allowDeselect={false}
        onChange={(value) => {
          if (value === 'light' || value === 'dark' || value === 'auto') setColorScheme(value)
        }}
      />
    </Stack>
  )
}
