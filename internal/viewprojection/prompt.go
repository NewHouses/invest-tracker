// Package viewprojection renders a 20-year projection of the portfolio: how
// the capital grows month by month given an initial annual salary, an expected
// monthly return and the automatic September raises.
package viewprojection

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"invest-tracker/internal/charts"
	"invest-tracker/internal/colors"
	"invest-tracker/internal/domain"
	"invest-tracker/internal/money"
	"invest-tracker/internal/prompts"
)

type Repo interface {
	FirstInvestmentMonth() (domain.YearMonth, bool, error)
}

const sep = "==================================================================="

// chartWidth limita as columnas da gráfica: os 240 meses da proxección non
// caberían nunha terminal a un punto por columna.
const chartWidth = 100

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	fmt.Fprintf(w, "\n--- Proxección a %d anos ---\n", domain.ProjectionYears)

	start, err := resolveStart(r, w, repo)
	if err != nil {
		return err
	}
	salary, err := prompts.NonNegativeAmount(r, w, "Salario anual inicial (USD)")
	if err != nil {
		return err
	}
	monthlyReturn, err := prompts.Percent(r, w,
		"Retorno mensual esperado (%)", domain.MinMonthlyReturnPct)
	if err != nil {
		return err
	}

	in := domain.ProjectionInput{
		AnnualSalary:  salary,
		MonthlyReturn: monthlyReturn,
		Start:         start,
	}
	months, err := domain.Project(in)
	if err != nil {
		// main.go xa prefixa "erro calculando a proxección".
		return err
	}

	renderReport(w, in, months)
	return nil
}

// resolveStart reutiliza a data de inicio dos investimentos xa rexistrada na
// aplicación (a do activo máis antigo). Se aínda non hai activos, pídella ao
// usuario.
func resolveStart(r *bufio.Reader, w io.Writer, repo Repo) (domain.YearMonth, error) {
	ym, ok, err := repo.FirstInvestmentMonth()
	if err != nil {
		return domain.YearMonth{}, fmt.Errorf("obtendo a data de inicio dos investimentos: %w", err)
	}
	if ok {
		fmt.Fprintf(w, "Data de inicio dos investimentos: %02d/%d (activo máis antigo)\n",
			ym.Month, ym.Year)
		return ym, nil
	}
	fmt.Fprintln(w, "Aínda non hai activos: indica a data de inicio dos investimentos.")
	month, err := prompts.Month(r, w)
	if err != nil {
		return domain.YearMonth{}, err
	}
	year, err := prompts.Year(r, w)
	if err != nil {
		return domain.YearMonth{}, err
	}
	return domain.YearMonth{Year: year, Month: month}, nil
}

func renderReport(w io.Writer, in domain.ProjectionInput, months []domain.ProjectionMonth) {
	if len(months) == 0 {
		return
	}
	first := months[0]
	last := months[len(months)-1]

	fmt.Fprintln(w)
	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "  Proxección a %d anos · %d meses · %02d/%d → %02d/%d\n",
		domain.ProjectionYears, len(months),
		first.Date.Month, first.Date.Year, last.Date.Month, last.Date.Year)
	fmt.Fprintln(w, sep)

	twH := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(twH, "  Salario anual inicial\t%s\n", money.USD(in.AnnualSalary))
	fmt.Fprintf(twH, "  Retorno mensual esperado\t%+.4f%%\n", in.MonthlyReturn*100)
	fmt.Fprintf(twH, "  Investimento mensual\t%.4f%% do salario mensual\n", domain.InvestmentRate*100)
	fmt.Fprintf(twH, "  Subida salarial\t+%.2f%% cada setembro\n", domain.SalaryRaiseRate*100)
	fmt.Fprintf(twH, "  Salario anual final\t%s\n", money.USD(last.AnnualSalary))
	fmt.Fprintf(twH, "  Investimento total\t%s\n", money.USD(last.TotalInvested))
	fmt.Fprintf(twH, "  Ganancias acumuladas\t%s\n", money.SignedUSD(last.TotalGains))
	fmt.Fprintf(twH, "  Capital final\t%s\n", money.USD(last.TotalCapital))
	twH.Flush()
	fmt.Fprintln(w, sep)

	renderChart(w, months)
	renderTable(w, months)
	fmt.Fprintln(w, sep)
}

// renderChart debuxa a evolución do capital total xunto co investimento
// acumulado e as ganancias, para ver canto do patrimonio vén de cada fonte.
func renderChart(w io.Writer, months []domain.ProjectionMonth) {
	dates := make([]domain.YearMonth, len(months))
	capital := make([]float64, len(months))
	invested := make([]float64, len(months))
	gains := make([]float64, len(months))
	for i, m := range months {
		dates[i] = m.Date
		capital[i] = m.TotalCapital
		invested[i] = m.TotalInvested
		gains[i] = m.TotalGains
	}
	series := []charts.Series{
		{Label: "Capital total", Values: capital},
		{Label: "Investimento acumulado", Values: invested},
		{Label: "Ganancias acumuladas", Values: gains},
	}
	charts.RenderLineWidth(w, series, dates, "Evolución do patrimonio", chartWidth)
	fmt.Fprintln(w)
}

func renderTable(w io.Writer, months []domain.ProjectionMonth) {
	var tbuf bytes.Buffer
	tw := tabwriter.NewWriter(&tbuf, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(tw, "  Mes\tData\tSalario anual\tSalario mensual\tInvest. mensual\t"+
		"Invest. acumulado\tRendemento\tGanancias acum.\tCapital total\t")
	for _, m := range months {
		fmt.Fprintf(tw, "  %d\t%02d/%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t\n",
			m.Index, m.Date.Month, m.Date.Year,
			money.USD(m.AnnualSalary), money.USD(m.MonthlySalary),
			money.USD(m.Investment), money.USD(m.TotalInvested),
			money.SignedUSD(m.Return), money.SignedUSD(m.TotalGains),
			money.USD(m.TotalCapital))
	}
	tw.Flush()
	writeColoredRows(w, tbuf.String(), months)
}

// writeColoredRows imprime as liñas xa formatadas: a primeira (cabeceira) sen
// cor, e cada fila de datos envolvida no código ANSI segundo as súas ganancias
// acumuladas.
func writeColoredRows(w io.Writer, formatted string, months []domain.ProjectionMonth) {
	lines := strings.Split(strings.TrimRight(formatted, "\n"), "\n")
	if len(lines) == 0 {
		return
	}
	fmt.Fprintln(w, lines[0])
	for i, line := range lines[1:] {
		if i < len(months) {
			fmt.Fprintln(w, colors.ForGain(months[i].TotalGains)+line+colors.Reset)
		} else {
			fmt.Fprintln(w, line)
		}
	}
}
