import { useQuery } from '@tanstack/react-query'

import { apiFetch } from '@/api/client'
import { queryKeys } from '@/api/queryKeys'
import type { Distribution, LineChartDTO } from '@/api/charts'
import type { Portfolio } from '@/api/portfolio'

export type Dashboard = {
  assetCount: number
  kpis: {
    totalInvested: number
    currentValue: number
    hasCurrentValue: boolean
    totalGain: number
    hasTotalGain: boolean
    totalDividends: number
    avgIndexPct: number
    hasAverages: boolean
  }
  evolution: LineChartDTO
  distribution: Distribution
  portfolio: Portfolio
}

export function useDashboard() {
  return useQuery({ queryKey: queryKeys.dashboard, queryFn: ({ signal }) => apiFetch<Dashboard>('/api/dashboard', { signal }) })
}
