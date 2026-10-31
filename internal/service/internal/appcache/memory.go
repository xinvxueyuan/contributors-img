package appcache

import (
	"bytes"
	"container/list"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"contrib.rocks/pkg/model"
)

var _ AppCache = &memoryCache{}

// DefaultMemoryCacheEntries caps how many entries one process keeps. A rendered
// image inlines every avatar as base64, so an entry is tens to hundreds of
// kilobytes; the same instance also serves concurrent requests, so the cache
// has to be both bounded and safe for concurrent use.
const DefaultMemoryCacheEntries = 300

type memoryCache struct {
	mu    sync.Mutex
	max   int
	ll    *list.List // front is the most recently used entry
	items map[string]*list.Element
}

type memoryEntry struct {
	key  string
	file model.FileHandle
	json []byte
}

func newMemoryCache() *memoryCache {
	return newMemoryCacheWithLimit(DefaultMemoryCacheEntries)
}

func newMemoryCacheWithLimit(max int) *memoryCache {
	return &memoryCache{
		max:   max,
		ll:    list.New(),
		items: make(map[string]*list.Element, max),
	}
}

// get returns a copy of the entry so callers never touch list elements after
// the lock is released.
func (c *memoryCache) get(key string) (memoryEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	el, ok := c.items[key]
	if !ok {
		return memoryEntry{}, false
	}
	c.ll.MoveToFront(el)
	return *el.Value.(*memoryEntry), true
}

func (c *memoryCache) put(entry memoryEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.items[entry.key]; ok {
		el.Value = &entry
		c.ll.MoveToFront(el)
		return
	}
	c.items[entry.key] = c.ll.PushFront(&entry)
	for c.ll.Len() > c.max {
		oldest := c.ll.Back()
		if oldest == nil {
			return
		}
		c.ll.Remove(oldest)
		delete(c.items, oldest.Value.(*memoryEntry).key)
	}
}

func fileKey(name string) string { return "file:" + name }
func jsonKey(name string) string { return "json:" + name }

func (c *memoryCache) Get(ctx context.Context, name string) (model.FileHandle, error) {
	entry, ok := c.get(fileKey(name))
	if !ok {
		return nil, nil
	}
	return entry.file, nil
}

func (c *memoryCache) GetJSON(ctx context.Context, name string, v any) error {
	entry, ok := c.get(jsonKey(name))
	if !ok || entry.json == nil {
		return nil
	}
	return json.Unmarshal(entry.json, v)
}

func (c *memoryCache) Save(ctx context.Context, name string, data []byte, contentType string) error {
	c.put(memoryEntry{key: fileKey(name), file: &memoryFileHandle{data, contentType}})
	return nil
}

func (c *memoryCache) SaveJSON(ctx context.Context, name string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	c.put(memoryEntry{key: jsonKey(name), json: data})
	return nil
}

var _ model.FileHandle = &memoryFileHandle{}

type memoryFileHandle struct {
	data        []byte
	contentType string
}

func (f *memoryFileHandle) ContentType() string {
	return f.contentType
}

func (f *memoryFileHandle) ETag() string {
	return fmt.Sprintf("%x", md5.Sum(f.data))
}

func (f *memoryFileHandle) Reader() io.ReadCloser {
	return io.NopCloser(bytes.NewBuffer(f.data))
}

func (f *memoryFileHandle) Size() int64 {
	return int64(len(f.data))
}
