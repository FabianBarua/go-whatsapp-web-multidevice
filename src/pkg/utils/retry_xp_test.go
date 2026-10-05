package utils

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestDownloadImageRetriesOn404(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) <= 2 {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Write([]byte("\x89PNG\r\n\x1a\nfake"))
	}))
	defer srv.Close()

	data, name, err := DownloadImageFromURL(srv.URL + "/foto.png")
	if err != nil {
		t.Fatalf("esperaba éxito tras reintentos, error: %v", err)
	}
	if name != "foto.png" || len(data) == 0 || calls != 3 {
		t.Fatalf("name=%s len=%d calls=%d", name, len(data), calls)
	}
}

func TestDownloadImageGivesUpAfterRetries(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		http.NotFound(w, r)
	}))
	defer srv.Close()

	if _, _, err := DownloadImageFromURL(srv.URL + "/foto.png"); err == nil {
		t.Fatal("esperaba error con 404 permanente")
	}
	if calls != 4 {
		t.Fatalf("esperaba 4 intentos, hubo %d", calls)
	}
}
