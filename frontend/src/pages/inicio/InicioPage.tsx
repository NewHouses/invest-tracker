import { Badge, Button, Card, Group, Loader, ScrollArea, SimpleGrid, Stack, Table, Text, Title } from '@mantine/core'
import { LineChart } from '@mantine/charts'
import { Link } from 'react-router'

import { useDashboard } from '@/api/dashboard'
import type { PortfolioRow } from '@/api/portfolio'
import { AssetTypeBadge } from '@/components/AssetTypeBadge'
import { EmptyState } from '@/components/EmptyState'
import { ErrorAlert } from '@/components/ErrorAlert'
import { PageHeader } from '@/components/PageHeader'
import { DistributionChart } from '@/features/charts/DistributionChart'
import { mapLineChartDTO } from '@/features/charts/lineChartMapper'
import { formatAxisUSD, formatPct, formatSignedUSD, formatUSD, formatYearMonth } from '@/lib/format'
import { gainColor } from '@/lib/gain'

function KpiCard({ label, value, color }: { label: string; value: string; color?: string }) {
  const valueProps = color ? { c: color } : {}
  return <Card withBorder><Text size="sm" c="dimmed">{label}</Text><Text fw={700} size="xl" {...valueProps}>{value}</Text></Card>
}

function PortfolioNumber({ value, available, pct = false, signed = false }: { value: number; available: boolean; pct?: boolean; signed?: boolean }) {
  if (!available) return <Text component="span" c="dimmed">—</Text>
  const text = pct ? formatPct(value) : signed ? formatSignedUSD(value) : formatUSD(value)
  const colorProps = signed || pct ? { c: gainColor(value) } : {}
  return <Text component="span" {...colorProps}>{text}</Text>
}

function PortfolioTable({ rows }: { rows: PortfolioRow[] }) {
  if (rows.length === 0) return <Text c="dimmed">Aínda non hai ativos na carteira.</Text>

  return (
    <ScrollArea type="auto">
      <Table verticalSpacing="sm" miw={760}>
        <Table.Thead>
          <Table.Tr>
            <Table.Th>Ativo</Table.Th>
            <Table.Th>Tipo</Table.Th>
            <Table.Th>Aportado</Table.Th>
            <Table.Th>Valor neto</Table.Th>
            <Table.Th>G/P</Table.Th>
            <Table.Th>%</Table.Th>
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>
          {rows.map((row) => (
            <Table.Tr key={row.asset.id}>
              <Table.Td>
                <Group gap="xs" wrap="nowrap">
                  <Text component={Link} to={`/activos/${row.asset.id}`} fw={600} c="teal.7" td="none">{row.asset.name}</Text>
                  {row.pending ? <Badge color="yellow" variant="light" size="sm">pendente</Badge> : null}
                </Group>
              </Table.Td>
              <Table.Td><AssetTypeBadge type={row.asset.type} /></Table.Td>
              <Table.Td>{formatUSD(row.totalInvested)}</Table.Td>
              <Table.Td><PortfolioNumber value={row.currentValue} available={row.hasCurrentValue} /></Table.Td>
              <Table.Td><PortfolioNumber value={row.gain} available={row.hasGain} signed /></Table.Td>
              <Table.Td><PortfolioNumber value={row.gainPct} available={row.hasGainPct} pct /></Table.Td>
            </Table.Tr>
          ))}
        </Table.Tbody>
      </Table>
    </ScrollArea>
  )
}

export function InicioPage() {
  const dashboard = useDashboard()
  if (dashboard.isLoading) return <Loader aria-label="Cargando inicio" />
  if (dashboard.isError) return <ErrorAlert error={dashboard.error} title="Non se puido cargar o inicio" />
  const data = dashboard.data
  if (!data) return null
  if (data.assetCount === 0) {
    return <EmptyState title="Aínda non hai ativos" description="Engade o primeiro ativo para comezar a ver o resumo." action={<Button component={Link} to="/activos">Ir a Portofolio</Button>} />
  }
  const evo = mapLineChartDTO(data.evolution)
  const pendingText = data.portfolio.pendingCount > 0
    ? `${data.portfolio.pendingCount} pendente(s) de pechar en ${formatYearMonth(data.portfolio.period)}`
    : `Todo pechado en ${formatYearMonth(data.portfolio.period)} ✓`

  return (
    <Stack>
      <PageHeader title="Inicio" description="Resumo da carteira" />
      <SimpleGrid cols={{ base: 1, sm: 2, lg: 5 }}>
        <KpiCard label="Aporte total" value={formatUSD(data.kpis.totalInvested)} />
        <KpiCard label="Valor neto" value={data.kpis.hasCurrentValue ? formatUSD(data.kpis.currentValue) : '—'} />
        <KpiCard label="G/P total" value={data.kpis.hasTotalGain ? formatSignedUSD(data.kpis.totalGain) : '—'} {...(data.kpis.totalGain > 0 ? { color: 'green' } : data.kpis.totalGain < 0 ? { color: 'red' } : {})} />
        <KpiCard label="Dividendos totais" value={formatUSD(data.kpis.totalDividends)} />
        <KpiCard label="Índice medio mensual" value={data.kpis.hasAverages ? formatPct(data.kpis.avgIndexPct) : '—'} />
      </SimpleGrid>
      <SimpleGrid cols={{ base: 1, lg: 2 }}>
        <Card withBorder><Stack><Title order={3}>Evolución total</Title>{data.evolution.months.length === 0 ? <Text c="dimmed">Aínda non hai resultados rexistrados.</Text> : <LineChart h={300} data={evo.data} dataKey="month" series={evo.series} valueFormatter={formatUSD} yAxisProps={{ tickFormatter: formatAxisUSD, width: 64 }} connectNulls={false} withLegend />}</Stack></Card>
        <Card withBorder><Stack><Title order={3}>Distribución</Title>{data.distribution.total <= 0 ? <Text c="dimmed">Aínda non hai aportes rexistrados.</Text> : <DistributionChart distribution={data.distribution} size={200} />}</Stack></Card>
      </SimpleGrid>
      <Card withBorder>
        <Stack>
          <Group justify="space-between" align="flex-start">
            <Stack gap={2}>
              <Title order={3}>Portofolio</Title>
              <Text c={data.portfolio.pendingCount > 0 ? 'yellow.8' : 'green'}>{pendingText}</Text>
            </Stack>
            <Button component={Link} to="/resultados?tab=pechar" variant={data.portfolio.pendingCount > 0 ? 'filled' : 'light'}>Pechar mes</Button>
          </Group>
          <PortfolioTable rows={data.portfolio.rows} />
        </Stack>
      </Card>
    </Stack>
  )
}
