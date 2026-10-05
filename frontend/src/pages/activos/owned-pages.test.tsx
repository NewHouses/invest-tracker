import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { useAssets } from '@/api/assets'
import { useCreateAsset } from '@/api/assetMutations'
import { useCreateTransaction } from '@/api/transactions'
import { TransactionForm } from '@/features/transactions/TransactionForm'
import { ActivosPage } from '@/pages/activos/ActivosPage'
import { TransaccionsPage } from '@/pages/transaccions/TransaccionsPage'
import { ResultadosPage } from '@/pages/resultados/ResultadosPage'
import { DividendosPage } from '@/pages/dividendos/DividendosPage'
import { renderWithProviders } from '@/test/utils'

const meta = { assetTypes: [{ value: 'accion', label: 'Acción' }, { value: 'indice', label: 'Índice' }, { value: 'fondo', label: 'Fondo' }, { value: 'copy_trading', label: 'Copy-trading' }], minYear: 1900, maxYear: 2200 }
const assetA = { id: 1, type: 'indice', name: 'SP500', amountUsd: 1000, year: 2026, month: 1 }
const assetB = { id: 2, type: 'accion', name: 'Novo', amountUsd: 200, year: 2026, month: 12 }

type RouteHandler = (path: string, init: RequestInit | undefined) => Response | Promise<Response>

function json(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } })
}

function noContent() {
  return new Response(null, { status: 204 })
}

function stubFetch(handler: RouteHandler) {
  const calls: Array<{ path: string; method: string; body: unknown }> = []
  vi.stubGlobal('fetch', vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = new URL(String(input), 'http://localhost')
    const method = init?.method ?? 'GET'
    const body = typeof init?.body === 'string' ? JSON.parse(init.body) : undefined
    calls.push({ path: `${url.pathname}${url.search}`, method, body })
    if (url.pathname === '/api/meta') return json(meta)
    return handler(`${url.pathname}${url.search}`, init)
  }))
  return calls
}

afterEach(() => {
  vi.restoreAllMocks()
})

function AssetCreateHarness() {
  const assets = useAssets()
  const create = useCreateAsset()
  return (
    <>
      <button type="button" onClick={() => void create.mutateAsync({ type: 'indice', name: 'SP500', amountUsd: 1000, year: 2026, month: 1 })}>
        Crear ativo
      </button>
      <ul>{(assets.data ?? []).map((asset) => <li key={asset.id}>{asset.name}</li>)}</ul>
    </>
  )
}

describe('asset pages', () => {
  it('creates an asset and refreshes the list', async () => {
    let assets = [] as typeof assetA[]
    const calls = stubFetch((path, init) => {
      if (path === '/api/assets' && (init?.method ?? 'GET') === 'GET') return json(assets)
      if (path === '/api/assets' && init?.method === 'POST') {
        assets = [assetA]
        return json(assetA, 201)
      }
      return json([])
    })

    renderWithProviders(<AssetCreateHarness />)
    await userEvent.click(screen.getByRole('button', { name: 'Crear ativo' }))

    await waitFor(() => expect(calls.some((call) => call.method === 'POST' && call.path === '/api/assets')).toBe(true))
    const post = calls.find((call) => call.method === 'POST' && call.path === '/api/assets')
    expect(post?.body).toMatchObject({ type: 'indice', name: 'SP500', amountUsd: 1000 })
    expect(await screen.findByText('SP500')).toBeInTheDocument()
    expect(calls.filter((call) => call.method === 'GET' && call.path === '/api/assets').length).toBeGreaterThan(1)
  })

  it('deletes an asset after confirmation', async () => {
    const calls = stubFetch((path, init) => {
      if (path === '/api/assets' && (init?.method ?? 'GET') === 'GET') return json([assetA])
      if (path === '/api/portfolio') return json({ period: { year: 2026, month: 10 }, pendingCount: 0, rows: [] })
      if (path === '/api/assets/1' && init?.method === 'DELETE') return noContent()
      return json([])
    })

    renderWithProviders(<ActivosPage />)
    expect(await screen.findByText('SP500')).toBeInTheDocument()
    await userEvent.click(screen.getByLabelText('Eliminar'))
    const modal = await screen.findByRole('dialog')
    await userEvent.click(within(modal).getByRole('button', { name: 'Eliminar' }))

    await waitFor(() => expect(calls.some((call) => call.method === 'DELETE' && call.path === '/api/assets/1')).toBe(true))
  })

  it('sorts the Portofolio by a clicked header', async () => {
    stubFetch((path) => {
      if (path === '/api/assets') return json([assetA, assetB])
      if (path === '/api/portfolio') return json({
        period: { year: 2026, month: 10 },
        pendingCount: 0,
        rows: [
          { asset: assetA, totalInvested: 1000, currentValue: 1200, hasCurrentValue: true, gain: 200, gainPct: 20, hasGain: true, hasGainPct: true, lastResult: { year: 2026, month: 10 }, pending: false },
          { asset: assetB, totalInvested: 250, currentValue: 0, hasCurrentValue: false, gain: 0, gainPct: 0, hasGain: false, hasGainPct: false, lastResult: null, pending: true },
        ],
      })
      return json([])
    })

    renderWithProviders(<ActivosPage />)
    expect(await screen.findByText('SP500')).toBeInTheDocument()

    const bodyRows = () => screen.getAllByRole('row').slice(1).map((row) => row.textContent ?? '')
    expect(bodyRows()[0]).toContain('SP500')
    expect(bodyRows()[1]).toContain('Novo')

    await userEvent.click(screen.getByRole('button', { name: /Nome/ }))
    expect(bodyRows()[0]).toContain('Novo')
    expect(bodyRows()[1]).toContain('SP500')

    await userEvent.click(screen.getByRole('button', { name: /Nome/ }))
    expect(bodyRows()[0]).toContain('SP500')
    expect(bodyRows()[1]).toContain('Novo')

    await userEvent.click(screen.getByRole('button', { name: /Nome/ }))
    expect(bodyRows()[0]).toContain('SP500')
    expect(bodyRows()[1]).toContain('Novo')
  })
})

