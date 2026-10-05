import type { ApiErrorBody } from '@/api/types'

export class ApiError extends Error {
  readonly status: number
  readonly fields: Record<string, string> | undefined
  readonly headers: Headers

  constructor(message: string, status: number, fields: Record<string, string> | undefined, headers: Headers) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.fields = fields
    this.headers = headers
  }
}

type ApiFetchOptions = {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  body?: unknown
  signal?: AbortSignal
}

type UnauthorizedHandler = () => void

const unauthorizedHandlers = new Set<UnauthorizedHandler>()

export function subscribeUnauthorized(handler: UnauthorizedHandler) {
  unauthorizedHandlers.add(handler)
  return () => {
    unauthorizedHandlers.delete(handler)
  }
}

function emitUnauthorized() {
  for (const handler of unauthorizedHandlers) {
    handler()
  }
}

async function parseErrorBody(response: Response): Promise<ApiErrorBody> {
  const text = await response.text()
  if (!text) {
    return {}
  }

  try {
    return JSON.parse(text) as ApiErrorBody
  } catch {
    return { error: 'Non se puido ler a resposta do servidor.' }
  }
}

export async function apiFetch<T>(path: string, options: ApiFetchOptions = {}): Promise<T> {
  const headers = new Headers({ Accept: 'application/json' })
  const init: RequestInit = {
    method: options.method ?? 'GET',
    credentials: 'same-origin',
    headers,
  }

  if (options.signal) {
    init.signal = options.signal
  }

  if (options.body !== undefined) {
    headers.set('Content-Type', 'application/json')
    init.body = JSON.stringify(options.body)
  }

  const response = await fetch(path, init)

  if (!response.ok) {
    const body = await parseErrorBody(response)
    const message = body.error ?? 'Produciuse un erro ao falar co servidor.'
    const error = new ApiError(message, response.status, body.fields, response.headers)
    if (response.status === 401 && !path.startsWith('/api/auth/')) {
      emitUnauthorized()
    }
    throw error
  }

  if (response.status === 204) {
    return undefined as T
  }

  const text = await response.text()
  if (!text) {
    return undefined as T
  }
  return JSON.parse(text) as T
}
