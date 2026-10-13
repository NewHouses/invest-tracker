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
import { t } from '@/i18n'
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
  if (assetId === null) return <EmptyState title={t('Escolle un activo')} />
  if (report.isLoading) return <LoadingReport />
  if (report.isError) return <ErrorAlert error={report.error} />
  const r = report.data
  if (!r) return null
  return <ReportCard><Title order={3}>{r.asset.name} · {formatYearMonth(r.period)}</Title><KvTable rows={[
    { label: t('Investido ata o mes'), value: formatUSD(r.summary.totalInvestedUpTo) },
    { label: t('Investido este mes'), value: formatUSD(r.summary.investedInMonth) },
    { label: t('No activo'), value: formatUSD(r.summary.estimatedHolding) },
    { label: t('Resultado'), value: dash(r.hasResult, formatUSD(r.summary.result)) },
    { label: t('Gañanzas/Perdas'), value: dash(r.hasResult, formatSignedUSD(r.gain)) },
    { label: t('Índice'), value: r.hasResult ? na(r.hasGainPct, formatPct(r.gainPct)) : '—' },
  ]} /></ReportCard>
}

function TypeMonth({ type, period }: { type: AssetType | null; period: YearMonth }) {
  const typeLabel = useAssetTypeLabel()
  const report = useTypeMonthReport(type, period)
  if (type === null) return <EmptyState title={t('Escolle un tipo')} />
  if (report.isLoading) return <LoadingReport />
  if (report.isError) return <ErrorAlert error={report.error} />
  const r = report.data
  if (!r) return null
  if (r.assets.length === 0) return <EmptyState title={t('Non hai activos de tipo {type}.', { type: typeLabel(type) })} />
  if (r.active.length === 0) return <EmptyState title={t('Non hai activos de tipo {type} con capital investido en {period}.', { type: typeLabel(type), period: formatYearMonth(r.period) })} />
  const partial = r.partial ? t(' ({done}/{total} activos)', { done: r.withResult, total: r.active.length }) : ''
  return <ReportCard><Title order={3}>{typeLabel(type)} · {formatYearMonth(r.period)}</Title><KvTable rows={[
    { label: t('Activos incluídos'), value: <Stack gap={2}>{r.active.map((e) => <Text key={e.asset.id}>{e.asset.name} {t('(no activo: {amount})', { amount: formatUSD(e.summary.estimatedHolding) })}</Text>)}</Stack> },
    { label: t('Investido ata o mes'), value: formatUSD(r.totalInvested) },
    { label: t('Investido este mes'), value: formatUSD(r.investedInMonth) },
    { label: t('No activo'), value: formatUSD(r.holding) },
    { label: r.partial ? t('Resultado (parc.)') : t('Resultado'), value: r.withResult === 0 ? '—' : `${formatUSD(r.resultSum)}${partial}` },
    { label: r.partial ? t('Gañanzas/Perdas (parc.)') : t('Gañanzas/Perdas'), value: r.withResult === 0 ? '—' : formatSignedUSD(r.gain) },
    { label: r.partial ? t('Índice (parc.)') : t('Índice'), value: r.withResult === 0 ? '—' : na(r.hasGainPct, formatPct(r.gainPct)) },
  ]} /></ReportCard>
}

