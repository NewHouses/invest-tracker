import { vi } from 'vitest'

// Utilidades compartidas para mockear a API nos tests de páxinas.

export const testAsset = { id: 1, type: 'accion', name: 'ACME', amountUsd: 1000, year: 2026, month: 1 } as const

export const testMeta = {
  assetTypes: [
    { value: 'accion', label: 'Acción' },
    { value: 'indice', label: 'Índice' },
    { value: 'copy_trading', label: 'Copy-trading' },
    { value: 'fondo', label: 'Fondo' },
  ],
  minYear: 2000,
  maxYear: 2100,
}

export function jsonResponse(data: unknown, status = 200) {
  return new Response(JSON.stringify(data), { status, headers: { 'Content-Type': 'application/json' } })
}

type Route = unknown | ((init: RequestInit) => unknown)

// installFetch substitúe fetch por un router simple de URL exacta → resposta
// (ou función que recibe o RequestInit). As rutas descoñecidas dan 404.
export function installFetch(routes: Record<string, Route>) {
  const calls: Array<{ url: string; init: RequestInit }> = []
  vi.stubGlobal(
    'fetch',
    vi.fn((input: RequestInfo | URL, init: RequestInit = {}) => {
      const url = String(input)
      calls.push({ url, init })
      const route = routes[url]
      if (route === undefined) return Promise.resolve(jsonResponse({ error: `sen ruta ${url}` }, 404))
      return Promise.resolve(jsonResponse(typeof route === 'function' ? (route as (init: RequestInit) => unknown)(init) : route))
    }),
  )
  return calls
}
