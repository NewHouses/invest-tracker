import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { apiFetch } from '@/api/client'
import { invalidatePortfolio } from '@/api/invalidate'
import { queryKeys } from '@/api/queryKeys'
import type { Asset, IdsResponse, Transaction, YearMonth } from '@/api/types'

export type TransactionKind = 'compra' | 'venda'

export type AssetTransactionRow = {
  isInitial: boolean
  id: number
  year: number
  month: number
  isVenda: boolean
  amount: number
}

export type AssetTransactionsResponse = {
  asset: Asset
  rows: AssetTransactionRow[]
  totals: {
    compra: number
    venda: number
    neto: number
  }
}

export type TransactionInput = YearMonth & {
  assetId: number
  kind: TransactionKind
  amountUsd: number
}

export type TransactionUpdateInput = YearMonth & {
  id: number
  kind: TransactionKind
  amountUsd: number
}

export type TransactionBatchInput = {
  assetId: number
  items: Array<YearMonth & { kind: TransactionKind; amountUsd: number }>
}

export type MonthAssetsResponse = {
  period: YearMonth
  eligible: Asset[]
  omitted: Asset[]
}

export type MonthTransactionInput = YearMonth & {
  items: Array<{ assetId: number; amountUsd: number }>
}

export function useAssetTransactions(assetId: number | null) {
  return useQuery({
    queryKey: queryKeys.assets.transactions(assetId ?? 0),
    queryFn: ({ signal }) => apiFetch<AssetTransactionsResponse>(`/api/assets/${assetId}/transactions`, { signal }),
    enabled: assetId !== null && assetId > 0,
  })
}

export function useMonthAssets(period: YearMonth) {
  return useQuery({
    queryKey: queryKeys.transactions.monthAssets(period),
    queryFn: ({ signal }) =>
      apiFetch<MonthAssetsResponse>(`/api/transactions/month-assets?year=${period.year}&month=${period.month}`, { signal }),
  })
}

export function useCreateTransaction() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (body: TransactionInput) => apiFetch<Transaction>('/api/transactions', { method: 'POST', body }),
    onSuccess: () => invalidatePortfolio(queryClient),
  })
}

export function useCreateTransactionBatch() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (body: TransactionBatchInput) => apiFetch<IdsResponse>('/api/transactions/batch', { method: 'POST', body }),
    onSuccess: () => invalidatePortfolio(queryClient),
  })
}

export function useCreateMonthTransactions() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (body: MonthTransactionInput) => apiFetch<IdsResponse>('/api/transactions/month', { method: 'POST', body }),
    onSuccess: () => invalidatePortfolio(queryClient),
  })
}

export function useUpdateTransaction() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, ...body }: TransactionUpdateInput) =>
      apiFetch<Transaction>(`/api/transactions/${id}`, { method: 'PUT', body }),
    onSuccess: () => invalidatePortfolio(queryClient),
  })
}

export function useDeleteTransaction() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => apiFetch<void>(`/api/transactions/${id}`, { method: 'DELETE' }),
    onSuccess: () => invalidatePortfolio(queryClient),
  })
}
