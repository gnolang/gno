package keyscli

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGnowebURLFromRemote(t *testing.T) {
	t.Parallel()

	t.Run("empty remote", func(t *testing.T) {
		t.Parallel()
		got := GnowebURLFromRemote("", "gno.land/r/demo/counter")
		if got != "" {
			t.Errorf("expected empty, got %q", got)
		}
	})

	t.Run("empty pkgPath", func(t *testing.T) {
		t.Parallel()
		got := GnowebURLFromRemote("127.0.0.1:26657", "")
		if got != "" {
			t.Errorf("expected empty, got %q", got)
		}
	})

	t.Run("unreachable remote returns empty", func(t *testing.T) {
		t.Parallel()
		got := GnowebURLFromRemote("192.0.2.1:26657", "gno.land/r/demo/counter")
		if got != "" {
			t.Errorf("expected empty for unreachable host, got %q", got)
		}
	})
}

func TestIsGnowebReachable(t *testing.T) {
	t.Parallel()

	t.Run("gnoweb instance", func(t *testing.T) {
		t.Parallel()
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `<html><title>gno.land - /r/gnoland/home</title></html>`)
		}))
		defer srv.Close()

		if !isGnowebReachable(srv.URL) {
			t.Error("expected reachable gnoweb")
		}
	})

	t.Run("non-gnoweb server", func(t *testing.T) {
		t.Parallel()
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `<html><title>Some Other App</title></html>`)
		}))
		defer srv.Close()

		if isGnowebReachable(srv.URL) {
			t.Error("expected non-gnoweb server to be rejected")
		}
	})

	t.Run("unreachable", func(t *testing.T) {
		t.Parallel()
		if isGnowebReachable("http://192.0.2.1:9999") {
			t.Error("expected unreachable")
		}
	})
}
