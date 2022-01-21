package plugin_ramcache

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name:    "should error if maxExpiry <= 1",
			cfg:     &Config{MaxExpiry: 1},
			wantErr: true,
		},
		{
			name:    "should be valid",
			cfg:     &Config{MaxExpiry: 300},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := New(context.Background(), nil, test.cfg, "ramcache")

			if test.wantErr && err == nil {
				t.Fatal("expected error on bad regexp format")
			}
		})
	}
}

func TestCache_ServeHTTP(t *testing.T) {

	next := func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Cache-Control", "max-age=20")
		rw.WriteHeader(http.StatusOK)
	}

	cfg := &Config{MaxExpiry: 10, AddStatusHeader: true}

	c, err := New(context.Background(), http.HandlerFunc(next), cfg, "ramcache")
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "http://localhost/some/path", nil)
	rw := httptest.NewRecorder()

	c.ServeHTTP(rw, req)

	if state := rw.Header().Get("Cache-Status"); state != "miss" {
		t.Errorf("unexprect cache state: want \"miss\", got: %q", state)
	}

	rw = httptest.NewRecorder()

	c.ServeHTTP(rw, req)

	if state := rw.Header().Get("Cache-Status"); state != "hit" {
		t.Errorf("unexprect cache state: want \"hit\", got: %q", state)
	}
}
