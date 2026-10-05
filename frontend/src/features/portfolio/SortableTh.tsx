import { Center, Group, Table, Text, UnstyledButton } from '@mantine/core'
import { IconChevronDown, IconChevronUp, IconSelector } from '@tabler/icons-react'
import type { ReactNode } from 'react'

import type { SortDirection } from '@/features/portfolio/sortRows'

type SortableThProps<T extends string> = {
  children: ReactNode
  sortKey: T
  activeKey: T | null
  direction: SortDirection
  onSort: (key: T) => void
  ta?: 'left' | 'right' | 'center'
}

export function SortableTh<T extends string>({ children, sortKey, activeKey, direction, onSort, ta = 'left' }: SortableThProps<T>) {
  const active = activeKey === sortKey && direction !== null
  const Icon = active ? (direction === 'asc' ? IconChevronUp : IconChevronDown) : IconSelector

  return (
    <Table.Th ta={ta}>
      <UnstyledButton onClick={() => onSort(sortKey)} w="100%">
        <Group gap="xs" justify={ta === 'right' ? 'flex-end' : 'flex-start'} wrap="nowrap">
          <Text component="span" fw={600} size="sm">{children}</Text>
          <Center c={active ? 'teal.7' : 'dimmed'}>
            <Icon size={14} stroke={1.8} />
          </Center>
        </Group>
      </UnstyledButton>
    </Table.Th>
  )
}
