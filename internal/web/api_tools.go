package web

import (
	"fmt"
	"math"
	"net/http"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/i18n"
	"invest-tracker/internal/repartoaporte"
)

// registerToolRoutes rexistra as rutas de ferramentas da API.
func (s *Server) registerToolRoutes() {
	s.handleAPI("POST /api/tools/allocation", s.toolAllocation)
	s.handleAPI("GET /api/tools/projection/defaults", s.toolProjectionDefaults)
	s.handleAPI("POST /api/tools/projection", s.toolProjection)
}

type allocationRequest struct {
	Total     float64               `json:"total"`
	Selection []allocationSelection `json:"selection"`
}

type allocationSelection struct {
	Type     domain.AssetType `json:"type"`
	AssetIDs []int64          `json:"assetIds"`
	// Percents é a porcentaxe do total para cada activo de AssetIDs, na mesma
	// orde. As porcentaxes son opcionais (sen elas repártese a partes
	// iguais), pero se se indica algunha hai que indicalas todas.
	Percents  []float64            `json:"percents"`
	NewAssets []allocationNewAsset `json:"newAssets"`
}

// allocationNewAsset é un activo que aínda non se creou: só serve para
// calcular o reparto da primeira vez que se inclúe.
type allocationNewAsset struct {
	Name    string   `json:"name"`
	Percent *float64 `json:"percent"`
}

type projectionRequest struct {
	Start               *domain.YearMonth     `json:"start"`
	Years               int                   `json:"years"`
	InitialInvestment   float64               `json:"initialInvestment"`
	MonthlyReturnPct    float64               `json:"monthlyReturnPct"`
	Mode                domain.ProjectionMode `json:"mode"`
	MonthlyContribution float64               `json:"monthlyContribution"`
	AnnualSalary        float64               `json:"annualSalary"`
	InvestmentRatePct   float64               `json:"investmentRatePct"`
	Rules               []domain.GrowthRule   `json:"rules"`
}

type projectionDefaultsResponse struct {
	Start             domain.YearMonth    `json:"start"`
	StartFromAssets   bool                `json:"startFromAssets"`
	Years             int                 `json:"years"`
	InvestmentRatePct float64             `json:"investmentRatePct"`
	SalaryRules       []domain.GrowthRule `json:"salaryRules"`
}

type projectionResponse struct {
	Input   projectionResponseInput   `json:"input"`
	Months  []domain.PlanMonth        `json:"months"`
	Summary projectionResponseSummary `json:"summary"`
}

type projectionResponseInput struct {
	Start               domain.YearMonth      `json:"start"`
	Years               int                   `json:"years"`
	InitialInvestment   float64               `json:"initialInvestment"`
	MonthlyReturnPct    float64               `json:"monthlyReturnPct"`
	Mode                domain.ProjectionMode `json:"mode"`
	MonthlyContribution float64               `json:"monthlyContribution"`
	AnnualSalary        float64               `json:"annualSalary"`
	InvestmentRatePct   float64               `json:"investmentRatePct"`
	Rules               []domain.GrowthRule   `json:"rules"`
}

type projectionResponseSummary struct {
	InitialInvestment  float64 `json:"initialInvestment"`
	FirstContribution  float64 `json:"firstContribution"`
	FinalContribution  float64 `json:"finalContribution"`
	FinalAnnualSalary  float64 `json:"finalAnnualSalary"`
	TotalContributions float64 `json:"totalContributions"`
	TotalInvested      float64 `json:"totalInvested"`
	TotalGains         float64 `json:"totalGains"`
	FinalCapital       float64 `json:"finalCapital"`
	Months             int     `json:"months"`
}

func (s *Server) toolAllocation(w http.ResponseWriter, r *http.Request) {
	lang := langOf(r)
	var req allocationRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	selection, fields, err := s.validateAllocation(req, lang)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	if len(fields) > 0 {
		writeFieldErrors(w, lang, fields)
		return
	}
	allocation, err := repartoaporte.Allocate(req.Total, selection, lang)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, allocationDTO(allocation))
}

