import { ActionIcon, Alert, Badge, Button, Checkbox, Group, NumberInput, Paper, Stack, Text, TextInput, Title } from '@mantine/core'
import { IconPlus, IconTrash } from '@tabler/icons-react'
import { useMemo, useRef, useState } from 'react'

import { ApiError } from '@/api/client'
import { useAssets } from '@/api/assets'
import { useAssetTypeLabel } from '@/api/meta'
import { useAllocationMutation } from '@/api/tools'
import type { AssetType } from '@/api/types'
import { AssetTypeSelect } from '@/components/AssetTypeSelect'
import { ErrorAlert } from '@/components/ErrorAlert'
import { parseNumber } from '@/features/shared/formErrors'
import {
  assetTypeOrder,
  buildAllocationRequest,
  buildGroups,
  equalPercents,
  existingKey,
  groupKeys,
  sumsTo100,
  validPercent,
} from '@/features/tools/allocation'
import type { AllocationGroup, NewAsset } from '@/features/tools/allocation'
import { t } from '@/i18n'
import { formatPercent, formatUSD, numberInputSeparators } from '@/lib/format'

type NumberValue = number | string

// Porcentaxes que cambiou o usuario, ligadas á selección na que se
// escribiron: se cambia a selección, vólvese á proposta a partes iguais.
type EditedPercents = { signature: string; values: Record<string, NumberValue> }

const fieldKeyRe = /\.(percents\[\d+\]|newAssets\[\d+\]\.(percent|name))$/

function fieldErrors(error: unknown): Record<string, string> {
  return error instanceof ApiError ? (error.fields ?? {}) : {}
}

