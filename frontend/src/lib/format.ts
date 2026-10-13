import type { YearMonth } from '@/api/types'
import { getLanguage } from '@/i18n'
import type { Language } from '@/i18n'

// Separadores de miles e decimais de cada idioma: galego "1.234,56" e inglés
// "1,234.56".
export function numberSeparators(language: Language = getLanguage()) {
  return language === 'en' ? { decimal: '.', group: ',' } : { decimal: ',', group: '.' }
}

// Props dos NumberInput de Mantine cos separadores do idioma actual.
export function numberInputSeparators() {
  const { decimal, group } = numberSeparators()
  return { decimalSeparator: decimal, thousandSeparator: group }
}

// Formato explícito en vez de Intl: o ICU dos navegadores Chromium (Edge,
// Chrome) non sempre inclúe datos de "gl" e formatearía en inglés
// ("$1,234.50"). Así sae igual en calquera navegador.
export function formatNumber(value: number, decimals = 2) {
  const { decimal, group } = numberSeparators()
  const [int = '0', frac] = Math.abs(value).toFixed(decimals).split('.')
  const grouped = int.replace(/\B(?=(\d{3})+(?!\d))/g, group)
  const sign = value < 0 && Number(Math.abs(value).toFixed(decimals)) !== 0 ? '-' : ''
  return `${sign}${grouped}${frac ? `${decimal}${frac}` : ''}`
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

// withDollar engade o símbolo do dólar a un importe xa formateado: en galego
// detrás ("1.234,50 $") e en inglés diante ("$1,234.50", "-$12.00").
function withDollar(amount: string) {
  if (getLanguage() !== 'en') return `${amount}\u00a0$`
  return amount.startsWith('-') ? `-$${amount.slice(1)}` : `$${amount}`
}

export function formatUSD(value: number) {
  return withDollar(formatNumber(value))
}

// Formato curto para os eixos das gráficas ("12,5 k $", "1,2 M $"; en inglés
// "$12.5k", "$1.2M"); o tooltip segue amosando o importe completo con
// formatUSD.
export function formatAxisUSD(value: number) {
  const abs = Math.abs(value)
  const short = (n: number) => formatNumber(n, Number.isInteger(n) ? 0 : 1)
  const scaled = (n: number, unit: string) =>
    getLanguage() === 'en' ? withDollar(`${short(n)}${unit}`) : `${short(n)}\u00a0${unit}\u00a0$`
  if (abs >= 1_000_000) return scaled(value / 1_000_000, 'M')
  if (abs >= 1_000) return scaled(value / 1_000, 'k')
  return withDollar(formatNumber(value, 0))
}

export function formatSignedUSD(value: number) {
  const sign = value > 0 ? '+' : value < 0 ? '−' : '+'
  return `${sign}${formatUSD(Math.abs(value))}`
}

export function formatPct(value: number) {
  const sign = value >= 0 ? '+' : '−'
  return `${sign}${formatNumber(Math.abs(value))}${getLanguage() === 'en' ? '%' : ' %'}`
}

// formatPercent formatea unha porcentaxe sen signo e só coas decimais que
// fan falta: galego "33,33 %", "12,5 %", "100 %"; inglés "33.33%".
export function formatPercent(value: number) {
  const hundredths = Math.round(value * 100)
  const decimals = hundredths % 100 === 0 ? 0 : hundredths % 10 === 0 ? 1 : 2
  return `${formatNumber(hundredths / 100, decimals)}${getLanguage() === 'en' ? '%' : ' %'}`
}

export function formatYearMonth(value: YearMonth) {
  return `${String(value.month).padStart(2, '0')}/${value.year}`
}
