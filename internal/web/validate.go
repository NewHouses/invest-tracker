package web

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/i18n"
)

const maxAssetNameRunes = 100

func validateName(fields map[string]string, lang i18n.Lang, key, value string) string {
	name := strings.TrimSpace(value)
	switch {
	case name == "":
		fields[key] = lang.T("o nome non pode estar baleiro")
	case utf8.RuneCountInString(name) > maxAssetNameRunes:
		fields[key] = lang.T("o nome non pode superar 100 caracteres")
	}
	return name
}

func validateAmount(fields map[string]string, lang i18n.Lang, key string, value float64) {
	if !domain.ValidAmount(value) {
		fields[key] = lang.T("o importe debe ser un número maior ca 0")
	}
}

func validateYearMonth(fields map[string]string, lang i18n.Lang, yearKey, monthKey string, year, month int) {
	if !domain.ValidYear(year) {
		fields[yearKey] = lang.Sprintf("o ano debe estar entre %d e %d", domain.MinYear, domain.MaxYear)
	}
	if !domain.ValidMonth(month) {
		fields[monthKey] = lang.T("o mes debe estar entre 1 e 12")
	}
}

func validateDateNotBefore(fields map[string]string, lang i18n.Lang, monthKey string, got, start domain.YearMonth) {
	if got.Valid() && got.Before(start) {
		fields[monthKey] = lang.Sprintf("a transacción non pode ser anterior á data do activo (%02d/%d)", start.Month, start.Year)
	}
}

func signedTransactionAmount(kind string, amount float64) (float64, bool) {
	switch kind {
	case "compra":
		return amount, true
	case "venda":
		return -amount, true
	}
	return 0, false
}

func (s *Server) bodyAsset(w http.ResponseWriter, r *http.Request, fields map[string]string, key string, id int64) (domain.Asset, bool, bool) {
	lang := langOf(r)
	if id <= 0 {
		fields[key] = lang.T("o activo non existe")
		return domain.Asset{}, false, false
	}
	asset, err := s.store.GetAsset(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fields[key] = lang.T("o activo non existe")
			return domain.Asset{}, false, false
		}
		s.internalError(w, r, err)
		return domain.Asset{}, false, true
	}
	return asset, true, false
}
