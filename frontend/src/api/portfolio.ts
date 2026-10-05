import { useQuery } from '@tanstack/react-query'

import { apiFetch } from '@/api/client'
import { queryKeys } from '@/api/queryKeys'
import type { Asset, YearMonth } from '@/api/types'

export type PortfolioRow = {
  asset: Asset
  totalInvested: number
  currentValue: number
  hasCurrentValue: boolean
  gain: number
  gainPct: number
  hasGain: boolean
  hasGainPct: boolean
  lastResult: YearMonth | null
  pending: boolean
}

export type Portfolio = {
  period: YearMonth
  rows: PortfolioRow[]
  pendingCount: number
}

export function usePortfolio() {
  return useQuery({
    queryKey: queryKeys.portfolio,
    queryFn: ({ signal }) => apiFetch<Portfolio>('/api/portfolio', { signal }),
  })
}
