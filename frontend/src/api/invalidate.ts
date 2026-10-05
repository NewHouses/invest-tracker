import type { QueryClient } from '@tanstack/react-query'

import { queryKeys } from '@/api/queryKeys'

// Calquera escritura (activos, transaccións, resultados ou dividendos) cambia
// informes, gráficas e inicio, así que se invalida todo o que depende da carteira.
export function invalidatePortfolio(queryClient: QueryClient) {
  const prefixes = [
    queryKeys.assets.all,
    queryKeys.transactions.all,
    queryKeys.results.all,
    queryKeys.dividends.all,
    queryKeys.reports.all,
    queryKeys.charts.all,
    queryKeys.tools.all,
    queryKeys.portfolio,
    queryKeys.dashboard,
  ]
  return Promise.all(prefixes.map((queryKey) => queryClient.invalidateQueries({ queryKey })))
}
