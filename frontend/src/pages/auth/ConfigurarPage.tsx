import { Alert, Anchor, Button, Center, PasswordInput, Paper, Stack, Text, Title } from '@mantine/core'
import { useForm } from '@mantine/form'
import { IconInfoCircle } from '@tabler/icons-react'
import { useEffect } from 'react'
import { useNavigate } from 'react-router'

import { ApiError } from '@/api/client'
import { useAuthStatus, useSetup } from '@/api/auth'
import { ErrorAlert } from '@/components/ErrorAlert'

export function ConfigurarPage() {
  const navigate = useNavigate()
  const status = useAuthStatus()
  const setup = useSetup()
  const form = useForm({
    initialValues: { password: '', confirm: '' },
    validate: {
      password: (value) => (value === '' ? 'O contrasinal non pode estar baleiro' : null),
      confirm: (value, values) => (value !== values.password ? 'Os contrasinais non coinciden' : null),
    },
  })

  useEffect(() => {
    if (status.data && !status.data.setupRequired && status.data.authenticated) {
      navigate('/', { replace: true })
    }
  }, [navigate, status.data])

  const submit = form.onSubmit((values) => {
    setup.mutate(values.password, {
      onSuccess: () => navigate('/', { replace: true }),
      onError: (error) => {
        if (error instanceof ApiError && error.fields) form.setErrors(error.fields)
      },
    })
  })

  const localUrl = `http://localhost:${window.location.port || '8080'}`

  return (
    <Center mih="100vh" p="md">
      <Paper withBorder shadow="sm" radius="md" p="xl" w="100%" maw={440}>
        <form onSubmit={submit}>
          <Stack>
            <Stack gap={4}>
              <Title order={1}>Configurar acceso</Title>
              <Text c="dimmed">Crea o contrasinal inicial para protexer a aplicación.</Text>
            </Stack>
            {status.data?.setupAllowed === false ? (
              <Alert color="yellow" icon={<IconInfoCircle size={18} />}>
                O contrasinal só se pode crear desde o propio ordenador: abre{' '}
                <Anchor href={localUrl}>{localUrl}</Anchor> nese equipo.
              </Alert>
            ) : null}
            {setup.error ? <ErrorAlert error={setup.error} /> : null}
            <PasswordInput label="Contrasinal" required {...form.getInputProps('password')} />
            <PasswordInput label="Confirmar contrasinal" required {...form.getInputProps('confirm')} />
            <Button type="submit" loading={setup.isPending} disabled={status.data?.setupAllowed === false}>
              Crear contrasinal
            </Button>
          </Stack>
        </form>
      </Paper>
    </Center>
  )
}