function TransactionCreateHarness() {
  const create = useCreateTransaction()
  return <TransactionForm assetId={1} submitting={create.isPending} error={create.error} onSubmit={async (values) => {
    if ('assetId' in values) await create.mutateAsync(values)
  }} />
}

describe('transaction page', () => {
  it('shows the server month field error in the Nova tab', async () => {
    stubFetch((path, init) => {
      if (path === '/api/assets') return json([assetA])
      if (path === '/api/transactions' && init?.method === 'POST') return json({ error: 'Erro de validación', fields: { month: 'a transacción non pode ser anterior á data do ativo (01/2026)' } }, 400)
      return json([])
    })

    renderWithProviders(<TransactionCreateHarness />)
    await userEvent.type(screen.getByLabelText(/^Importe/), '50')
    await userEvent.click(screen.getByRole('button', { name: 'Gardar transacción' }))

    expect(await screen.findByText(/a transacción non pode ser anterior/)).toBeInTheDocument()
  })

  it('lists omitted monthly assets and posts only non-empty amounts', async () => {
    const calls = stubFetch((path, init) => {
      if (path.startsWith('/api/transactions/month-assets')) return json({ period: { year: 2026, month: 10 }, eligible: [assetA], omitted: [assetB] })
      if (path === '/api/transactions/month' && init?.method === 'POST') return json({ ids: [10] }, 201)
      return json([])
    })

    renderWithProviders(<TransaccionsPage />, { initialEntries: ['/transaccions?tab=mes'] })
    expect(await screen.findByText(/Novo: omitido/)).toBeInTheDocument()
    await userEvent.type(screen.getByLabelText('Importe SP500'), '75')
    await userEvent.click(screen.getByRole('button', { name: 'Gardar transaccións do mes' }))

    await waitFor(() => expect(calls.some((call) => call.path === '/api/transactions/month' && call.method === 'POST')).toBe(true))
    const post = calls.find((call) => call.path === '/api/transactions/month' && call.method === 'POST')
    expect(post?.body).toMatchObject({ items: [{ assetId: 1, amountUsd: 75 }] })
  })
})

describe('results page', () => {
  it('closes a month with only non-empty results and shows G/P preview', async () => {
    const calls = stubFetch((path, init) => {
      if (path.startsWith('/api/results/eligible')) return json({ period: { year: 2026, month: 10 }, items: [
        { asset: assetA, holding: 100, result: 0, hasResult: false },
        { asset: assetB, holding: 200, result: 210, hasResult: true },
      ] })
      if (path === '/api/results/close-month' && init?.method === 'POST') return json({ ids: [1] }, 201)
      return json([])
    })

    renderWithProviders(<ResultadosPage />)
    await userEvent.type(await screen.findByLabelText('Resultado SP500'), '130')
    await waitFor(() => expect(screen.getAllByText(/\+30/).length).toBeGreaterThan(0))
    await userEvent.click(screen.getByRole('button', { name: 'Pechar mes' }))

    await waitFor(() => expect(calls.some((call) => call.path === '/api/results/close-month')).toBe(true))
    const post = calls.find((call) => call.path === '/api/results/close-month')
    expect(post?.body).toMatchObject({ items: [{ assetId: 1, resultUsd: 130 }] })
  })

  it('shows the CLI-like message when clearing a month deletes nothing', async () => {
    stubFetch((path, init) => {
      if (path.startsWith('/api/results?') && init?.method === 'DELETE') return json({ deleted: 0 })
      return json([])
    })

    renderWithProviders(<ResultadosPage />, { initialEntries: ['/resultados?tab=limpar'] })
    await userEvent.click(screen.getByRole('button', { name: 'Limpar mes' }))
    const modal = await screen.findByRole('dialog')
    await userEvent.click(within(modal).getByRole('button', { name: 'Eliminar' }))

    expect(await screen.findByText(/Non había resultados rexistrados/)).toBeInTheDocument()
  })
})

describe('dividends page', () => {
  it('renders the dividends total row', async () => {
    stubFetch((path) => {
      if (path === '/api/dividends') return json([
        { id: 1, amountUsd: 10, year: 2026, month: 1 },
        { id: 2, amountUsd: 15.5, year: 2026, month: 2 },
      ])
      return json([])
    })

    renderWithProviders(<DividendosPage />)
    expect(await screen.findByText('Total')).toBeInTheDocument()
    expect(screen.getByText(/25,50/)).toBeInTheDocument()
  })
})
