package go_cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var _ Cache = (*LocalCache)(nil)

type BuildInMapCacheOption func(cache *LocalCache)

type LocalCache struct {
	data map[string]*item

	mutex     sync.RWMutex
	close     chan struct{}
	onEvicted func(key string, val any)
}

func NewLocalCache(interval time.Duration, opts ...BuildInMapCacheOption) *LocalCache {
	res := &LocalCache{
		data:  make(map[string]*item, 100),
		close: make(chan struct{}),
		onEvicted: func(key string, val any) {

		},
	}

	for _, opt := range opts {
		opt(res)
	}

	go func() {
		ticker := time.NewTicker(interval)
		for {
			select {
			case t := <-ticker.C:
				res.mutex.Lock()
				i := 0
				for key, val := range res.data {
					if i > 10000 {
						break
					}
					if val.deadlineBefore(t) {
						res.delete(key)
					}
					i++
				}
				res.mutex.Unlock()
			case <-res.close:
				return
			}
		}
	}()

	return res
}

func BuildLocalCacheWithEvictedCallback(fn func(key string, val any)) BuildInMapCacheOption {
	return func(cache *LocalCache) {
		cache.onEvicted = fn
	}
}

func (b *LocalCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.set(key, val, expiration)
}

func (b *LocalCache) set(key string, val any, expiration time.Duration) error {
	var dl time.Time
	if expiration > 0 {
		dl = time.Now().Add(expiration)
	}
	b.data[key] = &item{
		val:      val,
		deadline: dl,
	}
	return nil
}

func (b *LocalCache) Get(ctx context.Context, key string) (any, error) {
	b.mutex.RLock()
	res, ok := b.data[key]
	b.mutex.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%w, key: %s", ErrKeyNotExists, key)
	}
	now := time.Now()
	if res.deadlineBefore(now) {
		b.mutex.Lock()
		defer b.mutex.Unlock()
		res, ok = b.data[key]
		if !ok {
			return nil, fmt.Errorf("%w, key: %s", ErrKeyNotExists, key)
		}
		if res.deadlineBefore(now) {
			b.delete(key)
			return nil, fmt.Errorf("%w, key: %s", ErrKeyNotExists, key)
		}
	}
	return res.val, nil
}

func (b *LocalCache) Delete(ctx context.Context, key string) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.delete(key)
	return nil
}

func (b *LocalCache) LoadAndDelete(ctx context.Context, key string) (any, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	val, ok := b.data[key]
	if !ok {
		return nil, ErrKeyNotExists
	}
	b.delete(key)
	return val.val, nil
}

func (b *LocalCache) delete(key string) {
	itm, ok := b.data[key]
	if !ok {
		return
	}
	delete(b.data, key)
	b.onEvicted(key, itm.val)
}

func (b *LocalCache) Close() error {
	select {
	case b.close <- struct{}{}:
	default:
		return errors.New("重复关闭")
	}
	return nil
}

type item struct {
	val      any
	deadline time.Time
}

func (i *item) deadlineBefore(t time.Time) bool {
	return !i.deadline.IsZero() && i.deadline.Before(t)
}
