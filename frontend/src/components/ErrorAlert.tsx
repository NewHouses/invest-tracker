import { Alert } from '@mantine/core'
import { IconAlertCircle } from '@tabler/icons-react'

import { ApiError } from '@/api/client'

type ErrorAlertProps = {
  error: unknown
  title?: string
}

export function ErrorAlert({ error, title = 'Erro' }: ErrorAlertProps) {
  const message = error instanceof ApiError ? error.message : 'Produciuse un erro inesperado.'

  return (
    <Alert color="red" icon={<IconAlertCircle size={18} />} title={title}>
      {message}
    </Alert>
  )
}
