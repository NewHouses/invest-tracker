import { Group, Loader, SegmentedControl, Stack, Text, Title } from '@mantine/core'
import { useMemo, useState } from 'react'
import { useSearchParams } from 'react-router'

import { useAssetMonthReport, useTotalHistoryReport, useTotalMonthReport, useTypeHistoryReport, useTypeMonthReport } from '@/api/reports'
import type { AssetType, YearMonth } from '@/api/types'
import { useAssets } from '@/api/assets'
import { useAssetTypeLabel } from '@/api/meta'
import { AssetSelect } from '@/components/AssetSelect'
import { AssetTypeSelect } from '@/components/AssetTypeSelect'
import { EmptyState } from '@/components/EmptyState'
import { ErrorAlert } from '@/components/ErrorAlert'
import { PageHeader } from '@/components/PageHeader'
import { YearMonthInput } from '@/components/YearMonthInput'
import { KvTable } from '@/components/ReportTable'
import { AssetHistoryPanel, LoadingReport, ReportCard, TotalHistoryView, TypeHistoryView } from '@/features/reports/AssetHistoryPanel'
import { formatPct, formatSignedUSD, formatUSD, formatYearMonth } from '@/lib/format'
import { currentYearMonth } from '@/lib/yearMonth'

const assetTypes: AssetType[] = ['accion', 'indice', 'copy_trading', 'fondo']

type Vista = 'mensual' | 'historico'
type Ambito = 'activo' | 'tipo' | 'total'

const dash = (ok: boolean, value: string) => (ok ? value : '—')
const na = (ok: boolean, value: string) => (ok ? value : 'n/a')

function useQueryState() {
  const [params, setParams] = useSearchParams()
  const vista = params.get('vista') === 'mensual' ? 'mensual' : 'historico'
  const ambitoParam = params.get('ambito')
  const ambito: Ambito = ambitoParam === 'tipo' || ambitoParam === 'activo' ? ambitoParam : 'total'
  const update = (next: Partial<{ vista: Vista; ambito: Ambito }>) => {
    const merged = new URLSearchParams(params)
    if (next.vista) merged.set('vista', next.vista)
    if (next.ambito) merged.set('ambito', next.ambito)
    setParams(merged, { replace: true })
  }
  return { vista, ambito, update }
}

function AssetMonth({ assetId, period }: { assetId: number | null; period: YearMonth }) {
  const report = useAssetMonthReport(assetId, period)
  if (assetId === null) return <EmptyState title="Escolle un ativo" />
  if (report.isLoading) return <LoadingReport />
  if (report.isError) return <ErrorAlert error={report.error} />
  const r = report.data
  if (!r) return null
  return <ReportCard><Title order={3}>{r.asset.name} · {formatYearMonth(r.period)}</Title><KvTable rows={[
    { label: 'Investido ata o mes', value: formatUSD(r.summary.totalInvestedUpTo) },
    { label: 'Investido este mes', value: formatUSD(r.summary.investedInMonth) },
    { label: 'No ativo', value: formatUSD(r.summary.estimatedHolding) },
    { label: 'Resultado', value: dash(r.hasResult, formatUSD(r.summary.result)) },
    { label: 'Gañanzas/Perdas', value: dash(r.hasResult, formatSignedUSD(r.gain)) },
    { label: 'Índice', value: r.hasResult ? na(r.hasGainPct, formatPct(r.gainPct)) : '—' },
  ]} /></ReportCard>
}

function TypeMonth({ type, period }: { type: AssetType | null; period: YearMonth }) {
  const typeLabel = useAssetTypeLabel()
  const report = useTypeMonthReport(type, period)
  if (type === null) return <EmptyState title="Escolle un tipo" />
  if (report.isLoading) return <LoadingReport />
  if (report.isError) return <ErrorAlert error={report.error} />
  const r = report.data
  if (!r) return null
  if (r.assets.length === 0) return <EmptyState title={`Non hai ativos de tipo ${typeLabel(type)}.`} />
  if (r.active.length === 0) return <EmptyState title={`Non hai ativos de tipo ${typeLabel(type)} con capital investido en ${formatYearMonth(r.period)}.`} />
  const partial = r.partial ? ` (${r.withResult}/${r.active.length} ativos)` : ''
  return <ReportCard><Title order={3}>{typeLabel(type)} · {formatYearMonth(r.period)}</Title><KvTable rows={[
    { label: 'Ativos incluídos', value: <Stack gap={2}>{r.active.map((e) => <Text key={e.asset.id}>{e.asset.name} (no ativo: {formatUSD(e.summary.estimatedHolding)})</Text>)}</Stack> },
    { label: 'Investido ata o mes', value: formatUSD(r.totalInvested) },
    { label: 'Investido este mes', value: formatUSD(r.investedInMonth) },
    { label: 'No ativo', value: formatUSD(r.holding) },
    { label: r.partial ? 'Resultado (parc.)' : 'Resultado', value: r.withResult === 0 ? '—' : `${formatUSD(r.resultSum)}${partial}` },
    { label: r.partial ? 'Gañanzas/Perdas (parc.)' : 'Gañanzas/Perdas', value: r.withResult === 0 ? '—' : formatSignedUSD(r.gain) },
    { label: r.partial ? 'Índice (parc.)' : 'Índice', value: r.withResult === 0 ? '—' : na(r.hasGainPct, formatPct(r.gainPct)) },
  ]} /></ReportCard>
}

