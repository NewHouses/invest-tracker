import dayjs from 'dayjs'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { detectLanguage, getLanguage, LANGUAGE_STORAGE_KEY, parseLanguage, setLanguage, t, translate } from '@/i18n'

afterEach(() => {
  vi.restoreAllMocks()
  localStorage.clear()
})

describe('idioma', () => {
  it('recoñece as etiquetas de idioma', () => {
    expect(parseLanguage('en-US')).toBe('en')
    expect(parseLanguage(' EN_gb ')).toBe('en')
    expect(parseLanguage('gl-ES')).toBe('gl')
    expect(parseLanguage('es-ES')).toBeNull()
    expect(parseLanguage(null)).toBeNull()
  })

  it('por defecto usa o inglés só se o navegador está en inglés', () => {
    localStorage.clear()
    const languages = vi.spyOn(navigator, 'languages', 'get')
    languages.mockReturnValue(['en-GB', 'gl'])
    expect(detectLanguage()).toBe('en')
    languages.mockReturnValue(['es-ES', 'en-US'])
    expect(detectLanguage()).toBe('gl')
    languages.mockReturnValue(['gl-ES'])
    expect(detectLanguage()).toBe('gl')
  })

  it('lembra o idioma escollido aínda que o navegador estea noutro', () => {
    vi.spyOn(navigator, 'languages', 'get').mockReturnValue(['en-US'])
    setLanguage('gl')
    expect(localStorage.getItem(LANGUAGE_STORAGE_KEY)).toBe('gl')
    expect(detectLanguage()).toBe('gl')
    expect(document.documentElement.lang).toBe('gl')

    setLanguage('en')
    expect(getLanguage()).toBe('en')
    expect(detectLanguage()).toBe('en')
    expect(document.documentElement.lang).toBe('en')
    expect(dayjs.locale()).toBe('en')
  })

  it('traduce co idioma actual', () => {
    setLanguage('gl')
    expect(t('Axustes')).toBe('Axustes')
    setLanguage('en')
    expect(t('Axustes')).toBe('Settings')
  })

  it('substitúe os parámetros e deixa en galego o que non ten tradución', () => {
    expect(translate('gl', '{x} de {y}', { x: 1, y: 'dous' })).toBe('1 de dous')
    expect(translate('gl', '{x} e {falta}', { x: 1 })).toBe('1 e {falta}')
    expect(translate('en', 'texto sen tradución')).toBe('texto sen tradución')
  })
})
