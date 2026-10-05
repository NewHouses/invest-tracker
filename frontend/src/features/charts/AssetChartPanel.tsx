import { ErrorAlert } from '@/components/ErrorAlert'
import { useAssetChart } from '@/api/charts'

import { LineChartPanel } from './common'

type AssetChartPanelProps = {
  assetId: number
}

export function AssetChartPanel({ assetId }: AssetChartPanelProps) {
  const chart = useAssetChart(assetId)
  if (chart.isError) return <ErrorAlert error={chart.error} title="Non se puido cargar a gráfica" />
  return <LineChartPanel chart={chart.data} loading={chart.isLoading} emptyNoResults="Aínda non hai resultados rexistrados para este activo." />
}
