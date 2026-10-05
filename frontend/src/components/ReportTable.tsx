import { alpha, Table, useMantineTheme } from '@mantine/core'
import type { ReactNode } from 'react'

import { gainColor } from '@/lib/gain'

type KvTableProps = { rows: Array<{ label: string; value: ReactNode }> }

// Táboa de dúas columnas (etiqueta → valor) para os resumos dos informes.
export function KvTable({ rows }: KvTableProps) {
  return (
    <Table withTableBorder withColumnBorders striped>
      <Table.Tbody>
        {rows.map((row) => (
          <Table.Tr key={row.label}>
            <Table.Th w="45%">{row.label}</Table.Th>
            <Table.Td>{row.value}</Table.Td>
          </Table.Tr>
        ))}
      </Table.Tbody>
    </Table>
  )
}

// Opacidade do tinguido das filas: abonda para distinguir ganancia/perda
// dun golpe de vista sen saturar, e funciona igual en modo claro e escuro.
const TINT_ALPHA = 0.22

// rowTintColor devolve o fondo dunha fila segundo o signo da G/P (verde,
// vermello ou amarelo, coma as cores da CLI), ou undefined sen métrica.
export function useRowTintColor() {
  const theme = useMantineTheme()
  return (gain: number | null) => {
    if (gain === null) return undefined
    const color = gainColor(gain)
    const base = color === 'yellow' ? theme.colors.yellow[5] : theme.colors[color][6]
    return base ? alpha(base, TINT_ALPHA) : undefined
  }
}

export function RowTint({ gain, children }: { gain: number | null; children: ReactNode }) {
  const tint = useRowTintColor()
  const bg = tint(gain)
  return bg ? <Table.Tr bg={bg}>{children}</Table.Tr> : <Table.Tr>{children}</Table.Tr>
}
