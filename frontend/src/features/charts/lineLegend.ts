import type { MantineLineSeries } from './lineChartMapper'

export type HiddenSeries = ReadonlySet<string>

export function isSeriesVisible(hiddenSeries: HiddenSeries, seriesName: string) {
  return !hiddenSeries.has(seriesName)
}

export function toggleSeriesVisibility(hiddenSeries: HiddenSeries, seriesName: string) {
  const next = new Set(hiddenSeries)
  if (next.has(seriesName)) {
    next.delete(seriesName)
  } else {
    next.add(seriesName)
  }
  return next
}

export function getVisibleSeries<T extends Pick<MantineLineSeries, 'name'>>(series: T[], hiddenSeries: HiddenSeries) {
  return series.filter((item) => isSeriesVisible(hiddenSeries, item.name))
}
