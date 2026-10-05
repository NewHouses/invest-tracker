import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { apiFetch } from '@/api/client'
import { invalidatePortfolio } from '@/api/invalidate'
import { queryKeys } from '@/api/queryKeys'
import type { Dividend, YearMonth } from '@/api/types'

export type DividendInput = YearMonth & {
  amountUsd: number
}

export function useDividends() {
  return useQuery({
    queryKey: queryKeys.dividends.list,
    queryFn: ({ signal }) => apiFetch<Dividend[]>('/api/dividends', { signal }),
  })
}

export function useCreateDividend() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (body: DividendInput) => apiFetch<Dividend>('/api/dividends', { method: 'POST', body }),
    onSuccess: () => invalidatePortfolio(queryClient),
  })
}

export function useDeleteDividend() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => apiFetch<void>(`/api/dividends/${id}`, { method: 'DELETE' }),
    onSuccess: () => invalidatePortfolio(queryClient),
  })
}
