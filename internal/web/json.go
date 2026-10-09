package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"invest-tracker/internal/domain"
)

type apiError struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if status == http.StatusNoContent {
		w.WriteHeader(status)
		return
	}
	// Serialízase antes de escribir a cabeceira: se falla (p.e. un importe
	// infinito ou NaN), mellor un 500 explícito ca un 200 co corpo baleiro.
	body, err := json.Marshal(v)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"erro interno"}` + "\n"))
		return
	}
	w.WriteHeader(status)
	_, _ = w.Write(append(body, '\n'))
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, apiError{Error: msg})
}

func writeFieldErrors(w http.ResponseWriter, fields map[string]string) {
	writeJSON(w, http.StatusBadRequest, apiError{Error: "datos non válidos", Fields: fields})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		var maxErr *http.MaxBytesError
		switch {
		case errors.As(err, &maxErr):
			return fmt.Errorf("o corpo JSON supera o tamaño máximo de 1 MiB")
		case errors.Is(err, io.EOF):
			return fmt.Errorf("o corpo JSON non pode estar baleiro")
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			field := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("campo descoñecido %s", field)
		default:
			return fmt.Errorf("JSON non válido: %w", err)
		}
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return fmt.Errorf("o corpo JSON debe conter un único valor")
		}
		return fmt.Errorf("JSON non válido: %w", err)
	}
	return nil
}

func pathID(r *http.Request, name string) (int64, error) {
	raw := r.PathValue(name)
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("identificador non válido")
	}
	return id, nil
}

func queryYearMonth(r *http.Request) (domain.YearMonth, map[string]string) {
	q := r.URL.Query()
	fields := map[string]string{}
	year, err := strconv.Atoi(q.Get("year"))
	if err != nil || !domain.ValidYear(year) {
		fields["year"] = fmt.Sprintf("o ano debe estar entre %d e %d", domain.MinYear, domain.MaxYear)
	}
	month, err := strconv.Atoi(q.Get("month"))
	if err != nil || !domain.ValidMonth(month) {
		fields["month"] = "o mes debe estar entre 1 e 12"
	}
	if len(fields) > 0 {
		return domain.YearMonth{}, fields
	}
	return domain.YearMonth{Year: year, Month: month}, nil
}
