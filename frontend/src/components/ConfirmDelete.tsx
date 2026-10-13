import { modals } from '@mantine/modals'

import { t } from '@/i18n'

type OpenConfirmDeleteInput = {
  title: string
  message: string
  onConfirm: () => void | Promise<void>
}

export function openConfirmDelete({ title, message, onConfirm }: OpenConfirmDeleteInput) {
  modals.openConfirmModal({
    title,
    children: message,
    confirmProps: { color: 'red' },
    labels: { confirm: t('Eliminar'), cancel: t('Cancelar') },
    onConfirm,
  })
}
