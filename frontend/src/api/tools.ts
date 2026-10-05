import { useMutation, useQuery } from '@tanstack/react-query'

import { apiFetch } from '@/api/client'
import { queryKeys } from '@/api/queryKeys'
import type { Asset, AssetType, YearMonth } from '@/api/types'

export type GrowthRule = { kind: 'fixed' | 'percent'; value: number; everyMonths: number; from: YearMonth }
export type ProjectionMode = 'contribution' | 'salary'
export type ProjectionDefaultsResponse = {
  start: YearMonth
  startFromAssets: boolean
  years: number
  investmentRatePct: number
  salaryRules: GrowthRule[]
}
export type ProjectionRequest = {
  start: YearMonth
  years: number
  initialInvestment: number
  monthlyReturnPct: number
  mode: ProjectionMode
  monthlyContribution: number
  annualSalary: number
  investmentRatePct: number
  rules: GrowthRule[]
}
export type ProjectionMonth = {
  index: number
  date: YearMonth
  annualSalary: number
  monthlySalary: number
  contribution: number
  totalInvested: number
  return: number
  totalGains: number
  totalCapital: number
}
export type ProjectionResponse = {
  input: ProjectionRequest
  months: ProjectionMonth[]
  summary: {
    initialInvestment: number
    firstContribution: number
    finalContribution: number
    finalAnnualSalary: number
    totalContributions: number
    totalInvested: number
    totalGains: number
    finalCapital: number
    months: number
  }
}

export type AllocationRequest = { total: number; selection: Array<{ type: AssetType; assetIds: number[] }> }
export type Allocation = {
  total: number
  types: Array<{ type: AssetType; label: string; amount: number; assets: Array<{ asset: Asset; amount: number }> }>
}

export function useProjectionDefaults() {
  return useQuery({
    queryKey: queryKeys.tools.projectionDefaults,
    queryFn: ({ signal }) => apiFetch<ProjectionDefaultsResponse>('/api/tools/projection/defaults', { signal }),
  })
}

export function useProjectionMutation() {
  return useMutation({ mutationFn: (body: ProjectionRequest) => apiFetch<ProjectionResponse>('/api/tools/projection', { method: 'POST', body }) })
}

export function useAllocationMutation() {
  return useMutation({ mutationFn: (body: AllocationRequest) => apiFetch<Allocation>('/api/tools/allocation', { method: 'POST', body }) })
}
