import { Select, Stack, useMantineColorScheme } from '@mantine/core'

import { PageHeader } from '@/components/PageHeader'

export function AxustesPage() {
  const { colorScheme, setColorScheme } = useMantineColorScheme()

  return (
    <Stack maw={520}>
      <PageHeader title="Axustes" description="Preferencias da aplicación" />
      <Select
        label="Esquema de cor"
        data={[
          { value: 'light', label: 'Claro' },
          { value: 'dark', label: 'Escuro' },
          { value: 'auto', label: 'Automático' },
        ]}
        value={colorScheme}
        onChange={(value) => {
          if (value === 'light' || value === 'dark' || value === 'auto') setColorScheme(value)
        }}
      />
    </Stack>
  )
}
