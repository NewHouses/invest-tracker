import { describe, expect, it } from 'vitest'

import { formatAxisUSD, formatNumber, formatPct, formatSignedUSD, formatUSD, formatYearMonth, galicianMonthNames } from '@/lib/format'

describe('format helpers', () => {
  it('formats USD amounts with Galician locale', () => {
    expect(formatUSD(1234.5)).toBe('1.234,50 $')
  })

  // Regresión: con Intl('gl-ES') Edge mostraba "$1,234.50" por non ter datos
  // de galego no seu ICU; o formato ten que ser idéntico en tódolos navegadores.
  it('formats numbers deterministically without depending on Intl locale data', () => {
    expect(formatNumber(1234567.891)).toBe('1.234.567,89')
    expect(formatNumber(999.999)).toBe('1.000,00')
    expect(formatNumber(0)).toBe('0,00')
    expect(formatNumber(-0.001)).toBe('0,00')
    expect(formatNumber(-1234.5)).toBe('-1.234,50')
    expect(formatNumber(18.8235, 4)).toBe('18,8235')
    expect(formatUSD(-12)).toBe('-12,00\u00a0$')
  })

  it('formats compact axis amounts', () => {
    expect(formatAxisUSD(10000)).toBe('10\u00a0k\u00a0$')
    expect(formatAxisUSD(12500)).toBe('12,5\u00a0k\u00a0$')
    expect(formatAxisUSD(2_400_000)).toBe('2,4\u00a0M\u00a0$')
    expect(formatAxisUSD(750)).toBe('750\u00a0$')
    expect(formatAxisUSD(-5000)).toBe('-5\u00a0k\u00a0$')
  })

  it('formats signed USD amounts', () => {
    expect(formatSignedUSD(12)).toBe('+12,00 $')
    expect(formatSignedUSD(-12)).toBe('−12,00 $')
  })

  it('formats percentages with an explicit sign', () => {
    expect(formatPct(7.33)).toBe('+7,33 %')
    expect(formatPct(-7.33)).toBe('−7,33 %')
  })

  it('formats and names months', () => {
    expect(formatYearMonth({ year: 2026, month: 3 })).toBe('03/2026')
    expect(galicianMonthNames[2]).toBe('marzo')
  })
})
