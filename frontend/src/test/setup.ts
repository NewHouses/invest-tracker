import '@testing-library/jest-dom/vitest'
import { beforeEach, vi } from 'vitest'

import { setLanguage } from '@/i18n'

// Os tests están escritos en galego, pero jsdom simula un navegador en inglés:
// cada test empeza en galego (os que proban o inglés cámbiano eles mesmos).
beforeEach(() => {
  setLanguage('gl')
})

Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
})

class MockResizeObserver implements ResizeObserver {
  observe = vi.fn()
  unobserve = vi.fn()
  disconnect = vi.fn()
}

globalThis.ResizeObserver = MockResizeObserver
