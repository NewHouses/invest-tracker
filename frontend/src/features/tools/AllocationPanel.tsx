import { Alert, Button, Checkbox, Group, NumberInput, Paper, Stack, Text, Title } from '@mantine/core'
import { useMemo, useState } from 'react'

import { ApiError } from '@/api/client'
import { useAllocationMutation } from '@/api/tools'
import { useAssets } from '@/api/assets'
import { useAssetTypeLabel } from '@/api/meta'
import type { AssetType } from '@/api/types'
import { EmptyState } from '@/components/EmptyState'
import { ErrorAlert } from '@/components/ErrorAlert'
import { formatUSD } from '@/lib/format'

const order: AssetType[] = ['accion', 'indice', 'copy_trading', 'fondo']

function field(error: unknown, key: string) { return error instanceof ApiError ? error.fields?.[key] : undefined }

export function AllocationPanel() {
  const assets = useAssets()
  const typeLabel = useAssetTypeLabel()
  const mutation = useAllocationMutation()
  const [total, setTotal] = useState<number | string>('')
  const [selectedTypes, setSelectedTypes] = useState<AssetType[]>([])
  const [selectedAssets, setSelectedAssets] = useState<Record<AssetType, number[]>>({ accion: [], indice: [], copy_trading: [], fondo: [] })
  const byType = useMemo(() => order.map((type) => ({ type, assets: (assets.data ?? []).filter((a) => a.type === type) })).filter((entry) => entry.assets.length > 0), [assets.data])
  const toggleType = (type: AssetType, checked: boolean) => {
    setSelectedTypes((prev) => checked ? [...prev, type] : prev.filter((t) => t !== type))
    if (checked && selectedAssets[type].length === 0) setSelectedAssets((prev) => ({ ...prev, [type]: byType.find((e) => e.type === type)?.assets.map((a) => a.id) ?? [] }))
  }
  const toggleAsset = (type: AssetType, id: number, checked: boolean) => setSelectedAssets((prev) => ({ ...prev, [type]: checked ? [...prev[type], id] : prev[type].filter((assetId) => assetId !== id) }))
  const submit = () => mutation.mutate({ total: typeof total === 'number' ? total : Number(total), selection: selectedTypes.map((type) => ({ type, assetIds: selectedAssets[type] })) })
  if (assets.isError) return <ErrorAlert error={assets.error} />
  if (!assets.isLoading && (assets.data ?? []).length === 0) return <EmptyState title="Aínda non hai activos" description="Engade activos antes de repartir aportes." />
  return (
    <Stack>
      <Paper withBorder p="md"><Stack>
        <NumberInput label="Cantidade total (USD)" min={0} value={total} onChange={setTotal} error={field(mutation.error, 'total')} />
        {field(mutation.error, 'selection') ? <Alert color="red">{field(mutation.error, 'selection')}</Alert> : null}
        {byType.map((entry, index) => <Paper withBorder p="sm" key={entry.type}><Stack gap="xs"><Checkbox label={typeLabel(entry.type)} checked={selectedTypes.includes(entry.type)} onChange={(e) => toggleType(entry.type, e.currentTarget.checked)} />{selectedTypes.includes(entry.type) ? <Stack pl="md">{field(mutation.error, `selection[${index}].assetIds`) ? <Alert color="red">{field(mutation.error, `selection[${index}].assetIds`)}</Alert> : null}{entry.assets.map((asset) => <Checkbox key={asset.id} label={asset.name} checked={selectedAssets[entry.type].includes(asset.id)} onChange={(e) => toggleAsset(entry.type, asset.id, e.currentTarget.checked)} />)}</Stack> : null}</Stack></Paper>)}
        <Group><Button onClick={submit} loading={mutation.isPending}>Repartir aporte</Button></Group>
      </Stack></Paper>
      {mutation.isError ? <ErrorAlert error={mutation.error} /> : null}
      {mutation.data ? <Paper withBorder p="md"><Stack><Title order={3}>Reparto de aporte mensual: {formatUSD(mutation.data.total)}</Title>{mutation.data.types.map((t) => <Stack key={t.type} gap={2}><Group justify="space-between"><Text fw={700}>{t.label}</Text><Text>{formatUSD(t.amount)}</Text></Group>{t.assets.map((a) => <Group key={a.asset.id} justify="space-between" pl="md"><Text>{a.asset.name}</Text><Text>{formatUSD(a.amount)}</Text></Group>)}</Stack>)}</Stack></Paper> : null}
    </Stack>
  )
}
