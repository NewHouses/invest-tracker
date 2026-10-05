import { Button, Group, Loader, Modal, NumberInput, Paper, Stack, Table } from '@mantine/core'
import { useForm } from '@mantine/form'
import { notifications } from '@mantine/notifications'
import { useEffect, useState } from 'react'

import { useCreateDividend, useDeleteDividend, useDividends } from '@/api/dividends'
import type { YearMonth } from '@/api/types'
import { EmptyState } from '@/components/EmptyState'
import { ErrorAlert } from '@/components/ErrorAlert'
import { MoneyText } from '@/components/MoneyText'
import { PageHeader } from '@/components/PageHeader'
import { YearMonthInput } from '@/components/YearMonthInput'
import { openConfirmDelete } from '@/components/ConfirmDelete'
import { RowActions } from '@/features/shared/RowActions'
import { fieldErrorMap, parseNumber } from '@/features/shared/formErrors'
import { formatPeriod } from '@/features/shared/PeriodCell'
import { currentYearMonth } from '@/lib/yearMonth'

type Values = { amountUsd: number | string; period: YearMonth | null }

function DividendForm({ submitting, error, onSubmit }: { submitting?: boolean; error?: unknown; onSubmit: (values: { amountUsd: number; year: number; month: number }) => Promise<void> | void }) {
  const form = useForm<Values>({
    initialValues: { amountUsd: '', period: currentYearMonth() },
    validate: {
      amountUsd: (value) => (parseNumber(value) > 0 ? null : 'O importe debe ser maior ca 0'),
      period: (value) => (value ? null : 'Escolle un mes'),
    },
  })

  const { setErrors } = form
  useEffect(() => {
    const errors = fieldErrorMap(error, { month: 'period', year: 'period' })
    if (errors) setErrors(errors)
  }, [error, setErrors])

  return (
    <form
      onSubmit={form.onSubmit((values) => {
        if (!values.period) return
        void onSubmit({ amountUsd: parseNumber(values.amountUsd), year: values.period.year, month: values.period.month })
      })}
    >
      <Stack>
        <NumberInput label="Importe" required min={0} decimalSeparator="," thousandSeparator="." {...form.getInputProps('amountUsd')} />
        <YearMonthInput label="Mes" required value={form.values.period} onChange={(value) => form.setFieldValue('period', value)} error={form.errors.period} />
        <Group justify="flex-end"><Button type="submit" loading={submitting ?? false}>Gardar dividendo</Button></Group>
      </Stack>
    </form>
  )
}

export function DividendosPage() {
  const dividends = useDividends()
  const create = useCreateDividend()
  const remove = useDeleteDividend()
  const [opened, setOpened] = useState(false)
  const rows = dividends.data ?? []
  const total = rows.reduce((sum, row) => sum + row.amountUsd, 0)

  return (
    <Stack>
      <PageHeader title="Dividendos" description="Rexistra os dividendos cobrados." actions={<Button onClick={() => setOpened(true)}>Novo dividendo</Button>} />
      {dividends.isLoading ? <Loader aria-label="Cargando dividendos" /> : null}
      {dividends.error ? <ErrorAlert error={dividends.error} /> : null}
      {!dividends.isLoading && rows.length === 0 ? <EmptyState title="Aínda non hai dividendos rexistrados." action={<Button onClick={() => setOpened(true)}>Novo dividendo</Button>} /> : null}
      {rows.length > 0 ? (
        <Paper withBorder radius="md">
          <Table verticalSpacing="sm">
            <Table.Thead><Table.Tr><Table.Th>Data</Table.Th><Table.Th>Importe</Table.Th><Table.Th ta="right">Accións</Table.Th></Table.Tr></Table.Thead>
            <Table.Tbody>
              {rows.map((dividend) => (
                <Table.Tr key={dividend.id}>
                  <Table.Td>{formatPeriod(dividend.year, dividend.month)}</Table.Td>
                  <Table.Td><MoneyText amount={dividend.amountUsd} /></Table.Td>
                  <Table.Td>
                    <RowActions
                      onDelete={() => openConfirmDelete({
                        title: 'Eliminar dividendo',
                        message: `Eliminarase o dividendo de ${formatPeriod(dividend.year, dividend.month)}.`,
                        onConfirm: async () => {
                          await remove.mutateAsync(dividend.id)
                          notifications.show({ color: 'green', message: 'Dividendo eliminado.' })
                        },
                      })}
                    />
                  </Table.Td>
                </Table.Tr>
              ))}
            </Table.Tbody>
            <Table.Tfoot><Table.Tr><Table.Th>Total</Table.Th><Table.Th><MoneyText amount={total} fw={700} /></Table.Th><Table.Th /></Table.Tr></Table.Tfoot>
          </Table>
        </Paper>
      ) : null}
      <Modal opened={opened} onClose={() => setOpened(false)} title="Novo dividendo" transitionProps={{ duration: 0 }}>
        {create.error ? <ErrorAlert error={create.error} /> : null}
        <DividendForm
          submitting={create.isPending}
          error={create.error}
          onSubmit={async (values) => {
            await create.mutateAsync(values)
            notifications.show({ color: 'green', message: 'Dividendo gardado.' })
            setOpened(false)
            create.reset()
          }}
        />
      </Modal>
    </Stack>
  )
}
