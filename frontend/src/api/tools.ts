import { useMutation, useQuery } from '@tanstack/react-query'

import { apiFetch } from '@/api/client'
import { queryKeys } from '@/api/queryKeys'
import type { Asset, AssetType, YearMonth } from '@/api/types'

export type ProjectionStartResponse = { start: YearMonth | null }
export type ProjectionRules = { investmentRate: number; salaryRaiseRate: number; salaryRaiseMonth: number; years: number }
export type ProjectionRequest = { annualSalary: number; monthlyReturnPct: number; start?: YearMonth }
export type ProjectionMonth = {
  index: number
  date: YearMonth
  annualSalary: number
  monthlySalary: number
  investment: number
  totalInvested: number
  return: number
  totalGains: number
  totalCapital: number
}
export type ProjectionResponse = {
  start: YearMonth
  startFromAssets: boolean
  input: { annualSalary: number; monthlyReturnPct: number }
  rules: ProjectionRules
  months: ProjectionMonth[]
  summary: { finalAnnualSalary: number; totalInvested: number; totalGains: number; finalCapital: number }
}

export type AllocationRequest = { total: number; selection: Array<{ type: AssetType; assetIds: number[] }> }
export type Allocation = {
  total: number
  types: Array<{ type: AssetType; label: string; amount: number; assets: Array<{ asset: Asset; amount: number }> }>
}

export function useProjectionStart() {
  return useQuery({
    queryKey: queryKeys.tools.projectionStart,
    queryFn: ({ signal }) => apiFetch<ProjectionStartResponse>('/api/tools/projection/start', { signal }),
  })
}

export function useProjectionMutation() {
  return useMutation({ mutationFn: (body: ProjectionRequest) => apiFetch<ProjectionResponse>('/api/tools/projection', { method: 'POST', body }) })
}

export function useAllocationMutation() {
  return useMutation({ mutationFn: (body: AllocationRequest) => apiFetch<Allocation>('/api/tools/allocation', { method: 'POST', body }) })
}