func (s *Server) validateAllocation(req allocationRequest, lang i18n.Lang) ([]repartoaporte.TypeSelection, map[string]string, error) {
	fields := map[string]string{}
	if !domain.ValidAmount(req.Total) {
		fields["total"] = lang.T("o total debe ser unha cantidade positiva")
	}
	if len(req.Selection) == 0 {
		fields["selection"] = lang.T("debe seleccionarse polo menos un tipo")
	}

	assets, err := s.store.ListAssets()
	if err != nil {
		return nil, nil, err
	}
	byID := make(map[int64]domain.Asset, len(assets))
	for _, a := range assets {
		byID[a.ID] = a
	}

	withPercents := false
	for _, sel := range req.Selection {
		withPercents = withPercents || sel.Percents != nil
		for _, n := range sel.NewAssets {
			withPercents = withPercents || n.Percent != nil
		}
	}
	percentSum, percentCount, percentsOK := 0.0, 0, true
	checkPercent := func(key string, p *float64) {
		switch {
		case p == nil:
			fields[key] = lang.T("indica a porcentaxe de cada activo")
			percentsOK = false
		case !repartoaporte.ValidPercent(*p):
			fields[key] = lang.T("a porcentaxe debe estar entre 0 e 100")
			percentsOK = false
		default:
			percentSum += *p
			percentCount++
		}
	}

	selection := make([]repartoaporte.TypeSelection, 0, len(req.Selection))
	seenTypes := map[domain.AssetType]bool{}
	for i, sel := range req.Selection {
		typeField := fmt.Sprintf("selection[%d].type", i)
		assetField := fmt.Sprintf("selection[%d].assetIds", i)
		percentsField := fmt.Sprintf("selection[%d].percents", i)
		if !sel.Type.Valid() {
			fields[typeField] = lang.T("tipo de activo non válido")
		}
		if seenTypes[sel.Type] {
			fields[typeField] = lang.T("tipo de activo duplicado")
		}
		seenTypes[sel.Type] = true
		if len(sel.AssetIDs) == 0 && len(sel.NewAssets) == 0 {
			fields[assetField] = lang.T("debe seleccionarse polo menos un activo")
		}
		picked := make([]domain.Asset, 0, len(sel.AssetIDs)+len(sel.NewAssets))
		seenAssets := map[int64]bool{}
		for _, id := range sel.AssetIDs {
			asset, ok := byID[id]
			if !ok || !sel.Type.Valid() || asset.Type != sel.Type || seenAssets[id] {
				fields[assetField] = lang.T("os activos seleccionados deben existir e pertencer ao tipo indicado")
				continue
			}
			seenAssets[id] = true
			picked = append(picked, asset)
		}
		var percents []float64
		if withPercents {
			if len(sel.Percents) != len(sel.AssetIDs) {
				fields[percentsField] = lang.T("indica a porcentaxe de cada activo")
				percentsOK = false
			}
			for j := range sel.Percents {
				checkPercent(fmt.Sprintf("%s[%d]", percentsField, j), &sel.Percents[j])
			}
			percents = append(percents, sel.Percents...)
		}
		// Os activos novos van con ID 0 e o tipo da selección.
		for k, n := range sel.NewAssets {
			prefix := fmt.Sprintf("selection[%d].newAssets[%d]", i, k)
			name := validateName(fields, lang, prefix+".name", n.Name)
			picked = append(picked, domain.Asset{Type: sel.Type, Name: name})
			if withPercents {
				checkPercent(prefix+".percent", n.Percent)
				if n.Percent != nil {
					percents = append(percents, *n.Percent)
				}
			}
		}
		selection = append(selection, repartoaporte.TypeSelection{Type: sel.Type, Assets: picked, Percents: percents})
	}
	if withPercents && percentsOK && len(req.Selection) > 0 && !repartoaporte.SumsTo100(percentSum, percentCount) {
		fields["percents"] = lang.T("as porcentaxes teñen que sumar 100 %")
	}
	return selection, fields, nil
}

func (s *Server) toolProjectionDefaults(w http.ResponseWriter, r *http.Request) {
	start, fromAssets, err := s.defaultProjectionStart()
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, projectionDefaultsResponse{
		Start:             start,
		StartFromAssets:   fromAssets,
		Years:             domain.ProjectionYears,
		InvestmentRatePct: domain.InvestmentRate * 100,
		SalaryRules:       []domain.GrowthRule{domain.DefaultSalaryRaiseRule(start)},
	})
}

