import { Text } from '@mantine/core'
import type { TextProps } from '@mantine/core'

import { formatSignedUSD, formatUSD } from '@/lib/format'
import { gainColor } from '@/lib/gain'

type MoneyTextProps = TextProps & {
  amount: number
  signed?: boolean
  colorByGain?: boolean
}

export function MoneyText({ amount, signed = false, colorByGain = false, ...props }: MoneyTextProps) {
  const colorProps = colorByGain ? { c: gainColor(amount) } : {}
  return (
    <Text component="span" {...props} {...colorProps}>
      {signed ? formatSignedUSD(amount) : formatUSD(amount)}
    </Text>
  )
}
