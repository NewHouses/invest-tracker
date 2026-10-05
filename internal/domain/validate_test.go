package domain_test

import (
	"encoding/json"
	"math"
	"reflect"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
)

func TestValidMonth(t *testing.T) {
	cases := []struct {
		month int
		want  bool
	}{
		{0, false},
		{1, true},
		{12, true},
		{13, false},
	}
	for _, c := range cases {
		if got := domain.ValidMonth(c.month); got != c.want {
			t.Errorf("ValidMonth(%d) = %v, esperabamos %v", c.month, got, c.want)
		}
	}
}

func TestValidYear(t *testing.T) {
	cases := []struct {
		year int
		want bool
	}{
		{domain.MinYear - 1, false},
		{domain.MinYear, true},
		{2026, true},
		{domain.MaxYear, true},
		{domain.MaxYear + 1, false},
	}
	for _, c := range cases {
		if got := domain.ValidYear(c.year); got != c.want {
			t.Errorf("ValidYear(%d) = %v, esperabamos %v", c.year, got, c.want)
		}
	}
}

func TestYearMonthValid(t *testing.T) {
	cases := []struct {
		date domain.YearMonth
		want bool
	}{
		{domain.YearMonth{Year: 2026, Month: 4}, true},
		{domain.YearMonth{Year: domain.MinYear - 1, Month: 4}, false},
		{domain.YearMonth{Year: 2026, Month: 0}, false},
		{domain.YearMonth{Year: domain.MaxYear + 1, Month: 13}, false},
	}
	for _, c := range cases {
		if got := c.date.Valid(); got != c.want {
			t.Errorf("%+v.Valid() = %v, esperabamos %v", c.date, got, c.want)
		}
	}
}

func TestValidAmount(t *testing.T) {
	cases := []struct {
		name string
		v    float64
		want bool
	}{
		{"positivo", 1, true},
		{"cero", 0, false},
		{"negativo", -1, false},
		{"NaN", math.NaN(), false},
		{"Inf", math.Inf(1), false},
		{"-Inf", math.Inf(-1), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := domain.ValidAmount(c.v); got != c.want {
				t.Errorf("ValidAmount(%v) = %v, esperabamos %v", c.v, got, c.want)
			}
		})
	}
}

func TestValidNonNegativeAmount(t *testing.T) {
	cases := []struct {
		name string
		v    float64
		want bool
	}{
		{"positivo", 1, true},
		{"cero", 0, true},
		{"negativo", -1, false},
		{"NaN", math.NaN(), false},
		{"Inf", math.Inf(1), false},
		{"-Inf", math.Inf(-1), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := domain.ValidNonNegativeAmount(c.v); got != c.want {
				t.Errorf("ValidNonNegativeAmount(%v) = %v, esperabamos %v", c.v, got, c.want)
			}
		})
	}
}

func TestYearMonthIndexAndBefore(t *testing.T) {
	december := domain.YearMonth{Year: 2026, Month: 12}
	january := domain.YearMonth{Year: 2027, Month: 1}
	if got, want := december.Index(), 2026*domain.MonthsPerYear+11; got != want {
		t.Errorf("Index() = %d, esperabamos %d", got, want)
	}
	if got, want := january.Index(), december.Index()+1; got != want {
		t.Errorf("xaneiro tras decembro Index() = %d, esperabamos %d", got, want)
	}
	if !december.Before(january) {
		t.Errorf("esperabamos que decembro fose anterior a xaneiro")
	}
	if january.Before(december) {
		t.Errorf("non esperabamos que xaneiro fose anterior a decembro")
	}
	if december.Before(december) {
		t.Errorf("un mes non debe ser anterior a si mesmo")
	}
}

func TestAssetStartAndCreatedBy(t *testing.T) {
	asset := domain.Asset{Year: 2026, Month: 12}
	if got, want := asset.Start(), (domain.YearMonth{Year: 2026, Month: 12}); got != want {
		t.Errorf("Start() = %+v, esperabamos %+v", got, want)
	}
	cases := []struct {
		date domain.YearMonth
		want bool
	}{
		{domain.YearMonth{Year: 2026, Month: 11}, false},
		{domain.YearMonth{Year: 2026, Month: 12}, true},
		{domain.YearMonth{Year: 2027, Month: 1}, true},
	}
	for _, c := range cases {
		if got := asset.CreatedBy(c.date); got != c.want {
			t.Errorf("CreatedBy(%+v) = %v, esperabamos %v", c.date, got, c.want)
		}
	}
}

func TestAssetTypesOrderAndCopy(t *testing.T) {
	want := []domain.AssetType{domain.Accion, domain.Indice, domain.CopyTrading, domain.Fondo}
	got := domain.AssetTypes()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("AssetTypes() = %+v, esperabamos %+v", got, want)
	}
	got[0] = domain.Fondo
	if again := domain.AssetTypes(); !reflect.DeepEqual(again, want) {
		t.Errorf("AssetTypes() tras mutar a copia = %+v, esperabamos %+v", again, want)
	}
}

func TestJSONTagsAssetAndYearMonth(t *testing.T) {
	asset := domain.Asset{ID: 1, Type: domain.Accion, Name: "AAPL", AmountUSD: 123.45, Month: 4, Year: 2026}
	data, err := json.Marshal(asset)
	if err != nil {
		t.Fatalf("Marshal(Asset): %v", err)
	}
	text := string(data)
	for _, want := range []string{"\"id\"", "\"type\"", "\"name\"", "\"amountUsd\"", "\"month\"", "\"year\""} {
		if !strings.Contains(text, want) {
			t.Errorf("JSON de Asset non contén %s: %s", want, text)
		}
	}
	if strings.Contains(text, "AmountUSD") {
		t.Errorf("JSON de Asset usa o nome Go en vez de camelCase: %s", text)
	}
	var decodedAsset domain.Asset
	if err := json.Unmarshal(data, &decodedAsset); err != nil {
		t.Fatalf("Unmarshal(Asset): %v", err)
	}
	if decodedAsset != asset {
		t.Errorf("Asset tras round-trip = %+v, esperabamos %+v", decodedAsset, asset)
	}

	date := domain.YearMonth{Year: 2026, Month: 4}
	data, err = json.Marshal(date)
	if err != nil {
		t.Fatalf("Marshal(YearMonth): %v", err)
	}
	text = string(data)
	for _, want := range []string{"\"year\"", "\"month\""} {
		if !strings.Contains(text, want) {
			t.Errorf("JSON de YearMonth non contén %s: %s", want, text)
		}
	}
	if strings.Contains(text, "Year") || strings.Contains(text, "Month") {
		t.Errorf("JSON de YearMonth usa nomes Go en vez de camelCase: %s", text)
	}
	var decodedDate domain.YearMonth
	if err := json.Unmarshal(data, &decodedDate); err != nil {
		t.Fatalf("Unmarshal(YearMonth): %v", err)
	}
	if decodedDate != date {
		t.Errorf("YearMonth tras round-trip = %+v, esperabamos %+v", decodedDate, date)
	}
}
