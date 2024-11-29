package go_cache

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/singleflight"
	"time"
)

type ReadThroughCache[T any] struct {
	Cache
	LoadFunc   func(ctx context.Context, key string) (any, error)
	Expiration time.Duration

	g singleflight.Group
}

func (r *ReadThroughCache[T]) Get(ctx context.Context, key string) (any, error) {
	val, err := r.Cache.Get(ctx, key)
	if errors.Is(err, ErrKeyNotExists) {
		val, err, _ = r.g.Do(key, func() (any, error) {
			v, er := r.LoadFunc(ctx, key)
			if er == nil {
				er = r.Cache.Set(ctx, key, val, r.Expiration)
				if er != nil {
					return v.(T), fmt.Errorf("%w, 原因：%s", ErrFailedToRefreshCache, er.Error())
				}
				return v.(T), er
			} else {
				return nil, er
			}
		})
	}
	return val.(T), err
}
