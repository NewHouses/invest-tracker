import { ActionIcon, Button, Group, Loader, NumberInput, Paper, SegmentedControl, Stack, Table, Tabs, Text } from '@mantine/core'
import { useForm } from '@mantine/form'
import { notifications } from '@mantine/notifications'
import { IconTrash } from '@tabler/icons-react'
import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router'

import { useCreateMonthTransactions, useCreateTransaction, useCreateTransactionBatch, useMonthAssets, type TransactionKind } from '@/api/transactions'
import type { YearMonth } from '@/api/types'
import { AssetSelect } from '@/components/AssetSelect'
import { EmptyState } from '@/components/EmptyState'
import { ErrorAlert } from '@/components/ErrorAlert'
import { PageHeader } from '@/components/PageHeader'
import { YearMonthInput } from '@/components/YearMonthInput'
import { AllocationPanel } from '@/features/tools/AllocationPanel'
import { TransactionForm } from '@/features/transactions/TransactionForm'
import { fieldErrorMap, parseNumber } from '@/features/shared/formErrors'
import { formatPeriod } from '@/features/shared/PeriodCell'
import { t } from '@/i18n'
import { formatUSD, formatYearMonth, numberInputSeparators } from '@/lib/format'
import { addMonths, currentYearMonth } from '@/lib/yearMonth'

type BatchMode = 'por-fila' | 'mes-unico' | 'unha-por-mes'
type BatchRow = { id: string; kind: TransactionKind; amountUsd: number | string; period: YearMonth | null }

type BatchValues = { assetId: number | null; mode: BatchMode; start: YearMonth | null; rows: BatchRow[] }

function rowId() {
  return crypto.randomUUID?.() ?? String(Date.now() + Math.random())
}

function newRow(period: YearMonth | null = currentYearMonth()): BatchRow {
  return { id: rowId(), kind: 'compra', amountUsd: '', period }
}

function tabValue(value: string | null) {
  return ['nova', 'serie', 'mes', 'repartir'].includes(value ?? '') ? value ?? 'nova' : 'nova'
}

