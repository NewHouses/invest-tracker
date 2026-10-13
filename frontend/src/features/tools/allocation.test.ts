import { describe, expect, it } from 'vitest'

import type { Asset } from '@/api/types'
import { buildAllocationRequest, buildGroups, equalPercents, sumsTo100 } from '@/features/tools/allocation'

const asset = (id: number, type: Asset['type'], name: string): Asset => ({ id, type, name, amountUsd: 100, year: 2026, month: 1 })
const acme = asset(1, 'accion', 'ACME')
const beta = asset(2, 'accion', 'BETA')
const fund = asset(3, 'fondo', 'FUND')

describe('reparto do aporte', () => {
  it('agrupa por tipo na orde de sempre, cos activos novos no seu tipo', () => {
    const groups = buildGroups([fund, acme], [{ key: 'n1', type: 'indice', name: 'Novo' }, { key: 'n2', type: 'accion', name: 'NVDA' }])
    expect(groups.map((group) => group.type)).toEqual(['accion', 'indice', 'fondo'])
    expect(groups[0]?.assets).toEqual([acme])
    expect(groups[0]?.newAssets.map((item) => item.name)).toEqual(['NVDA'])
  })

  it('propón partes iguais por tipo e por activo, con dúas decimais', () => {
    const groups = buildGroups([acme, beta, fund], [{ key: 'n1', type: 'indice', name: 'Novo' }])
    expect(equalPercents(groups)).toEqual({ a1: 16.67, a2: 16.67, n1: 33.33, a3: 33.33 })
  })

  it('admite o que se perde ao redondear cada porcentaxe, pero non máis', () => {
    expect(sumsTo100(99.99, 3)).toBe(true)
    expect(sumsTo100(100.02, 6)).toBe(true)
    expect(sumsTo100(99.98, 2)).toBe(false)
    expect(sumsTo100(105, 3)).toBe(false)
  })

  it('envía as porcentaxes só se se cambiaron', () => {
    const groups = buildGroups([acme], [{ key: 'n1', type: 'accion', name: 'NVDA' }])
    expect(buildAllocationRequest(100, groups, null)).toEqual({
      total: 100,
      selection: [{ type: 'accion', assetIds: [1], newAssets: [{ name: 'NVDA' }] }],
    })
    expect(buildAllocationRequest(100, groups, { a1: 70, n1: 30 })).toEqual({
      total: 100,
      selection: [{ type: 'accion', assetIds: [1], percents: [70], newAssets: [{ name: 'NVDA', percent: 30 }] }],
    })
  })
})
