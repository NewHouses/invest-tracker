import { Alert } from '@mantine/core'
import { IconAlertCircle } from '@tabler/icons-react'

import { ApiError } from '@/api/client'
import { t } from '@/i18n'

type ErrorAlertProps = {
  error: unknown
  title?: string
}

export function ErrorAlert({ error, title = t('Erro') }: ErrorAlertProps) {
  const message = error instanceof ApiError ? error.message : t('Produciuse un erro inesperado.')

  return (
    <Alert color="red" icon={<IconAlertCircle size={18} />} title={title}>
      {message}
    </Alert>
  )
}
