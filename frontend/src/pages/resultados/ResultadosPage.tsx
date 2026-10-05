import { Button, Group, Loader, NumberInput, Paper, Stack, Table, Tabs, Text } from '@mantine/core'
import { useForm } from '@mantine/form'
import { notifications } from '@mantine/notifications'
import { useEffect, useMemo, useState } from 'react'
import { useSearchParams } from 'react-router'

import { useCloseMonthResults, useCreateResult, useDeleteResultsByMonth, useEligibleResults } from '@/api/results'
import type { YearMonth } from '@/api/types'
import { AssetSelect } from '@/components/AssetSelect'
import { EmptyState } from '@/components/EmptyState'
import { ErrorAlert } from '@/components/ErrorAlert'
import { GainBadge } from '@/components/GainBadge'
import { MoneyText } from '@/components/MoneyText'
import { PageHeader } from '@/components/PageHeader'
import { YearMonthInput } from '@/components/YearMonthInput'
import { openConfirmDelete } from '@/components/ConfirmDelete'
import { fieldErrorMap, parseNumber } from '@/features/shared/formErrors'
import { formatPct, formatYearMonth } from '@/lib/format'
import { currentYearMonth } from '@/lib/yearMonth'

function tabValue(value: string | null) {
  return ['pechar', 'engadir', 'limpar'].includes(value ?? '') ? value ?? 'pechar' : 'pechar'
}

function gainPct(result: number, holding: number) {
  return holding > 0 ? ((result - holding) / holding) * 100 : 0
}

function CloseMonthForm() {
  const [period, setPeriod] = useState<YearMonth>(currentYearMonth())
  const [values, setValues] = useState<Record<number, number | string>>({})
  const eligible = useEligibleResults(period)
  const closeMonth = useCloseMonthResults()
  const items = eligible.data?.items ?? []

  return (
    <Stack>
      <YearMonthInput label="Mes" value={period} onChange={(value) => {
        if (!value) return
        setPeriod(value)
        setValues({})
      }} />
      {eligible.isLoading ? <Loader aria-label="Cargando ativos elixibles" /> : null}
      {eligible.error ? <ErrorAlert error={eligible.error} /> : null}
      {items.length === 0 && !eligible.isLoading ? <EmptyState title={`Non hai ativos con capital investido para ${formatYearMonth(period)}.`} /> : null}
      {items.length > 0 ? (
        <form
          onSubmit={async (event) => {
            event.preventDefault()
            const bodyItems = items
              .map((item) => ({ assetId: item.asset.id, resultUsd: parseNumber(values[item.asset.id]) }))
              .filter((item) => Number.isFinite(item.resultUsd) && item.resultUsd > 0)
            await closeMonth.mutateAsync({ ...period, items: bodyItems })
            notifications.show({ color: 'green', message: `Pechouse ${formatYearMonth(period)}: ${bodyItems.length} resultado(s) gardado(s)` })
            setValues({})
          }}
        >
          <Stack>
            {closeMonth.error ? <ErrorAlert error={closeMonth.error} /> : null}
            <Paper withBorder radius="md">
              <Table verticalSpacing="sm">
                <Table.Thead><Table.Tr><Table.Th>Ativo</Table.Th><Table.Th>No ativo</Table.Th><Table.Th>Resultado</Table.Th><Table.Th>G/P previsto</Table.Th></Table.Tr></Table.Thead>
                <Table.Tbody>
                  {items.map((item) => {
                    const result = parseNumber(values[item.asset.id])
                    const hasPreview = Number.isFinite(result) && result > 0
                    const gain = hasPreview ? result - item.holding : 0
                    return (
                      <Table.Tr key={item.asset.id}>
                        <Table.Td>{item.asset.name}</Table.Td>
                        <Table.Td><MoneyText amount={item.holding} /></Table.Td>
                        <Table.Td>
                          <NumberInput
                            aria-label={`Resultado ${item.asset.name}`}
                            min={0}
                            decimalSeparator=","
                            thousandSeparator="."
                            value={values[item.asset.id] ?? ''}
                            onChange={(value) => setValues((current) => ({ ...current, [item.asset.id]: value }))}
                            placeholder={item.hasResult ? `Actual: ${item.result}` : undefined}
                          />
                          {item.hasResult ? <Text size="xs" c="dimmed">Actual: <MoneyText amount={item.result} size="xs" /></Text> : null}
                        </Table.Td>
                        <Table.Td>{hasPreview ? <Group gap="xs"><GainBadge value={gain} kind="amount" /><Text size="sm">{formatPct(gainPct(result, item.holding))}</Text></Group> : '—'}</Table.Td>
                      </Table.Tr>
                    )
                  })}
                </Table.Tbody>
              </Table>
            </Paper>
            <Button type="submit" loading={closeMonth.isPending}>Pechar mes</Button>
          </Stack>
        </form>
      ) : null}
    </Stack>
  )
}

type ResultFormValues = { assetId: number | null; resultUsd: number | string }

