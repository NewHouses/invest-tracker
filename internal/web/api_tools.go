package web

import (
	"fmt"
	"math"
	"net/http"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/repartoaporte"
)

// registerToolRoutes rexistra as rutas de ferramentas da API.
func (s *Server) registerToolRoutes() {
	s.handleAPI("POST /api/tools/allocation", s.toolAllocation)
	s.handleAPI("GET /api/tools/projection/start", s.toolProjectionStart)
	s.handleAPI("POST /api/tools/projection", s.toolProjection)
}

type allocationRequest struct {
	Total     float64               `json:"total"`
	Selection []allocationSelection `json:"selection"`
}

type allocationSelection struct {
	Type     domain.AssetType `json:"type"`
	AssetIDs []int64          `json:"assetIds"`
}

type projectionRequest struct {
	AnnualSalary     float64           `json:"annualSalary"`
	MonthlyReturnPct float64           `json:"monthlyReturnPct"`
	Start            *domain.YearMonth `json:"start"`
}

type projectionResponse struct {
	Start           domain.YearMonth          `json:"start"`
	StartFromAssets bool                      `json:"startFromAssets"`
	Input           projectionResponseInput   `json:"input"`
	Rules           projectionResponseRules   `json:"rules"`
	Months          []domain.ProjectionMonth  `json:"months"`
	Summary         projectionResponseSummary `json:"summary"`
}

type projectionResponseInput struct {
	AnnualSalary     float64 `json:"annualSalary"`
	MonthlyReturnPct float64 `json:"monthlyReturnPct"`
}

type projectionResponseRules struct {
	InvestmentRate   float64 `json:"investmentRate"`
	SalaryRaiseRate  float64 `json:"salaryRaiseRate"`
	SalaryRaiseMonth int     `json:"salaryRaiseMonth"`
	Years            int     `json:"years"`
}

type projectionResponseSummary struct {
	FinalAnnualSalary float64 `json:"finalAnnualSalary"`
	TotalInvested     float64 `json:"totalInvested"`
	TotalGains        float64 `json:"totalGains"`
	FinalCapital      float64 `json:"finalCapital"`
}

func (s *Server) toolAllocation(w http.ResponseWriter, r *http.Request) {
	var req allocationRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	selection, fields, err := s.validateAllocation(req)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	if len(fields) > 0 {
		writeFieldErrors(w, fields)
		return
	}
	allocation, err := repartoaporte.Allocate(req.Total, selection)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, allocationDTO(allocation))
}

func (s *Server) validateAllocation(req allocationRequest) ([]repartoaporte.TypeSelection, map[string]string, error) {
	fields := map[string]string{}
	if !domain.ValidAmount(req.Total) {
		fields["total"] = "o total debe ser unha cantidade positiva"
	}
	if len(req.Selection) == 0 {
		fields["selection"] = "debe seleccionarse polo menos un tipo"
	}

	assets, err := s.store.ListAssets()
	if err != nil {
		return nil, nil, err
	}
	byID := make(map[int64]domain.Asset, len(assets))
	for _, a := range assets {
		byID[a.ID] = a
	}

	selection := make([]repartoaporte.TypeSelection, 0, len(req.Selection))
	seenTypes := map[domain.AssetType]bool{}
	for i, sel := range req.Selection {
		typeField := fmt.Sprintf("selection[%d].type", i)
		assetField := fmt.Sprintf("selection[%d].assetIds", i)
		if !sel.Type.Valid() {
			fields[typeField] = "tipo de activo non válido"
		}
		if seenTypes[sel.Type] {
			fields[typeField] = "tipo de activo duplicado"
		}
		seenTypes[sel.Type] = true
		if len(sel.AssetIDs) == 0 {
			fields[assetField] = "debe seleccionarse polo menos un activo"
		}
		picked := make([]domain.Asset, 0, len(sel.AssetIDs))
		seenAssets := map[int64]bool{}
		for _, id := range sel.AssetIDs {
			asset, ok := byID[id]
			if !ok || !sel.Type.Valid() || asset.Type != sel.Type || seenAssets[id] {
				fields[assetField] = "os activos seleccionados deben existir e pertencer ao tipo indicado"
				continue
			}
			seenAssets[id] = true
			picked = append(picked, asset)
		}
		selection = append(selection, repartoaporte.TypeSelection{Type: sel.Type, Assets: picked})
	}
	return selection, fields, nil
}

func (s *Server) toolProjectionStart(w http.ResponseWriter, r *http.Request) {
	start, ok, err := s.store.FirstInvestmentMonth()
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	var ptr *domain.YearMonth
	if ok {
		ptr = &start
	}
	writeJSON(w, http.StatusOK, map[string]*domain.YearMonth{"start": ptr})
}

func (s *Server) toolProjection(w http.ResponseWriter, r *http.Request) {
	var req projectionRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	start, fromAssets, fields, err := s.projectionStart(req)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	if !domain.ValidNonNegativeAmount(req.AnnualSalary) {
		fields["annualSalary"] = "o salario anual debe ser unha cantidade non negativa"
	}
	if math.IsNaN(req.MonthlyReturnPct) || math.IsInf(req.MonthlyReturnPct, 0) || req.MonthlyReturnPct <= domain.MinMonthlyReturnPct {
		fields["monthlyReturnPct"] = "o retorno mensual debe ser maior ca -100%"
	}
	if len(fields) > 0 {
		writeFieldErrors(w, fields)
		return
	}

	months, err := domain.Project(domain.ProjectionInput{
		AnnualSalary:  req.AnnualSalary,
		MonthlyReturn: req.MonthlyReturnPct / 100,
		Start:         start,
	})
	if err != nil {
		writeFieldErrors(w, map[string]string{"start": err.Error()})
		return
	}
	resp := projectionResponse{
		Start:           start,
		StartFromAssets: fromAssets,
		Input: projectionResponseInput{
			AnnualSalary:     req.AnnualSalary,
			MonthlyReturnPct: req.MonthlyReturnPct,
		},
		Rules: projectionResponseRules{
			InvestmentRate:   domain.InvestmentRate,
			SalaryRaiseRate:  domain.SalaryRaiseRate,
			SalaryRaiseMonth: domain.SalaryRaiseMonth,
			Years:            domain.ProjectionYears,
		},
		Months: nonNil(months),
	}
	if len(months) > 0 {
		last := months[len(months)-1]
		resp.Summary = projectionResponseSummary{
			FinalAnnualSalary: last.AnnualSalary,
			TotalInvested:     last.TotalInvested,
			TotalGains:        last.TotalGains,
			FinalCapital:      last.TotalCapital,
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) projectionStart(req projectionRequest) (domain.YearMonth, bool, map[string]string, error) {
	fields := map[string]string{}
	start, ok, err := s.store.FirstInvestmentMonth()
	if err != nil {
		return domain.YearMonth{}, false, nil, err
	}
	if ok {
		return start, true, fields, nil
	}
	if req.Start == nil {
		fields["start"] = "a data de inicio é obrigatoria se non hai activos"
		return domain.YearMonth{}, false, fields, nil
	}
	if !req.Start.Valid() {
		fields["start"] = "a data de inicio non é válida"
		return domain.YearMonth{}, false, fields, nil
	}
	return *req.Start, false, fields, nil
}
