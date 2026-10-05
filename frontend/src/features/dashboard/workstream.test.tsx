import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { LineChartDTO } from '@/api/charts'
import { AllocationPanel } from '@/features/tools/AllocationPanel'
import { mapLineChartDTO } from '@/features/charts/lineChartMapper'
import { AssetHistoryPanel } from '@/features/reports/AssetHistoryPanel'
import { InicioPage } from '@/pages/inicio/InicioPage'
import { InformesPage } from '@/pages/informes/InformesPage'
import { ProxeccionPage } from '@/pages/proxeccion/ProxeccionPage'
import { renderWithProviders } from '@/test/utils'

const asset = { id: 1, type: 'accion', name: 'ACME', amountUsd: 1000, year: 2026, month: 1 }
const meta = { assetTypes: [{ value: 'accion', label: 'Acción' }, { value: 'indice', label: 'Índice' }, { value: 'copy_trading', label: 'Copy-trading' }, { value: 'fondo', label: 'Fondo' }], minYear: 2000, maxYear: 2100 }

function json(data: unknown) {
  return new Response(JSON.stringify(data), { status: 200, headers: { 'Content-Type': 'application/json' } })
}
function installFetch(routes: Record<string, unknown | ((init: RequestInit) => unknown)>) {
  const calls: Array<{ url: string; init: RequestInit }> = []
  vi.stubGlobal('fetch', vi.fn((input: RequestInfo | URL, init: RequestInit = {}) => {
    const url = String(input)
    calls.push({ url, init })
    const route = routes[url]
    if (route === undefined) return Promise.resolve(new Response(JSON.stringify({ error: `sen ruta ${url}` }), { status: 404 }))
    return Promise.resolve(json(typeof route === 'function' ? route(init) : route))
  }))
  return calls
}

beforeEach(() => vi.restoreAllMocks())

describe('mapLineChartDTO', () => {
  it('mantén nulls e etiqueta meses como MM/AAAA', () => {
    const dto: LineChartDTO = { title: 't', hasAssets: true, months: [{ year: 2026, month: 1 }, { year: 2026, month: 2 }], series: [{ label: 'ACME', values: [10, null] }] }
    expect(mapLineChartDTO(dto)).toEqual({ data: [{ month: '01/2026', s0: 10 }, { month: '02/2026', s0: null }], series: [{ name: 's0', label: 'ACME', color: 'blue.6' }] })
  })

  // Regresión: Mantine corta as claves polos puntos, e unha etiqueta como
  // "Resultado + dividendos acum." deixaba a lenda baleira.
  it('usa claves seguras aínda que a etiqueta leve puntos', () => {
    const dto: LineChartDTO = { title: 't', hasAssets: true, months: [{ year: 2026, month: 1 }], series: [{ label: 'Resultado + dividendos acum.', values: [5] }, { label: 'Iberdrola S.A.', values: [7] }] }
    const mapped = mapLineChartDTO(dto)
    expect(mapped.series.map((s) => s.name)).toEqual(['s0', 's1'])
    expect(mapped.series.map((s) => s.label)).toEqual(['Resultado + dividendos acum.', 'Iberdrola S.A.'])
    expect(mapped.data[0]).toEqual({ month: '01/2026', s0: 5, s1: 7 })
  })
})

describe('InicioPage', () => {
  it('amosa KPIs con guión e pendentes', async () => {
    installFetch({ '/api/dashboard': { assetCount: 1, kpis: { totalInvested: 1000, currentValue: 0, hasCurrentValue: false, totalGain: 0, hasTotalGain: false, totalDividends: 12, avgIndexPct: 0, hasAverages: false }, evolution: { title: 'Total', hasAssets: true, months: [], series: [] }, distribution: { hasAssets: true, total: 1000, items: [{ asset, label: 'ACME', value: 1000, pct: 100 }] }, pending: { period: { year: 2026, month: 10 }, items: [{ asset, holding: 1200, result: 0, hasResult: false }] } } })
    renderWithProviders(<InicioPage />)
    expect(await screen.findByText('Aporte total')).toBeInTheDocument()
    expect(screen.getAllByText('—').length).toBeGreaterThanOrEqual(3)
    expect(screen.getByText('Pendente de pechar 10/2026')).toBeInTheDocument()
    // ACME aparece na lenda da distribución e nos pendentes (co seu "no activo").
    expect(screen.getAllByText('ACME')).toHaveLength(2)
    expect(screen.getByText(/No activo:/)).toHaveTextContent('1.200,00')
    expect(screen.getByRole('link', { name: /Pechar mes/ })).toHaveAttribute('href', '/resultados?tab=pechar')
  })

  it('amosa Todo pechado cando non hai pendentes', async () => {
    installFetch({ '/api/dashboard': { assetCount: 1, kpis: { totalInvested: 1000, currentValue: 1100, hasCurrentValue: true, totalGain: 100, hasTotalGain: true, totalDividends: 0, avgIndexPct: 1, hasAverages: true }, evolution: { title: 'Total', hasAssets: true, months: [], series: [] }, distribution: { hasAssets: true, total: 0, items: [] }, pending: { period: { year: 2026, month: 10 }, items: [] } } })
    renderWithProviders(<InicioPage />)
    expect(await screen.findByText('Todo pechado ✓')).toBeInTheDocument()
  })
})

