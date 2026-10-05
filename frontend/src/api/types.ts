export type AssetType = 'accion' | 'indice' | 'copy_trading' | 'fondo'

export type YearMonth = {
  year: number
  month: number
}

export type Asset = {
  id: number
  type: AssetType
  name: string
  amountUsd: number
  month: number
  year: number
}

export type Transaction = {
  id: number
  assetId: number
  amountUsd: number
  month: number
  year: number
}

export type MonthlyResult = {
  id: number
  assetId: number
  resultUsd: number
  month: number
  year: number
}

export type Dividend = {
  id: number
  amountUsd: number
  month: number
  year: number
}

export type MonthlySummary = {
  totalInvestedUpTo: number
  investedInMonth: number
  result: number
  hasResult: boolean
  estimatedHolding: number
  hasPrevResult: boolean
}

export type ApiErrorBody = {
  error?: string
  fields?: Record<string, string>
}

export type AuthStatusResponse = {
  setupRequired: boolean
  authenticated: boolean
  setupAllowed: boolean
}

export type AuthenticatedResponse = {
  authenticated: true
}

export type MetaResponse = {
  assetTypes: Array<{ value: AssetType; label: string }>
  minYear: number
  maxYear: number
}

// Activo con capital nun mes (pode recibir resultado). Usado en "Pechar mes",
// "Engadir resultado" e nos pendentes do inicio.
export type EligibleAsset = {
  asset: Asset
  holding: number
  result: number
  hasResult: boolean
}

export type EligibleResponse = {
  period: YearMonth
  items: EligibleAsset[]
}

export type IdsResponse = {
  ids: number[]
}