function BatchForm() {
  const mutation = useCreateTransactionBatch()
  const form = useForm<BatchValues>({
    initialValues: { assetId: null, mode: 'por-fila', start: currentYearMonth(), rows: [newRow()] },
    validate: {
      assetId: (value) => (value ? null : t('Escolle un activo')),
      start: (value, values) => (values.mode === 'por-fila' || value ? null : t('Escolle un mes')),
      rows: {
        amountUsd: (value) => (parseNumber(value) > 0 ? null : t('O importe debe ser maior ca 0')),
        period: (value, values) => (values.mode !== 'por-fila' || value ? null : t('Escolle un mes')),
      },
    },
  })

  const { setErrors } = form
  useEffect(() => {
    const errors = fieldErrorMap(mutation.error, { assetId: 'assetId' })
    if (!errors) return
    const mapped: Record<string, string> = {}
    for (const [key, value] of Object.entries(errors)) {
      const periodMatch = /^items\[(\d+)]\.(month|year)$/.exec(key)
      const fieldMatch = /^items\[(\d+)]\.(amountUsd|kind)$/.exec(key)
      if (periodMatch?.[1]) mapped[`rows.${periodMatch[1]}.period`] = value
      else if (fieldMatch?.[1] && fieldMatch[2]) mapped[`rows.${fieldMatch[1]}.${fieldMatch[2]}`] = value
      else mapped[key] = value
    }
    setErrors(mapped)
  }, [mutation.error, setErrors])

  const mode = form.values.mode
  const start = form.values.start

  return (
    <form
      onSubmit={form.onSubmit(async (values) => {
        if (!values.assetId) return
        const items = values.rows.map((row, index) => {
          const period = values.mode === 'por-fila' ? row.period : values.mode === 'mes-unico' ? values.start : values.start ? addMonths(values.start, index) : null
          if (!period) throw new Error(t('Mes incompleto'))
          return { kind: row.kind, amountUsd: parseNumber(row.amountUsd), year: period.year, month: period.month }
        })
        await mutation.mutateAsync({ assetId: values.assetId, items })
        notifications.show({ color: 'green', message: t('{count} transacción(s) engadidas.', { count: items.length }) })
        form.reset()
      })}
    >
      <Stack>
        {mutation.error ? <ErrorAlert error={mutation.error} /> : null}
        <AssetSelect label={t('Activo')} required value={form.values.assetId} onChange={(value) => form.setFieldValue('assetId', value)} error={form.errors.assetId} />
        <SegmentedControl
          aria-label={t('Modo de serie')}
        data={[{ label: t('Mes en cada fila'), value: 'por-fila' }, { label: t('Todas no mesmo mes'), value: 'mes-unico' }, { label: t('Unha por mes'), value: 'unha-por-mes' }]}
          {...form.getInputProps('mode')}
        />
        {mode !== 'por-fila' ? <YearMonthInput label={t('Mes inicial')} required value={form.values.start} onChange={(value) => form.setFieldValue('start', value)} error={form.errors.start} /> : null}
        <Table>
        <Table.Thead><Table.Tr><Table.Th>{t('Tipo')}</Table.Th><Table.Th>{t('Importe')}</Table.Th><Table.Th>{t('Mes')}</Table.Th><Table.Th /></Table.Tr></Table.Thead>
          <Table.Tbody>
            {form.values.rows.map((row, index) => {
              const computed = mode === 'unha-por-mes' && start ? addMonths(start, index) : null
              return (
                <Table.Tr key={row.id}>
                  <Table.Td>
                    <SegmentedControl data={[{ label: t('Compra'), value: 'compra' }, { label: t('Venda'), value: 'venda' }]} {...form.getInputProps(`rows.${index}.kind`)} />
                  </Table.Td>
                  <Table.Td><NumberInput aria-label={t('Importe fila {n}', { n: index + 1 })} min={0} {...numberInputSeparators()} {...form.getInputProps(`rows.${index}.amountUsd`)} /></Table.Td>
                  <Table.Td>{mode === 'por-fila' ? <YearMonthInput value={row.period} onChange={(value) => form.setFieldValue(`rows.${index}.period`, value)} error={form.errors[`rows.${index}.period`]} /> : computed ? formatYearMonth(computed) : start ? formatYearMonth(start) : '—'}</Table.Td>
                  <Table.Td>
                    <ActionIcon aria-label={t('Eliminar fila')} color="red" variant="subtle" disabled={form.values.rows.length === 1} onClick={() => form.removeListItem('rows', index)}>
                      <IconTrash size={18} />
                    </ActionIcon>
                  </Table.Td>
                </Table.Tr>
              )
            })}
          </Table.Tbody>
        </Table>
        <Group justify="space-between">
          <Button variant="light" onClick={() => form.insertListItem('rows', newRow(start))}>{t('Engadir fila')}</Button>
          <Button type="submit" loading={mutation.isPending}>{t('Gardar serie')}</Button>
        </Group>
      </Stack>
    </form>
  )
}

