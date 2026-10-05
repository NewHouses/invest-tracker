package web

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func strconvFormatInt(id int64) string {
	return strconv.FormatInt(id, 10)
}

func assertFieldErrors(t *testing.T, rr *httptest.ResponseRecorder, keys ...string) {
	t.Helper()
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("código=%d corpo=%s, esperabamos 400", rr.Code, rr.Body.String())
	}
	var got apiError
	decodeBody(t, rr, &got)
	if got.Error != "datos non válidos" {
		t.Fatalf("erro inesperado: %#v", got)
	}
	for _, key := range keys {
		if got.Fields[key] == "" {
			t.Fatalf("falta erro para %s en %#v", key, got.Fields)
		}
	}
}
