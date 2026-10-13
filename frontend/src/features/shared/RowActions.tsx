import { ActionIcon, Group } from '@mantine/core'
import { IconPencil, IconTrash } from '@tabler/icons-react'

import { t } from '@/i18n'

export function RowActions({ onEdit, onDelete, disabled }: { onEdit?: () => void; onDelete?: () => void; disabled?: boolean }) {
  return (
    <Group gap="xs" justify="flex-end" wrap="nowrap">
      {onEdit ? (
        <ActionIcon aria-label={t('Editar')} variant="subtle" onClick={onEdit} disabled={disabled ?? false}>
          <IconPencil size={18} />
        </ActionIcon>
      ) : null}
      {onDelete ? (
        <ActionIcon aria-label={t('Eliminar')} color="red" variant="subtle" onClick={onDelete} disabled={disabled ?? false}>
          <IconTrash size={18} />
        </ActionIcon>
      ) : null}
    </Group>
  )
}
