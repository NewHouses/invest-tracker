package web

import (
	"net/http"
	"testing"
)

func TestAPIMeta(t *testing.T) {
	s := testServer(t)
	h := s.Handler()
	if rr := doJSON(t, h, http.MethodGet, "/api/meta", nil, nil); rr.Code != http.StatusUnauthorized {
		t.Fatalf("sen cookie: código=%d, esperabamos 401", rr.Code)
	}
	rr := doJSON(t, h, http.MethodGet, "/api/meta", nil, authCookie(t, s))
	if rr.Code != http.StatusOK {
		t.Fatalf("código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	var got struct {
		AssetTypes []struct {
			Value string `json:"value"`
			Label string `json:"label"`
		} `json:"assetTypes"`
		MinYear int `json:"minYear"`
		MaxYear int `json:"maxYear"`
	}
	decodeBody(t, rr, &got)
	if got.MinYear != 1900 || got.MaxYear != 2100 || len(got.AssetTypes) != 4 {
		t.Fatalf("meta inesperada: %#v", got)
	}
	if got.AssetTypes[0].Value != "accion" || got.AssetTypes[0].Label != "Acción" {
		t.Fatalf("tipo inicial inesperado: %#v", got.AssetTypes[0])
	}
}
