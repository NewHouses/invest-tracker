import { modals } from '@mantine/modals'

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
    labels: { confirm: 'Eliminar', cancel: 'Cancelar' },
    onConfirm,
  })
}
