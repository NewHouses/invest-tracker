import { Button, Group, Menu, NumberInput, Paper, Select, SegmentedControl, Stack, Text } from '@mantine/core'
import type { ReactNode } from 'react'

import type { YearMonth } from '@/api/types'
import { YearMonthInput } from '@/components/YearMonthInput'
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
  { value: 'once', label: 'Unha vez' },
  { value: 'months', label: 'Cada N meses' },
  { value: 'years', label: 'Cada N anos' },
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
          <Text fw={700}>{title}</Text>
          <Menu shadow="md">
            <Menu.Target><Button variant="light">Engadir exemplo</Button></Menu.Target>
            <Menu.Dropdown>
              <Menu.Item onClick={() => addRule(emptyRule(start, { kind: 'fixed', value: 10, frequencyUnit: 'years', frequencyCount: 1, from: quickFrom }))}>+10 $ cada ano</Menu.Item>
              <Menu.Item onClick={() => addRule(emptyRule(start, { kind: 'percent', value: 5, frequencyUnit: 'years', frequencyCount: 1, from: quickFrom }))}>+5 % cada ano</Menu.Item>
              <Menu.Item onClick={() => addRule(emptyRule(start, { kind: 'fixed', value: '', frequencyUnit: 'months', frequencyCount: 6, from: start ? addMonths(start, 6) : null }))}>+x cada 6 meses</Menu.Item>
            </Menu.Dropdown>
          </Menu>
        </Group>
        {typeof errors.rules === 'string' ? <Text c="red" size="sm">{errors.rules}</Text> : null}
        {rules.length === 0 ? <Text c="dimmed" size="sm">Sen regras de evolución.</Text> : null}
        {rules.map((rule, index) => (
          <Paper withBorder p="sm" key={index}>
            <Stack gap="xs">
              <SegmentedControl
                aria-label={`Tipo da regra ${index + 1}`}
                data={[{ value: 'fixed', label: 'Cantidade fixa (USD)' }, { value: 'percent', label: 'Porcentaxe (%)' }]}
                value={rule.kind}
                onChange={(value) => onChange(updateRule(rules, index, { kind: value as GrowthRuleFormValue['kind'] }))}
              />
              <Group grow align="flex-start">
                <NumberInput
                  label="Valor"
                  decimalSeparator=","
                  thousandSeparator="."
                  value={rule.value}
                  onChange={(value) => onChange(updateRule(rules, index, { value }))}
                  error={errors[`rules.${index}.value`]}
                />
                <Select
                  label="Frecuencia"
                  data={frequencyOptions}
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
                  label="Desde"
                  value={rule.from}
                  onChange={(value) => onChange(updateRule(rules, index, { from: value }))}
                  error={errors[`rules.${index}.from`]}
                />
              </Group>
              <Group justify="space-between">
                <Text size="sm" c="dimmed">{describeRule(rule)}</Text>
                <Button variant="subtle" color="red" onClick={() => removeRule(index)}>Eliminar regra</Button>
              </Group>
            </Stack>
          </Paper>
        ))}
        <Group><Button variant="light" onClick={() => addRule()}>Engadir regra</Button></Group>
      </Stack>
    </Paper>
  )
}