function MonthTransactionsForm() {
  const [period, setPeriod] = useState<YearMonth>(currentYearMonth())
  const [amounts, setAmounts] = useState<Record<number, number | string>>({})
  const assets = useMonthAssets(period)
  const mutation = useCreateMonthTransactions()

  const eligible = assets.data?.eligible ?? []
  const omitted = assets.data?.omitted ?? []

  return (
    <Stack>
      <YearMonthInput label={t('Mes')} value={period} onChange={(value) => {
        if (!value) return
        setPeriod(value)
        setAmounts({})
      }} />
      {assets.isLoading ? <Loader aria-label={t('Cargando activos elixibles')} /> : null}
      {assets.error ? <ErrorAlert error={assets.error} /> : null}
      {eligible.length === 0 && !assets.isLoading ? <EmptyState title={t('Non hai activos creados en {period} ou antes.', { period: formatYearMonth(period) })} /> : null}
      {eligible.length > 0 ? (
        <form
          onSubmit={async (event) => {
            event.preventDefault()
            const items = eligible
              .map((asset) => ({ assetId: asset.id, amountUsd: parseNumber(amounts[asset.id]) }))
              .filter((item) => Number.isFinite(item.amountUsd) && item.amountUsd > 0)
            await mutation.mutateAsync({ ...period, items })
            const total = items.reduce((sum, item) => sum + item.amountUsd, 0)
            notifications.show({ color: 'green', message: t('{count} transacción(s) engadidas ({total} total)', { count: items.length, total: formatUSD(total) }) })
            setAmounts({})
          }}
        >
          <Stack>
            {mutation.error ? <ErrorAlert error={mutation.error} /> : null}
            <Paper withBorder radius="md">
              <Table>
                <Table.Thead><Table.Tr><Table.Th>{t('Activo')}</Table.Th><Table.Th>{t('Importe')}</Table.Th></Table.Tr></Table.Thead>
                <Table.Tbody>
                  {eligible.map((asset) => (
                    <Table.Tr key={asset.id}>
                      <Table.Td>{asset.name}</Table.Td>
                      <Table.Td><NumberInput aria-label={t('Importe {name}', { name: asset.name })} min={0} {...numberInputSeparators()} value={amounts[asset.id] ?? ''} onChange={(value) => setAmounts((current) => ({ ...current, [asset.id]: value }))} /></Table.Td>
                    </Table.Tr>
                  ))}
                </Table.Tbody>
              </Table>
            </Paper>
            <Button type="submit" loading={mutation.isPending}>{t('Gardar transaccións do mes')}</Button>
          </Stack>
        </form>
      ) : null}
      {omitted.length > 0 ? (
        <Stack gap={4}>
          <Text fw={600}>{t('Activos omitidos')}</Text>
          {omitted.map((asset) => <Text key={asset.id} c="dimmed">{t('{name}: omitido (creado en {period})', { name: asset.name, period: formatPeriod(asset.year, asset.month) })}</Text>)}
        </Stack>
      ) : null}
    </Stack>
  )
}

export function TransaccionsPage() {
  const [params, setParams] = useSearchParams()
  const active = tabValue(params.get('tab'))
  const create = useCreateTransaction()

  return (
    <Stack>
      <PageHeader title={t('Transaccións')} description={t('Rexistra compras, vendas e aportes mensuais.')} />
      <Tabs value={active} onChange={(value) => setParams(value && value !== 'nova' ? { tab: value } : {})}>
        <Tabs.List>
          <Tabs.Tab value="nova">{t('Nova')}</Tabs.Tab>
          <Tabs.Tab value="serie">{t('En serie')}</Tabs.Tab>
          <Tabs.Tab value="mes">{t('Do mes')}</Tabs.Tab>
          <Tabs.Tab value="repartir">{t('Repartir aporte')}</Tabs.Tab>
        </Tabs.List>
        <Tabs.Panel value="nova" pt="md">
          <Paper withBorder p="md" radius="md">
            {create.error ? <ErrorAlert error={create.error} /> : null}
            <TransactionForm
              submitting={create.isPending}
              error={create.error}
              onSubmit={async (values) => {
                if (!('assetId' in values)) return
                await create.mutateAsync(values)
                notifications.show({ color: 'green', message: t('Transacción gardada.') })
                create.reset()
              }}
            />
          </Paper>
        </Tabs.Panel>
        <Tabs.Panel value="serie" pt="md"><Paper withBorder p="md" radius="md"><BatchForm /></Paper></Tabs.Panel>
        <Tabs.Panel value="mes" pt="md"><MonthTransactionsForm /></Tabs.Panel>
        <Tabs.Panel value="repartir" pt="md"><AllocationPanel /></Tabs.Panel>
      </Tabs>
    </Stack>
  )
}
