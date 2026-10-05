import { Button, Card, Group, Loader, Stack, Text } from '@mantine/core'
import { useState } from 'react'
import { useSearchParams } from 'react-router'

import { useAssets } from '@/api/assets'
import { useAllAssetsChart, useAssetChart, useDistributionChart, useTotalChart, useTypeAssetsChart, useTypeChart, useTypesChart } from '@/api/charts'
import type { AssetType } from '@/api/types'
import { AssetSelect } from '@/components/AssetSelect'
import { AssetTypeSelect } from '@/components/AssetTypeSelect'
import { EmptyState } from '@/components/EmptyState'
import { ErrorAlert } from '@/components/ErrorAlert'
import { PageHeader } from '@/components/PageHeader'
import { LineChartPanel } from '@/features/charts/common'
import { DistributionChart as DistributionChartView } from '@/features/charts/DistributionChart'
import { formatUSD } from '@/lib/format'

type ChartKind = 'distribution' | 'asset' | 'type' | 'typeAssets' | 'types' | 'allAssets' | 'total'

const chartOptions: Array<{ label: string; value: ChartKind }> = [
  { label: 'Distribución de aportes', value: 'distribution' },
  { label: 'Evolución dun ativo', value: 'asset' },
  { label: 'Evolución dun tipo (agregada)', value: 'type' },
  { label: 'Evolución dos ativos dun tipo', value: 'typeAssets' },
  { label: 'Evolución dos tipos', value: 'types' },
  { label: 'Evolución de todos os ativos', value: 'allAssets' },
  { label: 'Evolución do resultado total', value: 'total' },
]

function isChartKind(value: string | null): value is ChartKind {
  return chartOptions.some((option) => option.value === value)
}

function DistributionChart() {
  const query = useDistributionChart()
  if (query.isLoading) return <Loader />
  if (query.isError) return <ErrorAlert error={query.error} />
  const data = query.data
  if (!data) return null
  if (!data.hasAssets) return <EmptyState title="Aínda non hai ativos" />
  if (data.total === 0) return <EmptyState title="Aínda non hai aportes rexistrados" />
  return (
    <Card withBorder>
      <Text fw={700} mb="md">Distribución de aportes · Total aportado: {formatUSD(data.total)}</Text>
      <DistributionChartView distribution={data} size={260} />
    </Card>
  )
}

function AssetEvolution({ assetId }: { assetId: number | null }) {
  const query = useAssetChart(assetId)
  if (assetId === null) return <EmptyState title="Escolle un ativo" />
  if (query.isError) return <ErrorAlert error={query.error} />
  return <LineChartPanel chart={query.data} loading={query.isLoading} title="Evolución dun ativo" emptyNoResults="Aínda non hai resultados rexistrados para este ativo." />
}

function TypeEvolution({ type, assets }: { type: AssetType | null; assets: boolean }) {
  const aggregated = useTypeChart(assets ? null : type)
  const perAsset = useTypeAssetsChart(assets ? type : null)
  const query = assets ? perAsset : aggregated
  if (type === null) return <EmptyState title="Escolle un tipo" />
  if (query.isError) return <ErrorAlert error={query.error} />
  return <LineChartPanel chart={query.data} loading={query.isLoading} title={assets ? 'Evolución dos ativos dun tipo' : 'Evolución dun tipo (agregada)'} emptyNoAssets="Non hai ativos deste tipo" emptyNoResults="Aínda non hai resultados rexistrados para este tipo." />
}

function TypesEvolution() {
  const query = useTypesChart()
  if (query.isError) return <ErrorAlert error={query.error} />
  return <LineChartPanel chart={query.data} loading={query.isLoading} title="Evolución dos tipos" emptyNoResults="Aínda non hai resultados rexistrados." />
}

function AllAssetsEvolution() {
  const query = useAllAssetsChart()
  if (query.isError) return <ErrorAlert error={query.error} />
  return <LineChartPanel chart={query.data} loading={query.isLoading} title="Evolución de todos os ativos" emptyNoAssets="Aínda non hai ativos" emptyNoResults="Aínda non hai resultados rexistrados." />
}

function TotalEvolution() {
  const query = useTotalChart()
  if (query.isError) return <ErrorAlert error={query.error} />
  return <LineChartPanel chart={query.data} loading={query.isLoading} title="Evolución do resultado total" emptyNoResults="Aínda non hai resultados rexistrados." />
}

export function GraficasPage() {
  const assets = useAssets()
  const [searchParams, setSearchParams] = useSearchParams()
  const [kind, setKind] = useState<ChartKind>(() => {
    const initialKind = searchParams.get('grafica')
    return isChartKind(initialKind) ? initialKind : 'distribution'
  })
  const [assetId, setAssetId] = useState<number | null>(null)
  const [type, setType] = useState<AssetType | null>('accion')
  const effectiveAssetId = assetId ?? assets.data?.[0]?.id ?? null
  const selectKind = (nextKind: ChartKind) => {
    setKind(nextKind)
    setSearchParams((current) => {
      const next = new URLSearchParams(current)
      next.set('grafica', nextKind)
      return next
    }, { replace: true })
  }

  return (
    <Stack>
      <PageHeader title="Gráficas" description="Visualizacións da carteira" />
      <Group align="end" wrap="wrap">
        <Group role="radiogroup" aria-label="Tipo de gráfica" gap="xs" wrap="wrap">
          {chartOptions.map((option) => (
            <Button key={option.value} type="button" size="xs" variant={kind === option.value ? 'filled' : 'light'} aria-pressed={kind === option.value} onClick={() => selectKind(option.value)}>
              {option.label}
            </Button>
          ))}
        </Group>
        {kind === 'asset' ? <AssetSelect value={effectiveAssetId} onChange={setAssetId} label="Ativo" /> : null}
        {kind === 'type' || kind === 'typeAssets' ? <AssetTypeSelect value={type} onChange={setType} label="Tipo" /> : null}
      </Group>
      {assets.isLoading ? <Text c="dimmed">Cargando ativos…</Text> : null}
      {kind === 'distribution' ? <DistributionChart /> : null}
      {kind === 'asset' ? <AssetEvolution assetId={effectiveAssetId} /> : null}
      {kind === 'type' ? <TypeEvolution type={type} assets={false} /> : null}
      {kind === 'typeAssets' ? <TypeEvolution type={type} assets /> : null}
      {kind === 'types' ? <TypesEvolution /> : null}
      {kind === 'allAssets' ? <AllAssetsEvolution /> : null}
      {kind === 'total' ? <TotalEvolution /> : null}
    </Stack>
  )
}
