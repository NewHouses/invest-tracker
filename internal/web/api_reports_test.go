package web

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"testing"
	"time"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/viewtotalhistory"
	"invest-tracker/internal/viewtotalreport"
)

func seedSkippedMonth(t *testing.T, s *Server) (domain.Asset, domain.Asset) {
	t.Helper()
	a := domain.Asset{Type: domain.Accion, Name: "A", AmountUSD: 1000, Month: 1, Year: 2026}
	b := domain.Asset{Type: domain.Accion, Name: "B", AmountUSD: 500, Month: 1, Year: 2026}
	id, err := s.store.InsertAsset(a)
	if err != nil {
		t.Fatalf("InsertAsset A: %v", err)
	}
	a.ID = id
	id, err = s.store.InsertAsset(b)
	if err != nil {
		t.Fatalf("InsertAsset B: %v", err)
	}
	b.ID = id
	for _, r := range []domain.MonthlyResult{
		{AssetID: a.ID, ResultUSD: 1000, Month: 1, Year: 2026},
		{AssetID: b.ID, ResultUSD: 500, Month: 1, Year: 2026},
		{AssetID: b.ID, ResultUSD: 500, Month: 2, Year: 2026},
		{AssetID: a.ID, ResultUSD: 1100, Month: 3, Year: 2026},
		{AssetID: b.ID, ResultUSD: 510, Month: 3, Year: 2026},
	} {
		if _, err := s.store.InsertMonthlyResult(r); err != nil {
			t.Fatalf("InsertMonthlyResult: %v", err)
		}
	}
	return a, b
}

func compactJSON(t *testing.T, b []byte) []byte {
	t.Helper()
	var out bytes.Buffer
	if err := json.Compact(&out, b); err != nil {
		t.Fatalf("compactando JSON: %v\n%s", err, string(b))
	}
	return out.Bytes()
}

func TestReportsRequireSession(t *testing.T) {
	s := testServer(t)
	rr := doJSON(t, s.Handler(), http.MethodGet, "/api/reports/total/history", nil, nil)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("código=%d, esperabamos 401", rr.Code)
	}
}

