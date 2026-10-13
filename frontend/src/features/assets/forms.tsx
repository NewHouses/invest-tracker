import { Button, Group, NumberInput, Stack, TextInput } from '@mantine/core'
import { useForm } from '@mantine/form'
import { useEffect } from 'react'

import type { CreateAssetInput, UpdateAssetInput } from '@/api/assetMutations'
import type { Asset, AssetType, YearMonth } from '@/api/types'
import { AssetTypeSelect } from '@/components/AssetTypeSelect'
import { YearMonthInput } from '@/components/YearMonthInput'
import { fieldErrorMap, parseNumber } from '@/features/shared/formErrors'
import { t } from '@/i18n'
import { numberInputSeparators } from '@/lib/format'
import { currentYearMonth } from '@/lib/yearMonth'

type AssetFormValues = {
  type: AssetType | null
  name: string
  amountUsd: number | string
  period: YearMonth | null
}

const monthAliases = { month: 'period', year: 'period' }

type CreateAssetFormProps = {
  submitting?: boolean
  error?: unknown
  onSubmit: (values: CreateAssetInput) => Promise<void> | void
}

export function CreateAssetForm({ submitting, error, onSubmit }: CreateAssetFormProps) {
  const form = useForm<AssetFormValues>({
    initialValues: { type: null, name: '', amountUsd: '', period: currentYearMonth() },
    validate: {
      type: (value) => (value ? null : t('Escolle un tipo')),
      name: (value) => (value.trim() ? null : t('O nome non pode estar baleiro')),
      amountUsd: (value) => (parseNumber(value) > 0 ? null : t('O importe debe ser maior ca 0')),
      period: (value) => (value ? null : t('Escolle un mes')),
    },
  })

  // setErrors é estable; o obxecto form cambia en cada render, así que non
  // pode ser dependencia (provocaría un bucle de renders).
  const { setErrors } = form
  useEffect(() => {
    const errors = fieldErrorMap(error, monthAliases)
    if (errors) setErrors(errors)
  }, [error, setErrors])

  return (
    <form
      onSubmit={form.onSubmit((values) => {
        if (!values.type || !values.period) return
        void onSubmit({
          type: values.type,
          name: values.name.trim(),
          amountUsd: parseNumber(values.amountUsd),
          year: values.period.year,
          month: values.period.month,
        })
      })}
    >
      <Stack>
        <AssetTypeSelect label={t('Tipo')} required value={form.values.type} onChange={(value) => form.setFieldValue('type', value)} error={form.errors.type} />
        <TextInput label={t('Nome')} required maxLength={100} {...form.getInputProps('name')} />
        <NumberInput label={t('Importe inicial')} required min={0} {...numberInputSeparators()} {...form.getInputProps('amountUsd')} />
        <YearMonthInput label={t('Data inicial')} required value={form.values.period} onChange={(value) => form.setFieldValue('period', value)} error={form.errors.period} />
        <Group justify="flex-end">
          <Button type="submit" loading={submitting ?? false}>{t('Gardar activo')}</Button>
        </Group>
      </Stack>
    </form>
  )
}

type EditAssetFormProps = {
  asset: Asset
  typeLabel: string
  submitting?: boolean
  error?: unknown
  onSubmit: (values: UpdateAssetInput) => Promise<void> | void
}

export function EditAssetForm({ asset, typeLabel, submitting, error, onSubmit }: EditAssetFormProps) {
  const form = useForm<AssetFormValues>({
    initialValues: { type: asset.type, name: asset.name, amountUsd: asset.amountUsd, period: { year: asset.year, month: asset.month } },
    validate: {
      name: (value) => (value.trim() ? null : t('O nome non pode estar baleiro')),
      amountUsd: (value) => (parseNumber(value) > 0 ? null : t('O importe debe ser maior ca 0')),
      period: (value) => (value ? null : t('Escolle un mes')),
    },
  })

  const { setErrors } = form
  useEffect(() => {
    const errors = fieldErrorMap(error, monthAliases)
    if (errors) setErrors(errors)
  }, [error, setErrors])

  return (
    <form
      onSubmit={form.onSubmit((values) => {
        if (!values.period) return
        void onSubmit({
          id: asset.id,
          name: values.name.trim(),
          amountUsd: parseNumber(values.amountUsd),
          year: values.period.year,
          month: values.period.month,
        })
      })}
    >
      <Stack>
        <TextInput label={t('Tipo')} value={typeLabel} readOnly />
        <TextInput label={t('Nome')} required maxLength={100} {...form.getInputProps('name')} />
        <NumberInput label={t('Importe inicial')} required min={0} {...numberInputSeparators()} {...form.getInputProps('amountUsd')} />
        <YearMonthInput label={t('Data inicial')} required value={form.values.period} onChange={(value) => form.setFieldValue('period', value)} error={form.errors.period} />
        <Group justify="flex-end">
          <Button type="submit" loading={submitting ?? false}>{t('Gardar activo')}</Button>
        </Group>
      </Stack>
    </form>
  )
}
