package viewtypereport

import (
	"bufio"
	"fmt"
	"io"
	"text/tabwriter"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/money"
	"invest-tracker/internal/prompts"
)

type Repo interface {
	ListAssets() ([]domain.Asset, error)
	MonthlySummary(assetID int64, year, month int) (domain.MonthlySummary, error)
}

const sep = "========================================================="

type Entry struct {
	Asset   domain.Asset          `json:"asset"`
	Summary domain.MonthlySummary `json:"summary"`
}

type Report struct {
	Type             domain.AssetType `json:"type"`
	Period           domain.YearMonth `json:"period"`
	Assets           []domain.Asset   `json:"assets"`
	Active           []Entry          `json:"active"`
	TotalInvested    float64          `json:"totalInvested"`
	InvestedInMonth  float64          `json:"investedInMonth"`
	Holding          float64          `json:"holding"`
	ResultSum        float64          `json:"resultSum"`
	HoldingForResult float64          `json:"holdingForResult"`
	WithResult       int              `json:"withResult"`
	Gain             float64          `json:"gain"`
	GainPct          float64          `json:"gainPct"`
	HasGainPct       bool             `json:"hasGainPct"`
	Partial          bool             `json:"partial"`
}

func Build(repo Repo, typ domain.AssetType, year, month int) (Report, error) {
	assets, err := repo.ListAssets()
	if err != nil {
		return Report{}, fmt.Errorf("listando activos: %w", err)
	}
	return buildFromAssets(repo, typ, assetsOfType(assets, typ), year, month)
}

func assetsOfType(assets []domain.Asset, typ domain.AssetType) []domain.Asset {
	var ofType []domain.Asset
	for _, a := range assets {
		if a.Type == typ {
			ofType = append(ofType, a)
		}
	}
	return ofType
}

func buildFromAssets(repo Repo, typ domain.AssetType, assets []domain.Asset, year, month int) (Report, error) {
	report := Report{
		Type:   typ,
		Period: domain.YearMonth{Year: year, Month: month},
		Assets: assets,
	}
	for _, a := range assets {
		sum, err := repo.MonthlySummary(a.ID, year, month)
		if err != nil {
			return report, fmt.Errorf("calculando resumo de %s: %w", a.Name, err)
		}
		if sum.EstimatedHolding <= 0 {
			continue
		}
		report.Active = append(report.Active, Entry{Asset: a, Summary: sum})
		report.TotalInvested += sum.TotalInvestedUpTo
		report.InvestedInMonth += sum.InvestedInMonth
		report.Holding += sum.EstimatedHolding
		if sum.HasResult {
			report.ResultSum += sum.Result
			report.HoldingForResult += sum.EstimatedHolding
			report.WithResult++
		}
	}
	if report.WithResult > 0 {
		report.Gain = report.ResultSum - report.HoldingForResult
		if report.HoldingForResult > 0 {
			report.GainPct = report.Gain / report.HoldingForResult * 100
			report.HasGainPct = true
		}
	}
	report.Partial = report.WithResult > 0 && report.WithResult < len(report.Active)
	return report, nil
}

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	fmt.Fprint(w, "\n--- Informe mensual por tipo ---\n")

	typ, err := prompts.SelectAssetType(r, w)
	if err != nil {
		return err
	}

	assets, err := repo.ListAssets()
	if err != nil {
		return fmt.Errorf("listando activos: %w", err)
	}

	ofType := assetsOfType(assets, typ)
	if len(ofType) == 0 {
		fmt.Fprintf(w, "Non hai activos de tipo %s.\n", typ.Display())
		return nil
	}

	fmt.Fprintf(w, "Activos de tipo %s:\n", typ.Display())
	for _, a := range ofType {
		fmt.Fprintf(w, "  - %s\n", a.Name)
	}

	month, err := prompts.Month(r, w)
	if err != nil {
		return err
	}
	year, err := prompts.Year(r, w)
	if err != nil {
		return err
	}

	report, err := buildFromAssets(repo, typ, ofType, year, month)
	if err != nil {
		return err
	}

	if len(report.Active) == 0 {
		fmt.Fprintf(w, "Non hai activos de tipo %s con capital investido en %02d/%d.\n",
			typ.Display(), month, year)
		return nil
	}

	renderTable(w, report)
	return nil
}

func renderTable(w io.Writer, report Report) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "  Tipo: %s · %02d/%d\n", report.Type.Display(), report.Period.Month, report.Period.Year)
	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "  Activos incluídos: %d\n", len(report.Active))
	for _, e := range report.Active {
		fmt.Fprintf(w, "    - %s (no activo: %s)\n", e.Asset.Name, money.USD(e.Summary.EstimatedHolding))
	}
	fmt.Fprintln(w, sep)

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "  Investido ata o mes\t%s\n", money.USD(report.TotalInvested))
	fmt.Fprintf(tw, "  Investido este mes\t%s\n", money.USD(report.InvestedInMonth))
	fmt.Fprintf(tw, "  No activo\t%s\n", money.USD(report.Holding))

	switch {
	case report.WithResult == 0:
		fmt.Fprintln(tw, "  Resultado\t—")
		fmt.Fprintln(tw, "  Gañanzas/Perdas\t—")
		fmt.Fprintln(tw, "  Índice\t—")
	case !report.Partial:
		fmt.Fprintf(tw, "  Resultado\t%s\n", money.USD(report.ResultSum))
		fmt.Fprintf(tw, "  Gañanzas/Perdas\t%s\n", money.SignedUSD(report.Gain))
		if report.HasGainPct {
			fmt.Fprintf(tw, "  Índice\t%+.2f%%\n", report.GainPct)
		} else {
			fmt.Fprintln(tw, "  Índice\tn/a")
		}
	default:
		fmt.Fprintf(tw, "  Resultado (parc.)\t%s  (%d/%d activos)\n", money.USD(report.ResultSum), report.WithResult, len(report.Active))
		fmt.Fprintf(tw, "  Gañanzas/Perdas (parc.)\t%s\n", money.SignedUSD(report.Gain))
		if report.HasGainPct {
			fmt.Fprintf(tw, "  Índice (parc.)\t%+.2f%%\n", report.GainPct)
		} else {
			fmt.Fprintln(tw, "  Índice (parc.)\tn/a")
		}
	}
	tw.Flush()

	fmt.Fprintln(w, sep)
}
