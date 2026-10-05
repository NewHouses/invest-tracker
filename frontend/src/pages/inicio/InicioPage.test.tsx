import { screen, within } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { InicioPage } from '@/pages/inicio/InicioPage'
import { installFetch, testAsset as asset } from '@/test/fetchMock'
import { renderWithProviders } from '@/test/utils'

beforeEach(() => vi.restoreAllMocks())

const secondAsset = { id: 2, type: 'indice', name: 'SP500', amountUsd: 500, year: 2026, month: 2 } as const

describe('InicioPage', () => {
  it('amosa o Portofolio con pendentes, guións e ligazón de peche', async () => {
    installFetch({
      '/api/dashboard': {
        assetCount: 2,
        kpis: { totalInvested: 1500, currentValue: 0, hasCurrentValue: false, totalGain: 0, hasTotalGain: false, totalDividends: 12, avgIndexPct: 0, hasAverages: false },
        evolution: {
          title: 'Total',
          hasAssets: true,
          months: [{ year: 2026, month: 9 }],
          series: [
            { label: 'Resultado + dividendos acum.', values: [120] },
            { label: 'Aporte acumulado', values: [1500] },
          ],
        },
        distribution: { hasAssets: true, total: 1000, items: [{ asset, label: 'ACME', value: 1000, pct: 100 }] },
        portfolio: {
          period: { year: 2026, month: 10 },
          pendingCount: 1,
          rows: [
            { asset, totalInvested: 1200, currentValue: 0, hasCurrentValue: false, gain: 0, gainPct: 0, hasGain: false, hasGainPct: false, lastResult: null, pending: true },
            { asset: secondAsset, totalInvested: 500, currentValue: 550, hasCurrentValue: true, gain: 50, gainPct: 10, hasGain: true, hasGainPct: true, lastResult: { year: 2026, month: 9 }, pending: false },
          ],
        },
      },
    })
    renderWithProviders(<InicioPage />)

    expect(await screen.findByRole('heading', { name: 'Portofolio' })).toBeInTheDocument()
    expect(screen.getByText('1 pendente(s) de pechar en 10/2026')).toBeInTheDocument()
    expect(screen.getByText('pendente')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /Pechar mes/ })).toHaveAttribute('href', '/resultados?tab=pechar')
    expect(screen.getByRole('link', { name: 'ACME' })).toHaveAttribute('href', '/activos/1')

    const acmeRow = screen.getByRole('link', { name: 'ACME' }).closest('tr')
    expect(acmeRow).not.toBeNull()
    expect(within(acmeRow as HTMLTableRowElement).getByText(/1\.200,00\s\$/)).toBeInTheDocument()
    expect(within(acmeRow as HTMLTableRowElement).getAllByText('—')).toHaveLength(3)
  })

  it('amosa o estado pechado cando non hai pendentes', async () => {
    installFetch({
      '/api/dashboard': {
        assetCount: 1,
        kpis: { totalInvested: 1000, currentValue: 1100, hasCurrentValue: true, totalGain: 100, hasTotalGain: true, totalDividends: 0, avgIndexPct: 1, hasAverages: true },
        evolution: { title: 'Total', hasAssets: true, months: [], series: [] },
        distribution: { hasAssets: true, total: 0, items: [] },
        portfolio: { period: { year: 2026, month: 10 }, pendingCount: 0, rows: [{ asset, totalInvested: 1000, currentValue: 1100, hasCurrentValue: true, gain: 100, gainPct: 10, hasGain: true, hasGainPct: true, lastResult: { year: 2026, month: 10 }, pending: false }] },
      },
    })
    renderWithProviders(<InicioPage />)
    expect(await screen.findByText('Todo pechado en 10/2026 ✓')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /Pechar mes/ })).toHaveAttribute('href', '/resultados?tab=pechar')
  })
})
