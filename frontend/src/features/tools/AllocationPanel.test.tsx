import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { AllocationPanel } from '@/features/tools/AllocationPanel'
import { installFetch, testAsset as asset, testMeta as meta } from '@/test/fetchMock'
import { renderWithProviders } from '@/test/utils'

beforeEach(() => vi.restoreAllMocks())

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
