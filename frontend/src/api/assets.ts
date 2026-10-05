import { useQuery } from '@tanstack/react-query'

import { apiFetch } from '@/api/client'
import { queryKeys } from '@/api/queryKeys'
import type { Asset } from '@/api/types'

export function useAssets() {
  return useQuery({
    queryKey: queryKeys.assets.list,
    queryFn: ({ signal }) => apiFetch<Asset[]>('/api/assets', { signal }),
  })
}

export function useAsset(id: number | null) {
  return useQuery({
    queryKey: queryKeys.assets.detail(id ?? 0),
    queryFn: ({ signal }) => apiFetch<Asset>(`/api/assets/${id}`, { signal }),
    enabled: id !== null && id > 0,
  })
}
