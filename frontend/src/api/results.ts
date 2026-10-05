import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { apiFetch } from '@/api/client'
import { invalidatePortfolio } from '@/api/invalidate'
import { queryKeys } from '@/api/queryKeys'
import type { EligibleResponse, IdsResponse, MonthlyResult, YearMonth } from '@/api/types'

export type ResultInput = YearMonth & {
  assetId: number
  resultUsd: number
}

export type CloseMonthInput = YearMonth & {
  items: Array<{ assetId: number; resultUsd: number }>
}

export type DeletedResponse = {
  deleted: number
}

export function useEligibleResults(period: YearMonth) {
  return useQuery({
    queryKey: queryKeys.results.eligible(period),
    queryFn: ({ signal }) => apiFetch<EligibleResponse>(`/api/results/eligible?year=${period.year}&month=${period.month}`, { signal }),
  })
}

export function useAssetResults(assetId: number | null) {
  return useQuery({
    queryKey: queryKeys.assets.results(assetId ?? 0),
    queryFn: ({ signal }) => apiFetch<MonthlyResult[]>(`/api/assets/${assetId}/results`, { signal }),
    enabled: assetId !== null && assetId > 0,
  })
}

export function useCreateResult() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (body: ResultInput) => apiFetch<MonthlyResult>('/api/results', { method: 'POST', body }),
    onSuccess: () => invalidatePortfolio(queryClient),
  })
}

export function useCloseMonthResults() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (body: CloseMonthInput) => apiFetch<IdsResponse>('/api/results/close-month', { method: 'POST', body }),
    onSuccess: () => invalidatePortfolio(queryClient),
  })
}

export function useDeleteResult() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => apiFetch<void>(`/api/results/${id}`, { method: 'DELETE' }),
    onSuccess: () => invalidatePortfolio(queryClient),
  })
}

export function useDeleteResultsByMonth() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (period: YearMonth) =>
      apiFetch<DeletedResponse>(`/api/results?year=${period.year}&month=${period.month}`, { method: 'DELETE' }),
    onSuccess: () => invalidatePortfolio(queryClient),
  })
}
