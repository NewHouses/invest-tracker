import { ErrorAlert } from '@/components/ErrorAlert'
import { useAssetChart } from '@/api/charts'
import { t, tKey } from '@/i18n'

import { LineChartPanel } from './common'

type AssetChartPanelProps = {
  assetId: number
}

export function AssetChartPanel({ assetId }: AssetChartPanelProps) {
  const chart = useAssetChart(assetId)
  if (chart.isError) return <ErrorAlert error={chart.error} title={t('Non se puido cargar a gráfica')} />
  return <LineChartPanel chart={chart.data} loading={chart.isLoading} title={tKey('Evolución dun activo')} emptyNoResults={tKey('Aínda non hai resultados rexistrados para este activo.')} />
}
