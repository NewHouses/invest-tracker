import { Badge, Group, Loader, Paper, Stack, Table, Text, Title } from '@mantine/core'
import type { ReactNode } from 'react'

import type { AssetHistory, HistoryRow, TotalHistory, TotalHistoryRow, TypeHistory } from '@/api/reports'
import { EmptyState } from '@/components/EmptyState'
import { ErrorAlert } from '@/components/ErrorAlert'
import { useAssetTypeLabel } from '@/api/meta'
import { formatPct, formatSignedUSD, formatUSD, formatYearMonth } from '@/lib/format'
import { useAssetHistoryReport } from '@/api/reports'
import { KvTable, RowTint } from '@/features/charts/common'

function dash(value: boolean, render: () => string) {
  return value ? render() : '—'
}

function metricPct(row: { hasMetrics: boolean; gainPct: number }) {
  return row.hasMetrics ? formatPct(row.gainPct) : 'n/a'
}
function metricGain(row: { hasMetrics: boolean; gain: number }) {
  return row.hasMetrics ? formatSignedUSD(row.gain) : '—'
}

function AssetTypeSummary({ history }: { history: AssetHistory | TypeHistory }) {
  return (
    <KvTable
      rows={[
        { label: 'Total Aportado', value: formatUSD(history.totalInvested) },
        { label: 'Índice Medio Mensual', value: dash(history.hasAverages, () => formatPct(history.avgIndexPct)) },
        { label: 'Gañanzas/Perdas Medias Mensuais', value: dash(history.hasAverages, () => formatSignedUSD(history.avgGain)) },
        { label: 'Total Gañanzas/Perdas', value: dash(history.hasTotalGain, () => formatSignedUSD(history.totalGain)) },
      ]}
    />
  )
}

function AssetTypeRows({ rows }: { rows: HistoryRow[] }) {
  return (
    <Table withTableBorder striped highlightOnHover>
      <Table.Thead>
        <Table.Tr><Table.Th>Ano</Table.Th><Table.Th>Mes</Table.Th><Table.Th>Aporte Mensual</Table.Th><Table.Th>No activo</Table.Th><Table.Th>Índice</Table.Th><Table.Th>G/P</Table.Th><Table.Th>Resultado</Table.Th></Table.Tr>
      </Table.Thead>
      <Table.Tbody>
        {rows.map((row) => (
          <RowTint key={`${row.period.year}-${row.period.month}`} gain={row.hasMetrics ? row.gain : null}>
            <Table.Td>{row.period.year}</Table.Td><Table.Td>{row.period.month}</Table.Td><Table.Td>{formatUSD(row.aporte)}</Table.Td><Table.Td>{formatUSD(row.holding)}</Table.Td><Table.Td>{metricPct(row)}</Table.Td><Table.Td>{metricGain(row)}</Table.Td><Table.Td>{formatUSD(row.result)}</Table.Td>
          </RowTint>
        ))}
      </Table.Tbody>
    </Table>
  )
}

export function AssetHistoryView({ history }: { history: AssetHistory }) {
  const typeLabel = useAssetTypeLabel()
  if (history.rows.length === 0) return <EmptyState title="Aínda non hai resultados rexistrados para este activo." description="Engade un coa operación 'Engadir resultado'." />
  return (
    <Stack>
      <Group><Title order={3}>{history.asset.name}</Title><Badge>{typeLabel(history.asset.type)}</Badge><Text c="dimmed">{history.rows.length} mes(es) con resultado</Text></Group>
      <AssetTypeSummary history={history} />
      <AssetTypeRows rows={history.rows} />
    </Stack>
  )
}

export function TypeHistoryView({ history }: { history: TypeHistory }) {
  const typeLabel = useAssetTypeLabel()
  if (history.assetCount === 0) return <EmptyState title={`Non hai activos de tipo ${typeLabel(history.type)}.`} />
  if (history.rows.length === 0) return <EmptyState title={`Aínda non hai resultados rexistrados para activos de tipo ${typeLabel(history.type)}.`} />
  return (
    <Stack>
      <Group><Title order={3}>{typeLabel(history.type)}</Title><Text c="dimmed">{history.assetCount} activo(s) · {history.rows.length} mes(es) con resultado</Text></Group>
      <AssetTypeSummary history={history} />
      <AssetTypeRows rows={history.rows} />
    </Stack>
  )
}

export function TotalHistoryView({ history }: { history: TotalHistory }) {
  if (history.assetCount === 0) return <EmptyState title="Aínda non hai activos." description="Engade un primeiro coa operación 'Engadir activo'." />
  if (history.rows.length === 0) return <EmptyState title="Aínda non hai resultados rexistrados." description="Engade resultados mensuais coa operación 'Engadir resultado' ou 'Pechar mes'." />
  return (
    <Stack>
      <KvTable rows={[
        { label: 'Aporte histórico total', value: formatUSD(history.lifetimeAporte) },
        { label: 'Índice Medio', value: dash(history.hasAverages, () => formatPct(history.avgIndexPct)) },
        { label: 'G/P Media', value: dash(history.hasAverages, () => formatSignedUSD(history.avgGain)) },
        { label: 'G/P Total', value: dash(history.hasTotalGain, () => formatSignedUSD(history.totalGain)) },
        { label: 'Dividendos totais', value: formatUSD(history.totalDividends) },
      ]} />
      <Table withTableBorder striped highlightOnHover>
        <Table.Thead><Table.Tr><Table.Th>Ano</Table.Th><Table.Th>Mes</Table.Th><Table.Th>Aporte Mensual</Table.Th><Table.Th>Fondos</Table.Th><Table.Th>Índice</Table.Th><Table.Th>G/P</Table.Th><Table.Th>Dividendos</Table.Th><Table.Th>Resultado</Table.Th></Table.Tr></Table.Thead>
        <Table.Tbody>{history.rows.map((row: TotalHistoryRow) => <RowTint key={formatYearMonth(row.period)} gain={row.hasMetrics ? row.gain : null}><Table.Td>{row.period.year}</Table.Td><Table.Td>{row.period.month}</Table.Td><Table.Td>{formatSignedUSD(row.aporte)}</Table.Td><Table.Td>{formatUSD(row.fondos)}</Table.Td><Table.Td>{metricPct(row)}</Table.Td><Table.Td>{metricGain(row)}</Table.Td><Table.Td>{formatUSD(row.dividends)}</Table.Td><Table.Td>{formatUSD(row.result)}</Table.Td></RowTint>)}</Table.Tbody>
      </Table>
    </Stack>
  )
}

export function LoadingReport() { return <Loader aria-label="Cargando informe" /> }

export function ReportCard({ children }: { children: ReactNode }) {
  return <Paper withBorder p="md" radius="md"><Stack>{children}</Stack></Paper>
}

export function AssetHistoryPanel({ assetId }: { assetId: number }) {
  const history = useAssetHistoryReport(assetId)
  if (history.isLoading) return <LoadingReport />
  if (history.isError) return <ErrorAlert error={history.error} />
  return history.data ? <AssetHistoryView history={history.data} /> : null
}
