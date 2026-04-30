package addresultbatch_test

import (
	"bufio"
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"invest-tracker/internal/addresultbatch"
	"invest-tracker/internal/domain"
	"invest-tracker/internal/store"
)

func TestRun_EndToEnd_AddsSeriesOfResults(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	assetID, err := s.InsertAsset(domain.Asset{
		Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 3, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertAsset: %v", err)
	}

	// asset=1, mes inicial=3 ano=2026
	// 03/2026: holding=1000, resultado 1100, sí
	// 04/2026: holding=1100, resultado 1200, sí
	// 05/2026: holding=1200, resultado 1300, n
	r := bufio.NewReader(strings.NewReader("1\n3\n2026\n1100\ns\n1200\ns\n1300\nn\n"))
	var out bytes.Buffer
	if err := addresultbatch.Run(r, &out, s); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	s2, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("re-Open: %v", err)
	}
	t.Cleanup(func() { _ = s2.Close() })

	results, err := s2.ListMonthlyResultsByAsset(assetID)
	if err != nil {
		t.Fatalf("ListMonthlyResultsByAsset: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("got %d resultados, esperabamos 3", len(results))
	}

	want := []struct {
		amount      float64
		year, month int
	}{
		{1100, 2026, 3},
		{1200, 2026, 4},
		{1300, 2026, 5},
	}
	for i, w := range want {
		if results[i].ResultUSD != w.amount || results[i].Year != w.year || results[i].Month != w.month {
			t.Errorf("results[%d] = %+v, esperabamos %+v", i, results[i], w)
		}
	}

	output := out.String()
	if !strings.Contains(output, "3 resultado(s) e saltáronse 0") {
		t.Errorf("saída non contén resumo correcto:\n%s", output)
	}
}
