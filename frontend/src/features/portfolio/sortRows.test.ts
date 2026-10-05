import { describe, expect, it } from 'vitest'

import type { PortfolioRow } from '@/api/portfolio'
import type { Asset } from '@/api/types'
import { sortRows, type PortfolioTableRow, type PortfolioSortKey } from '@/features/portfolio/sortRows'

const asset = (id: number, overrides: Partial<Asset> = {}): Asset => ({
  id,
  type: 'accion',
  name: `Ativo ${id}`,
  amountUsd: id * 100,
  year: 2026,
  month: id,
  ...overrides,
})

const portfolio = (assetValue: Asset, overrides: Partial<PortfolioRow> = {}): PortfolioRow => ({
  asset: assetValue,
  totalInvested: assetValue.amountUsd + 50,
  currentValue: assetValue.amountUsd + 100,
  hasCurrentValue: true,
  gain: 50,
  gainPct: 5,
  hasGain: true,
  hasGainPct: true,
  lastResult: { year: 2026, month: assetValue.month },
  pending: false,
  ...overrides,
})

const row = (id: number, overrides: Partial<Asset> = {}, rowOverrides: Partial<PortfolioTableRow> = {}): PortfolioTableRow => {
  const current = asset(id, overrides)
  return {
    asset: current,
    originalIndex: id,
    typeLabel: overrides.type === 'indice' ? 'Índice' : overrides.type === 'fondo' ? 'Fondo' : overrides.type === 'copy_trading' ? 'Copy-trading' : 'Acción',
    portfolio: portfolio(current),
    ...rowOverrides,
  }
}

const values = (rows: PortfolioTableRow[]) => rows.map((item) => item.asset.name)

describe('sortRows', () => {
  it.each([
    ['type', [row(0, { type: 'indice', name: 'B' }), row(1, { type: 'accion', name: 'A' })], ['A', 'B'], ['B', 'A']],
    ['name', [row(0, { name: 'beta' }), row(1, { name: 'Álfa' })], ['Álfa', 'beta'], ['beta', 'Álfa']],
    ['period', [row(0, { name: 'Dec', year: 2026, month: 12 }), row(1, { name: 'Jan', year: 2026, month: 1 })], ['Jan', 'Dec'], ['Dec', 'Jan']],
    ['initialAmount', [row(0, { name: 'High', amountUsd: 300 }), row(1, { name: 'Low', amountUsd: 100 })], ['Low', 'High'], ['High', 'Low']],
    ['totalInvested', [row(0, { name: 'High' }, { portfolio: portfolio(asset(0), { totalInvested: 300 }) }), row(1, { name: 'Low' }, { portfolio: portfolio(asset(1), { totalInvested: 100 }) })], ['Low', 'High'], ['High', 'Low']],
    ['currentValue', [row(0, { name: 'High' }, { portfolio: portfolio(asset(0), { currentValue: 300 }) }), row(1, { name: 'Low' }, { portfolio: portfolio(asset(1), { currentValue: 100 }) })], ['Low', 'High'], ['High', 'Low']],
    ['gain', [row(0, { name: 'High' }, { portfolio: portfolio(asset(0), { gain: 30 }) }), row(1, { name: 'Low' }, { portfolio: portfolio(asset(1), { gain: -5 }) })], ['Low', 'High'], ['High', 'Low']],
    ['gainPct', [row(0, { name: 'High' }, { portfolio: portfolio(asset(0), { gainPct: 30 }) }), row(1, { name: 'Low' }, { portfolio: portfolio(asset(1), { gainPct: -5 }) })], ['Low', 'High'], ['High', 'Low']],
  ] as Array<[PortfolioSortKey, PortfolioTableRow[], string[], string[]]>)('sorts %s asc and desc', (key, rows, asc, desc) => {
    expect(values(sortRows(rows, key, 'asc'))).toEqual(asc)
    expect(values(sortRows(rows, key, 'desc'))).toEqual(desc)
  })

  it('restores original order when direction is none', () => {
    const rows = [row(1, { name: 'B' }), row(0, { name: 'A' })]
    expect(values(sortRows(rows, 'name', null))).toEqual(['A', 'B'])
  })

  it('keeps unavailable values last in both directions', () => {
    const missing = row(0, { name: 'Missing' }, { portfolio: portfolio(asset(0), { hasCurrentValue: false }) })
    const low = row(1, { name: 'Low' }, { portfolio: portfolio(asset(1), { currentValue: 100 }) })
    const high = row(2, { name: 'High' }, { portfolio: portfolio(asset(2), { currentValue: 300 }) })
    expect(values(sortRows([missing, high, low], 'currentValue', 'asc'))).toEqual(['Low', 'High', 'Missing'])
    expect(values(sortRows([missing, high, low], 'currentValue', 'desc'))).toEqual(['High', 'Low', 'Missing'])
  })

  it('keeps rows without portfolio data last', () => {
    const missing = row(0, { name: 'Missing' }, { portfolio: undefined })
    const present = row(1, { name: 'Present' }, { portfolio: portfolio(asset(1), { totalInvested: 10 }) })
    expect(values(sortRows([missing, present], 'totalInvested', 'asc'))).toEqual(['Present', 'Missing'])
  })
})
