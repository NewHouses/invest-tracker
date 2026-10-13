import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { AllocationRequest } from '@/api/tools'
import { AllocationPanel } from '@/features/tools/AllocationPanel'
import { setLanguage } from '@/i18n'
import { installFetch, jsonResponse, testAsset as asset, testMeta as meta } from '@/test/fetchMock'
import { renderWithProviders } from '@/test/utils'

beforeEach(() => vi.restoreAllMocks())

const beta = { ...asset, id: 2, name: 'BETA' }
const idx = { ...asset, id: 3, type: 'indice', name: 'IDX' }

const acmeResult = { total: 100, types: [{ type: 'accion', label: 'Acción', percent: 100, amount: 100, assets: [{ asset, percent: 100, amount: 100, new: false }] }] }

describe('AllocationPanel', () => {
  it('envía selección e amosa reparto', async () => {
    let posted = ''
    installFetch({ '/api/assets': [asset], '/api/meta': meta, '/api/tools/allocation': (init: RequestInit) => { posted = String(init.body); return acmeResult } })
    const user = userEvent.setup()
    renderWithProviders(<AllocationPanel />)
    await user.type(await screen.findByLabelText('Cantidade total (USD)'), '100')
    await user.click(screen.getByLabelText('Acción'))
    await user.click(screen.getByRole('button', { name: 'Repartir aporte' }))
    await waitFor(() => expect(posted).toContain('"total":100'))
    expect(posted).toContain('"assetIds":[1]')
    // Sen cambiar a proposta, o servidor reparte a partes iguais exactas.
    expect(posted).not.toContain('percents')
    expect(await screen.findByText(/Reparto de aporte mensual:/)).toBeInTheDocument()
    expect(screen.getAllByText('ACME').length).toBeGreaterThan(0)
  })

  it('propón partes iguais e deixa escoller a porcentaxe de cada activo', async () => {
    let posted: AllocationRequest | null = null
    installFetch({ '/api/assets': [asset, beta, idx], '/api/meta': meta, '/api/tools/allocation': (init: RequestInit) => { posted = JSON.parse(String(init.body)) as AllocationRequest; return acmeResult } })
    const user = userEvent.setup()
    renderWithProviders(<AllocationPanel />)
    await user.type(await screen.findByLabelText('Cantidade total (USD)'), '1000')
    await user.click(screen.getByLabelText('Acción'))
    await user.click(screen.getByLabelText('Índice'))

    expect(screen.getByLabelText('Porcentaxe de ACME')).toHaveValue('25')
    expect(screen.getByLabelText('Porcentaxe de BETA')).toHaveValue('25')
    expect(screen.getByLabelText('Porcentaxe de IDX')).toHaveValue('50')
    expect(screen.getByText('Suma: 100 %')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Repartir a partes iguais' })).toBeDisabled()

    await user.clear(screen.getByLabelText('Porcentaxe de ACME'))
    await user.type(screen.getByLabelText('Porcentaxe de ACME'), '40')
    expect(screen.getByText('As porcentaxes suman 115 %; teñen que sumar 100 %.')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Repartir aporte' })).toBeDisabled()

    await user.clear(screen.getByLabelText('Porcentaxe de BETA'))
    await user.type(screen.getByLabelText('Porcentaxe de BETA'), '10')
    expect(screen.getByText('Suma: 100 %')).toBeInTheDocument()
    expect(screen.getByText('≈ 400,00 $')).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Repartir aporte' }))
    await waitFor(() => expect(posted).not.toBeNull())
    expect(posted).toEqual({
      total: 1000,
      selection: [
        { type: 'accion', assetIds: [1, 2], percents: [40, 10], newAssets: [] },
        { type: 'indice', assetIds: [3], percents: [50], newAssets: [] },
      ],
    })

    await user.click(screen.getByRole('button', { name: 'Repartir a partes iguais' }))
    expect(screen.getByLabelText('Porcentaxe de ACME')).toHaveValue('25')
  })

  it('reparte tamén entre activos novos, aínda sen ningún activo creado', async () => {
    let posted: AllocationRequest | null = null
    const nvda = { id: 0, type: 'accion', name: 'NVDA', amountUsd: 0, year: 0, month: 0 }
    installFetch({
      '/api/assets': [],
      '/api/meta': meta,
      '/api/tools/allocation': (init: RequestInit) => {
        posted = JSON.parse(String(init.body)) as AllocationRequest
        return { total: 100, types: [{ type: 'accion', label: 'Acción', percent: 100, amount: 100, assets: [{ asset: nvda, percent: 100, amount: 100, new: true }] }] }
      },
    })
    const user = userEvent.setup()
    renderWithProviders(<AllocationPanel />)
    expect(await screen.findByText(/Aínda non tes activos creados/)).toBeInTheDocument()
    await user.type(screen.getByLabelText('Cantidade total (USD)'), '100')
    await user.click(screen.getByRole('button', { name: 'Engadir activo novo' }))
    await user.type(screen.getByLabelText('Nome do activo novo'), 'NVDA')
    expect(screen.getByLabelText('Porcentaxe de NVDA')).toHaveValue('100')
    await user.click(screen.getByRole('button', { name: 'Repartir aporte' }))
    await waitFor(() => expect(posted).not.toBeNull())
    expect(posted).toEqual({ total: 100, selection: [{ type: 'accion', assetIds: [], newAssets: [{ name: 'NVDA' }] }] })
    expect(await screen.findByText('novo')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Quitar activo novo' }))
    expect(screen.queryByLabelText('Nome do activo novo')).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Repartir aporte' })).toBeDisabled()
  })

  it('amosa cada erro do servidor na súa fila', async () => {
    vi.stubGlobal('fetch', vi.fn((input: RequestInfo | URL) => {
      const url = String(input)
      if (url === '/api/assets') return Promise.resolve(jsonResponse([asset]))
      if (url === '/api/meta') return Promise.resolve(jsonResponse(meta))
      return Promise.resolve(jsonResponse({ error: 'datos non válidos', fields: { 'selection[0].newAssets[0].name': 'o nome non pode estar baleiro', percents: 'as porcentaxes teñen que sumar 100 %' } }, 400))
    }))
    const user = userEvent.setup()
    renderWithProviders(<AllocationPanel />)
    await user.type(await screen.findByLabelText('Cantidade total (USD)'), '100')
    await user.click(screen.getByLabelText('Acción'))
    await user.click(screen.getByRole('button', { name: 'Engadir activo novo' }))
    await user.click(screen.getByRole('button', { name: 'Repartir aporte' }))
    expect(await screen.findByText('o nome non pode estar baleiro')).toBeInTheDocument()
    expect(screen.getByText('as porcentaxes teñen que sumar 100 %')).toBeInTheDocument()
  })

  it('está en inglés', async () => {
    setLanguage('en')
    installFetch({ '/api/assets': [asset], '/api/meta': meta })
    const user = userEvent.setup()
    renderWithProviders(<AllocationPanel />)
    await user.click(await screen.findByLabelText('Acción'))
    expect(screen.getByLabelText('Percentage for ACME')).toHaveValue('100')
    expect(screen.getByText('Total: 100%')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Add new asset' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Split equally' })).toBeInTheDocument()
  })
})