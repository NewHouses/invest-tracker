import { Button, Card, Group, Loader, SimpleGrid, Stack, Text, Title } from '@mantine/core'
import { LineChart } from '@mantine/charts'
import { Link } from 'react-router'

import { useDashboard } from '@/api/dashboard'
import { EmptyState } from '@/components/EmptyState'
import { ErrorAlert } from '@/components/ErrorAlert'
import { PageHeader } from '@/components/PageHeader'
import { formatAxisUSD, formatPct, formatSignedUSD, formatUSD, formatYearMonth } from '@/lib/format'
import { DistributionChart } from '@/features/charts/DistributionChart'
import { mapLineChartDTO } from '@/features/charts/lineChartMapper'

function KpiCard({ label, value, color }: { label: string; value: string; color?: string }) {
  const valueProps = color ? { c: color } : {}
  return <Card withBorder><Text size="sm" c="dimmed">{label}</Text><Text fw={700} size="xl" {...valueProps}>{value}</Text></Card>
}

export function InicioPage() {
  const dashboard = useDashboard()
  if (dashboard.isLoading) return <Loader aria-label="Cargando inicio" />
  if (dashboard.isError) return <ErrorAlert error={dashboard.error} title="Non se puido cargar o inicio" />
  const data = dashboard.data
  if (!data) return null
  if (data.assetCount === 0) {
    return <EmptyState title="Aínda non hai activos" description="Engade o primeiro activo para comezar a ver o resumo." action={<Button component={Link} to="/activos">Ir a activos</Button>} />
  }
  const evo = mapLineChartDTO(data.evolution)
  return (
    <Stack>
      <PageHeader title="Inicio" description="Resumo da carteira" />
      <SimpleGrid cols={{ base: 1, sm: 2, lg: 5 }}>
        <KpiCard label="Aporte total" value={formatUSD(data.kpis.totalInvested)} />
        <KpiCard label="Valor actual" value={data.kpis.hasCurrentValue ? formatUSD(data.kpis.currentValue) : '—'} />
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
          <Group justify="space-between"><Title order={3}>Pendente de pechar {formatYearMonth(data.pending.period)}</Title>{data.pending.items.length > 0 ? <Button component={Link} to="/resultados?tab=pechar">Pechar mes</Button> : null}</Group>
          {data.pending.items.length === 0 ? <Text c="green">Todo pechado ✓</Text> : data.pending.items.map((item) => <Group key={item.asset.id} justify="space-between"><Text>{item.asset.name}</Text><Text c="dimmed">No activo: {formatUSD(item.holding)}</Text></Group>)}
        </Stack>
      </Card>
    </Stack>
  )
}
