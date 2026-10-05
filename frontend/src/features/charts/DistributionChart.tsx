import { DonutChart } from '@mantine/charts'
import { ColorSwatch, Group, Stack, Table, Text, useMantineTheme } from '@mantine/core'

import type { Distribution } from '@/api/charts'
import { formatNumber, formatUSD } from '@/lib/format'

const colors = ['blue.6', 'green.6', 'orange.6', 'grape.6', 'cyan.6', 'red.6', 'yellow.7', 'teal.6']

type DistributionChartProps = {
  distribution: Distribution
  size?: number
}

// Rosca da distribución de aportes cunha lenda en táboa (activo, importe e
// porcentaxe), xa que as etiquetas da propia rosca non indican o activo.
export function DistributionChart({ distribution, size = 220 }: DistributionChartProps) {
  const theme = useMantineTheme()
  const items = distribution.items.map((item, index) => ({ ...item, color: colors[index % colors.length] ?? 'blue.6' }))
  const swatch = (color: string) => {
    const [name, shade] = color.split('.')
    return theme.colors[name ?? 'blue']?.[Number(shade ?? 6)] ?? color
  }

  return (
    <Group align="center" justify="center" gap="xl" wrap="wrap">
      <DonutChart
        size={size}
        thickness={28}
        data={items.map((item) => ({ name: item.label, value: item.value, color: item.color }))}
        valueFormatter={formatUSD}
        chartLabel={formatUSD(distribution.total)}
        withTooltip
      />
      <Stack gap={0} style={{ flex: '1 1 340px', minWidth: 0 }}>
        <Table verticalSpacing={6}>
          <Table.Tbody>
            {items.map((item) => (
              <Table.Tr key={item.asset.id}>
                <Table.Td>
                  <Group gap="xs" wrap="nowrap">
                    <ColorSwatch color={swatch(item.color)} size={12} withShadow={false} />
                    <Text size="sm" truncate="end">{item.label}</Text>
                  </Group>
                </Table.Td>
                <Table.Td ta="right"><Text size="sm">{formatUSD(item.value)}</Text></Table.Td>
                <Table.Td ta="right" w={80}><Text size="sm" c="dimmed">{formatNumber(item.pct, 1)} %</Text></Table.Td>
              </Table.Tr>
            ))}
          </Table.Tbody>
        </Table>
      </Stack>
    </Group>
  )
}
