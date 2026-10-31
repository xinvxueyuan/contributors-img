package appcache

import (
	"context"
	"sync"
	"testing"
)

// One instance serves many concurrent requests, so every access has to go
// through the mutex. Before this test existed the cache was two bare maps and
// a concurrent write crashed the process.
func TestMemoryCacheIsSafeForConcurrentUse(t *testing.T) {
	cache := newMemoryCacheWithLimit(64)
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				key := string(rune('a' + i))
				_ = cache.Save(ctx, key, []byte("data"), "image/svg+xml")
				_, _ = cache.Get(ctx, key)
				_ = cache.SaveJSON(ctx, key, map[string]int{"i": i})
				var v map[string]int
				_ = cache.GetJSON(ctx, key, &v)
			}
		}(i)
	}
	wg.Wait()
}

func TestMemoryCacheEvictsLeastRecentlyUsed(t *testing.T) {
	cache := newMemoryCacheWithLimit(2)
	ctx := context.Background()

	_ = cache.Save(ctx, "a", []byte("a"), "image/svg+xml")
	_ = cache.Save(ctx, "b", []byte("b"), "image/svg+xml")

	if h, _ := cache.Get(ctx, "a"); h == nil {
		t.Fatal("expected a to be cached")
	}

	_ = cache.Save(ctx, "c", []byte("c"), "image/svg+xml")

	if h, _ := cache.Get(ctx, "b"); h != nil {
		t.Fatal("expected b (least recently used) to be evicted")
	}
	if h, _ := cache.Get(ctx, "a"); h == nil {
		t.Fatal("expected a to survive: it was used after b")
	}
	if h, _ := cache.Get(ctx, "c"); h == nil {
		t.Fatal("expected c to be cached")
	}
}

func TestMemoryCacheHonoursTheLimit(t *testing.T) {
	cache := newMemoryCacheWithLimit(3)
	ctx := context.Background()

	for _, k := range []string{"a", "b", "c", "d", "e"} {
		_ = cache.Save(ctx, k, []byte(k), "image/svg+xml")
	}

	cache.mu.Lock()
	size := len(cache.items)
	cache.mu.Unlock()

	if size != 3 {
		t.Fatalf("expected at most 3 entries, got %d", size)
	}
}

func TestMemoryCacheRoundTripsJSON(t *testing.T) {
	cache := newMemoryCache()
	ctx := context.Background()

	want := map[string]string{"owner": "angular", "repo": "angular-ja"}
	if err := cache.SaveJSON(ctx, "repo", want); err != nil {
		t.Fatalf("SaveJSON: %v", err)
	}

	var got map[string]string
	if err := cache.GetJSON(ctx, "repo", &got); err != nil {
		t.Fatalf("GetJSON: %v", err)
	}
	if got["owner"] != want["owner"] || got["repo"] != want["repo"] {
		t.Fatalf("expected %v, got %v", want, got)
	}

	// A miss must leave the destination untouched instead of erroring.
	var miss map[string]string
	if err := cache.GetJSON(ctx, "absent", &miss); err != nil {
		t.Fatalf("GetJSON on a miss returned %v", err)
	}
	if miss != nil {
		t.Fatalf("expected the miss to be nil, got %v", miss)
	}
}
