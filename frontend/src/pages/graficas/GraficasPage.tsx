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
import { t, tKey } from '@/i18n'
import { formatUSD } from '@/lib/format'

type ChartKind = 'distribution' | 'asset' | 'type' | 'typeAssets' | 'types' | 'allAssets' | 'total'

const chartOptions: Array<{ label: string; value: ChartKind }> = [
  { label: tKey('Distribución de aportes'), value: 'distribution' },
  { label: tKey('Evolución dun activo'), value: 'asset' },
  { label: tKey('Evolución dun tipo (agregada)'), value: 'type' },
  { label: tKey('Evolución dos activos dun tipo'), value: 'typeAssets' },
  { label: tKey('Evolución dos tipos'), value: 'types' },
  { label: tKey('Evolución de todos os activos'), value: 'allAssets' },
  { label: tKey('Evolución do resultado total'), value: 'total' },
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
  if (!data.hasAssets) return <EmptyState title={t('Aínda non hai activos')} />
  if (data.total === 0) return <EmptyState title={t('Aínda non hai aportes rexistrados')} />
  return (
    <Card withBorder>
      <Text fw={700} mb="md">{t('Distribución de aportes · Total aportado: {total}', { total: formatUSD(data.total) })}</Text>
      <DistributionChartView distribution={data} size={260} />
    </Card>
  )
}

function AssetEvolution({ assetId }: { assetId: number | null }) {
  const query = useAssetChart(assetId)
  if (assetId === null) return <EmptyState title={t('Escolle un activo')} />
  if (query.isError) return <ErrorAlert error={query.error} />
  return <LineChartPanel chart={query.data} loading={query.isLoading} title={tKey('Evolución dun activo')} emptyNoResults={tKey('Aínda non hai resultados rexistrados para este activo.')} />
}

function TypeEvolution({ type, assets }: { type: AssetType | null; assets: boolean }) {
  const aggregated = useTypeChart(assets ? null : type)
  const perAsset = useTypeAssetsChart(assets ? type : null)
  const query = assets ? perAsset : aggregated
  if (type === null) return <EmptyState title={t('Escolle un tipo')} />
  if (query.isError) return <ErrorAlert error={query.error} />
  return <LineChartPanel chart={query.data} loading={query.isLoading} title={assets ? tKey('Evolución dos activos dun tipo') : tKey('Evolución dun tipo (agregada)')} emptyNoAssets={tKey('Non hai activos deste tipo')} emptyNoResults={tKey('Aínda non hai resultados rexistrados para este tipo.')} />
}

function TypesEvolution() {
  const query = useTypesChart()
  if (query.isError) return <ErrorAlert error={query.error} />
  return <LineChartPanel chart={query.data} loading={query.isLoading} title={tKey('Evolución dos tipos')} emptyNoResults={tKey('Aínda non hai resultados rexistrados.')} />
}

function AllAssetsEvolution() {
  const query = useAllAssetsChart()
  if (query.isError) return <ErrorAlert error={query.error} />
  return <LineChartPanel chart={query.data} loading={query.isLoading} title={tKey('Evolución de todos os activos')} emptyNoAssets={tKey('Aínda non hai activos')} emptyNoResults={tKey('Aínda non hai resultados rexistrados.')} />
}

function TotalEvolution() {
  const query = useTotalChart()
  if (query.isError) return <ErrorAlert error={query.error} />
  return <LineChartPanel chart={query.data} loading={query.isLoading} title={tKey('Evolución do resultado total')} emptyNoResults={tKey('Aínda non hai resultados rexistrados.')} />
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
      <PageHeader title={t('Gráficas')} description={t('Visualizacións da carteira')} />
      <Group align="end" wrap="wrap">
        <Group role="radiogroup" aria-label={t('Tipo de gráfica')} gap="xs" wrap="wrap">
          {chartOptions.map((option) => (
            <Button key={option.value} type="button" size="xs" variant={kind === option.value ? 'filled' : 'light'} aria-pressed={kind === option.value} onClick={() => selectKind(option.value)}>
              {t(option.label)}
            </Button>
          ))}
        </Group>
        {kind === 'asset' ? <AssetSelect value={effectiveAssetId} onChange={setAssetId} label={t('Activo')} /> : null}
        {kind === 'type' || kind === 'typeAssets' ? <AssetTypeSelect value={type} onChange={setType} label={t('Tipo')} /> : null}
      </Group>
      {assets.isLoading ? <Text c="dimmed">{t('Cargando activos…')}</Text> : null}
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
