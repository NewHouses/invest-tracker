import { Badge } from '@mantine/core'

import { formatPct, formatSignedUSD } from '@/lib/format'
import { gainColor } from '@/lib/gain'

type GainBadgeProps = {
  value: number
  kind?: 'pct' | 'amount'
}

export function GainBadge({ value, kind = 'pct' }: GainBadgeProps) {
  return <Badge color={gainColor(value)}>{kind === 'pct' ? formatPct(value) : formatSignedUSD(value)}</Badge>
}
