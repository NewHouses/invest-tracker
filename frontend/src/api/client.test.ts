import { afterEach, describe, expect, it, vi } from 'vitest'

import { ApiError, apiFetch } from '@/api/client'

afterEach(() => {
  vi.restoreAllMocks()
})

describe('apiFetch', () => {
  it('parses JSON error bodies', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        new Response(JSON.stringify({ error: 'Erro de validación', fields: { password: 'Curto' } }), {
          status: 400,
          headers: { 'Content-Type': 'application/json' },
        }),
      ),
    )

    await expect(apiFetch('/api/auth/setup', { method: 'POST', body: { password: 'x' } })).rejects.toMatchObject({
      status: 400,
      message: 'Erro de validación',
      fields: { password: 'Curto' },
    })
  })

  it('uses a Galician fallback for non-JSON errors', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response('<html></html>', { status: 500 })))

    await expect(apiFetch('/api/meta')).rejects.toThrow('Non se puido ler a resposta do servidor.')
  })

  it('returns undefined for 204 responses', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response(null, { status: 204 })))

    await expect(apiFetch<void>('/api/auth/logout', { method: 'POST' })).resolves.toBeUndefined()
  })

  it('exposes response headers on ApiError', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        new Response(JSON.stringify({ error: 'Demasiados intentos' }), {
          status: 429,
          headers: { 'Retry-After': '10' },
        }),
      ),
    )

    try {
      await apiFetch('/api/auth/login', { method: 'POST', body: { password: 'x', remember: false } })
    } catch (error) {
      expect(error).toBeInstanceOf(ApiError)
      expect((error as ApiError).headers.get('Retry-After')).toBe('10')
    }
  })
})
