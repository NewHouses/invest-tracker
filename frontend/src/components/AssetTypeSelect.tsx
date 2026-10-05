import { Select } from '@mantine/core'
import type { ReactNode } from 'react'

import { useMeta } from '@/api/meta'
import type { AssetType } from '@/api/types'

type AssetTypeSelectProps = {
  value: AssetType | null
  onChange: (type: AssetType | null) => void
  label?: ReactNode
  placeholder?: string
  error?: ReactNode
  required?: boolean
  disabled?: boolean
  clearable?: boolean
  // Restrinxe as opcións (p.ex. só os tipos presentes nos activos).
  only?: AssetType[]
}

export function AssetTypeSelect({ value, onChange, only, placeholder, ...props }: AssetTypeSelectProps) {
  const meta = useMeta()
  const options = (meta.data?.assetTypes ?? []).filter((t) => !only || only.includes(t.value))

  return (
    <Select
      data={options}
      value={value}
      onChange={(v) => onChange(options.find((t) => t.value === v)?.value ?? null)}
      placeholder={placeholder ?? 'Escolle un tipo'}
      {...props}
    />
  )
}
