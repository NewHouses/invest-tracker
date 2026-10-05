import { Alert, Loader, Paper, Stack, Table, Text, useMantineTheme } from '@mantine/core'
import { LineChart } from '@mantine/charts'
import type { ReactNode } from 'react'

import type { LineChartDTO } from '@/api/charts'
import { EmptyState } from '@/components/EmptyState'
import { formatAxisUSD, formatUSD } from '@/lib/format'

import { mapLineChartDTO } from './lineChartMapper'

type LineChartPanelProps = {
  chart: LineChartDTO | undefined
  loading?: boolean
  emptyNoAssets?: string
  emptyNoResults?: string
  h?: number
}

export function LineChartPanel({ chart, loading, emptyNoAssets, emptyNoResults, h = 320 }: LineChartPanelProps) {
  if (loading) return <Loader aria-label="Cargando gráfica" />
  if (!chart) return null
  if (!chart.hasAssets) return <EmptyState title={emptyNoAssets ?? 'Aínda non hai activos'} />
  if (chart.months.length === 0) return <EmptyState title={emptyNoResults ?? 'Aínda non hai resultados rexistrados.'} />
  const mapped = mapLineChartDTO(chart)
  return (
    <Paper withBorder p="md" radius="md">
      <Stack>
        <Text fw={700}>{chart.title}</Text>
        <LineChart
          h={h}
          data={mapped.data}
          dataKey="month"
          series={mapped.series}
          valueFormatter={formatUSD}
          yAxisProps={{ tickFormatter: formatAxisUSD, width: 64 }}
          connectNulls={false}
          withLegend
        />
      </Stack>
    </Paper>
  )
}

type KvTableProps = { rows: Array<{ label: string; value: ReactNode }> }
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

export function RowTint({ gain, children }: { gain: number | null; children: React.ReactNode }) {
  const theme = useMantineTheme()
  const bg = gain === null ? undefined : gain > 0 ? theme.colors.green[0] : gain < 0 ? theme.colors.red[0] : theme.colors.yellow[0]
  return bg ? <Table.Tr bg={bg}>{children}</Table.Tr> : <Table.Tr>{children}</Table.Tr>
}

export function FieldError({ message }: { message: string | undefined }) {
  return message ? <Alert color="red">{message}</Alert> : null
}
