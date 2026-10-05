import { Badge, Button, Group, Loader, Modal, Paper, Stack, Table, Tabs, Text } from '@mantine/core'
import { notifications } from '@mantine/notifications'
import { useState } from 'react'
import { Link, useParams } from 'react-router'

import { ApiError } from '@/api/client'
import { useAsset } from '@/api/assets'
import { useAssetTypeLabel } from '@/api/meta'
import {
  useAssetTransactions,
  useCreateTransaction,
  useDeleteTransaction,
  useUpdateTransaction,
  type AssetTransactionRow,
  type TransactionKind,
} from '@/api/transactions'
import { useAssetResults, useDeleteResult } from '@/api/results'
import { EmptyState } from '@/components/EmptyState'
import { ErrorAlert } from '@/components/ErrorAlert'
import { MoneyText } from '@/components/MoneyText'
import { PageHeader } from '@/components/PageHeader'
import { openConfirmDelete } from '@/components/ConfirmDelete'
import { AssetChartPanel } from '@/features/charts/AssetChartPanel'
import { AssetHistoryPanel } from '@/features/reports/AssetHistoryPanel'
import { TransactionForm } from '@/features/transactions/TransactionForm'
import { RowActions } from '@/features/shared/RowActions'
import { formatPeriod } from '@/features/shared/PeriodCell'
import { formatYearMonth } from '@/lib/format'

function txKind(row: AssetTransactionRow) {
  if (row.isInitial) return 'Compra inicial'
  return row.isVenda ? 'Venda' : 'Compra'
}

