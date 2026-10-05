import type { AssetType, YearMonth } from '@/api/types'

const ym = (p: YearMonth) => [p.year, p.month] as const

// Prefixos xerárquicos: invalidar ['reports'] refresca todos os informes.
export const queryKeys = {
  auth: {
    status: ['auth', 'status'] as const,
  },
  meta: ['meta'] as const,
  assets: {
    all: ['assets'] as const,
    list: ['assets', 'list'] as const,
    detail: (id: number) => ['assets', 'detail', id] as const,
    transactions: (id: number) => ['assets', 'detail', id, 'transactions'] as const,
    results: (id: number) => ['assets', 'detail', id, 'results'] as const,
  },
  transactions: {
    all: ['transactions'] as const,
    monthAssets: (p: YearMonth) => ['transactions', 'month-assets', ...ym(p)] as const,
  },
  results: {
    all: ['results'] as const,
    eligible: (p: YearMonth) => ['results', 'eligible', ...ym(p)] as const,
  },
  dividends: {
    all: ['dividends'] as const,
    list: ['dividends', 'list'] as const,
  },
  reports: {
    all: ['reports'] as const,
    assetMonth: (id: number, p: YearMonth) => ['reports', 'asset', id, 'month', ...ym(p)] as const,
    typeMonth: (t: AssetType, p: YearMonth) => ['reports', 'type', t, 'month', ...ym(p)] as const,
    totalMonth: (p: YearMonth) => ['reports', 'total', 'month', ...ym(p)] as const,
    assetHistory: (id: number) => ['reports', 'asset', id, 'history'] as const,
    typeHistory: (t: AssetType) => ['reports', 'type', t, 'history'] as const,
    totalHistory: ['reports', 'total', 'history'] as const,
  },
  charts: {
    all: ['charts'] as const,
    distribution: ['charts', 'distribution'] as const,
    asset: (id: number) => ['charts', 'asset', id] as const,
    type: (t: AssetType) => ['charts', 'type', t] as const,
    typeAssets: (t: AssetType) => ['charts', 'type', t, 'assets'] as const,
    types: ['charts', 'types'] as const,
    total: ['charts', 'total'] as const,
  },
  tools: {
    all: ['tools'] as const,
    projectionStart: ['tools', 'projection', 'start'] as const,
  },
  dashboard: ['dashboard'] as const,
} as const
