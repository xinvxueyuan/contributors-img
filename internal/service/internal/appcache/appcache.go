package appcache

import (
	"context"

	"contrib.rocks/pkg/model"
)

type AppCache interface {
	Get(c context.Context, name string) (model.FileHandle, error)
	GetJSON(c context.Context, name string, v any) error
	Save(c context.Context, name string, data []byte, contentType string) error
	SaveJSON(c context.Context, name string, v any) error
}

func NewMemoryCache() AppCache {
	return newMemoryCache()
}
