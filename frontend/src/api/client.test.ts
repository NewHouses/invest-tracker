import { afterEach, describe, expect, it, vi } from 'vitest'

import { apiFetch } from '@/api/client'
import { setLanguage } from '@/i18n'

afterEach(() => {
  vi.restoreAllMocks()
})

describe('apiFetch', () => {
  it('parses JSON error bodies', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        new Response(JSON.stringify({ error: 'datos non válidos', fields: { amountUsd: 'o importe debe ser un número maior ca 0' } }), {
          status: 400,
          headers: { 'Content-Type': 'application/json' },
        }),
      ),
    )

    await expect(apiFetch('/api/dividends', { method: 'POST', body: { amountUsd: 0, year: 2026, month: 10 } })).rejects.toMatchObject({
      status: 400,
      message: 'datos non válidos',
      fields: { amountUsd: 'o importe debe ser un número maior ca 0' },
    })
  })

  it('uses a Galician fallback for non-JSON errors', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response('<html></html>', { status: 500 })))

    await expect(apiFetch('/api/meta')).rejects.toThrow('Non se puido ler a resposta do servidor.')
  })

  it('returns undefined for 204 responses', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response(null, { status: 204 })))

    await expect(apiFetch<void>('/api/dividends/1', { method: 'DELETE' })).resolves.toBeUndefined()
  })

  it('asks the server to answer in the current language', async () => {
    const fetchMock = vi.fn<(input: RequestInfo | URL, init?: RequestInit) => Promise<Response>>(
      async () => new Response('{}', { status: 200 }),
    )
    vi.stubGlobal('fetch', fetchMock)

    await apiFetch('/api/meta')
    setLanguage('en')
    await apiFetch('/api/meta')

    const languages = fetchMock.mock.calls.map(([, init]) => new Headers(init?.headers).get('Accept-Language'))
    expect(languages).toEqual(['gl', 'en'])
  })

  it('uses an English fallback for non-JSON errors in English', async () => {
    setLanguage('en')
    vi.stubGlobal('fetch', vi.fn(async () => new Response('<html></html>', { status: 500 })))

    await expect(apiFetch('/api/meta')).rejects.toThrow('Could not read the server response.')
  })
})