function TotalMonth({ period }: { period: YearMonth }) {
  const report = useTotalMonthReport(period)
  if (report.isLoading) return <LoadingReport />
  if (report.isError) return <ErrorAlert error={report.error} />
  const r = report.data
  if (!r) return null
  if (r.totalAssets === 0) return <EmptyState title={t('Aínda non hai activos')} />
  if (r.assetsActive === 0) return <EmptyState title={t('Non hai activos con capital investido en {period}.', { period: formatYearMonth(r.period) })} />
  const coverage = r.partial ? t(' ({done}/{total} activos con resultado)', { done: r.assetsWithResult, total: r.assetsActive }) : ''
  return <ReportCard><Title order={3}>Informe total · {formatYearMonth(r.period)}</Title><KvTable rows={[
    { label: t('Total investido ata o mes'), value: formatUSD(r.totalInvested) },
    { label: t('Investido este mes'), value: formatUSD(r.investedInMonth) },
    { label: t('Investimento + dividendos prev. mes'), value: formatUSD(r.investedPlusPrevDividends) },
    { label: t('No activo (sen div)'), value: formatUSD(r.holdingNoDiv) },
    { label: t('No activo (con div)'), value: formatUSD(r.holdingWithDiv) },
    { label: t('Dividendos este mes'), value: formatUSD(r.dividends) },
    { label: t('Resultado (sen div)'), value: r.hasResults ? `${formatUSD(r.resultNoDiv)}${coverage}` : '—' },
    { label: t('Resultado total (con div)'), value: r.hasResults ? formatUSD(r.resultWithDiv) : '—' },
    { label: t('Gañanzas/Perdas'), value: r.hasResults ? na(r.hasMetrics, formatSignedUSD(r.gainNoDiv)) : '—' },
    { label: t('Gañanzas/Perdas (con div)'), value: r.hasResults ? na(r.hasMetrics, formatSignedUSD(r.gainWithDiv)) : '—' },
    { label: t('Índice'), value: r.hasResults ? na(r.hasMetrics, formatPct(r.pctNoDiv)) : '—' },
    { label: t('Índice (con div)'), value: r.hasResults ? na(r.hasPctWithDiv, formatPct(r.pctWithDiv)) : '—' },
  ]} /><Title order={4}>{r.averages.months === 0 ? t('Promedios mensuais: sen meses con resultados.') : t('Promedios mensuais ({count} mes(es) con resultado ata {period})', { count: r.averages.months, period: formatYearMonth(r.period) })}</Title>{r.averages.months > 0 ? <KvTable rows={[
    { label: t('Índice medio mensual (sen div)'), value: formatPct(r.averages.pctNoDiv) },
    { label: t('Gañanza media mensual (sen div)'), value: formatSignedUSD(r.averages.gainNoDiv) },
    { label: t('Índice medio mensual (con div)'), value: formatPct(r.averages.pctWithDiv) },
    { label: t('Gañanza media mensual (con div)'), value: formatSignedUSD(r.averages.gainWithDiv) },
  ]} /> : null}</ReportCard>
}

function TypeHistoryPanel({ type }: { type: AssetType | null }) {
  const history = useTypeHistoryReport(type)
  if (type === null) return <EmptyState title={t('Escolle un tipo')} />
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
      <PageHeader title={t('Informes')} description={t('Informes mensuais e históricos')} />
      <Group align="end"><SegmentedControl value={vista} onChange={(v) => update({ vista: v as Vista })} data={[{ label: t('Mensual'), value: 'mensual' }, { label: t('Histórico'), value: 'historico' }]} /><SegmentedControl value={ambito} onChange={(v) => update({ ambito: v as Ambito })} data={[{ label: t('Total'), value: 'total' }, { label: t('Tipo'), value: 'tipo' }, { label: t('Activo'), value: 'activo' }]} />{ambito === 'activo' ? <AssetSelect value={effectiveAssetId} onChange={setAssetId} label={t('Activo')} /> : null}{ambito === 'tipo' ? <AssetTypeSelect value={type} onChange={setType} label={t('Tipo')} {...(presentTypes.length > 0 ? { only: presentTypes } : {})} /> : null}{vista === 'mensual' ? <YearMonthInput value={period} onChange={(v) => v && setPeriod(v)} label={t('Mes')} required /> : null}</Group>
      {assets.isLoading ? <Loader /> : null}
      {vista === 'mensual' && ambito === 'activo' ? <AssetMonth assetId={effectiveAssetId} period={period} /> : null}
      {vista === 'mensual' && ambito === 'tipo' ? <TypeMonth type={type} period={period} /> : null}
      {vista === 'mensual' && ambito === 'total' ? <TotalMonth period={period} /> : null}
      {vista === 'historico' && ambito === 'activo' && effectiveAssetId !== null ? <AssetHistoryPanel assetId={effectiveAssetId} /> : null}
      {vista === 'historico' && ambito === 'activo' && effectiveAssetId === null ? <EmptyState title={t('Escolle un activo')} /> : null}
      {vista === 'historico' && ambito === 'tipo' ? <TypeHistoryPanel type={type} /> : null}
      {vista === 'historico' && ambito === 'total' ? <TotalHistoryPanel /> : null}
    </Stack>
  )
}
