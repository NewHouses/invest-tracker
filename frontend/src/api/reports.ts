import { useQuery } from '@tanstack/react-query'

import { apiFetch } from '@/api/client'
import { queryKeys } from '@/api/queryKeys'
import type { Asset, AssetType, MonthlySummary, YearMonth } from '@/api/types'

const monthQuery = (period: YearMonth) => `year=${period.year}&month=${period.month}`

export type AssetMonthReport = {
  asset: Asset
  period: YearMonth
  summary: MonthlySummary
  gain: number
  gainPct: number
  hasResult: boolean
  hasGainPct: boolean
}

export type TypeMonthReport = {
  type: AssetType
  period: YearMonth
  assets: Asset[]
  active: Array<{ asset: Asset; summary: MonthlySummary }>
  totalInvested: number
  investedInMonth: number
  holding: number
  resultSum: number
  holdingForResult: number
  withResult: number
  gain: number
  gainPct: number
  hasGainPct: boolean
  partial: boolean
}

export type TotalMonthReport = {
  period: YearMonth
  totalAssets: number
  assetsActive: number
  assetsWithResult: number
  partial: boolean
  totalInvested: number
  investedInMonth: number
  investedPlusPrevDividends: number
  dividends: number
  dividendsPrev: number
  holdingNoDiv: number
  holdingWithDiv: number
  baseNoDiv: number
  baseWithDiv: number
  resultNoDiv: number
  resultWithDiv: number
  gainNoDiv: number
  gainWithDiv: number
  pctNoDiv: number
  pctWithDiv: number
  hasResults: boolean
  hasMetrics: boolean
  hasPctWithDiv: boolean
  averages: {
    months: number
    pctNoDiv: number
    gainNoDiv: number
    pctWithDiv: number
    gainWithDiv: number
  }
}

export type HistoryRow = {
  period: YearMonth
  aporte: number
  holding: number
  result: number
  gain: number
  gainPct: number
  hasMetrics: boolean
}

export type AssetHistory = {
  asset: Asset
  rows: HistoryRow[]
  current: HistoryRow | null
  totalInvested: number
  avgIndexPct: number
  avgGain: number
  hasAverages: boolean
  totalGain: number
  hasTotalGain: boolean
}

export type TypeHistory = Omit<AssetHistory, 'asset'> & {
  type: AssetType
  assetCount: number
}

export type TotalHistoryRow = {
  period: YearMonth
  aporte: number
  fondos: number
  dividends: number
  result: number
  gain: number
  gainPct: number
  hasMetrics: boolean
}

export type TotalHistory = {
  assetCount: number
  rows: TotalHistoryRow[]
  current: TotalHistoryRow | null
  lifetimeAporte: number
  avgIndexPct: number
  avgGain: number
  hasAverages: boolean
  totalGain: number
  hasTotalGain: boolean
  totalDividends: number
  currentValue: number
  hasCurrentValue: boolean
}

export function useAssetMonthReport(assetId: number | null, period: YearMonth) {
  return useQuery({
    queryKey: queryKeys.reports.assetMonth(assetId ?? 0, period),
    queryFn: ({ signal }) => apiFetch<AssetMonthReport>(`/api/reports/asset/${assetId}/month?${monthQuery(period)}`, { signal }),
    enabled: assetId !== null && assetId > 0,
  })
}

export function useTypeMonthReport(type: AssetType | null, period: YearMonth) {
  return useQuery({
    queryKey: queryKeys.reports.typeMonth(type ?? 'accion', period),
    queryFn: ({ signal }) => apiFetch<TypeMonthReport>(`/api/reports/type/${type}/month?${monthQuery(period)}`, { signal }),
    enabled: type !== null,
  })
}

export function useTotalMonthReport(period: YearMonth) {
  return useQuery({
    queryKey: queryKeys.reports.totalMonth(period),
    queryFn: ({ signal }) => apiFetch<TotalMonthReport>(`/api/reports/total/month?${monthQuery(period)}`, { signal }),
  })
}

export function useAssetHistoryReport(assetId: number | null) {
  return useQuery({
    queryKey: queryKeys.reports.assetHistory(assetId ?? 0),
    queryFn: ({ signal }) => apiFetch<AssetHistory>(`/api/reports/asset/${assetId}/history`, { signal }),
    enabled: assetId !== null && assetId > 0,
  })
}

export function useTypeHistoryReport(type: AssetType | null) {
  return useQuery({
    queryKey: queryKeys.reports.typeHistory(type ?? 'accion'),
    queryFn: ({ signal }) => apiFetch<TypeHistory>(`/api/reports/type/${type}/history`, { signal }),
    enabled: type !== null,
  })
}

export function useTotalHistoryReport() {
  return useQuery({
    queryKey: queryKeys.reports.totalHistory,
    queryFn: ({ signal }) => apiFetch<TotalHistory>('/api/reports/total/history', { signal }),
  })
}
