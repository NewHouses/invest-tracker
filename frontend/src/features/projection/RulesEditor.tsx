import { Button, Group, Menu, NumberInput, Paper, Select, SegmentedControl, Stack, Text } from '@mantine/core'
import type { ReactNode } from 'react'

import type { YearMonth } from '@/api/types'
import { YearMonthInput } from '@/components/YearMonthInput'
import { t, tKey } from '@/i18n'
import { numberInputSeparators } from '@/lib/format'
import { addMonths } from '@/lib/yearMonth'

import { describeRule, emptyRule } from './rules'
import type { GrowthRuleFormValue, RuleFrequencyUnit } from './rules'

type RulesEditorProps = {
  title: string
  rules: GrowthRuleFormValue[]
  start: YearMonth | null
  errors?: Record<string, ReactNode>
  onChange: (rules: GrowthRuleFormValue[]) => void
}

const frequencyOptions = [
  { value: 'once', label: tKey('Unha vez') },
  { value: 'months', label: tKey('Cada N meses') },
  { value: 'years', label: tKey('Cada N anos') },
]

function updateRule(rules: GrowthRuleFormValue[], index: number, patch: Partial<GrowthRuleFormValue>) {
  return rules.map((rule, i) => (i === index ? { ...rule, ...patch } : rule))
}

export function RulesEditor({ title, rules, start, errors = {}, onChange }: RulesEditorProps) {
  const addRule = (rule: GrowthRuleFormValue = emptyRule(start)) => onChange([...rules, rule])
  const removeRule = (index: number) => onChange(rules.filter((_, i) => i !== index))
  const quickFrom = start ? addMonths(start, 12) : null

  return (
    <Paper withBorder p="md">
      <Stack>
        <Group justify="space-between" align="center">
          <Text fw={700}>{t(title)}</Text>
          <Menu shadow="md">
            <Menu.Target><Button variant="light">{t('Engadir exemplo')}</Button></Menu.Target>
            <Menu.Dropdown>
              <Menu.Item onClick={() => addRule(emptyRule(start, { kind: 'fixed', value: 10, frequencyUnit: 'years', frequencyCount: 1, from: quickFrom }))}>{t('+10 $ cada ano')}</Menu.Item>
              <Menu.Item onClick={() => addRule(emptyRule(start, { kind: 'percent', value: 5, frequencyUnit: 'years', frequencyCount: 1, from: quickFrom }))}>{t('+5 % cada ano')}</Menu.Item>
              <Menu.Item onClick={() => addRule(emptyRule(start, { kind: 'fixed', value: '', frequencyUnit: 'months', frequencyCount: 6, from: start ? addMonths(start, 6) : null }))}>{t('+x cada 6 meses')}</Menu.Item>
            </Menu.Dropdown>
          </Menu>
        </Group>
        {typeof errors.rules === 'string' ? <Text c="red" size="sm">{errors.rules}</Text> : null}
        {rules.length === 0 ? <Text c="dimmed" size="sm">{t('Sen regras de evolución.')}</Text> : null}
        {rules.map((rule, index) => (
          <Paper withBorder p="sm" key={index}>
            <Stack gap="xs">
              <SegmentedControl
                aria-label={t('Tipo da regra {number}', { number: index + 1 })}
                data={[{ value: 'fixed', label: t('Cantidade fixa (USD)') }, { value: 'percent', label: t('Porcentaxe (%)') }]}
                value={rule.kind}
                onChange={(value) => onChange(updateRule(rules, index, { kind: value as GrowthRuleFormValue['kind'] }))}
              />
              <Group grow align="flex-start">
                <NumberInput
                  label={t('Valor')}
                  {...numberInputSeparators()}
                  value={rule.value}
                  onChange={(value) => onChange(updateRule(rules, index, { value }))}
                  error={errors[`rules.${index}.value`]}
                />
                <Select
                  label={t('Frecuencia')}
                  data={frequencyOptions.map((option) => ({ ...option, label: t(option.label) }))}
                  value={rule.frequencyUnit}
                  allowDeselect={false}
                  onChange={(value) => onChange(updateRule(rules, index, { frequencyUnit: (value ?? 'once') as RuleFrequencyUnit }))}
                  error={errors[`rules.${index}.everyMonths`]}
                />
                {rule.frequencyUnit !== 'once' ? (
                  <NumberInput
                    label="N"
                    min={1}
                    step={1}
                    value={rule.frequencyCount}
                    onChange={(value) => onChange(updateRule(rules, index, { frequencyCount: value }))}
                    error={errors[`rules.${index}.everyMonths`]}
                  />
                ) : null}
                <YearMonthInput
                  label={t('Desde')}
                  value={rule.from}
                  onChange={(value) => onChange(updateRule(rules, index, { from: value }))}
                  error={errors[`rules.${index}.from`]}
                />
              </Group>
              <Group justify="space-between">
                <Text size="sm" c="dimmed">{describeRule(rule)}</Text>
                <Button variant="subtle" color="red" onClick={() => removeRule(index)}>{t('Eliminar regra')}</Button>
              </Group>
            </Stack>
          </Paper>
        ))}
        <Group><Button variant="light" onClick={() => addRule()}>{t('Engadir regra')}</Button></Group>
      </Stack>
    </Paper>
  )
}
