// Package plugin-ramcache is a plugin to cache responses to disk.
package plugin_ramcache

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/pquerna/cachecontrol"
)

// Config configures the middleware.
type Config struct {
	MaxExpiry          int      `json:"maxExpiry" yaml:"maxExpiry" toml:"maxExpiry"`
	RefreshTime        int      `json:"refreshTime" yaml:"refreshTime" toml:"refreshTime"`
	AddStatusHeader    bool     `json:"addStatusHeader" yaml:"addStatusHeader" toml:"addStatusHeader"`
	CacheQueryParams   bool     `json:"cacheQueryParams" yaml:"cacheQueryParams" toml:"cacheQueryParams"`
	ForceNoCacheHeader bool     `json:"forceNoCacheHeader" yaml:"forceNoCacheHeader" toml:"forceNoCacheHeader"`
	BlacklistedHeaders []string `json:"blacklistedHeaders" yaml:"blacklistedHeaders" toml:"blacklistedHeaders"`
}

// CreateConfig returns a config instance.
func CreateConfig() *Config {
	return &Config{
		MaxExpiry:          int((5 * time.Minute).Seconds()),
		RefreshTime:        5,
		AddStatusHeader:    true,
		CacheQueryParams:   false,
		ForceNoCacheHeader: false,
		BlacklistedHeaders: []string{},
	}
}

const (
	cacheHeader      = "Cache-Status"
	cacheHitStatus   = "hit"
	cacheMissStatus  = "miss"
	cacheErrorStatus = "error"
)

type cacheHandler struct {
	name  string
	cache *ramCache
	cfg   *Config
	next  http.Handler
}

// New returns a plugin instance.
func New(_ context.Context, next http.Handler, cfg *Config, name string) (http.Handler, error) {
	if cfg.MaxExpiry <= 1 {
		return nil, errors.New("MaxExpiry must be greater than 1")
	}

	rc, err := newRAMCache(cfg.MaxExpiry)
	if err != nil {
		return nil, err
	}

	m := &cacheHandler{
		name:  name,
		cache: rc,
		cfg:   cfg,
		next:  next,
	}

	return m, nil
}

type cacheData struct {
	Status  int
	Headers map[string][]string
	Body    []byte
}

// ServeHTTP serves an HTTP request.
func (m *cacheHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cs := cacheMissStatus

	for _, header := range m.cfg.BlacklistedHeaders {
		if r.Header.Get(header) != "" {
			rw := &responseWriter{ResponseWriter: w}
			m.next.ServeHTTP(rw, r)

			return
		}
	}

	key := m.cacheKey(r)

	b, found := m.cache.Get(key)
	// cache hit
	if found {
		var data cacheData

		err := json.Unmarshal(b, &data)
		if err != nil {
			cs = cacheErrorStatus
		} else {
			for key, vals := range data.Headers {
				for _, val := range vals {
					w.Header().Add(key, val)
				}
			}
			if m.cfg.AddStatusHeader {
				w.Header().Set(cacheHeader, cacheHitStatus)
			}

			if m.cfg.ForceNoCacheHeader {
				w.Header().Set("Cache-Control", "no-cache")
			}

			w.WriteHeader(data.Status)
			_, _ = w.Write(data.Body)

			return
		}
	}
	// cache miss
	if m.cfg.AddStatusHeader {
		w.Header().Set(cacheHeader, cs)
	}

	// to next middleware
	rw := &responseWriter{ResponseWriter: w}
	m.next.ServeHTTP(rw, r)

	expiry, ok := m.cacheable(r, w, rw.status)
	if !ok {
		return
	}

	data := cacheData{
		Status:  rw.status,
		Headers: w.Header(),
		Body:    rw.body,
	}

	b, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error serializing cache item: %v", err)
	}

	if err = m.cache.Set(key, b, expiry); err != nil {
		log.Printf("Error setting cache item: %v", err)
	}
}

func (m *cacheHandler) cacheable(r *http.Request, w http.ResponseWriter, status int) (int, bool) {
	reasons, expireBy, err := cachecontrol.CachableResponseWriter(r, status, w, cachecontrol.Options{})
	if err != nil || len(reasons) > 0 {
		return 0, false
	}

	expiry := time.Until(expireBy)
	maxExpiry := time.Duration(m.cfg.MaxExpiry) * time.Second

	if maxExpiry < expiry {
		expiry = maxExpiry
	}

	return int(expiry / time.Second), true
}

func (m *cacheHandler) cacheKey(r *http.Request) string {
	if m.cfg.CacheQueryParams {
		return r.Method + r.Host + r.URL.Path + r.URL.RawQuery
	}

	return r.Method + r.Host + r.URL.Path
}

type responseWriter struct {
	http.ResponseWriter
	status int
	body   []byte
}

func (rw *responseWriter) Header() http.Header {
	return rw.ResponseWriter.Header()
}

func (rw *responseWriter) Write(p []byte) (int, error) {
	rw.body = append(rw.body, p...)
	return rw.ResponseWriter.Write(p)
}

func (rw *responseWriter) WriteHeader(s int) {
	rw.status = s
	rw.ResponseWriter.WriteHeader(s)
}
