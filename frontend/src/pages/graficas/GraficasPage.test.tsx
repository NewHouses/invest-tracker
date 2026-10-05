import { fireEvent, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { Distribution, LineChartDTO } from '@/api/charts'
import { GraficasPage } from '@/pages/graficas/GraficasPage'
import { installFetch, testAsset as asset } from '@/test/fetchMock'
import { renderWithProviders } from '@/test/utils'

vi.mock('@mantine/charts', () => ({
  LineChart: ({ series }: { series: Array<{ name: string; label?: string }> }) => (
    <div data-testid="line-chart">
      {series.map((item) => <span key={item.name}>{item.label ?? item.name}</span>)}
    </div>
  ),
  DonutChart: ({ data, chartLabel }: { data: Array<{ name: string; value: number; color: string }>; chartLabel?: string | number }) => (
    <div data-testid="donut-chart" data-label={chartLabel}>
      {data.map((item) => <span key={item.name} data-color={item.color}>{item.name}</span>)}
    </div>
  ),
}))

beforeEach(() => vi.restoreAllMocks())

const distribution: Distribution = {
  hasAssets: true,
  total: 1500,
  items: [
    { asset, label: 'ACME', value: 1000, pct: 66.7 },
    { asset: { ...asset, id: 2, name: 'Beta' }, label: 'Beta', value: 500, pct: 33.3 },
  ],
}

const allAssetsChart: LineChartDTO = {
  title: 'Todos os ativos',
  hasAssets: true,
  months: [{ year: 2026, month: 1 }, { year: 2026, month: 2 }],
  series: [
    { label: 'ACME', values: [1000, null] },
    { label: 'Beta', values: [500, 550] },
  ],
}

describe('GraficasPage', () => {
  it('pide a gráfica de todos os ativos ao seleccionar esa opción', async () => {
    const calls = installFetch({ '/api/assets': [asset], '/api/charts/distribution': distribution, '/api/charts/assets': allAssetsChart })
    const user = userEvent.setup()

    renderWithProviders(<GraficasPage />, { initialEntries: ['/graficas'] })
    await user.click(await screen.findByRole('button', { name: 'Evolución de todos os ativos' }))

    await waitFor(() => expect(calls.some((call) => call.url === '/api/charts/assets')).toBe(true))
    expect(screen.getByRole('button', { name: 'Evolución de todos os ativos' })).toHaveAttribute('aria-pressed', 'true')
  })

  it('permite ocultar unha serie desde a lenda', async () => {
    installFetch({ '/api/assets': [asset], '/api/charts/distribution': distribution, '/api/charts/assets': allAssetsChart })
    const user = userEvent.setup()

    renderWithProviders(<GraficasPage />, { initialEntries: ['/graficas?grafica=allAssets'] })

    const acmeLegend = await screen.findByRole('button', { name: 'ACME' })
    expect(acmeLegend).toHaveAttribute('aria-pressed', 'true')

    await user.click(acmeLegend)

    expect(acmeLegend).toHaveAttribute('aria-pressed', 'false')
  })

  it('resalta a fila da lenda de distribución ao pasar o rato', async () => {
    installFetch({ '/api/assets': [asset], '/api/charts/distribution': distribution })

    renderWithProviders(<GraficasPage />, { initialEntries: ['/graficas'] })

    const acmeLabels = await screen.findAllByText('ACME')
    const row = acmeLabels.map((label) => label.closest('tr')).find((candidate) => candidate !== null)
    expect(row).toBeDefined()
    fireEvent.mouseEnter(row as HTMLTableRowElement)

    expect(row).toHaveAttribute('data-highlighted', 'true')
  })
})
