package go_cache

import (
	"context"
	"log"
	"time"
)

type WriteThroughCache[T any] struct {
	Cache
	StoreFunc func(ctx context.Context, key string, val T) error
}

func (w *WriteThroughCache[T]) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	err := w.StoreFunc(ctx, key, val)
	go func() {
		er := w.Cache.Set(ctx, key, val, expiration)
		if er != nil {
			log.Fatalln(er)
		}
	}()
	return err
}
