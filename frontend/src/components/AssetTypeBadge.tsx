import { Badge } from '@mantine/core'
import type { BadgeProps } from '@mantine/core'

import { useAssetTypeLabel } from '@/api/meta'
import type { AssetType } from '@/api/types'

// Cor fixa por tipo de ativo, para recoñecelos dun golpe de vista en
// tódalas táboas.
const typeColors: Record<AssetType, string> = {
  accion: 'blue',
  indice: 'teal',
  copy_trading: 'orange',
  fondo: 'grape',
}

export function assetTypeColor(type: AssetType) {
  return typeColors[type]
}

type AssetTypeBadgeProps = Omit<BadgeProps, 'color' | 'children'> & { type: AssetType }

export function AssetTypeBadge({ type, variant = 'light', ...props }: AssetTypeBadgeProps) {
  const typeLabel = useAssetTypeLabel()
  return (
    <Badge color={assetTypeColor(type)} variant={variant} {...props}>
      {typeLabel(type)}
    </Badge>
  )
}
