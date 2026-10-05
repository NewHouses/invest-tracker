import type { LineChartDTO } from '@/api/charts'
import { formatYearMonth } from '@/lib/format'

export type MantineLineDatum = { month: string } & Record<string, number | null | string>
export type MantineLineSeries = { name: string; label: string; color: string }

const colors = ['blue.6', 'green.6', 'orange.6', 'grape.6', 'cyan.6', 'red.6', 'yellow.7', 'teal.6']

// As claves dos datos son sintéticas (s0, s1…): Mantine/Recharts interpretan
// os puntos dunha clave como rutas aniñadas, e etiquetas como
// "Resultado + dividendos acum." ou nomes de ativos con "S.A." romperían a
// lenda. A etiqueta visible vai en `label`.
export function mapLineChartDTO(dto: LineChartDTO): { data: MantineLineDatum[]; series: MantineLineSeries[] } {
  const keyed = dto.series.map((serie, index) => ({ ...serie, key: `s${index}` }))
  const data = dto.months.map((month, index) => {
    const row: MantineLineDatum = { month: formatYearMonth(month) }
    for (const serie of keyed) {
      row[serie.key] = serie.values[index] ?? null
    }
    return row
  })
  return {
    data,
    series: keyed.map((serie, index) => ({ name: serie.key, label: serie.label, color: colors[index % colors.length] ?? 'blue.6' })),
  }
}
