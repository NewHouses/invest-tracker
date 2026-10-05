import { Button, Group, PasswordInput, Select, Stack } from '@mantine/core'
import { useForm } from '@mantine/form'
import { notifications } from '@mantine/notifications'
import { IconLogout } from '@tabler/icons-react'
import { useNavigate } from 'react-router'
import { useMantineColorScheme } from '@mantine/core'

import { ApiError } from '@/api/client'
import { useChangePassword, useLogout } from '@/api/auth'
import { ErrorAlert } from '@/components/ErrorAlert'
import { PageHeader } from '@/components/PageHeader'

export function AxustesPage() {
  const navigate = useNavigate()
  const changePassword = useChangePassword()
  const logout = useLogout()
  const { colorScheme, setColorScheme } = useMantineColorScheme()
  const form = useForm({
    initialValues: { currentPassword: '', newPassword: '', confirm: '' },
    validate: {
      newPassword: (value) => (value === '' ? 'O contrasinal non pode estar baleiro' : null),
      confirm: (value, values) => (value !== values.newPassword ? 'Os contrasinais non coinciden' : null),
    },
  })

  const submit = form.onSubmit((values) => {
    changePassword.mutate(
      { currentPassword: values.currentPassword, newPassword: values.newPassword },
      {
        onSuccess: () => {
          form.reset()
          notifications.show({ color: 'green', message: 'Contrasinal actualizado.' })
        },
        onError: (error) => {
          if (error instanceof ApiError && error.fields) {
            form.setErrors(error.fields)
          }
        },
      },
    )
  })

  const handleLogout = () => {
    logout.mutate(undefined, { onSettled: () => navigate('/login', { replace: true }) })
  }

  return (
    <Stack maw={520}>
      <PageHeader title="Axustes" description="Preferencias e seguridade da aplicación" />
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
      <form onSubmit={submit}>
        <Stack>
          {changePassword.error ? <ErrorAlert error={changePassword.error} /> : null}
          <PasswordInput label="Contrasinal actual" required {...form.getInputProps('currentPassword')} />
          <PasswordInput label="Novo contrasinal" required {...form.getInputProps('newPassword')} />
          <PasswordInput label="Confirmar novo contrasinal" required {...form.getInputProps('confirm')} />
          <Group>
            <Button type="submit" loading={changePassword.isPending}>
              Cambiar contrasinal
            </Button>
            <Button variant="light" color="red" leftSection={<IconLogout size={16} />} onClick={handleLogout} loading={logout.isPending}>
              Pechar sesión
            </Button>
          </Group>
        </Stack>
      </form>
    </Stack>
  )
}