function TotalMonth({ period }: { period: YearMonth }) {
  const report = useTotalMonthReport(period)
  if (report.isLoading) return <LoadingReport />
  if (report.isError) return <ErrorAlert error={report.error} />
  const r = report.data
  if (!r) return null
  if (r.totalAssets === 0) return <EmptyState title="Aínda non hai ativos" />
  if (r.assetsActive === 0) return <EmptyState title={`Non hai ativos con capital investido en ${formatYearMonth(r.period)}.`} />
  const coverage = r.partial ? ` (${r.assetsWithResult}/${r.assetsActive} ativos con resultado)` : ''
  return <ReportCard><Title order={3}>Informe total · {formatYearMonth(r.period)}</Title><KvTable rows={[
    { label: 'Total investido ata o mes', value: formatUSD(r.totalInvested) },
    { label: 'Investido este mes', value: formatUSD(r.investedInMonth) },
    { label: 'Investimento + dividendos prev. mes', value: formatUSD(r.investedPlusPrevDividends) },
    { label: 'No ativo (sen div)', value: formatUSD(r.holdingNoDiv) },
    { label: 'No ativo (con div)', value: formatUSD(r.holdingWithDiv) },
    { label: 'Dividendos este mes', value: formatUSD(r.dividends) },
    { label: 'Resultado (sen div)', value: r.hasResults ? `${formatUSD(r.resultNoDiv)}${coverage}` : '—' },
    { label: 'Resultado total (con div)', value: r.hasResults ? formatUSD(r.resultWithDiv) : '—' },
    { label: 'Gañanzas/Perdas', value: r.hasResults ? na(r.hasMetrics, formatSignedUSD(r.gainNoDiv)) : '—' },
    { label: 'Gañanzas/Perdas (con div)', value: r.hasResults ? na(r.hasMetrics, formatSignedUSD(r.gainWithDiv)) : '—' },
    { label: 'Índice', value: r.hasResults ? na(r.hasMetrics, formatPct(r.pctNoDiv)) : '—' },
    { label: 'Índice (con div)', value: r.hasResults ? na(r.hasPctWithDiv, formatPct(r.pctWithDiv)) : '—' },
  ]} /><Title order={4}>{r.averages.months === 0 ? 'Promedios mensuais: sen meses con resultados.' : `Promedios mensuais (${r.averages.months} mes(es) con resultado ata ${formatYearMonth(r.period)})`}</Title>{r.averages.months > 0 ? <KvTable rows={[
    { label: 'Índice medio mensual (sen div)', value: formatPct(r.averages.pctNoDiv) },
    { label: 'Gañanza media mensual (sen div)', value: formatSignedUSD(r.averages.gainNoDiv) },
    { label: 'Índice medio mensual (con div)', value: formatPct(r.averages.pctWithDiv) },
    { label: 'Gañanza media mensual (con div)', value: formatSignedUSD(r.averages.gainWithDiv) },
  ]} /> : null}</ReportCard>
}

function TypeHistoryPanel({ type }: { type: AssetType | null }) {
  const history = useTypeHistoryReport(type)
  if (type === null) return <EmptyState title="Escolle un tipo" />
  if (history.isLoading) return <LoadingReport />
  if (history.isError) return <ErrorAlert error={history.error} />
  return history.data ? <TypeHistoryView history={history.data} /> : null
}
function TotalHistoryPanel() {
  const history = useTotalHistoryReport()
  if (history.isLoading) return <LoadingReport />
  if (history.isError) return <ErrorAlert error={history.error} />
  return history.data ? <TotalHistoryView history={history.data} /> : null
}

export function InformesPage() {
  const { vista, ambito, update } = useQueryState()
  const assets = useAssets()
  const [assetId, setAssetId] = useState<number | null>(null)
  const [type, setType] = useState<AssetType | null>('accion')
  const [period, setPeriod] = useState<YearMonth>(currentYearMonth())
  const effectiveAssetId = assetId ?? assets.data?.[0]?.id ?? null
  const presentTypes = useMemo(() => assetTypes.filter((t) => assets.data?.some((a) => a.type === t)), [assets.data])
  return (
    <Stack>
      <PageHeader title="Informes" description="Informes mensuais e históricos" />
      <Group align="end"><SegmentedControl value={vista} onChange={(v) => update({ vista: v as Vista })} data={[{ label: 'Mensual', value: 'mensual' }, { label: 'Histórico', value: 'historico' }]} /><SegmentedControl value={ambito} onChange={(v) => update({ ambito: v as Ambito })} data={[{ label: 'Total', value: 'total' }, { label: 'Tipo', value: 'tipo' }, { label: 'Ativo', value: 'activo' }]} />{ambito === 'activo' ? <AssetSelect value={effectiveAssetId} onChange={setAssetId} label="Ativo" /> : null}{ambito === 'tipo' ? <AssetTypeSelect value={type} onChange={setType} label="Tipo" {...(presentTypes.length > 0 ? { only: presentTypes } : {})} /> : null}{vista === 'mensual' ? <YearMonthInput value={period} onChange={(v) => v && setPeriod(v)} label="Mes" required /> : null}</Group>
      {assets.isLoading ? <Loader /> : null}
      {vista === 'mensual' && ambito === 'activo' ? <AssetMonth assetId={effectiveAssetId} period={period} /> : null}
      {vista === 'mensual' && ambito === 'tipo' ? <TypeMonth type={type} period={period} /> : null}
      {vista === 'mensual' && ambito === 'total' ? <TotalMonth period={period} /> : null}
      {vista === 'historico' && ambito === 'activo' && effectiveAssetId !== null ? <AssetHistoryPanel assetId={effectiveAssetId} /> : null}
      {vista === 'historico' && ambito === 'activo' && effectiveAssetId === null ? <EmptyState title="Escolle un ativo" /> : null}
      {vista === 'historico' && ambito === 'tipo' ? <TypeHistoryPanel type={type} /> : null}
      {vista === 'historico' && ambito === 'total' ? <TotalHistoryPanel /> : null}
    </Stack>
  )
}
