import { Select } from '@mantine/core'
import type { ReactNode } from 'react'

import { useAssets } from '@/api/assets'
import { useAssetTypeLabel } from '@/api/meta'
import type { Asset, AssetType } from '@/api/types'

type AssetSelectProps = {
  value: number | null
  onChange: (id: number | null) => void
  label?: ReactNode
  placeholder?: string
  error?: ReactNode
  required?: boolean
  disabled?: boolean
  clearable?: boolean
  // Só activos deste tipo.
  typeFilter?: AssetType
  // Lista explícita de activos (p.ex. os elixibles nun mes) en vez de todos.
  assets?: Asset[]
}

export function AssetSelect({ value, onChange, typeFilter, assets, placeholder, ...props }: AssetSelectProps) {
  const all = useAssets()
  const typeLabel = useAssetTypeLabel()
  const source = assets ?? all.data ?? []
  const options = source
    .filter((a) => !typeFilter || a.type === typeFilter)
    .map((a) => ({ value: String(a.id), label: `${typeLabel(a.type)} — ${a.name}` }))

  return (
    <Select
      data={options}
      value={value === null ? null : String(value)}
      onChange={(v) => onChange(v === null ? null : Number(v))}
      searchable
      nothingFoundMessage="Non hai ativos"
      placeholder={placeholder ?? (assets === undefined && all.isLoading ? 'Cargando…' : 'Escolle un ativo')}
      {...props}
    />
  )
}
