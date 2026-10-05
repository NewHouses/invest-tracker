import { describe, expect, it } from 'vitest'

import type { LineChartDTO } from '@/api/charts'
import { mapLineChartDTO } from '@/features/charts/lineChartMapper'

describe('mapLineChartDTO', () => {
  it('mantén nulls e etiqueta meses como MM/AAAA', () => {
    const dto: LineChartDTO = { title: 't', hasAssets: true, months: [{ year: 2026, month: 1 }, { year: 2026, month: 2 }], series: [{ label: 'ACME', values: [10, null] }] }
    expect(mapLineChartDTO(dto)).toEqual({ data: [{ month: '01/2026', s0: 10 }, { month: '02/2026', s0: null }], series: [{ name: 's0', label: 'ACME', color: 'blue.6' }] })
  })

  // Regresión: Mantine corta as claves polos puntos, e unha etiqueta como
  // "Resultado + dividendos acum." deixaba a lenda baleira.
  it('usa claves seguras aínda que a etiqueta leve puntos', () => {
    const dto: LineChartDTO = { title: 't', hasAssets: true, months: [{ year: 2026, month: 1 }], series: [{ label: 'Resultado + dividendos acum.', values: [5] }, { label: 'Iberdrola S.A.', values: [7] }] }
    const mapped = mapLineChartDTO(dto)
    expect(mapped.series.map((s) => s.name)).toEqual(['s0', 's1'])
    expect(mapped.series.map((s) => s.label)).toEqual(['Resultado + dividendos acum.', 'Iberdrola S.A.'])
    expect(mapped.data[0]).toEqual({ month: '01/2026', s0: 5, s1: 7 })
  })
})
