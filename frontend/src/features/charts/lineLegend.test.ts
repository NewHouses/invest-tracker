import { describe, expect, it } from 'vitest'

import { getVisibleSeries, isSeriesVisible, toggleSeriesVisibility } from './lineLegend'

const series = [
  { name: 's0', label: 'ACME', color: 'blue.6' },
  { name: 's1', label: 'Beta', color: 'green.6' },
]

describe('lineLegend', () => {
  it('oculta e volve amosar unha serie ao alternala', () => {
    const hidden = toggleSeriesVisibility(new Set<string>(), 's0')
    expect(isSeriesVisible(hidden, 's0')).toBe(false)
    expect(getVisibleSeries(series, hidden).map((item) => item.name)).toEqual(['s1'])

    const visibleAgain = toggleSeriesVisibility(hidden, 's0')
    expect(isSeriesVisible(visibleAgain, 's0')).toBe(true)
    expect(getVisibleSeries(series, visibleAgain).map((item) => item.name)).toEqual(['s0', 's1'])
  })

  it('mantén a orde e as cores estables ao filtrar', () => {
    const hidden = new Set(['s0'])
    expect(getVisibleSeries(series, hidden)).toEqual([{ name: 's1', label: 'Beta', color: 'green.6' }])
  })
})
