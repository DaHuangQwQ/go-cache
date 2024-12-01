package go_cache

import (
	"context"
)

type BloomFilterCache[T any] struct {
	ReadThroughCache[T]
}

func NewBloomFilterCache[T any](cache Cache, bf BloomFilter,
	loadFunc func(ctx context.Context, key string) (any, error)) *BloomFilterCache[T] {
	return &BloomFilterCache[T]{
		ReadThroughCache: ReadThroughCache[T]{
			Cache: cache,
			LoadFunc: func(ctx context.Context, key string) (any, error) {
				if !bf.HasKey(ctx, key) {
					return nil, ErrKeyNotExists
				}
				return loadFunc(ctx, key)
			},
		},
	}
}

type BloomFilter interface {
	HasKey(ctx context.Context, key string) bool
}
