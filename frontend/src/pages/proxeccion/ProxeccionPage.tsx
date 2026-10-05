import { Button, Group, Loader, NumberInput, Paper, ScrollArea, SegmentedControl, Stack, Table, Text, Title } from '@mantine/core'
import { LineChart } from '@mantine/charts'
import { useForm } from '@mantine/form'
import { useEffect, useRef } from 'react'

import { ApiError } from '@/api/client'
import type { GrowthRule, ProjectionMode, ProjectionMonth } from '@/api/tools'
import { useProjectionDefaults, useProjectionMutation } from '@/api/tools'
import { ErrorAlert } from '@/components/ErrorAlert'
import { PageHeader } from '@/components/PageHeader'
import { YearMonthInput } from '@/components/YearMonthInput'
import { KvTable, RowTint } from '@/components/ReportTable'
import { RulesEditor } from '@/features/projection/RulesEditor'
import { buildProjectionRequest, defaultProjectionFormValues, ruleToForm } from '@/features/projection/rules'
import type { GrowthRuleFormValue } from '@/features/projection/rules'
import { parseNumber } from '@/features/shared/formErrors'
import { formatAxisUSD, formatSignedUSD, formatUSD, formatYearMonth } from '@/lib/format'
import { currentYearMonth } from '@/lib/yearMonth'

function projectionData(months: ProjectionMonth[]) {
  return months.map((m) => ({ month: formatYearMonth(m.date), 'Capital total': m.totalCapital, 'Investimento acumulado': m.totalInvested, 'Ganancias acumuladas': m.totalGains }))
}

function apiErrors(error: unknown) {
  if (!(error instanceof ApiError) || !error.fields) return {}
  return Object.fromEntries(Object.entries(error.fields).map(([key, value]) => [key.replace(/\[(\d+)\]/g, '.$1'), value]))
}

function rulesFromApi(rules: GrowthRule[]) {
  return rules.map(ruleToForm)
}

