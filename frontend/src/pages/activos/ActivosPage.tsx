import { Badge, Button, Loader, Modal, Paper, ScrollArea, Stack, Table, Text } from '@mantine/core'
import { notifications } from '@mantine/notifications'
import { useState } from 'react'
import { Link, useNavigate } from 'react-router'

import { useAssets } from '@/api/assets'
import { useCreateAsset, useDeleteAsset, useUpdateAsset } from '@/api/assetMutations'
import type { Asset } from '@/api/types'
import { EmptyState } from '@/components/EmptyState'
import { ErrorAlert } from '@/components/ErrorAlert'
import { MoneyText } from '@/components/MoneyText'
import { PageHeader } from '@/components/PageHeader'
import { openConfirmDelete } from '@/components/ConfirmDelete'
import { useAssetTypeLabel } from '@/api/meta'
import { CreateAssetForm, EditAssetForm } from '@/features/assets/forms'
import { RowActions } from '@/features/shared/RowActions'
import { formatPeriod } from '@/features/shared/PeriodCell'

export function ActivosPage() {
  const assets = useAssets()
  const typeLabel = useAssetTypeLabel()
  const navigate = useNavigate()
  const [creating, setCreating] = useState(false)
  const [editing, setEditing] = useState<Asset | null>(null)
  const createAsset = useCreateAsset()
  const updateAsset = useUpdateAsset()
  const deleteAsset = useDeleteAsset()

  const rows = assets.data ?? []

  return (
    <Stack>
      <PageHeader
        title="Activos"
        description="Xestiona os activos da carteira."
        actions={<Button onClick={() => setCreating(true)}>Novo activo</Button>}
      />

      {assets.isLoading ? <Loader aria-label="Cargando activos" /> : null}
      {assets.error ? <ErrorAlert error={assets.error} /> : null}
      {!assets.isLoading && !assets.error && rows.length === 0 ? (
        <EmptyState
          title="Aínda non hai activos. Engade o primeiro."
          action={<Button onClick={() => setCreating(true)}>Novo activo</Button>}
        />
      ) : null}

      {rows.length > 0 ? (
        <Paper withBorder radius="md">
          <ScrollArea>
            <Table highlightOnHover verticalSpacing="sm" miw={720}>
              <Table.Thead>
                <Table.Tr>
                  <Table.Th>Tipo</Table.Th>
                  <Table.Th>Nome</Table.Th>
                  <Table.Th>Data inicial</Table.Th>
                  <Table.Th>Importe inicial</Table.Th>
                  <Table.Th ta="right">Accións</Table.Th>
                </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                {rows.map((asset) => (
                  <Table.Tr key={asset.id} onClick={() => navigate(`/activos/${asset.id}`)} style={{ cursor: 'pointer' }}>
                    <Table.Td><Badge variant="light">{typeLabel(asset.type)}</Badge></Table.Td>
                    <Table.Td>
                      <Text component={Link} to={`/activos/${asset.id}`} fw={600} c="teal.7" td="none" onClick={(event) => event.stopPropagation()}>
                        {asset.name}
                      </Text>
                    </Table.Td>
                    <Table.Td>{formatPeriod(asset.year, asset.month)}</Table.Td>
                    <Table.Td><MoneyText amount={asset.amountUsd} /></Table.Td>
                    <Table.Td onClick={(event) => event.stopPropagation()}>
                      <RowActions
                        onEdit={() => setEditing(asset)}
                        onDelete={() =>
                          openConfirmDelete({
                            title: `Eliminar ${asset.name}`,
                            message: 'Eliminarase o activo e tamén todas as súas transaccións e resultados mensuais.',
                            onConfirm: async () => {
                              await deleteAsset.mutateAsync(asset.id)
                              notifications.show({ color: 'green', message: 'Activo eliminado.' })
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

      <Modal opened={creating} onClose={() => setCreating(false)} title="Novo activo" transitionProps={{ duration: 0 }}>
        {createAsset.error ? <ErrorAlert error={createAsset.error} /> : null}
        <CreateAssetForm
          submitting={createAsset.isPending}
          error={createAsset.error}
          onSubmit={async (values) => {
            await createAsset.mutateAsync(values)
            notifications.show({ color: 'green', message: 'Activo creado.' })
            setCreating(false)
            createAsset.reset()
          }}
        />
      </Modal>

      <Modal opened={editing !== null} onClose={() => setEditing(null)} title="Editar activo" transitionProps={{ duration: 0 }}>
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
                notifications.show({ color: 'green', message: 'Activo actualizado.' })
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
