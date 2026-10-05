import { DonutChart } from '@mantine/charts'
import { alpha, ColorSwatch, Group, Paper, Stack, Table, Text, useMantineTheme } from '@mantine/core'
import { useState } from 'react'

import type { Distribution } from '@/api/charts'
import { formatNumber, formatUSD } from '@/lib/format'

const colors = ['blue.6', 'green.6', 'orange.6', 'grape.6', 'cyan.6', 'red.6', 'yellow.7', 'teal.6']

type DistributionChartProps = {
  distribution: Distribution
  size?: number
}

type TooltipPayload = {
  name?: string
  value?: number
}

type DistributionTooltipProps = {
  active?: boolean
  payload?: TooltipPayload[]
  items: Array<{ label: string; pct: number }>
}

function DistributionTooltip({ active, payload, items }: DistributionTooltipProps) {
  const hovered = payload?.[0]
  if (!active || hovered?.name === undefined || hovered.value === undefined) return null
  const pct = items.find((item) => item.label === hovered.name)?.pct ?? 0
  return (
    <Paper withBorder shadow="sm" p="xs" radius="sm">
      <Text size="sm" fw={600}>{hovered.name}</Text>
      <Text size="sm">{formatUSD(hovered.value)}</Text>
      <Text size="xs" c="dimmed">{formatNumber(pct, 1)} %</Text>
    </Paper>
  )
}

export function DistributionChart({ distribution, size = 220 }: DistributionChartProps) {
  const theme = useMantineTheme()
  const [activeIndex, setActiveIndex] = useState<number | null>(null)
  const items = distribution.items.map((item, index) => ({ ...item, color: colors[index % colors.length] ?? 'blue.6' }))
  const swatch = (color: string) => {
    const [name, shade] = color.split('.')
    return theme.colors[name ?? 'blue']?.[Number(shade ?? 6)] ?? color
  }
  const isActive = (index: number) => activeIndex === null || activeIndex === index

  return (
    <Group align="center" justify="center" gap="xl" wrap="wrap">
      <DonutChart
        size={size}
        thickness={28}
        data={items.map((item, index) => ({ name: item.label, value: item.value, color: isActive(index) ? item.color : alpha(swatch(item.color), 0.25) }))}
        valueFormatter={formatUSD}
        chartLabel={formatUSD(distribution.total)}
        withTooltip
        tooltipDataSource="segment"
        tooltipProps={{ content: <DistributionTooltip items={items} /> }}
        pieProps={{
          onMouseEnter: (_entry, index) => setActiveIndex(index),
          onMouseLeave: () => setActiveIndex(null),
        }}
      />
      <Stack gap={0} style={{ flex: '1 1 340px', minWidth: 0 }}>
        <Table verticalSpacing={6}>
          <Table.Tbody>
            {items.map((item, index) => {
              const highlighted = activeIndex === index
              return (
                <Table.Tr
                  key={item.asset.id}
                  data-highlighted={highlighted ? 'true' : undefined}
                  onMouseEnter={() => setActiveIndex(index)}
                  onMouseLeave={() => setActiveIndex(null)}
                  style={{ backgroundColor: highlighted ? theme.colors.blue[0] : undefined }}
                >
                  <Table.Td>
                    <Group gap="xs" wrap="nowrap">
                      <ColorSwatch color={isActive(index) ? swatch(item.color) : alpha(swatch(item.color), 0.25)} size={12} withShadow={false} />
                      <Text size="sm" truncate="end">{item.label}</Text>
                    </Group>
                  </Table.Td>
                  <Table.Td ta="right"><Text size="sm">{formatUSD(item.value)}</Text></Table.Td>
                  <Table.Td ta="right" w={80}><Text size="sm" c="dimmed">{formatNumber(item.pct, 1)} %</Text></Table.Td>
                </Table.Tr>
              )
            })}
          </Table.Tbody>
        </Table>
      </Stack>
    </Group>
  )
}