func (s *Server) toolProjection(w http.ResponseWriter, r *http.Request) {
	lang := langOf(r)
	var req projectionRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	plan, fields := projectionPlanFromRequest(req, lang)
	if len(fields) > 0 {
		writeFieldErrors(w, lang, fields)
		return
	}

	months, err := domain.ProjectPlan(plan)
	if err != nil {
		writeError(w, http.StatusBadRequest, lang.T(err.Error()))
		return
	}
	resp := projectionResponse{
		Input: projectionResponseInput{
			Start:               plan.Start,
			Years:               plan.Years,
			InitialInvestment:   plan.InitialInvestment,
			MonthlyReturnPct:    req.MonthlyReturnPct,
			Mode:                plan.Mode,
			MonthlyContribution: plan.MonthlyContribution,
			AnnualSalary:        plan.AnnualSalary,
			InvestmentRatePct:   req.InvestmentRatePct,
			Rules:               nonNil(plan.Rules),
		},
		Months:  nonNil(months),
		Summary: projectionSummary(plan.InitialInvestment, months),
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) defaultProjectionStart() (domain.YearMonth, bool, error) {
	start, ok, err := s.store.FirstInvestmentMonth()
	if err != nil {
		return domain.YearMonth{}, false, err
	}
	if ok {
		return start, true, nil
	}
	now := s.now()
	return domain.YearMonth{Year: now.Year(), Month: int(now.Month())}, false, nil
}

func projectionPlanFromRequest(req projectionRequest, lang i18n.Lang) (domain.PlanInput, map[string]string) {
	fields := map[string]string{}
	plan := domain.PlanInput{
		Years:               req.Years,
		InitialInvestment:   req.InitialInvestment,
		MonthlyReturn:       req.MonthlyReturnPct / 100,
		Mode:                req.Mode,
		MonthlyContribution: req.MonthlyContribution,
		AnnualSalary:        req.AnnualSalary,
		InvestmentRate:      req.InvestmentRatePct / 100,
		Rules:               req.Rules,
	}
	if req.Start == nil {
		fields["start"] = lang.T("a data de inicio é obrigatoria")
	} else if !req.Start.Valid() {
		fields["start"] = lang.T("a data de inicio non é válida")
	} else {
		plan.Start = *req.Start
	}
	if req.Years < 1 || req.Years > domain.MaxProjectionPlanYears {
		fields["years"] = lang.Sprintf("os anos deben estar entre 1 e %d", domain.MaxProjectionPlanYears)
	}
	if !finiteNumber(req.InitialInvestment) || req.InitialInvestment < 0 {
		fields["initialInvestment"] = lang.T("o investimento inicial debe ser unha cantidade non negativa")
	}
	if !finiteNumber(req.MonthlyReturnPct) || req.MonthlyReturnPct <= domain.MinMonthlyReturnPct {
		fields["monthlyReturnPct"] = lang.T("o retorno mensual debe ser maior ca -100%")
	}
	if req.Mode != domain.ProjectionModeContribution && req.Mode != domain.ProjectionModeSalary {
		fields["mode"] = lang.T("o modo debe ser contribution ou salary")
	}
	if req.Mode == domain.ProjectionModeContribution && (!finiteNumber(req.MonthlyContribution) || req.MonthlyContribution < 0) {
		fields["monthlyContribution"] = lang.T("o aporte mensual debe ser unha cantidade non negativa")
	}
	if req.Mode == domain.ProjectionModeSalary {
		if !finiteNumber(req.AnnualSalary) || req.AnnualSalary < 0 {
			fields["annualSalary"] = lang.T("o salario anual debe ser unha cantidade non negativa")
		}
		if !finiteNumber(req.InvestmentRatePct) || req.InvestmentRatePct <= 0 || req.InvestmentRatePct > 100 {
			fields["investmentRatePct"] = lang.T("a porcentaxe de investimento debe estar entre 0 e 100")
		}
	}
	validateGrowthRules(req.Rules, fields, lang)
	return plan, fields
}

func validateGrowthRules(rules []domain.GrowthRule, fields map[string]string, lang i18n.Lang) {
	if len(rules) > domain.MaxGrowthRules {
		fields["rules"] = lang.Sprintf("non se poden definir máis de %d regras", domain.MaxGrowthRules)
	}
	for i, rule := range rules {
		prefix := fmt.Sprintf("rules[%d]", i)
		if rule.Kind != domain.GrowthKindFixed && rule.Kind != domain.GrowthKindPercent {
			fields[prefix+".kind"] = lang.T("o tipo de regra debe ser fixed ou percent")
		}
		if !finiteNumber(rule.Value) || (rule.Kind == domain.GrowthKindPercent && rule.Value <= -100) {
			fields[prefix+".value"] = lang.T("o valor da regra non é válido")
		}
		if rule.EveryMonths < 0 || rule.EveryMonths > domain.MaxRuleEveryMonths {
			fields[prefix+".everyMonths"] = lang.Sprintf("a periodicidade debe estar entre 0 e %d meses", domain.MaxRuleEveryMonths)
		}
		if !rule.From.Valid() {
			fields[prefix+".from"] = lang.T("a data de inicio da regra non é válida")
		}
	}
}

func projectionSummary(initial float64, months []domain.PlanMonth) projectionResponseSummary {
	summary := projectionResponseSummary{InitialInvestment: initial, Months: len(months)}
	if len(months) == 0 {
		return summary
	}
	first := months[0]
	last := months[len(months)-1]
	summary.FirstContribution = first.Contribution
	summary.FinalContribution = last.Contribution
	summary.FinalAnnualSalary = last.AnnualSalary
	summary.TotalInvested = last.TotalInvested
	summary.TotalContributions = last.TotalInvested - initial
	summary.TotalGains = last.TotalGains
	summary.FinalCapital = last.TotalCapital
	return summary
}

func finiteNumber(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}