export function ProxeccionPage() {
  const defaults = useProjectionDefaults()
  const mutation = useProjectionMutation()
  const form = useForm({
    initialValues: defaultProjectionFormValues(currentYearMonth()),
    validate: {
      start: (value) => (value ? null : 'Escolle a data inicial'),
      years: (value) => (parseNumber(value) >= 1 && parseNumber(value) <= 60 ? null : 'Os anos deben estar entre 1 e 60'),
      initialInvestment: (value) => (parseNumber(value) >= 0 ? null : 'O investimento inicial debe ser unha cantidade non negativa'),
      monthlyReturnPct: (value) => (parseNumber(value) > -100 ? null : 'O retorno mensual debe ser maior ca -100 %'),
      monthlyContribution: (value, values) => (values.mode === 'contribution' && !(parseNumber(value) >= 0) ? 'O aporte mensual debe ser non negativo' : null),
      annualSalary: (value, values) => (values.mode === 'salary' && !(parseNumber(value) >= 0) ? 'O salario anual debe ser non negativo' : null),
      investmentRatePct: (value, values) => (values.mode === 'salary' && (parseNumber(value) <= 0 || parseNumber(value) > 100) ? 'A porcentaxe debe estar entre 0 e 100' : null),
    },
  })
  const defaultsApplied = useRef(false)
  const contributionRules = useRef<GrowthRuleFormValue[]>([])
  const salaryRules = useRef<GrowthRuleFormValue[]>([])
  const salaryRulesInitialized = useRef(false)

  const { setErrors, setValues } = form
  useEffect(() => {
    if (!defaults.data || defaultsApplied.current) return
    const defaultSalaryRules = rulesFromApi(defaults.data.salaryRules)
    setValues((values) => ({
      ...values,
      start: defaults.data.start,
      years: defaults.data.years,
      investmentRatePct: defaults.data.investmentRatePct,
    }))
    salaryRules.current = defaultSalaryRules
    defaultsApplied.current = true
  }, [defaults.data, setValues])

  useEffect(() => {
    setErrors(apiErrors(mutation.error))
  }, [mutation.error, setErrors])

  const setRules = (rules: GrowthRuleFormValue[]) => {
    form.setFieldValue('rules', rules)
    if (form.values.mode === 'salary') salaryRules.current = rules
    else contributionRules.current = rules
  }

  const setMode = (mode: ProjectionMode) => {
    if (mode === form.values.mode) return
    if (form.values.mode === 'salary') salaryRules.current = form.values.rules
    else contributionRules.current = form.values.rules

    let nextRules = contributionRules.current
    if (mode === 'salary') {
      nextRules = salaryRulesInitialized.current ? salaryRules.current : rulesFromApi(defaults.data?.salaryRules ?? [])
      salaryRules.current = nextRules
      salaryRulesInitialized.current = true
    }
    form.setValues({ mode, rules: nextRules })
  }

  const submit = form.onSubmit((values) => mutation.mutate(buildProjectionRequest(values)))
  const errors = { ...form.errors, ...apiErrors(mutation.error) }
  const result = mutation.data
  const salaryMode = form.values.mode === 'salary'

  return (
    <Stack>
      <PageHeader title="Proxeccións" description="Simula a evolución do teu investimento" />
      {defaults.isLoading ? <Loader aria-label="Cargando valores por defecto" /> : null}
      {defaults.isError ? <ErrorAlert error={defaults.error} /> : null}
      <Paper withBorder p="md">
        <form onSubmit={submit}>
          <Stack>
            <YearMonthInput label="Data inicial" value={form.values.start} onChange={(value) => form.setFieldValue('start', value)} error={form.errors.start} required />
            {defaults.data?.startFromAssets ? <Text size="sm" c="dimmed">(ativo máis antigo)</Text> : null}
            <NumberInput label="Anos a proxectar" min={1} max={60} step={1} {...form.getInputProps('years')} />
            <NumberInput label="Investimento inicial (USD)" min={0} decimalSeparator="," thousandSeparator="." {...form.getInputProps('initialInvestment')} />
            <NumberInput label="Retorno mensual esperado (%)" decimalSeparator="," thousandSeparator="." {...form.getInputProps('monthlyReturnPct')} />
            <SegmentedControl
              aria-label="Modo"
              data={[{ value: 'contribution', label: 'Aporte mensual' }, { value: 'salary', label: 'A partir do salario' }]}
              value={form.values.mode}
              onChange={(value) => setMode(value as ProjectionMode)}
            />
            {salaryMode ? (
              <>
                <NumberInput label="Salario anual neto inicial (USD)" min={0} decimalSeparator="," thousandSeparator="." {...form.getInputProps('annualSalary')} />
                <NumberInput label="Porcentaxe que se inviste (%)" min={0} max={100} decimalSeparator="," thousandSeparator="." {...form.getInputProps('investmentRatePct')} />
                <RulesEditor title="Evolución salarial" rules={form.values.rules} start={form.values.start} errors={errors} onChange={setRules} />
              </>
            ) : (
              <>
                <NumberInput label="Aporte mensual (USD)" min={0} decimalSeparator="," thousandSeparator="." {...form.getInputProps('monthlyContribution')} />
                <RulesEditor title="Evolución do aporte" rules={form.values.rules} start={form.values.start} errors={errors} onChange={setRules} />
              </>
            )}
            <Group><Button type="submit" loading={mutation.isPending}>Calcular proxección</Button></Group>
          </Stack>
        </form>
      </Paper>
      {mutation.isError ? <ErrorAlert error={mutation.error} title="Non se puido calcular" /> : null}
      {result ? <Stack>
        <Paper withBorder p="md"><Stack><Title order={3}>Resumo</Title><KvTable rows={[
          { label: 'Investimento inicial', value: formatUSD(result.summary.initialInvestment) },
          { label: 'Aporte mensual inicial', value: formatUSD(result.summary.firstContribution) },
          { label: 'Aporte mensual final', value: formatUSD(result.summary.finalContribution) },
          ...(result.input.mode === 'salary' ? [{ label: 'Salario anual final', value: formatUSD(result.summary.finalAnnualSalary) }] : []),
          { label: 'Total aportado', value: formatUSD(result.summary.totalContributions) },
          { label: 'Investimento total', value: formatUSD(result.summary.totalInvested) },
          { label: 'Ganancias acumuladas', value: formatSignedUSD(result.summary.totalGains) },
          { label: 'Capital final', value: formatUSD(result.summary.finalCapital) },
        ]} /></Stack></Paper>
        <Paper withBorder p="md"><LineChart h={320} data={projectionData(result.months)} dataKey="month" series={[{ name: 'Capital total', color: 'blue.6' }, { name: 'Investimento acumulado', color: 'green.6' }, { name: 'Ganancias acumuladas', color: 'orange.6' }]} valueFormatter={formatUSD} yAxisProps={{ tickFormatter: formatAxisUSD, width: 64 }} withDots={false} withLegend /></Paper>
        <ScrollArea><Table withTableBorder striped style={{ whiteSpace: 'nowrap' }}><Table.Thead><Table.Tr><Table.Th>Mes</Table.Th><Table.Th>Data</Table.Th>{result.input.mode === 'salary' ? <Table.Th>Salario anual</Table.Th> : null}<Table.Th>Aporte</Table.Th><Table.Th>Invest. acumulado</Table.Th><Table.Th>Rendemento</Table.Th><Table.Th>Ganancias acum.</Table.Th><Table.Th>Capital total</Table.Th></Table.Tr></Table.Thead><Table.Tbody>{result.months.map((m) => <RowTint key={m.index} gain={m.totalGains}><Table.Td>{m.index}</Table.Td><Table.Td>{formatYearMonth(m.date)}</Table.Td>{result.input.mode === 'salary' ? <Table.Td>{formatUSD(m.annualSalary)}</Table.Td> : null}<Table.Td>{formatUSD(m.contribution)}</Table.Td><Table.Td>{formatUSD(m.totalInvested)}</Table.Td><Table.Td>{formatSignedUSD(m.return)}</Table.Td><Table.Td>{formatSignedUSD(m.totalGains)}</Table.Td><Table.Td>{formatUSD(m.totalCapital)}</Table.Td></RowTint>)}</Table.Tbody></Table></ScrollArea>
      </Stack> : null}
    </Stack>
  )
}
