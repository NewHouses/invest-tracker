import { useQuery } from '@tanstack/react-query'
import { useCallback } from 'react'

import { apiFetch } from '@/api/client'
import { queryKeys } from '@/api/queryKeys'
import type { AssetType, MetaResponse } from '@/api/types'

export function useMeta() {
  return useQuery({
    queryKey: queryKeys.meta,
    queryFn: ({ signal }) => apiFetch<MetaResponse>('/api/meta', { signal }),
    staleTime: Infinity,
  })
}

// useAssetTypeLabel devolve unha función que traduce o tipo á súa etiqueta
// (p.ex. "copy_trading" → "Copy-trading") coas etiquetas que envía o servidor.
export function useAssetTypeLabel() {
  const meta = useMeta()
  const types = meta.data?.assetTypes
  return useCallback(
    (type: AssetType) => types?.find((t) => t.value === type)?.label ?? type,
    [types],
  )
}