function AddResultForm() {
  const [period, setPeriod] = useState<YearMonth>(currentYearMonth())
  const eligible = useEligibleResults(period)
  const create = useCreateResult()
  const assets = useMemo(() => (eligible.data?.items ?? []).map((item) => item.asset), [eligible.data?.items])
  const form = useForm<ResultFormValues>({
    initialValues: { assetId: null, resultUsd: '' },
    validate: {
      assetId: (value) => (value ? null : 'Escolle un ativo'),
      resultUsd: (value) => (parseNumber(value) > 0 ? null : 'O resultado debe ser maior ca 0'),
    },
  })

  // Ao cambiar de mes cambian os ativos elixibles: a selección anterior
  // deixa de ser válida.
  const changePeriod = (value: YearMonth | null) => {
    if (!value) return
    setPeriod(value)
    form.setFieldValue('assetId', null)
  }

  const { setErrors } = form
  useEffect(() => {
    const errors = fieldErrorMap(create.error, { month: 'period', year: 'period' })
    if (errors) setErrors(errors)
  }, [create.error, setErrors])

  const selected = (eligible.data?.items ?? []).find((item) => item.asset.id === form.values.assetId)
  const preview = selected && parseNumber(form.values.resultUsd) > 0 ? parseNumber(form.values.resultUsd) - selected.holding : null

  return (
    <form
      onSubmit={form.onSubmit(async (values) => {
        if (!values.assetId) return
        await create.mutateAsync({ assetId: values.assetId, resultUsd: parseNumber(values.resultUsd), year: period.year, month: period.month })
        notifications.show({ color: 'green', message: 'Resultado gardado.' })
        form.setFieldValue('resultUsd', '')
      })}
    >
      <Stack>
        {eligible.error ? <ErrorAlert error={eligible.error} /> : null}
        {create.error ? <ErrorAlert error={create.error} /> : null}
        <YearMonthInput label="Mes" value={period} onChange={changePeriod} error={form.errors.period} />
        {eligible.isLoading ? <Loader aria-label="Cargando ativos elixibles" /> : null}
        <AssetSelect label="Ativo" required assets={assets} value={form.values.assetId} onChange={(value) => form.setFieldValue('assetId', value)} error={form.errors.assetId} />
        <NumberInput label="Resultado" required min={0} decimalSeparator="," thousandSeparator="." {...form.getInputProps('resultUsd')} />
        {preview !== null && selected ? <Text>G/P resultante: <GainBadge value={preview} kind="amount" /> ({formatPct(gainPct(parseNumber(form.values.resultUsd), selected.holding))})</Text> : null}
        <Button type="submit" loading={create.isPending}>Engadir resultado</Button>
      </Stack>
    </form>
  )
}

function ClearMonthForm() {
  const [period, setPeriod] = useState<YearMonth>(currentYearMonth())
  const mutation = useDeleteResultsByMonth()

  return (
    <Stack align="flex-start">
      {mutation.error ? <ErrorAlert error={mutation.error} /> : null}
      <YearMonthInput label="Mes" value={period} onChange={(value) => value && setPeriod(value)} />
      <Button
        color="red"
        loading={mutation.isPending}
        onClick={() => openConfirmDelete({
          title: `Limpar ${formatYearMonth(period)}`,
          message: `Eliminaranse todos os resultados de ${formatYearMonth(period)}.`,
          onConfirm: async () => {
            const result = await mutation.mutateAsync(period)
            notifications.show({
              color: result.deleted > 0 ? 'green' : 'blue',
              message: result.deleted > 0 ? `${result.deleted} resultado(s) eliminados.` : `Non había resultados rexistrados en ${formatYearMonth(period)}.`,
            })
          },
        })}
      >
        Limpar mes
      </Button>
    </Stack>
  )
}

export function ResultadosPage() {
  const [params, setParams] = useSearchParams()
  const active = tabValue(params.get('tab'))

  return (
    <Stack>
      <PageHeader title="Resultados" description="Pecha meses e mantén os valores actuais dos ativos." />
      <Tabs value={active} onChange={(value) => setParams(value && value !== 'pechar' ? { tab: value } : {})}>
        <Tabs.List>
          <Tabs.Tab value="pechar">Pechar mes</Tabs.Tab>
          <Tabs.Tab value="engadir">Engadir resultado</Tabs.Tab>
          <Tabs.Tab value="limpar">Limpar mes</Tabs.Tab>
        </Tabs.List>
        <Tabs.Panel value="pechar" pt="md"><CloseMonthForm /></Tabs.Panel>
        <Tabs.Panel value="engadir" pt="md"><Paper withBorder p="md" radius="md"><AddResultForm /></Paper></Tabs.Panel>
        <Tabs.Panel value="limpar" pt="md"><Paper withBorder p="md" radius="md"><ClearMonthForm /></Paper></Tabs.Panel>
      </Tabs>
    </Stack>
  )
}
