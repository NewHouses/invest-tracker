import { Badge, Group, Loader, Paper, Stack, Table, Text, Title } from '@mantine/core'
import type { ReactNode } from 'react'

import type { AssetHistory, HistoryRow, TotalHistory, TotalHistoryRow, TypeHistory } from '@/api/reports'
import { EmptyState } from '@/components/EmptyState'
import { ErrorAlert } from '@/components/ErrorAlert'
import { useAssetTypeLabel } from '@/api/meta'
import { AssetTypeBadge } from '@/components/AssetTypeBadge'
import { formatPct, formatSignedUSD, formatUSD, formatYearMonth } from '@/lib/format'
import { useAssetHistoryReport } from '@/api/reports'
import { KvTable, RowTint } from '@/components/ReportTable'

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

function CurrentMonthCell({ month }: { month: number }) {
  return (
    <Group gap="xs" wrap="nowrap">
      <Text span fs="italic">{month}</Text>
      <Badge size="xs" variant="light">en curso</Badge>
    </Group>
  )
}

function PendingRowsNote() {
  return (
    <Text c="dimmed" size="sm">
      Aínda non hai resultados rexistrados; móstrase o mes en curso cos importes xa aportados.
    </Text>
  )
}

function AssetTypeCurrentRow({ row }: { row: HistoryRow }) {
  return (
    <Table.Tr c="dimmed" fs="italic" data-testid="history-current-row">
      <Table.Td>{row.period.year}</Table.Td>
      <Table.Td><CurrentMonthCell month={row.period.month} /></Table.Td>
      <Table.Td>{formatUSD(row.aporte)}</Table.Td>
      <Table.Td>{formatUSD(row.holding)}</Table.Td>
      <Table.Td>—</Table.Td>
      <Table.Td data-bold="true" data-testid="history-gain-cell" fw={600}>—</Table.Td>
      <Table.Td data-bold="true" data-testid="history-result-cell" fw={600}>—</Table.Td>
    </Table.Tr>
  )
}

function AssetTypeRows({ rows, current }: { rows: HistoryRow[]; current: HistoryRow | null }) {
  return (
    <Table.ScrollContainer minWidth={640}>
    <Table withTableBorder striped highlightOnHover style={{ whiteSpace: 'nowrap' }}>
      <Table.Thead>
        <Table.Tr>
          <Table.Th>Ano</Table.Th>
          <Table.Th>Mes</Table.Th>
          <Table.Th>Aporte Mensual</Table.Th>
          <Table.Th>No ativo</Table.Th>
          <Table.Th>Índice</Table.Th>
          <Table.Th>G/P</Table.Th>
          <Table.Th>Resultado</Table.Th>
        </Table.Tr>
      </Table.Thead>
      <Table.Tbody>
        {rows.map((row) => (
          <RowTint key={`${row.period.year}-${row.period.month}`} gain={row.hasMetrics ? row.gain : null}>
            <Table.Td>{row.period.year}</Table.Td>
            <Table.Td>{row.period.month}</Table.Td>
            <Table.Td>{formatUSD(row.aporte)}</Table.Td>
            <Table.Td>{formatUSD(row.holding)}</Table.Td>
            <Table.Td>{metricPct(row)}</Table.Td>
            <Table.Td data-bold="true" data-testid="history-gain-cell" fw={600}>{metricGain(row)}</Table.Td>
            <Table.Td data-bold="true" data-testid="history-result-cell" fw={600}>{formatUSD(row.result)}</Table.Td>
          </RowTint>
        ))}
        {current ? <AssetTypeCurrentRow row={current} /> : null}
      </Table.Tbody>
    </Table>
    </Table.ScrollContainer>
  )
}

export function AssetHistoryView({ history }: { history: AssetHistory }) {
  if (history.rows.length === 0 && !history.current) {
    return <EmptyState title="Aínda non hai resultados rexistrados para este ativo." description="Engade un coa operación 'Engadir resultado'." />
  }
  return (
    <Stack>
      <Group>
        <Title order={3}>{history.asset.name}</Title>
        <AssetTypeBadge type={history.asset.type} />
        <Text c="dimmed">{history.rows.length} mes(es) con resultado</Text>
      </Group>
      <AssetTypeSummary history={history} />
      {history.rows.length === 0 ? <PendingRowsNote /> : null}
      <AssetTypeRows rows={history.rows} current={history.current} />
    </Stack>
  )
}

