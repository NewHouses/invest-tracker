import { Card, Group, Loader, SegmentedControl, Stack, Text } from '@mantine/core'
import { useState } from 'react'

import { useAssets } from '@/api/assets'
import { useAssetChart, useDistributionChart, useTotalChart, useTypeAssetsChart, useTypeChart, useTypesChart } from '@/api/charts'
import type { AssetType } from '@/api/types'
import { AssetSelect } from '@/components/AssetSelect'
import { AssetTypeSelect } from '@/components/AssetTypeSelect'
import { EmptyState } from '@/components/EmptyState'
import { ErrorAlert } from '@/components/ErrorAlert'
import { PageHeader } from '@/components/PageHeader'
import { LineChartPanel } from '@/features/charts/common'
import { DistributionChart as DistributionChartView } from '@/features/charts/DistributionChart'
import { formatUSD } from '@/lib/format'

type ChartKind = 'distribution' | 'asset' | 'type' | 'typeAssets' | 'types' | 'total'

function DistributionChart() {
  const query = useDistributionChart()
  if (query.isLoading) return <Loader />
  if (query.isError) return <ErrorAlert error={query.error} />
  const data = query.data
  if (!data) return null
  if (!data.hasAssets) return <EmptyState title="Aínda non hai activos" />
  if (data.total === 0) return <EmptyState title="Aínda non hai aportes rexistrados" />
  return <Card withBorder><Text fw={700} mb="md">Total aportado: {formatUSD(data.total)}</Text><DistributionChartView distribution={data} size={260} /></Card>
}
function AssetEvolution({ assetId }: { assetId: number | null }) {
  const query = useAssetChart(assetId)
  if (assetId === null) return <EmptyState title="Escolle un activo" />
  if (query.isError) return <ErrorAlert error={query.error} />
  return <LineChartPanel chart={query.data} loading={query.isLoading} emptyNoResults="Aínda non hai resultados rexistrados para este activo." />
}
function TypeEvolution({ type, assets }: { type: AssetType | null; assets: boolean }) {
  const aggregated = useTypeChart(assets ? null : type)
  const perAsset = useTypeAssetsChart(assets ? type : null)
  const query = assets ? perAsset : aggregated
  if (type === null) return <EmptyState title="Escolle un tipo" />
  if (query.isError) return <ErrorAlert error={query.error} />
  return <LineChartPanel chart={query.data} loading={query.isLoading} emptyNoAssets="Non hai activos deste tipo" emptyNoResults="Aínda non hai resultados rexistrados para este tipo." />
}
function TypesEvolution() {
  const query = useTypesChart()
  if (query.isError) return <ErrorAlert error={query.error} />
  return <LineChartPanel chart={query.data} loading={query.isLoading} emptyNoResults="Aínda non hai resultados rexistrados." />
}
function TotalEvolution() {
  const query = useTotalChart()
  if (query.isError) return <ErrorAlert error={query.error} />
  return <LineChartPanel chart={query.data} loading={query.isLoading} emptyNoResults="Aínda non hai resultados rexistrados." />
}

export function GraficasPage() {
  const assets = useAssets()
  const [kind, setKind] = useState<ChartKind>('distribution')
  const [assetId, setAssetId] = useState<number | null>(null)
  const [type, setType] = useState<AssetType | null>('accion')
  const effectiveAssetId = assetId ?? assets.data?.[0]?.id ?? null
  return (
    <Stack>
      <PageHeader title="Gráficas" description="Visualizacións da carteira" />
      <Group align="end"><SegmentedControl value={kind} onChange={(v) => setKind(v as ChartKind)} data={[
        { label: 'Distribución de aportes', value: 'distribution' },
        { label: 'Evolución dun activo', value: 'asset' },
        { label: 'Evolución dun tipo (agregada)', value: 'type' },
        { label: 'Evolución dos activos dun tipo', value: 'typeAssets' },
        { label: 'Evolución dos tipos', value: 'types' },
        { label: 'Evolución do resultado total', value: 'total' },
      ]} />{kind === 'asset' ? <AssetSelect value={effectiveAssetId} onChange={setAssetId} label="Activo" /> : null}{kind === 'type' || kind === 'typeAssets' ? <AssetTypeSelect value={type} onChange={setType} label="Tipo" /> : null}</Group>
      {assets.isLoading ? <Text c="dimmed">Cargando activos…</Text> : null}
      {kind === 'distribution' ? <DistributionChart /> : null}
      {kind === 'asset' ? <AssetEvolution assetId={effectiveAssetId} /> : null}
      {kind === 'type' ? <TypeEvolution type={type} assets={false} /> : null}
      {kind === 'typeAssets' ? <TypeEvolution type={type} assets /> : null}
      {kind === 'types' ? <TypesEvolution /> : null}
      {kind === 'total' ? <TotalEvolution /> : null}
    </Stack>
  )
}
