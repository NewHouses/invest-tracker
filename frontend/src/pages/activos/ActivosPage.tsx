import { Button, Loader, Modal, Paper, ScrollArea, Stack, Table, Text } from '@mantine/core'
import { notifications } from '@mantine/notifications'
import { useMemo, useState } from 'react'
import { Link, useNavigate } from 'react-router'

import { useAssets } from '@/api/assets'
import { useCreateAsset, useDeleteAsset, useUpdateAsset } from '@/api/assetMutations'
import { useAssetTypeLabel } from '@/api/meta'
import { usePortfolio } from '@/api/portfolio'
import type { Asset } from '@/api/types'
import { AssetTypeBadge } from '@/components/AssetTypeBadge'
import { EmptyState } from '@/components/EmptyState'
import { ErrorAlert } from '@/components/ErrorAlert'
import { MoneyText } from '@/components/MoneyText'
import { PageHeader } from '@/components/PageHeader'
import { openConfirmDelete } from '@/components/ConfirmDelete'
import { CreateAssetForm, EditAssetForm } from '@/features/assets/forms'
import { SortableTh } from '@/features/portfolio/SortableTh'
import { sortRows, type PortfolioSortKey, type SortDirection } from '@/features/portfolio/sortRows'
import { RowActions } from '@/features/shared/RowActions'
import { formatPeriod } from '@/features/shared/PeriodCell'
import { formatPct, formatSignedUSD, formatUSD } from '@/lib/format'
import { gainColor } from '@/lib/gain'

function ValueCell({ value, available, pct = false, signed = false }: { value: number; available: boolean; pct?: boolean; signed?: boolean }) {
  if (!available) return <Text component="span" c="dimmed">—</Text>
  const text = pct ? formatPct(value) : signed ? formatSignedUSD(value) : formatUSD(value)
  const colorProps = pct || signed ? { c: gainColor(value) } : {}
  return <Text component="span" {...colorProps}>{text}</Text>
}

function nextDirection(currentKey: PortfolioSortKey | null, currentDirection: SortDirection, key: PortfolioSortKey): SortDirection {
  if (currentKey !== key) return 'asc'
  if (currentDirection === 'asc') return 'desc'
  if (currentDirection === 'desc') return null
  return 'asc'
}

