import { useQuery } from '@tanstack/react-query'

import { apiFetch } from '@/api/client'
import { queryKeys } from '@/api/queryKeys'
import type { Asset, AssetType, YearMonth } from '@/api/types'

export type DistributionItem = { asset: Asset; label: string; value: number; pct: number }
export type Distribution = { items: DistributionItem[]; total: number; hasAssets: boolean }
export type LineChartDTO = {
  title: string
  months: YearMonth[]
  series: Array<{ label: string; values: Array<number | null> }>
  hasAssets: boolean
}

export function useDistributionChart() {
  return useQuery({
    queryKey: queryKeys.charts.distribution,
    queryFn: ({ signal }) => apiFetch<Distribution>('/api/charts/distribution', { signal }),
  })
}

export function useAssetChart(assetId: number | null) {
  return useQuery({
    queryKey: queryKeys.charts.asset(assetId ?? 0),
    queryFn: ({ signal }) => apiFetch<LineChartDTO>(`/api/charts/asset/${assetId}`, { signal }),
    enabled: assetId !== null && assetId > 0,
  })
}

export function useTypeChart(type: AssetType | null) {
  return useQuery({
    queryKey: queryKeys.charts.type(type ?? 'accion'),
    queryFn: ({ signal }) => apiFetch<LineChartDTO>(`/api/charts/type/${type}`, { signal }),
    enabled: type !== null,
  })
}

export function useTypeAssetsChart(type: AssetType | null) {
  return useQuery({
    queryKey: queryKeys.charts.typeAssets(type ?? 'accion'),
    queryFn: ({ signal }) => apiFetch<LineChartDTO>(`/api/charts/type/${type}/assets`, { signal }),
    enabled: type !== null,
  })
}

export function useTypesChart() {
  return useQuery({ queryKey: queryKeys.charts.types, queryFn: ({ signal }) => apiFetch<LineChartDTO>('/api/charts/types', { signal }) })
}

export function useAllAssetsChart() {
  return useQuery({ queryKey: queryKeys.charts.allAssets, queryFn: ({ signal }) => apiFetch<LineChartDTO>('/api/charts/assets', { signal }) })
}

export function useTotalChart() {
  return useQuery({ queryKey: queryKeys.charts.total, queryFn: ({ signal }) => apiFetch<LineChartDTO>('/api/charts/total', { signal }) })
}
