import { Button, Group, NumberInput, SegmentedControl, Stack } from '@mantine/core'
import { useForm } from '@mantine/form'
import { useEffect } from 'react'

import type { TransactionInput, TransactionKind, TransactionUpdateInput } from '@/api/transactions'
import type { YearMonth } from '@/api/types'
import { AssetSelect } from '@/components/AssetSelect'
import { YearMonthInput } from '@/components/YearMonthInput'
import { fieldErrorMap, parseNumber } from '@/features/shared/formErrors'
import { currentYearMonth } from '@/lib/yearMonth'

type Values = {
  assetId: number | null
  kind: TransactionKind
  amountUsd: number | string
  period: YearMonth | null
}

type TransactionFormProps = {
  assetId?: number
  initial?: { id?: number; kind: TransactionKind; amountUsd: number; period: YearMonth }
  submitLabel?: string
  submitting?: boolean
  error?: unknown
  onSubmit: (values: TransactionInput | TransactionUpdateInput) => Promise<void> | void
}

export function TransactionForm({ assetId, initial, submitLabel = 'Gardar transacción', submitting, error, onSubmit }: TransactionFormProps) {
  const form = useForm<Values>({
    initialValues: {
      assetId: assetId ?? null,
      kind: initial?.kind ?? 'compra',
      amountUsd: initial?.amountUsd ?? '',
      period: initial?.period ?? currentYearMonth(),
    },
    validate: {
      assetId: (value) => (value ? null : 'Escolle un activo'),
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
        if (!values.assetId || !values.period) return
        const body = {
          assetId: values.assetId,
          kind: values.kind,
          amountUsd: parseNumber(values.amountUsd),
          year: values.period.year,
          month: values.period.month,
        }
        void Promise.resolve(onSubmit(initial?.id ? { id: initial.id, ...body } : body)).catch(() => undefined)
      })}
    >
      <Stack>
        <AssetSelect label="Activo" required disabled={assetId !== undefined} value={form.values.assetId} onChange={(value) => form.setFieldValue('assetId', value)} error={form.errors.assetId} />
        <SegmentedControl
          aria-label="Tipo de transacción"
          data={[{ label: 'Compra', value: 'compra' }, { label: 'Venda', value: 'venda' }]}
          {...form.getInputProps('kind')}
        />
        <NumberInput label="Importe" required min={0} decimalSeparator="," thousandSeparator="." {...form.getInputProps('amountUsd')} />
        <YearMonthInput label="Mes" required value={form.values.period} onChange={(value) => form.setFieldValue('period', value)} error={form.errors.period} />
        <Group justify="flex-end">
          <Button type="submit" loading={submitting ?? false}>{submitLabel}</Button>
        </Group>
      </Stack>
    </form>
  )
}
