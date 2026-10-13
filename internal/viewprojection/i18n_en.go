package viewprojection

import "invest-tracker/internal/i18n"

func init() {
	i18n.Register(map[string]string{
		"\n--- Proxección a %d anos ---\n":                                  "\n--- %d-year projection ---\n",
		"Salario anual inicial (USD)":                                       "Initial annual salary (USD)",
		"Retorno mensual esperado (%)":                                      "Expected monthly return (%)",
		"obtendo a data de inicio dos investimentos: %w":                    "getting the investment start date: %w",
		"Data de inicio dos investimentos: %02d/%d (activo máis antigo)\n":  "Investment start date: %02d/%d (oldest asset)\n",
		"Aínda non hai activos: indica a data de inicio dos investimentos.": "There are no assets yet: enter the investment start date.",
		"  Proxección a %d anos · %d meses · %02d/%d → %02d/%d\n":           "  %d-year projection · %d months · %02d/%d → %02d/%d\n",
		"  Salario anual inicial\t%s\n":                                     "  Initial annual salary\t%s\n",
		"  Retorno mensual esperado\t%+.4f%%\n":                             "  Expected monthly return\t%+.4f%%\n",
		"  Investimento mensual\t%.4f%% do salario mensual\n":               "  Monthly investment\t%.4f%% of monthly salary\n",
		"  Subida salarial\t+%.2f%% cada setembro\n":                        "  Salary raise\t+%.2f%% every September\n",
		"  Salario anual final\t%s\n":                                       "  Final annual salary\t%s\n",
		"  Investimento total\t%s\n":                                        "  Total investment\t%s\n",
		"  Ganancias acumuladas\t%s\n":                                      "  Cumulative gains\t%s\n",
		"  Capital final\t%s\n":                                             "  Final capital\t%s\n",
		"Capital total":                                                     "Total capital",
		"Investimento acumulado":                                            "Cumulative investment",
		"Ganancias acumuladas":                                              "Cumulative gains",
		"Evolución do patrimonio":                                           "Net worth evolution",
		"  Mes\tData\tSalario anual\tSalario mensual\tInvest. mensual\tInvest. acumulado\tRendemento\tGanancias acum.\tCapital total\t": "  Month\tDate\tAnnual salary\tMonthly salary\tMonthly invest.\tCumulative invest.\tReturn\tCum. gains\tTotal capital\t",
	})
}
