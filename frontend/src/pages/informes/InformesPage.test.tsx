import { screen, within } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { TotalHistory } from '@/api/reports'
import { AssetHistoryPanel } from '@/features/reports/AssetHistoryPanel'
import { InformesPage } from '@/pages/informes/InformesPage'
import { installFetch, testAsset as asset, testMeta as meta } from '@/test/fetchMock'
import { renderWithProviders } from '@/test/utils'

beforeEach(() => vi.restoreAllMocks())

const totalHistory = (overrides: Partial<TotalHistory> = {}): TotalHistory => ({
  assetCount: 1,
  rows: [{ period: { year: 2026, month: 4 }, aporte: 100, fondos: 1300, dividends: 25, result: 1400, gain: 100, gainPct: 7.69, hasMetrics: true }],
  current: null,
  lifetimeAporte: 1200,
  avgIndexPct: 7.69,
  avgGain: 100,
  hasAverages: true,
  totalGain: 100,
  hasTotalGain: true,
  totalDividends: 25,
  currentValue: 1300,
  hasCurrentValue: true,
  ...overrides,
})

describe('InformesPage', () => {
  it('renderiza informe mensual total con cobertura parcial e guión sen resultados', async () => {
    installFetch({ '/api/assets': [asset], '/api/reports/total/month?year=2026&month=10': { period: { year: 2026, month: 10 }, totalAssets: 2, assetsActive: 2, assetsWithResult: 1, partial: true, totalInvested: 1000, investedInMonth: 100, investedPlusPrevDividends: 105, dividends: 3, dividendsPrev: 5, holdingNoDiv: 1200, holdingWithDiv: 1205, baseNoDiv: 600, baseWithDiv: 605, resultNoDiv: 660, resultWithDiv: 663, gainNoDiv: 60, gainWithDiv: 58, pctNoDiv: 10, pctWithDiv: 9.59, hasResults: true, hasMetrics: true, hasPctWithDiv: true, averages: { months: 2, pctNoDiv: 5, gainNoDiv: 30, pctWithDiv: 4, gainWithDiv: 25 } } })
    renderWithProviders(<InformesPage />, { initialEntries: ['/informes?vista=mensual&ambito=total'] })
    expect(await screen.findByText('Resultado (sen div)')).toBeInTheDocument()
    expect(screen.getByText(/1\/2 ativos con resultado/)).toBeInTheDocument()
    expect(screen.getByText('Promedios mensuais (2 mes(es) con resultado ata 10/2026)')).toBeInTheDocument()
  })

  it('renderiza filas históricas n/a', async () => {
    installFetch({ '/api/assets': [asset], '/api/meta': meta, '/api/reports/asset/1/history': { asset, totalInvested: 1000, avgIndexPct: 0, avgGain: 0, hasAverages: false, totalGain: 0, hasTotalGain: false, rows: [{ period: { year: 2026, month: 1 }, aporte: 100, holding: 0, result: 0, gain: 0, gainPct: 0, hasMetrics: false }], current: null } })
    renderWithProviders(<AssetHistoryPanel assetId={1} />)
    expect(await screen.findByText('n/a')).toBeInTheDocument()
    expect(screen.getAllByText('—').length).toBeGreaterThan(0)
  })

  it('abre por defecto o histórico total', async () => {
    installFetch({ '/api/assets': [asset], '/api/reports/total/history': totalHistory() })
    renderWithProviders(<InformesPage />, { initialEntries: ['/informes'] })
    expect(await screen.findByText('Aporte histórico total')).toBeInTheDocument()
    expect(screen.queryByLabelText('Mes')).not.toBeInTheDocument()
  })

  it('ordena o selector de ámbito como Total, Tipo e Ativo', async () => {
    installFetch({ '/api/assets': [asset], '/api/reports/total/history': totalHistory() })
    renderWithProviders(<InformesPage />, { initialEntries: ['/informes'] })
    await screen.findByText('Aporte histórico total')

    const total = screen.getByText('Total')
    const tipo = screen.getByText('Tipo')
    const ativo = screen.getByText('Ativo')

    expect(total.compareDocumentPosition(tipo) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy()
    expect(tipo.compareDocumentPosition(ativo) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy()
  })

  it('amosa a fila do mes en curso ao final con resultado pendente', async () => {
    installFetch({
      '/api/assets': [asset],
      '/api/reports/total/history': totalHistory({
        current: { period: { year: 2026, month: 5 }, aporte: 75, fondos: 1300, dividends: 25, result: 0, gain: 0, gainPct: 0, hasMetrics: false },
      }),
    })
    renderWithProviders(<InformesPage />, { initialEntries: ['/informes'] })

    const currentRow = await screen.findByTestId('history-current-row')
    expect(currentRow.nextElementSibling).toBeNull()
    expect(within(currentRow).getByText('en curso')).toBeInTheDocument()
    expect(within(currentRow).getAllByRole('cell').at(-1)).toHaveTextContent('—')
  })

  it('marca en negra as celas de Resultado no histórico', async () => {
    installFetch({ '/api/assets': [asset], '/api/reports/total/history': totalHistory() })
    renderWithProviders(<InformesPage />, { initialEntries: ['/informes'] })

    const resultCells = await screen.findAllByTestId('history-result-cell')
    expect(resultCells[0]).toHaveAttribute('data-bold', 'true')
  })
})
