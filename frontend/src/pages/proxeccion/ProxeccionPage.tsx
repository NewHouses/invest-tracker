import { Button, Group, Loader, NumberInput, Paper, ScrollArea, Stack, Table, Text, TextInput, Title } from '@mantine/core'
import { LineChart } from '@mantine/charts'
import { useState } from 'react'

import { ApiError } from '@/api/client'
import type { ProjectionMonth } from '@/api/tools'
import { useProjectionMutation, useProjectionStart } from '@/api/tools'
import type { YearMonth } from '@/api/types'
import { ErrorAlert } from '@/components/ErrorAlert'
import { PageHeader } from '@/components/PageHeader'
import { YearMonthInput } from '@/components/YearMonthInput'
import { KvTable, RowTint } from '@/features/charts/common'
import { formatAxisUSD, formatNumber, formatPct, formatSignedUSD, formatUSD, formatYearMonth, galicianMonthNames } from '@/lib/format'
import { currentYearMonth } from '@/lib/yearMonth'

function parseDecimal(value: string) {
  const n = Number(value.replace(',', '.'))
  return Number.isFinite(n) ? n : Number.NaN
}
function field(error: unknown, name: string) { return error instanceof ApiError ? error.fields?.[name] : undefined }

function projectionData(months: ProjectionMonth[]) {
  return months.map((m) => ({ month: formatYearMonth(m.date), 'Capital total': m.totalCapital, 'Investimento acumulado': m.totalInvested, 'Ganancias acumuladas': m.totalGains }))
}

export function ProxeccionPage() {
  const start = useProjectionStart()
  const mutation = useProjectionMutation()
  const [annualSalary, setAnnualSalary] = useState<number | string>('')
  const [monthlyReturn, setMonthlyReturn] = useState('')
  const [manualStart, setManualStart] = useState<YearMonth>(currentYearMonth())
  const hasAssets = start.data?.start !== null && start.data?.start !== undefined
  const submit = () => {
    const body = { annualSalary: typeof annualSalary === 'number' ? annualSalary : parseDecimal(String(annualSalary)), monthlyReturnPct: parseDecimal(monthlyReturn), ...(hasAssets ? {} : { start: manualStart }) }
    mutation.mutate(body)
  }
  return (
    <Stack>
      <PageHeader title="Proxección" description="Simulación a longo prazo" />
      {start.isLoading ? <Loader /> : null}
      {start.isError ? <ErrorAlert error={start.error} /> : null}
      <Paper withBorder p="md"><Stack>
        <NumberInput label="Salario anual inicial (USD)" min={0} decimalSeparator="," thousandSeparator="." value={annualSalary} onChange={setAnnualSalary} error={field(mutation.error, 'annualSalary')} />
        <TextInput label="Retorno mensual esperado (%)" value={monthlyReturn} onChange={(e) => setMonthlyReturn(e.currentTarget.value)} error={field(mutation.error, 'monthlyReturnPct')} />
        {hasAssets && start.data?.start ? <Text>Data de inicio dos investimentos: {formatYearMonth(start.data.start)} (activo máis antigo)</Text> : <YearMonthInput label="Data de inicio" value={manualStart} onChange={(v) => v && setManualStart(v)} error={field(mutation.error, 'start')} required />}
        <Group><Button onClick={submit} loading={mutation.isPending}>Calcular proxección</Button></Group>
      </Stack></Paper>
      {mutation.isError ? <ErrorAlert error={mutation.error} title="Non se puido calcular" /> : null}
      {mutation.data ? <Stack>
        <Paper withBorder p="md"><Stack><Title order={3}>Resumo</Title><KvTable rows={[
          { label: 'Salario anual inicial', value: formatUSD(mutation.data.input.annualSalary) },
          { label: 'Retorno mensual esperado', value: formatPct(mutation.data.input.monthlyReturnPct) },
          { label: 'Investimento mensual', value: `${formatNumber(mutation.data.rules.investmentRate * 100, 4)} % do salario mensual` },
          { label: 'Subida salarial', value: `+${formatNumber(mutation.data.rules.salaryRaiseRate * 100)} % cada ${galicianMonthNames[mutation.data.rules.salaryRaiseMonth - 1] ?? mutation.data.rules.salaryRaiseMonth}` },
          { label: 'Salario anual final', value: formatUSD(mutation.data.summary.finalAnnualSalary) },
          { label: 'Investimento total', value: formatUSD(mutation.data.summary.totalInvested) },
          { label: 'Ganancias acumuladas', value: formatSignedUSD(mutation.data.summary.totalGains) },
          { label: 'Capital final', value: formatUSD(mutation.data.summary.finalCapital) },
        ]} /></Stack></Paper>
        <Paper withBorder p="md"><LineChart h={320} data={projectionData(mutation.data.months)} dataKey="month" series={[{ name: 'Capital total', color: 'blue.6' }, { name: 'Investimento acumulado', color: 'green.6' }, { name: 'Ganancias acumuladas', color: 'orange.6' }]} valueFormatter={formatUSD} yAxisProps={{ tickFormatter: formatAxisUSD, width: 64 }} withLegend /></Paper>
        <ScrollArea><Table withTableBorder striped><Table.Thead><Table.Tr><Table.Th>Mes</Table.Th><Table.Th>Data</Table.Th><Table.Th>Salario anual</Table.Th><Table.Th>Salario mensual</Table.Th><Table.Th>Invest. mensual</Table.Th><Table.Th>Invest. acumulado</Table.Th><Table.Th>Rendemento</Table.Th><Table.Th>Ganancias acum.</Table.Th><Table.Th>Capital total</Table.Th></Table.Tr></Table.Thead><Table.Tbody>{mutation.data.months.map((m) => <RowTint key={m.index} gain={m.totalGains}><Table.Td>{m.index}</Table.Td><Table.Td>{formatYearMonth(m.date)}</Table.Td><Table.Td>{formatUSD(m.annualSalary)}</Table.Td><Table.Td>{formatUSD(m.monthlySalary)}</Table.Td><Table.Td>{formatUSD(m.investment)}</Table.Td><Table.Td>{formatUSD(m.totalInvested)}</Table.Td><Table.Td>{formatSignedUSD(m.return)}</Table.Td><Table.Td>{formatSignedUSD(m.totalGains)}</Table.Td><Table.Td>{formatUSD(m.totalCapital)}</Table.Td></RowTint>)}</Table.Tbody></Table></ScrollArea>
      </Stack> : null}
    </Stack>
  )
}
