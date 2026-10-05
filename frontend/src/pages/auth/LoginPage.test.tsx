import { MantineProvider } from '@mantine/core'
import { DatesProvider } from '@mantine/dates'
import { ModalsProvider } from '@mantine/modals'
import { Notifications } from '@mantine/notifications'
import { QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { createMemoryRouter, RouterProvider } from 'react-router'

import { LoginPage } from '@/pages/auth/LoginPage'
import { RequireAuth } from '@/router'
import { createTestQueryClient } from '@/test/utils'
import { theme } from '@/theme'

function renderRouter(initialEntry: string, routes: Parameters<typeof createMemoryRouter>[0]) {
  const queryClient = createTestQueryClient()
  const router = createMemoryRouter(routes, { initialEntries: [initialEntry] })
  render(
    <MantineProvider theme={theme} defaultColorScheme="light">
      <DatesProvider settings={{ locale: 'gl', firstDayOfWeek: 1 }}>
        <QueryClientProvider client={queryClient}>
          <ModalsProvider>
            <Notifications />
            <RouterProvider router={router} />
          </ModalsProvider>
        </QueryClientProvider>
      </DatesProvider>
    </MantineProvider>,
  )
  return router
}

afterEach(() => {
  vi.restoreAllMocks()
})

describe('LoginPage', () => {
  it('shows the wrong password error', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async (input: RequestInfo | URL) => {
        const path = String(input)
        if (path === '/api/auth/status') {
          return new Response(JSON.stringify({ setupRequired: false, authenticated: false, setupAllowed: true }))
        }
        return new Response(JSON.stringify({ error: 'Contrasinal incorrecto' }), { status: 401 })
      }),
    )
    renderRouter('/login', [{ path: '/login', element: <LoginPage /> }])

    await userEvent.type(await screen.findByLabelText('Contrasinal'), 'incorrecto')
    await userEvent.click(screen.getByRole('button', { name: 'Entrar' }))

    expect(await screen.findByText('Contrasinal incorrecto')).toBeInTheDocument()
  })

  it('navigates after a successful login', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async (input: RequestInfo | URL) => {
        const path = String(input)
        if (path === '/api/auth/status') {
          return new Response(JSON.stringify({ setupRequired: false, authenticated: false, setupAllowed: true }))
        }
        return new Response(JSON.stringify({ authenticated: true }))
      }),
    )
    renderRouter('/login?next=%2Factivos', [
      { path: '/login', element: <LoginPage /> },
      { path: '/activos', element: <div>Destino activos</div> },
    ])

    await userEvent.type(await screen.findByLabelText('Contrasinal'), 'correcto')
    await userEvent.click(screen.getByRole('button', { name: 'Entrar' }))

    expect(await screen.findByText('Destino activos')).toBeInTheDocument()
  })
})

describe('RequireAuth', () => {
  function routes() {
    return [
      {
        element: <RequireAuth />,
        children: [{ path: '/privado', element: <div>Zona privada</div> }],
      },
      { path: '/login', element: <div>Páxina login</div> },
      { path: '/configurar', element: <div>Páxina configurar</div> },
    ]
  }

  it('redirects to setup when setup is required', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        new Response(JSON.stringify({ setupRequired: true, authenticated: false, setupAllowed: true })),
      ),
    )
    renderRouter('/privado', routes())

    expect(await screen.findByText('Páxina configurar')).toBeInTheDocument()
  })

  it('redirects unauthenticated users to login', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        new Response(JSON.stringify({ setupRequired: false, authenticated: false, setupAllowed: true })),
      ),
    )
    const router = renderRouter('/privado', routes())

    expect(await screen.findByText('Páxina login')).toBeInTheDocument()
    await waitFor(() => expect(router.state.location.search).toBe('?next=%2Fprivado'))
  })

  it('renders private routes for authenticated users', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        new Response(JSON.stringify({ setupRequired: false, authenticated: true, setupAllowed: true })),
      ),
    )
    renderRouter('/privado', routes())

    expect(await screen.findByText('Zona privada')).toBeInTheDocument()
  })
})
