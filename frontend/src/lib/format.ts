import type { YearMonth } from '@/api/types'

// Formato galego explícito (miles con ".", decimais con ","): o ICU dos
// navegadores Chromium (Edge, Chrome) non sempre inclúe datos de "gl" e
// formatearía en inglés ("$1,234.50"). Así sae igual en calquera navegador.
export function formatNumber(value: number, decimals = 2) {
  const [int = '0', frac] = Math.abs(value).toFixed(decimals).split('.')
  const grouped = int.replace(/\B(?=(\d{3})+(?!\d))/g, '.')
  const sign = value < 0 && Number(Math.abs(value).toFixed(decimals)) !== 0 ? '-' : ''
  return `${sign}${grouped}${frac ? `,${frac}` : ''}`
}

export const galicianMonthNames = [
  'xaneiro',
  'febreiro',
  'marzo',
  'abril',
  'maio',
  'xuño',
  'xullo',
  'agosto',
  'setembro',
  'outubro',
  'novembro',
  'decembro',
] as const

export function formatUSD(value: number) {
  return `${formatNumber(value)}\u00a0$`
}

// Formato curto para os eixos das gráficas ("12,5 k $", "1,2 M $"); o
// tooltip segue amosando o importe completo con formatUSD.
export function formatAxisUSD(value: number) {
  const abs = Math.abs(value)
  const short = (n: number) => formatNumber(n, Number.isInteger(n) ? 0 : 1)
  if (abs >= 1_000_000) return `${short(value / 1_000_000)}\u00a0M\u00a0$`
  if (abs >= 1_000) return `${short(value / 1_000)}\u00a0k\u00a0$`
  return `${formatNumber(value, 0)}\u00a0$`
}

export function formatSignedUSD(value: number) {
  const sign = value > 0 ? '+' : value < 0 ? '−' : '+'
  return `${sign}${formatUSD(Math.abs(value))}`
}

export function formatPct(value: number) {
  const sign = value >= 0 ? '+' : '−'
  return `${sign}${formatNumber(Math.abs(value))} %`
}

export function formatYearMonth(value: YearMonth) {
  return `${String(value.month).padStart(2, '0')}/${value.year}`
}
