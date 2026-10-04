package viewprojection_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/prompts"
	"invest-tracker/internal/viewprojection"
)

type fakeRepo struct {
	start domain.YearMonth
	ok    bool
	err   error
}

func (f *fakeRepo) FirstInvestmentMonth() (domain.YearMonth, bool, error) {
	return f.start, f.ok, f.err
}

// withStart devolve un repo que xa ten data de inicio rexistrada (01/2026).
func withStart() *fakeRepo {
	return &fakeRepo{start: domain.YearMonth{Year: 2026, Month: 1}, ok: true}
}

func runWith(repo viewprojection.Repo, input string) (string, error) {
	var buf bytes.Buffer
	r := bufio.NewReader(strings.NewReader(input))
	err := viewprojection.Run(r, &buf, repo)
	return buf.String(), err
}

func TestRun_ReusesStoredStartDate(t *testing.T) {
	// Só pide salario e retorno: a data xa está no proxecto.
	out, err := runWith(withStart(), "40000\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Proxección a 20 anos",
		"Data de inicio dos investimentos: 01/2026 (activo máis antigo)",
		"240 meses",
		"01/2026 → 12/2045",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, truncate(out))
		}
	}
	if strings.Contains(out, "Mes (1-12):") {
		t.Error("non debería pedir a data cando xa existe no proxecto")
	}
}

func TestRun_PromptsForStartWhenNoAssets(t *testing.T) {
	// mes, ano, salario, retorno.
	out, err := runWith(&fakeRepo{}, "3\n2026\n40000\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Aínda non hai activos: indica a data de inicio dos investimentos.",
		"Mes (1-12):",
		"Ano:",
		"03/2026 → 02/2046",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, truncate(out))
		}
	}
}

func TestRun_RendersSummary(t *testing.T) {
	out, err := runWith(withStart(), "40000\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Salario anual inicial", "$40000.00",
		"Retorno mensual esperado", "+1.0000%",
		"Investimento mensual", "18.8235% do salario mensual",
		"Subida salarial", "+10.00% cada setembro",
		// 40000 · 1.1^20 = 269100.00
		"Salario anual final", "$269100.00",
		"Investimento total",
		"Ganancias acumuladas",
		"Capital final",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, truncate(out))
		}
	}
}

func TestRun_RendersChartSeries(t *testing.T) {
	out, err := runWith(withStart(), "40000\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Evolución do patrimonio · 01/2026 → 12/2045 (240 meses)") {
		t.Errorf("saída non contén o caption da gráfica:\n%s", truncate(out))
	}
	// A lenda debe listar as tres series na mesma liña; comprobámola aparte da
	// táboa e do resumo, que repiten algunhas destas etiquetas.
	legend := findLine(out, "Capital total", "Investimento acumulado", "Ganancias acumuladas")
	if legend == "" {
		t.Errorf("non se atopou a lenda coas tres series:\n%s", truncate(out))
	}
}

// findLine devolve a primeira liña que contén todas as subcadeas indicadas.
func findLine(out string, wants ...string) string {
	for _, line := range strings.Split(out, "\n") {
		all := true
		for _, w := range wants {
			if !strings.Contains(line, w) {
				all = false
				break
			}
		}
		if all {
			return line
		}
	}
	return ""
}

func TestRun_RendersMonthlyTable(t *testing.T) {
	out, err := runWith(withStart(), "40000\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		// Cabeceiras da táboa
		"Mes", "Data", "Salario anual", "Salario mensual", "Invest. mensual",
		"Invest. acumulado", "Rendemento", "Ganancias acum.", "Capital total",
		// Mes 1: investimento 627.45, rendemento +6.27, capital 633.72
		"$627.45", "+$6.27", "$633.72",
		// Setembro do primeiro ano: salario 44000 e investimento 690.20
		"$44000.00", "$690.20",
		"12/2045",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, truncate(out))
		}
	}
	// Unha fila por mes proxectado.
	if got := strings.Count(out, "$627.45"); got < 1 {
		t.Errorf("esperabamos o investimento do primeiro mes na táboa")
	}
	for _, month := range []string{"01/2026", "09/2026", "12/2045"} {
		if !strings.Contains(out, month) {
			t.Errorf("a táboa non contén o mes %q", month)
		}
	}
}

func TestRun_RejectsNegativeSalary(t *testing.T) {
	out, err := runWith(withStart(), "-1\n40000\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Cantidade non válida") {
		t.Errorf("esperabamos rexeitar o salario negativo:\n%s", truncate(out))
	}
}

func TestRun_RejectsEmptyOrNonNumericSalary(t *testing.T) {
	out, err := runWith(withStart(), "\nabc\n40000\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got := strings.Count(out, "Cantidade non válida"); got != 2 {
		t.Errorf("esperabamos 2 avisos de campo obrigatorio, got %d:\n%s", got, truncate(out))
	}
}

func TestRun_AcceptsZeroSalary(t *testing.T) {
	out, err := runWith(withStart(), "0\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if strings.Contains(out, "Cantidade non válida") {
		t.Errorf("un salario de 0 debe aceptarse:\n%s", truncate(out))
	}
}

func TestRun_RejectsReturnAtOrBelowMinus100(t *testing.T) {
	out, err := runWith(withStart(), "40000\n-100\n-150\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got := strings.Count(out, "Porcentaxe non válida"); got != 2 {
		t.Errorf("esperabamos 2 rexeitamentos de retorno, got %d:\n%s", got, truncate(out))
	}
	if !strings.Contains(out, "maior ca -100%") {
		t.Errorf("a mensaxe debe indicar o límite -100%%:\n%s", truncate(out))
	}
}

func TestRun_AcceptsNegativeReturnAboveMinus100(t *testing.T) {
	out, err := runWith(withStart(), "40000\n-5\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "-5.0000%") {
		t.Errorf("saída non contén o retorno negativo:\n%s", truncate(out))
	}
}

func TestRun_AcceptsCommaDecimalReturn(t *testing.T) {
	out, err := runWith(withStart(), "40000\n1,5\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "+1.5000%") {
		t.Errorf("saída non contén o retorno con coma decimal:\n%s", truncate(out))
	}
}

func TestRun_PropagatesCancellation(t *testing.T) {
	for _, input := range []string{":q\n", "40000\n:q\n"} {
		if _, err := runWith(withStart(), input); !errors.Is(err, prompts.ErrCancelled) {
			t.Errorf("input %q: esperabamos ErrCancelled, got %v", input, err)
		}
	}
}

func TestRun_EOFReturnsError(t *testing.T) {
	if _, err := runWith(withStart(), ""); !errors.Is(err, io.EOF) {
		t.Errorf("esperabamos io.EOF, got %v", err)
	}
}

func TestRun_RepoErrorPropagates(t *testing.T) {
	boom := errors.New("boom")
	_, err := runWith(&fakeRepo{err: boom}, "40000\n1\n")
	if !errors.Is(err, boom) {
		t.Errorf("esperabamos o erro do repo, got %v", err)
	}
}

// truncate acurta a saída nas mensaxes de erro: a táboa ten 240 filas.
func truncate(s string) string {
	const max = 2000
	if len(s) <= max {
		return s
	}
	return s[:max] + "\n… (saída truncada)"
}
