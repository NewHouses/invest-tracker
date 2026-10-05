import { Alert, Button, Center, Checkbox, PasswordInput, Paper, Stack, Text, Title } from '@mantine/core'
import { useForm } from '@mantine/form'
import { IconAlertCircle } from '@tabler/icons-react'
import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router'

import { ApiError } from '@/api/client'
import { useAuthStatus, useLogin } from '@/api/auth'

function loginErrorMessage(error: unknown) {
  if (!(error instanceof ApiError)) return 'Non se puido iniciar sesión.'
  if (error.status === 401) return 'Contrasinal incorrecto'
  if (error.status === 429) {
    const retryAfter = error.headers.get('Retry-After')
    return retryAfter
      ? `Demasiados intentos. Agarda ${retryAfter} segundos e téntao de novo.`
      : 'Demasiados intentos. Agarda un pouco e téntao de novo.'
  }
  return error.message
}

export function LoginPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const login = useLogin()
  const status = useAuthStatus()
  const [message, setMessage] = useState<string | null>(null)
  const form = useForm({ initialValues: { password: '', remember: false } })
  const next = searchParams.get('next') || '/'

  useEffect(() => {
    if (status.data?.setupRequired) navigate('/configurar', { replace: true })
    if (status.data?.authenticated) navigate(next, { replace: true })
  }, [navigate, next, status.data])

  const submit = form.onSubmit((values) => {
    setMessage(null)
    login.mutate(values, {
      onSuccess: () => navigate(next, { replace: true }),
      onError: (error) => {
        if (error instanceof ApiError && error.status === 409) {
          navigate('/configurar', { replace: true })
          return
        }
        setMessage(loginErrorMessage(error))
      },
    })
  })

  return (
    <Center mih="100vh" p="md">
      <Paper withBorder shadow="sm" radius="md" p="xl" w="100%" maw={420}>
        <form onSubmit={submit}>
          <Stack>
            <Stack gap={4}>
              <Title order={1}>Entrar</Title>
              <Text c="dimmed">Introduce o contrasinal para continuar.</Text>
            </Stack>
            {message ? (
              <Alert color="red" icon={<IconAlertCircle size={18} />}>
                {message}
              </Alert>
            ) : null}
            <PasswordInput label="Contrasinal" aria-label="Contrasinal" required autoFocus {...form.getInputProps('password')} />
            <Checkbox label="Lembrarme neste dispositivo" {...form.getInputProps('remember', { type: 'checkbox' })} />
            <Button type="submit" loading={login.isPending}>
              Entrar
            </Button>
          </Stack>
        </form>
      </Paper>
    </Center>
  )
}