export function TypeHistoryView({ history }: { history: TypeHistory }) {
  const typeLabel = useAssetTypeLabel()
  if (history.assetCount === 0) return <EmptyState title={`Non hai ativos de tipo ${typeLabel(history.type)}.`} />
  if (history.rows.length === 0 && !history.current) return <EmptyState title={`Aínda non hai resultados rexistrados para ativos de tipo ${typeLabel(history.type)}.`} />
  return (
    <Stack>
      <Group><Title order={3}>{typeLabel(history.type)}</Title><Text c="dimmed">{history.assetCount} ativo(s) · {history.rows.length} mes(es) con resultado</Text></Group>
      <AssetTypeSummary history={history} />
      {history.rows.length === 0 ? <PendingRowsNote /> : null}
      <AssetTypeRows rows={history.rows} current={history.current} />
    </Stack>
  )
}

function TotalCurrentRow({ row }: { row: TotalHistoryRow }) {
  return (
    <Table.Tr c="dimmed" fs="italic" data-testid="history-current-row">
      <Table.Td>{row.period.year}</Table.Td>
      <Table.Td><CurrentMonthCell month={row.period.month} /></Table.Td>
      <Table.Td>{formatSignedUSD(row.aporte)}</Table.Td>
      <Table.Td>{formatUSD(row.fondos)}</Table.Td>
      <Table.Td>—</Table.Td>
      <Table.Td data-bold="true" data-testid="history-gain-cell" fw={600}>—</Table.Td>
      <Table.Td>{formatUSD(row.dividends)}</Table.Td>
      <Table.Td data-bold="true" data-testid="history-result-cell" fw={600}>—</Table.Td>
    </Table.Tr>
  )
}

export function TotalHistoryView({ history }: { history: TotalHistory }) {
  if (history.assetCount === 0 && !history.current) return <EmptyState title="Aínda non hai ativos." description="Engade un primeiro coa operación 'Engadir ativo'." />
  if (history.rows.length === 0 && !history.current) return <EmptyState title="Aínda non hai resultados rexistrados." description="Engade resultados mensuais coa operación 'Engadir resultado' ou 'Pechar mes'." />
  return (
    <Stack>
      <KvTable rows={[
        { label: 'Aporte histórico total', value: formatUSD(history.lifetimeAporte) },
        { label: 'Índice Medio', value: dash(history.hasAverages, () => formatPct(history.avgIndexPct)) },
        { label: 'G/P Media', value: dash(history.hasAverages, () => formatSignedUSD(history.avgGain)) },
        { label: 'G/P Total', value: dash(history.hasTotalGain, () => formatSignedUSD(history.totalGain)) },
        { label: 'Dividendos totais', value: formatUSD(history.totalDividends) },
      ]} />
      {history.rows.length === 0 ? <PendingRowsNote /> : null}
      <Table.ScrollContainer minWidth={720}>
      <Table withTableBorder striped highlightOnHover style={{ whiteSpace: 'nowrap' }}>
        <Table.Thead>
          <Table.Tr>
            <Table.Th>Ano</Table.Th>
            <Table.Th>Mes</Table.Th>
            <Table.Th>Aporte Mensual</Table.Th>
            <Table.Th>Fondos</Table.Th>
            <Table.Th>Índice</Table.Th>
            <Table.Th>G/P</Table.Th>
            <Table.Th>Dividendos</Table.Th>
            <Table.Th>Resultado</Table.Th>
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>
          {history.rows.map((row: TotalHistoryRow) => (
            <RowTint key={formatYearMonth(row.period)} gain={row.hasMetrics ? row.gain : null}>
              <Table.Td>{row.period.year}</Table.Td>
              <Table.Td>{row.period.month}</Table.Td>
              <Table.Td>{formatSignedUSD(row.aporte)}</Table.Td>
              <Table.Td>{formatUSD(row.fondos)}</Table.Td>
              <Table.Td>{metricPct(row)}</Table.Td>
              <Table.Td data-bold="true" data-testid="history-gain-cell" fw={600}>{metricGain(row)}</Table.Td>
              <Table.Td>{formatUSD(row.dividends)}</Table.Td>
              <Table.Td data-bold="true" data-testid="history-result-cell" fw={600}>{formatUSD(row.result)}</Table.Td>
            </RowTint>
          ))}
          {history.current ? <TotalCurrentRow row={history.current} /> : null}
        </Table.Tbody>
      </Table>
      </Table.ScrollContainer>
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
