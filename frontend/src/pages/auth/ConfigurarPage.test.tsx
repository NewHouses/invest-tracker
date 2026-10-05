import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Route, Routes } from 'react-router'

import { ConfigurarPage } from '@/pages/auth/ConfigurarPage'
import { renderWithProviders } from '@/test/utils'

afterEach(() => {
  vi.restoreAllMocks()
  vi.unstubAllGlobals()
})

function installFetch(onSetup: (body: unknown) => Response) {
  vi.stubGlobal(
    'fetch',
    vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const path = String(input)
      if (path === '/api/auth/status') {
        return new Response(JSON.stringify({ setupRequired: true, authenticated: false, setupAllowed: true }))
      }
      if (path === '/api/auth/setup') {
        return onSetup(JSON.parse(String(init?.body)))
      }
      return new Response('{}', { status: 404 })
    }),
  )
}

function renderPage() {
  renderWithProviders(
    <Routes>
      <Route path="/configurar" element={<ConfigurarPage />} />
      <Route path="/" element={<div>Inicio cargado</div>} />
    </Routes>,
    { initialEntries: ['/configurar'] },
  )
}

describe('ConfigurarPage', () => {
  // O contrasinal non ten requisitos: un só carácter é válido.
  it('acepta un contrasinal curto sen requisitos de lonxitude', async () => {
    let posted: unknown
    installFetch((body) => {
      posted = body
      return new Response(JSON.stringify({ authenticated: true }), { status: 201 })
    })
    renderPage()

    await userEvent.type(await screen.findByLabelText(/^Contrasinal/), 'x')
    await userEvent.type(screen.getByLabelText(/Confirmar contrasinal/), 'x')
    await userEvent.click(screen.getByRole('button', { name: 'Crear contrasinal' }))

    expect(await screen.findByText('Inicio cargado')).toBeInTheDocument()
    expect(posted).toEqual({ password: 'x' })
  })

  it('non envía un contrasinal baleiro', async () => {
    const setup = vi.fn(() => new Response('{}', { status: 201 }))
    installFetch(setup)
    renderPage()

    const password = await screen.findByLabelText(/^Contrasinal/)
    await userEvent.click(screen.getByRole('button', { name: 'Crear contrasinal' }))

    // O campo é obrigatorio: o navegador bloquea o envío antes de chegar á API.
    expect(password).toBeRequired()
    expect(password).toBeInvalid()
    expect(setup).not.toHaveBeenCalled()
  })
})