export function ActivoDetailPage() {
  const params = useParams()
  const id = Number(params.id)
  const assetId = Number.isFinite(id) && id > 0 ? id : null
  const asset = useAsset(assetId)
  const typeLabel = useAssetTypeLabel()
  const transactions = useAssetTransactions(assetId)
  const results = useAssetResults(assetId)
  const createTransaction = useCreateTransaction()
  const updateTransaction = useUpdateTransaction()
  const deleteTransaction = useDeleteTransaction()
  const deleteResult = useDeleteResult()
  const [creatingTx, setCreatingTx] = useState(false)
  const [editingTx, setEditingTx] = useState<AssetTransactionRow | null>(null)

  if (asset.isLoading) return <Loader aria-label="Cargando activo" />
  if ((asset.error instanceof ApiError && asset.error.status === 404) || (!asset.isLoading && !asset.data)) {
    return <EmptyState title="Activo non atopado" action={<Button component={Link} to="/activos">Volver aos activos</Button>} />
  }
  if (asset.error) return <ErrorAlert error={asset.error} />

  const current = asset.data
  if (!current || assetId === null) return null

  return (
    <Stack>
      <PageHeader
        title={current.name}
        description={`${formatPeriod(current.year, current.month)} · ${typeLabel(current.type)}`}
        actions={<Button component={Link} to="/activos" variant="light">Volver</Button>}
      />
      <Paper withBorder p="md" radius="md">
        <Group gap="md">
          <Badge variant="light">{typeLabel(current.type)}</Badge>
          <Text>Data inicial: {formatPeriod(current.year, current.month)}</Text>
          <Text>Importe inicial: <MoneyText amount={current.amountUsd} /></Text>
        </Group>
      </Paper>

      <Tabs defaultValue="historico">
        <Tabs.List>
          <Tabs.Tab value="historico">Histórico</Tabs.Tab>
          <Tabs.Tab value="transaccions">Transaccións</Tabs.Tab>
          <Tabs.Tab value="resultados">Resultados</Tabs.Tab>
          <Tabs.Tab value="grafica">Gráfica</Tabs.Tab>
        </Tabs.List>

        <Tabs.Panel value="historico" pt="md"><AssetHistoryPanel assetId={current.id} /></Tabs.Panel>
        <Tabs.Panel value="grafica" pt="md"><AssetChartPanel assetId={current.id} /></Tabs.Panel>

        <Tabs.Panel value="transaccions" pt="md">
          <Stack>
            <Group justify="space-between"><Text fw={600}>Transaccións do activo</Text><Button onClick={() => setCreatingTx(true)}>Nova transacción</Button></Group>
            {transactions.isLoading ? <Loader aria-label="Cargando transaccións" /> : null}
            {transactions.error ? <ErrorAlert error={transactions.error} /> : null}
            {transactions.data ? (
              <Paper withBorder radius="md">
                <Table verticalSpacing="sm">
                  <Table.Thead><Table.Tr><Table.Th>Data</Table.Th><Table.Th>Tipo</Table.Th><Table.Th>Importe</Table.Th><Table.Th ta="right">Accións</Table.Th></Table.Tr></Table.Thead>
                  <Table.Tbody>
                    {transactions.data.rows.map((row) => (
                      <Table.Tr key={`${row.isInitial ? 'initial' : 'tx'}-${row.id}`}>
                        <Table.Td>{formatPeriod(row.year, row.month)}</Table.Td>
                        <Table.Td>{txKind(row)}</Table.Td>
                        <Table.Td><MoneyText amount={row.amount} colorByGain={row.isVenda} signed={row.isVenda} /></Table.Td>
                        <Table.Td>
                          {row.isInitial ? null : (
                            <RowActions
                              onEdit={() => setEditingTx(row)}
                              onDelete={() => openConfirmDelete({
                                title: 'Eliminar transacción',
                                message: `Eliminarase a transacción de ${formatPeriod(row.year, row.month)}.`,
                                onConfirm: async () => {
                                  await deleteTransaction.mutateAsync(row.id)
                                  notifications.show({ color: 'green', message: 'Transacción eliminada.' })
                                },
                              })}
                            />
                          )}
                        </Table.Td>
                      </Table.Tr>
                    ))}
                  </Table.Tbody>
                  <Table.Tfoot>
                    <Table.Tr><Table.Th>Total compra</Table.Th><Table.Th /><Table.Th><MoneyText amount={transactions.data.totals.compra} /></Table.Th><Table.Th /></Table.Tr>
                    <Table.Tr><Table.Th>Total venda</Table.Th><Table.Th /><Table.Th><MoneyText amount={transactions.data.totals.venda} /></Table.Th><Table.Th /></Table.Tr>
                    <Table.Tr><Table.Th>Neto</Table.Th><Table.Th /><Table.Th><MoneyText amount={transactions.data.totals.neto} /></Table.Th><Table.Th /></Table.Tr>
                  </Table.Tfoot>
                </Table>
              </Paper>
            ) : null}
          </Stack>
        </Tabs.Panel>

        <Tabs.Panel value="resultados" pt="md">
          <Stack>
            {results.isLoading ? <Loader aria-label="Cargando resultados" /> : null}
            {results.error ? <ErrorAlert error={results.error} /> : null}
            {results.data && results.data.length === 0 ? <EmptyState title="Aínda non hai resultados rexistrados para este activo." /> : null}
            {results.data && results.data.length > 0 ? (
              <Paper withBorder radius="md">
                <Table verticalSpacing="sm">
                  <Table.Thead><Table.Tr><Table.Th>Data</Table.Th><Table.Th>Resultado</Table.Th><Table.Th ta="right">Accións</Table.Th></Table.Tr></Table.Thead>
                  <Table.Tbody>
                    {results.data.map((result) => (
                      <Table.Tr key={result.id}>
                        <Table.Td>{formatYearMonth(result)}</Table.Td>
                        <Table.Td><MoneyText amount={result.resultUsd} /></Table.Td>
                        <Table.Td>
                          <RowActions
                            onDelete={() => openConfirmDelete({
                              title: 'Eliminar resultado',
                              message: `Eliminarase o resultado de ${formatPeriod(result.year, result.month)}.`,
                              onConfirm: async () => {
                                await deleteResult.mutateAsync(result.id)
                                notifications.show({ color: 'green', message: 'Resultado eliminado.' })
                              },
                            })}
                          />
                        </Table.Td>
                      </Table.Tr>
                    ))}
                  </Table.Tbody>
                </Table>
              </Paper>
            ) : null}
          </Stack>
        </Tabs.Panel>
      </Tabs>

      <Modal opened={creatingTx} onClose={() => setCreatingTx(false)} title="Nova transacción" transitionProps={{ duration: 0 }}>
        {createTransaction.error ? <ErrorAlert error={createTransaction.error} /> : null}
        <TransactionForm
          assetId={current.id}
          submitting={createTransaction.isPending}
          error={createTransaction.error}
          onSubmit={async (values) => {
            if (!('assetId' in values)) return
            await createTransaction.mutateAsync(values)
            notifications.show({ color: 'green', message: 'Transacción gardada.' })
            setCreatingTx(false)
            createTransaction.reset()
          }}
        />
      </Modal>

      <Modal opened={editingTx !== null} onClose={() => setEditingTx(null)} title="Editar transacción" transitionProps={{ duration: 0 }}>
        {editingTx ? (
          <>
            {updateTransaction.error ? <ErrorAlert error={updateTransaction.error} /> : null}
            <TransactionForm
              assetId={current.id}
              initial={{ id: editingTx.id, kind: (editingTx.isVenda ? 'venda' : 'compra') as TransactionKind, amountUsd: editingTx.amount, period: { year: editingTx.year, month: editingTx.month } }}
              submitting={updateTransaction.isPending}
              error={updateTransaction.error}
              onSubmit={async (values) => {
                await updateTransaction.mutateAsync(values as Parameters<typeof updateTransaction.mutateAsync>[0])
                notifications.show({ color: 'green', message: 'Transacción actualizada.' })
                setEditingTx(null)
                updateTransaction.reset()
              }}
            />
          </>
        ) : null}
      </Modal>
    </Stack>
  )
}