export function AllocationPanel() {
  const assets = useAssets()
  const typeLabel = useAssetTypeLabel()
  const mutation = useAllocationMutation()
  const [total, setTotal] = useState<NumberValue>('')
  const [selectedTypes, setSelectedTypes] = useState<AssetType[]>([])
  const [selectedAssets, setSelectedAssets] = useState<Record<AssetType, number[]>>({ accion: [], indice: [], copy_trading: [], fondo: [] })
  const [newAssets, setNewAssets] = useState<NewAsset[]>([])
  const [edited, setEdited] = useState<EditedPercents | null>(null)
  const [submittedGroups, setSubmittedGroups] = useState<AllocationGroup[]>([])
  const nextKey = useRef(0)

  const byType = useMemo(
    () => assetTypeOrder.map((type) => ({ type, assets: (assets.data ?? []).filter((a) => a.type === type) })).filter((entry) => entry.assets.length > 0),
    [assets.data],
  )
  const selected = byType.flatMap((entry) => (selectedTypes.includes(entry.type) ? entry.assets.filter((a) => selectedAssets[entry.type].includes(a.id)) : []))
  const groups = buildGroups(selected, newAssets)
  const keys = groups.flatMap(groupKeys)
  const signature = groups.map((group) => `${group.type}:${groupKeys(group).join(',')}`).join('|')
  const custom = edited?.signature === signature ? edited.values : null
  const values: Record<string, NumberValue> = custom ?? equalPercents(groups)
  const numbers = keys.map((key) => parseNumber(values[key]))
  const sum = numbers.reduce((acc, value) => acc + value, 0)
  const allValid = numbers.every(validPercent)
  const percentsValid = keys.length > 0 && allValid && sumsTo100(sum, keys.length)
  const totalNumber = parseNumber(total)

  // Os erros do servidor refírense á petición enviada: pásanse á clave de
  // cada fila para que sigan no seu sitio aínda que cambie a selección.
  const errors = fieldErrors(mutation.error)
  const percentErrors: Record<string, string> = {}
  const nameErrors: Record<string, string> = {}
  submittedGroups.forEach((group, i) => {
    group.assets.forEach((asset, j) => {
      const message = errors[`selection[${i}].percents[${j}]`]
      if (message) percentErrors[existingKey(asset.id)] = message
    })
    group.newAssets.forEach((item, k) => {
      const percent = errors[`selection[${i}].newAssets[${k}].percent`]
      const name = errors[`selection[${i}].newAssets[${k}].name`]
      if (percent) percentErrors[item.key] = percent
      if (name) nameErrors[item.key] = name
    })
  })
  const generalErrors = [...new Set(Object.entries(errors).filter(([key]) => key !== 'total' && !fieldKeyRe.test(key)).map(([, message]) => message))]

  const toggleType = (type: AssetType, checked: boolean) => {
    setSelectedTypes((prev) => (checked ? [...prev, type] : prev.filter((item) => item !== type)))
    if (checked && selectedAssets[type].length === 0) {
      setSelectedAssets((prev) => ({ ...prev, [type]: byType.find((entry) => entry.type === type)?.assets.map((a) => a.id) ?? [] }))
    }
  }
  const toggleAsset = (type: AssetType, id: number, checked: boolean) =>
    setSelectedAssets((prev) => ({ ...prev, [type]: checked ? [...prev[type], id] : prev[type].filter((assetId) => assetId !== id) }))
  const setPercent = (key: string, value: NumberValue) => setEdited({ signature, values: { ...values, [key]: value } })
  const addNewAsset = () => {
    nextKey.current += 1
    const key = `n${nextKey.current}`
    setNewAssets((prev) => [...prev, { key, type: prev.at(-1)?.type ?? 'accion', name: '' }])
  }
  const updateNewAsset = (key: string, change: Partial<Omit<NewAsset, 'key'>>) =>
    setNewAssets((prev) => prev.map((item) => (item.key === key ? { ...item, ...change } : item)))
  const removeNewAsset = (key: string) => setNewAssets((prev) => prev.filter((item) => item.key !== key))
  const submit = () => {
    const percents = custom ? Object.fromEntries(keys.map((key, i) => [key, numbers[i] ?? Number.NaN])) : null
    setSubmittedGroups(groups)
    mutation.mutate(buildAllocationRequest(totalNumber, groups, percents))
  }

  if (assets.isError) return <ErrorAlert error={assets.error} />

  const percentField = (key: string, name: string) => {
    const percent = parseNumber(values[key])
    const amount = Number.isFinite(totalNumber) && totalNumber > 0 && validPercent(percent) ? (totalNumber * percent) / (percentsValid ? sum : 100) : null
    return (
      <Group gap="xs" wrap="nowrap">
        <NumberInput
          aria-label={t('Porcentaxe de {name}', { name })}
          w={110}
          min={0}
          max={100}
          decimalScale={2}
          rightSection="%"
          value={values[key] ?? ''}
          onChange={(value) => setPercent(key, value)}
          error={percentErrors[key]}
          {...numberInputSeparators()}
        />
        <Text size="sm" c="dimmed" w={110} ta="right">
          {amount === null ? '' : `≈ ${formatUSD(amount)}`}
        </Text>
      </Group>
    )
  }

  const status = keys.length === 0
    ? { color: 'dimmed', text: t('Escolle os activos entre os que repartir o aporte.') }
    : !allValid
      ? { color: 'red', text: t('Indica unha porcentaxe entre 0 e 100 para cada activo.') }
      : percentsValid
        ? { color: 'green', text: t('Suma: {sum}', { sum: formatPercent(100) }) }
        : { color: 'red', text: t('As porcentaxes suman {sum}; teñen que sumar 100 %.', { sum: formatPercent(sum) }) }

  return (
    <Stack>
      <Paper withBorder p="md">
        <Stack>
          <NumberInput label={t('Cantidade total (USD)')} min={0} value={total} onChange={setTotal} error={errors.total} {...numberInputSeparators()} />
          {!assets.isLoading && byType.length === 0 ? (
            <Text c="dimmed">{t('Aínda non tes activos creados: engade abaixo os activos novos entre os que repartir.')}</Text>
          ) : null}
          {byType.map((entry) => {
            const checked = selectedTypes.includes(entry.type)
            return (
              <Paper withBorder p="sm" key={entry.type}>
                <Stack gap="xs">
                  <Checkbox label={typeLabel(entry.type)} checked={checked} onChange={(e) => toggleType(entry.type, e.currentTarget.checked)} />
                  {checked ? (
                    <Stack pl="md" gap="xs">
                      {entry.assets.map((asset) => {
                        const isSelected = selectedAssets[entry.type].includes(asset.id)
                        return (
                          <Group key={asset.id} justify="space-between" wrap="nowrap" mih={36}>
                            <Checkbox label={asset.name} checked={isSelected} onChange={(e) => toggleAsset(entry.type, asset.id, e.currentTarget.checked)} />
                            {isSelected ? percentField(existingKey(asset.id), asset.name) : null}
                          </Group>
                        )
                      })}
                      {selectedAssets[entry.type].length === 0 ? <Text size="sm" c="dimmed">{t('Escolle polo menos un activo deste tipo.')}</Text> : null}
                    </Stack>
                  ) : null}
                </Stack>
              </Paper>
            )
          })}
          <Paper withBorder p="sm">
            <Stack gap="xs">
              <Group justify="space-between">
                <Stack gap={0}>
                  <Text fw={600}>{t('Activos novos')}</Text>
                  <Text size="sm" c="dimmed">{t('Aínda non están creados: só serven para calcular o reparto da primeira vez.')}</Text>
                </Stack>
                <Button variant="light" leftSection={<IconPlus size={16} />} onClick={addNewAsset}>
                  {t('Engadir activo novo')}
                </Button>
              </Group>
              {newAssets.map((item) => (
                <Group key={item.key} align="flex-start" justify="space-between">
                  <Group align="flex-start" gap="xs" style={{ flex: 1 }}>
                    <AssetTypeSelect
                      aria-label={t('Tipo do activo novo')}
                      w={160}
                      allowDeselect={false}
                      value={item.type}
                      onChange={(type) => {
                        if (type) updateNewAsset(item.key, { type })
                      }}
                    />
                    <TextInput
                      aria-label={t('Nome do activo novo')}
                      placeholder={t('Nome')}
                      value={item.name}
                      onChange={(e) => updateNewAsset(item.key, { name: e.currentTarget.value })}
                      error={nameErrors[item.key]}
                      style={{ flex: 1, minWidth: 140 }}
                    />
                  </Group>
                  <Group gap="xs" wrap="nowrap">
                    {percentField(item.key, item.name.trim() || t('activo novo'))}
                    <ActionIcon variant="subtle" color="red" aria-label={t('Quitar activo novo')} onClick={() => removeNewAsset(item.key)} mt={4}>
                      <IconTrash size={18} />
                    </ActionIcon>
                  </Group>
                </Group>
              ))}
            </Stack>
          </Paper>
          <Group justify="space-between">
            <Text c={status.color} fw={500}>{status.text}</Text>
            <Button variant="subtle" onClick={() => setEdited(null)} disabled={!custom}>
              {t('Repartir a partes iguais')}
            </Button>
          </Group>
          {generalErrors.length > 0 ? <Alert color="red">{generalErrors.join(' ')}</Alert> : null}
          <Group>
            <Button onClick={submit} loading={mutation.isPending} disabled={!percentsValid}>
              {t('Repartir aporte')}
            </Button>
          </Group>
        </Stack>
      </Paper>
      {mutation.isError ? <ErrorAlert error={mutation.error} /> : null}
      {mutation.data ? (
        <Paper withBorder p="md">
          <Stack>
            <Title order={3}>{t('Reparto de aporte mensual: {amount}', { amount: formatUSD(mutation.data.total) })}</Title>
            {mutation.data.types.map((typeResult) => (
              <Stack key={typeResult.type} gap={2}>
                <Group justify="space-between">
                  <Text fw={700}>{typeResult.label}</Text>
                  <Group gap="md">
                    <Text c="dimmed">{formatPercent(typeResult.percent)}</Text>
                    <Text fw={700}>{formatUSD(typeResult.amount)}</Text>
                  </Group>
                </Group>
                {typeResult.assets.map((assetResult, index) => (
                  <Group key={assetResult.new ? `novo-${index}` : assetResult.asset.id} justify="space-between" pl="md">
                    <Group gap="xs">
                      <Text>{assetResult.asset.name}</Text>
                      {assetResult.new ? (
                        <Badge size="sm" variant="light">
                          {t('novo')}
                        </Badge>
                      ) : null}
                    </Group>
                    <Group gap="md">
                      <Text c="dimmed">{formatPercent(assetResult.percent)}</Text>
                      <Text>{formatUSD(assetResult.amount)}</Text>
                    </Group>
                  </Group>
                ))}
              </Stack>
            ))}
          </Stack>
        </Paper>
      ) : null}
    </Stack>
  )
}