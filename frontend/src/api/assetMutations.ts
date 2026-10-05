import { useMutation, useQueryClient } from '@tanstack/react-query'

import { apiFetch } from '@/api/client'
import { invalidatePortfolio } from '@/api/invalidate'
import type { Asset, AssetType, YearMonth } from '@/api/types'

export type CreateAssetInput = YearMonth & {
  type: AssetType
  name: string
  amountUsd: number
}

export type UpdateAssetInput = YearMonth & {
  id: number
  name: string
  amountUsd: number
}

export function useCreateAsset() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (body: CreateAssetInput) => apiFetch<Asset>('/api/assets', { method: 'POST', body }),
    onSuccess: () => invalidatePortfolio(queryClient),
  })
}

export function useUpdateAsset() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, ...body }: UpdateAssetInput) => apiFetch<Asset>(`/api/assets/${id}`, { method: 'PUT', body }),
    onSuccess: () => invalidatePortfolio(queryClient),
  })
}

export function useDeleteAsset() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => apiFetch<void>(`/api/assets/${id}`, { method: 'DELETE' }),
    onSuccess: () => invalidatePortfolio(queryClient),
  })
}