func TestReportsHappyAndParity(t *testing.T) {
	s := testServer(t)
	a, _ := seedSkippedMonth(t, s)
	h := s.Handler()
	cookie := authCookie(t, s)

	cases := []string{
		"/api/reports/asset/" + jsonID(a.ID) + "/month?year=2026&month=3",
		"/api/reports/type/accion/month?year=2026&month=3",
		"/api/reports/asset/" + jsonID(a.ID) + "/history",
		"/api/reports/type/accion/history",
	}
	for _, path := range cases {
		rr := doJSON(t, h, http.MethodGet, path, nil, cookie)
		if rr.Code != http.StatusOK {
			t.Fatalf("%s: código=%d corpo=%s", path, rr.Code, rr.Body.String())
		}
	}

	rr := doJSON(t, h, http.MethodGet, "/api/reports/total/month?year=2026&month=3", nil, cookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("total/month: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	var month viewtotalreport.Report
	decodeBody(t, rr, &month)
	if !month.HasMetrics || month.GainNoDiv != 110 || month.PctNoDiv < 7.33 || month.PctNoDiv > 7.34 {
		t.Fatalf("métricas total/month inesperadas: %#v", month)
	}
	expectedMonth, err := viewtotalreport.Build(s.store, 2026, 3)
	if err != nil {
		t.Fatalf("Build total/month: %v", err)
	}
	expectedMonthJSON, _ := json.Marshal(expectedMonth)
	if !bytes.Equal(compactJSON(t, rr.Body.Bytes()), expectedMonthJSON) {
		t.Fatalf("JSON total/month non coincide\napi=%s\nexp=%s", compactJSON(t, rr.Body.Bytes()), expectedMonthJSON)
	}

	rr = doJSON(t, h, http.MethodGet, "/api/reports/total/history", nil, cookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("total/history: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	var historyResp totalHistoryResponse
	decodeBody(t, rr, &historyResp)
	if !historyResp.HasAverages || historyResp.AvgIndexPct < 2.44 || historyResp.AvgIndexPct > 2.45 {
		t.Fatalf("media histórica inesperada: %#v", historyResp.History)
	}
	expectedHistory, err := viewtotalhistory.Build(s.store)
	if err != nil {
		t.Fatalf("Build total/history: %v", err)
	}
	apiHistoryJSON, _ := json.Marshal(historyResp.History)
	expectedHistoryJSON, _ := json.Marshal(totalHistoryDTO(expectedHistory))
	if !bytes.Equal(apiHistoryJSON, expectedHistoryJSON) {
		t.Fatalf("JSON total/history non coincide\napi=%s\nexp=%s", apiHistoryJSON, expectedHistoryJSON)
	}
	if historyResp.Current == nil {
		t.Fatal("esperabamos current no mes aberto")
	}
}

func TestReportsValidationNotFoundAndEmptySlices(t *testing.T) {
	s := testServer(t)
	h := s.Handler()
	cookie := authCookie(t, s)

	if rr := doJSON(t, h, http.MethodGet, "/api/reports/asset/999/month?year=2026&month=3", nil, cookie); rr.Code != http.StatusNotFound {
		t.Fatalf("activo descoñecido: código=%d", rr.Code)
	}
	if rr := doJSON(t, h, http.MethodGet, "/api/reports/type/mal/month?year=2026&month=3", nil, cookie); rr.Code != http.StatusBadRequest {
		t.Fatalf("tipo inválido: código=%d", rr.Code)
	}
	rr := doJSON(t, h, http.MethodGet, "/api/reports/total/month?year=2026", nil, cookie)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("query inválida: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	var apiErr apiError
	decodeBody(t, rr, &apiErr)
	if apiErr.Fields["month"] == "" {
		t.Fatalf("faltou fields.month: %#v", apiErr)
	}

	rr = doJSON(t, h, http.MethodGet, "/api/reports/total/history", nil, cookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("empty total/history: código=%d", rr.Code)
	}
	if !bytes.Contains(rr.Body.Bytes(), []byte(`"rows":[]`)) {
		t.Fatalf("rows debe ser []: %s", rr.Body.String())
	}
	rr = doJSON(t, h, http.MethodGet, "/api/reports/type/accion/month?year=2026&month=3", nil, cookie)
	if !bytes.Contains(rr.Body.Bytes(), []byte(`"assets":[]`)) || !bytes.Contains(rr.Body.Bytes(), []byte(`"active":[]`)) {
		t.Fatalf("assets/active deben ser []: %s", rr.Body.String())
	}
}

func testReportServerAt(t *testing.T, now time.Time) *Server {
	t.Helper()
	return New(testStore(t), testStaticFS(), Options{Logger: log.New(io.Discard, "", 0), Now: func() time.Time { return now }})
}

func TestReportHistoryCurrent_MesAberto(t *testing.T) {
	s := testReportServerAt(t, time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC))
	assetID, err := s.store.InsertAsset(domain.Asset{Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 1, Year: 2026})
	if err != nil {
		t.Fatalf("InsertAsset: %v", err)
	}
	if _, err := s.store.InsertMonthlyResult(domain.MonthlyResult{AssetID: assetID, ResultUSD: 1200, Month: 4, Year: 2026}); err != nil {
		t.Fatalf("InsertMonthlyResult: %v", err)
	}
	if _, err := s.store.InsertTransaction(domain.Transaction{AssetID: assetID, AmountUSD: 100, Month: 5, Year: 2026}); err != nil {
		t.Fatalf("InsertTransaction: %v", err)
	}
	if _, err := s.store.InsertDividend(domain.Dividend{AmountUSD: 25, Month: 5, Year: 2026}); err != nil {
		t.Fatalf("InsertDividend: %v", err)
	}

	h := s.Handler()
	cookie := authCookie(t, s)

	assetRR := doJSON(t, h, http.MethodGet, "/api/reports/asset/"+jsonID(assetID)+"/history", nil, cookie)
	if assetRR.Code != http.StatusOK {
		t.Fatalf("asset/history: código=%d corpo=%s", assetRR.Code, assetRR.Body.String())
	}
	var assetResp assetHistoryResponse
	decodeBody(t, assetRR, &assetResp)
	if assetResp.Current == nil {
		t.Fatal("asset/history sen current")
	}
	pretoWeb(t, assetResp.Current.Aporte, 100)
	pretoWeb(t, assetResp.Current.Holding, 1300)
	if assetResp.Current.HasMetrics {
		t.Fatalf("asset current non debe ter métricas: %#v", assetResp.Current)
	}

	typeRR := doJSON(t, h, http.MethodGet, "/api/reports/type/accion/history", nil, cookie)
	if typeRR.Code != http.StatusOK {
		t.Fatalf("type/history: código=%d corpo=%s", typeRR.Code, typeRR.Body.String())
	}
	var typeResp typeHistoryResponse
	decodeBody(t, typeRR, &typeResp)
	if typeResp.Current == nil {
		t.Fatal("type/history sen current")
	}
	pretoWeb(t, typeResp.Current.Aporte, 100)
	pretoWeb(t, typeResp.Current.Holding, 1300)

	totalRR := doJSON(t, h, http.MethodGet, "/api/reports/total/history", nil, cookie)
	if totalRR.Code != http.StatusOK {
		t.Fatalf("total/history: código=%d corpo=%s", totalRR.Code, totalRR.Body.String())
	}
	var totalResp totalHistoryResponse
	decodeBody(t, totalRR, &totalResp)
	if totalResp.Current == nil {
		t.Fatal("total/history sen current")
	}
	pretoWeb(t, totalResp.Current.Aporte, 75)
	pretoWeb(t, totalResp.Current.Fondos, 1300)
	pretoWeb(t, totalResp.Current.Dividends, 25)
	if totalResp.Current.HasMetrics || totalResp.Current.Result != 0 || totalResp.Current.Gain != 0 || totalResp.Current.GainPct != 0 {
		t.Fatalf("total current non debe ter métricas: %#v", totalResp.Current)
	}
}

func TestReportHistoryCurrent_NullCandoMesXaTenResultados(t *testing.T) {
	s := testReportServerAt(t, time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC))
	assetID, err := s.store.InsertAsset(domain.Asset{Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 1, Year: 2026})
	if err != nil {
		t.Fatalf("InsertAsset: %v", err)
	}
	if _, err := s.store.InsertMonthlyResult(domain.MonthlyResult{AssetID: assetID, ResultUSD: 1200, Month: 5, Year: 2026}); err != nil {
		t.Fatalf("InsertMonthlyResult: %v", err)
	}

	h := s.Handler()
	cookie := authCookie(t, s)

	for _, tc := range []struct {
		path string
		into any
	}{
		{"/api/reports/asset/" + jsonID(assetID) + "/history", &assetHistoryResponse{}},
		{"/api/reports/type/accion/history", &typeHistoryResponse{}},
		{"/api/reports/total/history", &totalHistoryResponse{}},
	} {
		rr := doJSON(t, h, http.MethodGet, tc.path, nil, cookie)
		if rr.Code != http.StatusOK {
			t.Fatalf("%s: código=%d corpo=%s", tc.path, rr.Code, rr.Body.String())
		}
		if !bytes.Contains(rr.Body.Bytes(), []byte(`"current":null`)) {
			t.Fatalf("%s: current debe ser null: %s", tc.path, rr.Body.String())
		}
		decodeBody(t, rr, tc.into)
	}
}

func pretoWeb(t *testing.T, got, want float64) {
	t.Helper()
	if got < want-0.01 || got > want+0.01 {
		t.Fatalf("got %.4f, want %.4f", got, want)
	}
}

func jsonID(id int64) string {
	return strconv.FormatInt(id, 10)
}
