import { MantineProvider } from '@mantine/core'
import { ModalsProvider } from '@mantine/modals'
import { Notifications } from '@mantine/notifications'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render } from '@testing-library/react'
import type { RenderOptions } from '@testing-library/react'
import type { ReactElement, ReactNode } from 'react'
import { MemoryRouter } from 'react-router'

import { ApiError } from '@/api/client'
import { LanguageProvider } from '@/i18n'
import { theme } from '@/theme'

export function createTestQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        refetchOnWindowFocus: false,
      },
      mutations: {
        retry: false,
      },
    },
  })
}

type ProvidersProps = {
  children: ReactNode
  queryClient?: QueryClient | undefined
  initialEntries?: string[] | undefined
}

export function TestProviders({ children, queryClient = createTestQueryClient(), initialEntries = ['/'] }: ProvidersProps) {
  return (
    <MantineProvider theme={theme} defaultColorScheme="light">
      <QueryClientProvider client={queryClient}>
        <LanguageProvider>
          <ModalsProvider>
            <Notifications />
            <MemoryRouter initialEntries={initialEntries}>{children}</MemoryRouter>
          </ModalsProvider>
        </LanguageProvider>
      </QueryClientProvider>
    </MantineProvider>
  )
}

type RenderWithProvidersOptions = Omit<RenderOptions, 'wrapper'> & {
  queryClient?: QueryClient
  initialEntries?: string[]
}

export function renderWithProviders(ui: ReactElement, options: RenderWithProvidersOptions = {}) {
  const { queryClient, initialEntries, ...renderOptions } = options
  return render(ui, {
    wrapper: ({ children }) => (
      <TestProviders queryClient={queryClient} initialEntries={initialEntries}>
        {children}
      </TestProviders>
    ),
    ...renderOptions,
  })
}

export function makeApiError(status: number, body: string | object, headers?: HeadersInit) {
  return new Response(typeof body === 'string' ? body : JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json', ...headers },
  })
}

export function expectApiError(error: unknown): asserts error is ApiError {
  expect(error).toBeInstanceOf(ApiError)
}
