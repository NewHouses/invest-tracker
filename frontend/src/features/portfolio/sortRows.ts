import type { PortfolioRow } from '@/api/portfolio'
import type { Asset } from '@/api/types'

export type PortfolioSortKey = 'type' | 'name' | 'period' | 'initialAmount' | 'totalInvested' | 'currentValue' | 'gain' | 'gainPct'
export type SortDirection = 'asc' | 'desc' | null

export type PortfolioTableRow = {
  asset: Asset
  originalIndex: number
  typeLabel: string
  portfolio?: PortfolioRow | undefined
}

type SortValue = string | number | null

function periodValue(asset: Asset) {
  return asset.year * 100 + asset.month
}

function valueFor(row: PortfolioTableRow, key: PortfolioSortKey): SortValue {
  switch (key) {
    case 'type':
      return row.typeLabel
    case 'name':
      return row.asset.name
    case 'period':
      return periodValue(row.asset)
    case 'initialAmount':
      return row.asset.amountUsd
    case 'totalInvested':
      return row.portfolio?.totalInvested ?? null
    case 'currentValue':
      return row.portfolio?.hasCurrentValue ? row.portfolio.currentValue : null
    case 'gain':
      return row.portfolio?.hasGain ? row.portfolio.gain : null
    case 'gainPct':
      return row.portfolio?.hasGainPct ? row.portfolio.gainPct : null
  }
}

function compareValues(a: SortValue, b: SortValue, direction: Exclude<SortDirection, null>) {
  const aMissing = a === null
  const bMissing = b === null
  if (aMissing && bMissing) return 0
  if (aMissing) return 1
  if (bMissing) return -1

  const result = typeof a === 'string' && typeof b === 'string'
    ? a.localeCompare(b, 'gl', { sensitivity: 'base' })
    : Number(a) - Number(b)

  return direction === 'asc' ? result : -result
}

export function sortRows(rows: PortfolioTableRow[], key: PortfolioSortKey | null, direction: SortDirection) {
  if (!key || direction === null) return [...rows].sort((a, b) => a.originalIndex - b.originalIndex)

  return [...rows].sort((a, b) => {
    const valueCompare = compareValues(valueFor(a, key), valueFor(b, key), direction)
    return valueCompare === 0 ? a.originalIndex - b.originalIndex : valueCompare
  })
}
