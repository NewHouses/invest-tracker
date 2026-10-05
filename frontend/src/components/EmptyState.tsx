import { Paper, Stack, Text, ThemeIcon, Title } from '@mantine/core'
import { IconInbox } from '@tabler/icons-react'
import type { ReactNode } from 'react'

type EmptyStateProps = {
  title: string
  description?: string
  action?: ReactNode
}

export function EmptyState({ title, description, action }: EmptyStateProps) {
  return (
    <Paper withBorder p="xl" radius="md">
      <Stack align="center" ta="center" gap="sm">
        <ThemeIcon variant="light" size="xl" radius="xl">
          <IconInbox size={28} />
        </ThemeIcon>
        <Title order={3}>{title}</Title>
        {description ? <Text c="dimmed">{description}</Text> : null}
        {action}
      </Stack>
    </Paper>
  )
}
