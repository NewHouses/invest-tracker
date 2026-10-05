import { LineChart } from '@mantine/charts'
import { ColorSwatch, Group, Loader, Paper, Stack, Text, UnstyledButton, useMantineTheme } from '@mantine/core'
import { useMemo, useState } from 'react'

import type { LineChartDTO } from '@/api/charts'
import { EmptyState } from '@/components/EmptyState'
import { formatAxisUSD, formatUSD } from '@/lib/format'

import { getVisibleSeries, isSeriesVisible, toggleSeriesVisibility } from './lineLegend'
import { mapLineChartDTO } from './lineChartMapper'
import type { MantineLineSeries } from './lineChartMapper'

type LineChartPanelProps = {
  chart: LineChartDTO | undefined
  loading?: boolean
  emptyNoAssets?: string
  emptyNoResults?: string
  h?: number
  title?: string | null
}

function resolveColor(theme: ReturnType<typeof useMantineTheme>, color: string) {
  const [name, shade] = color.split('.')
  return theme.colors[name ?? 'blue']?.[Number(shade ?? 6)] ?? color
}

type LineChartLegendProps = {
  series: MantineLineSeries[]
  hiddenSeries: ReadonlySet<string>
  onToggle: (seriesName: string) => void
}

function LineChartLegend({ series, hiddenSeries, onToggle }: LineChartLegendProps) {
  const theme = useMantineTheme()
  return (
    <Group gap="xs" wrap="wrap" aria-label="Lenda da gráfica">
      {series.map((item) => {
        const visible = isSeriesVisible(hiddenSeries, item.name)
        return (
          <UnstyledButton
            key={item.name}
            type="button"
            aria-pressed={visible}
            onClick={() => onToggle(item.name)}
            style={{
              border: `1px solid ${theme.colors.gray[3]}`,
              borderRadius: theme.radius.xl,
              padding: '4px 10px',
              opacity: visible ? 1 : 0.55,
              textDecoration: visible ? 'none' : 'line-through',
            }}
          >
            <Group gap={6} wrap="nowrap">
              <ColorSwatch color={resolveColor(theme, item.color)} size={10} withShadow={false} />
              <Text size="sm">{item.label}</Text>
            </Group>
          </UnstyledButton>
        )
      })}
    </Group>
  )
}

type LineChartContentProps = {
  mapped: ReturnType<typeof mapLineChartDTO>
  h: number
  title: string | null
}

function LineChartContent({ mapped, h, title }: LineChartContentProps) {
  const [hiddenSeries, setHiddenSeries] = useState<Set<string>>(() => new Set())
  const visibleSeries = getVisibleSeries(mapped.series, hiddenSeries)
  const hasCustomLegend = mapped.series.length > 1

  return (
    <Paper withBorder p="md" radius="md">
      <Stack>
        {title === null ? null : <Text fw={700}>{title}</Text>}
        {hasCustomLegend ? <LineChartLegend series={mapped.series} hiddenSeries={hiddenSeries} onToggle={(seriesName) => setHiddenSeries((current) => toggleSeriesVisibility(current, seriesName))} /> : null}
        {visibleSeries.length === 0 ? <Text c="dimmed" size="sm">Escolle polo menos unha serie para ver a evolución.</Text> : null}
        <LineChart
          h={h}
          data={mapped.data}
          dataKey="month"
          series={visibleSeries}
          valueFormatter={formatUSD}
          yAxisProps={{ tickFormatter: formatAxisUSD, width: 64 }}
          connectNulls={false}
          withLegend={!hasCustomLegend}
        />
      </Stack>
    </Paper>
  )
}

export function LineChartPanel({ chart, loading, emptyNoAssets, emptyNoResults, h = 320, title = 'Evolución' }: LineChartPanelProps) {
  const mapped = useMemo(() => (chart ? mapLineChartDTO(chart) : null), [chart])

  if (loading) return <Loader aria-label="Cargando gráfica" />
  if (!chart || !mapped) return null
  if (!chart.hasAssets) return <EmptyState title={emptyNoAssets ?? 'Aínda non hai ativos'} />
  if (chart.months.length === 0) return <EmptyState title={emptyNoResults ?? 'Aínda non hai resultados rexistrados.'} />

  const contentKey = `${title ?? 'sen-titulo'}:${chart.months.length}:${chart.series.map((item) => item.label).join('|')}`
  return <LineChartContent key={contentKey} mapped={mapped} h={h} title={title} />
}