export function ActivosPage() {
  const assets = useAssets()
  const portfolio = usePortfolio()
  const typeLabel = useAssetTypeLabel()
  const navigate = useNavigate()
  const [creating, setCreating] = useState(false)
  const [editing, setEditing] = useState<Asset | null>(null)
  const [sortKey, setSortKey] = useState<PortfolioSortKey | null>(null)
  const [sortDirection, setSortDirection] = useState<SortDirection>(null)
  const createAsset = useCreateAsset()
  const updateAsset = useUpdateAsset()
  const deleteAsset = useDeleteAsset()

  const rows = useMemo(() => {
    const portfolioById = new Map((portfolio.data?.rows ?? []).map((row) => [row.asset.id, row]))
    return sortRows((assets.data ?? []).map((asset, index) => ({ asset, originalIndex: index, typeLabel: typeLabel(asset.type), portfolio: portfolioById.get(asset.id) })), sortKey, sortDirection)
  }, [assets.data, portfolio.data?.rows, sortDirection, sortKey, typeLabel])

  const handleSort = (key: PortfolioSortKey) => {
    const direction = nextDirection(sortKey, sortDirection, key)
    setSortKey(direction === null ? null : key)
    setSortDirection(direction)
  }

  return (
    <Stack>
      <PageHeader
        title="Portofolio"
        description="Ativos da carteira, valor e rendemento."
        actions={<Button onClick={() => setCreating(true)}>Novo ativo</Button>}
      />

      {assets.isLoading ? <Loader aria-label="Cargando ativos" /> : null}
      {assets.error ? <ErrorAlert error={assets.error} /> : null}
      {portfolio.error ? <ErrorAlert error={portfolio.error} title="Non se puido cargar o Portofolio" /> : null}
      {!assets.isLoading && !assets.error && rows.length === 0 ? (
        <EmptyState
          title="Aínda non hai ativos. Engade o primeiro."
          action={<Button onClick={() => setCreating(true)}>Novo ativo</Button>}
        />
      ) : null}

      {rows.length > 0 ? (
        <Paper withBorder radius="md">
          <ScrollArea>
            <Table highlightOnHover verticalSpacing="sm" miw={980}>
              <Table.Thead>
                <Table.Tr>
                  <SortableTh sortKey="type" activeKey={sortKey} direction={sortDirection} onSort={handleSort}>Tipo</SortableTh>
                  <SortableTh sortKey="name" activeKey={sortKey} direction={sortDirection} onSort={handleSort}>Nome</SortableTh>
                  <SortableTh sortKey="period" activeKey={sortKey} direction={sortDirection} onSort={handleSort}>Data inicial</SortableTh>
                  <SortableTh sortKey="initialAmount" activeKey={sortKey} direction={sortDirection} onSort={handleSort}>Importe inicial</SortableTh>
                  <SortableTh sortKey="totalInvested" activeKey={sortKey} direction={sortDirection} onSort={handleSort}>Aportado</SortableTh>
                  <SortableTh sortKey="currentValue" activeKey={sortKey} direction={sortDirection} onSort={handleSort}>Valor neto</SortableTh>
                  <SortableTh sortKey="gain" activeKey={sortKey} direction={sortDirection} onSort={handleSort}>G/P</SortableTh>
                  <SortableTh sortKey="gainPct" activeKey={sortKey} direction={sortDirection} onSort={handleSort}>%</SortableTh>
                  <Table.Th ta="right">Accións</Table.Th>
                </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                {rows.map(({ asset, portfolio: row }) => (
                  <Table.Tr key={asset.id} onClick={() => navigate(`/activos/${asset.id}`)} style={{ cursor: 'pointer' }}>
                    <Table.Td><AssetTypeBadge type={asset.type} /></Table.Td>
                    <Table.Td>
                      <Text component={Link} to={`/activos/${asset.id}`} fw={600} c="teal.7" td="none" onClick={(event) => event.stopPropagation()}>
                        {asset.name}
                      </Text>
                    </Table.Td>
                    <Table.Td>{formatPeriod(asset.year, asset.month)}</Table.Td>
                    <Table.Td><MoneyText amount={asset.amountUsd} /></Table.Td>
                    <Table.Td>{row ? formatUSD(row.totalInvested) : '—'}</Table.Td>
                    <Table.Td><ValueCell value={row?.currentValue ?? 0} available={row?.hasCurrentValue ?? false} /></Table.Td>
                    <Table.Td><ValueCell value={row?.gain ?? 0} available={row?.hasGain ?? false} signed /></Table.Td>
                    <Table.Td><ValueCell value={row?.gainPct ?? 0} available={row?.hasGainPct ?? false} pct /></Table.Td>
                    <Table.Td onClick={(event) => event.stopPropagation()}>
                      <RowActions
                        onEdit={() => setEditing(asset)}
                        onDelete={() =>
                          openConfirmDelete({
                            title: `Eliminar ${asset.name}`,
                            message: 'Eliminarase o ativo e tamén todas as súas transaccións e resultados mensuais.',
                            onConfirm: async () => {
                              await deleteAsset.mutateAsync(asset.id)
                              notifications.show({ color: 'green', message: 'Ativo eliminado.' })
                            },
                          })
                        }
                      />
                    </Table.Td>
                  </Table.Tr>
                ))}
              </Table.Tbody>
            </Table>
          </ScrollArea>
        </Paper>
      ) : null}

      <Modal opened={creating} onClose={() => setCreating(false)} title="Novo ativo" transitionProps={{ duration: 0 }}>
        {createAsset.error ? <ErrorAlert error={createAsset.error} /> : null}
        <CreateAssetForm
          submitting={createAsset.isPending}
          error={createAsset.error}
          onSubmit={async (values) => {
            await createAsset.mutateAsync(values)
            notifications.show({ color: 'green', message: 'Ativo creado.' })
            setCreating(false)
            createAsset.reset()
          }}
        />
      </Modal>

      <Modal opened={editing !== null} onClose={() => setEditing(null)} title="Editar ativo" transitionProps={{ duration: 0 }}>
        {editing ? (
          <>
            {updateAsset.error ? <ErrorAlert error={updateAsset.error} /> : null}
            <EditAssetForm
              key={editing.id}
              asset={editing}
              typeLabel={typeLabel(editing.type)}
              submitting={updateAsset.isPending}
              error={updateAsset.error}
              onSubmit={async (values) => {
                await updateAsset.mutateAsync(values)
                notifications.show({ color: 'green', message: 'Ativo actualizado.' })
                setEditing(null)
                updateAsset.reset()
              }}
            />
          </>
        ) : null}
      </Modal>
    </Stack>
  )
}