describe('InformesPage', () => {
  it('renderiza informe mensual total con cobertura parcial e guión sen resultados', async () => {
    installFetch({ '/api/assets': [asset], '/api/reports/total/month?year=2026&month=10': { period: { year: 2026, month: 10 }, totalAssets: 2, assetsActive: 2, assetsWithResult: 1, partial: true, totalInvested: 1000, investedInMonth: 100, investedPlusPrevDividends: 105, dividends: 3, dividendsPrev: 5, holdingNoDiv: 1200, holdingWithDiv: 1205, baseNoDiv: 600, baseWithDiv: 605, resultNoDiv: 660, resultWithDiv: 663, gainNoDiv: 60, gainWithDiv: 58, pctNoDiv: 10, pctWithDiv: 9.59, hasResults: true, hasMetrics: true, hasPctWithDiv: true, averages: { months: 2, pctNoDiv: 5, gainNoDiv: 30, pctWithDiv: 4, gainWithDiv: 25 } } })
    renderWithProviders(<InformesPage />, { initialEntries: ['/informes?vista=mensual&ambito=total'] })
    expect(await screen.findByText('Resultado (sen div)')).toBeInTheDocument()
    expect(screen.getByText(/1\/2 activos con resultado/)).toBeInTheDocument()
    expect(screen.getByText('Promedios mensuais (2 mes(es) con resultado ata 10/2026)')).toBeInTheDocument()
  })

  it('renderiza filas históricas n/a', async () => {
    installFetch({ '/api/assets': [asset], '/api/meta': meta, '/api/reports/asset/1/history': { asset, totalInvested: 1000, avgIndexPct: 0, avgGain: 0, hasAverages: false, totalGain: 0, hasTotalGain: false, rows: [{ period: { year: 2026, month: 1 }, aporte: 100, holding: 0, result: 0, gain: 0, gainPct: 0, hasMetrics: false }] } })
    renderWithProviders(<AssetHistoryPanel assetId={1} />)
    expect(await screen.findByText('n/a')).toBeInTheDocument()
    expect(screen.getAllByText('—').length).toBeGreaterThan(0)
  })
})

describe('ProxeccionPage', () => {
  it('amosa start manual sen activos, envía payload e renderiza rules', async () => {
    let posted = ''
    installFetch({ '/api/tools/projection/start': { start: null }, '/api/tools/projection': (init: RequestInit) => { posted = String(init.body); return { start: { year: 2026, month: 10 }, startFromAssets: false, input: { annualSalary: 120000, monthlyReturnPct: 1.5 }, rules: { investmentRate: 0.188235, salaryRaiseRate: 0.10, salaryRaiseMonth: 9, years: 20 }, months: [{ index: 1, date: { year: 2026, month: 10 }, annualSalary: 120000, monthlySalary: 10000, investment: 1882.35, totalInvested: 1882.35, return: 28.24, totalGains: 28.24, totalCapital: 1910.59 }], summary: { finalAnnualSalary: 120000, totalInvested: 1882.35, totalGains: 28.24, finalCapital: 1910.59 } } } })
    const user = userEvent.setup()
    renderWithProviders(<ProxeccionPage />)
    expect(await screen.findByText('Data de inicio')).toBeInTheDocument()
    await user.type(screen.getByLabelText('Salario anual inicial (USD)'), '120000')
    await user.type(screen.getByLabelText('Retorno mensual esperado (%)'), '1,5')
    await user.click(screen.getByRole('button', { name: 'Calcular proxección' }))
    await waitFor(() => expect(posted).toContain('"monthlyReturnPct":1.5'))
    expect(posted).toContain('"start"')
    expect(await screen.findByText('Investimento mensual')).toBeInTheDocument()
    expect(screen.getByText(/18,8235 % do salario mensual/)).toBeInTheDocument()
    expect(screen.getByText(/\+10,00 % cada setembro/)).toBeInTheDocument()
  })

  it('oculta start manual cando hai activos', async () => {
    installFetch({ '/api/tools/projection/start': { start: { year: 2026, month: 1 } } })
    renderWithProviders(<ProxeccionPage />)
    expect(await screen.findByText('Data de inicio dos investimentos: 01/2026 (activo máis antigo)')).toBeInTheDocument()
    expect(screen.queryByLabelText('Data de inicio')).not.toBeInTheDocument()
  })
})

describe('AllocationPanel', () => {
  it('envía selección e amosa reparto', async () => {
    let posted = ''
    installFetch({ '/api/assets': [asset], '/api/meta': meta, '/api/tools/allocation': (init: RequestInit) => { posted = String(init.body); return { total: 100, types: [{ type: 'accion', label: 'Acción', amount: 100, assets: [{ asset, amount: 100 }] }] } } })
    const user = userEvent.setup()
    renderWithProviders(<AllocationPanel />)
    await user.type(await screen.findByLabelText('Cantidade total (USD)'), '100')
    await user.click(screen.getByLabelText('Acción'))
    await user.click(screen.getByRole('button', { name: 'Repartir aporte' }))
    await waitFor(() => expect(posted).toContain('"total":100'))
    expect(posted).toContain('"assetIds":[1]')
    expect(await screen.findByText(/Reparto de aporte mensual:/)).toBeInTheDocument()
    expect(screen.getAllByText('ACME').length).toBeGreaterThan(0)
  })
})
