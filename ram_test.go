package plugin_ramcache

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

const testCacheKey = "GETlocalhost:8080/test/path"

func TestRAMCache(t *testing.T) {
	rc, err := newRAMCache(30)
	if err != nil {
		t.Errorf("unexpected newRAMCache error: %v", err)
	}
	_, found := rc.Get(testCacheKey)
	if found {
		t.Errorf("unexpected cache content")
	}

	cacheContent := []byte("some random cache content that should be exact")

	err = rc.Set(testCacheKey, cacheContent, 3)
	if err != nil {
		t.Errorf("unexpected cache set error: %v", err)
	}
	// cache get
	got, found := rc.Get(testCacheKey)
	if !found {
		t.Errorf("unexpected cache get error: %v", err)
	}

	if !bytes.Equal(got, cacheContent) {
		t.Errorf("unexpected cache content: want %s, got %s", cacheContent, got)
	}

	// cache expired
	time.Sleep(3 * time.Second)
	got, found = rc.Get(testCacheKey)
	if found {
		t.Errorf("should miss when cache expired! but got: %s", got)
	}
}

func TestRAMCache_ConcurrentAccess(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	defer func() {
		if r := recover(); r != nil {
			t.Fatal(r)
		}
	}()

	rc, err := newRAMCache(30)
	if err != nil {
		t.Errorf("unexpected newRAMCache error: %v", err)
	}

	cacheContent := []byte("some random cache content that should be exact")

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()

		for {
			got, _ := rc.Get(testCacheKey)
			if got != nil && !bytes.Equal(got, cacheContent) {
				panic(fmt.Errorf("unexpected cache content: want %s, got %s", cacheContent, got))
			}

			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}()

	go func() {
		defer wg.Done()

		for {
			err = rc.Set(testCacheKey, cacheContent, 30)
			if err != nil {
				panic(fmt.Errorf("unexpected cache set error: %w", err))
			}

			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}()

	wg.Wait()
}

func BenchmarkRAMCache_Get(b *testing.B) {

	rc, err := newRAMCache(30)
	if err != nil {
		b.Errorf("unexpected newRAMCache error: %v", err)
	}

	_ = rc.Set(testCacheKey, []byte("some random cache content that should be exact"), 30)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = rc.Get(testCacheKey)
	}
}
