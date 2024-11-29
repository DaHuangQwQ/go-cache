package go_cache

import (
	"errors"
)

var (
	ErrKeyNotExists = errors.New("go-cache: key not exists")

	ErrRedisSetFail = errors.New("go-cache: redis set fail")

	ErrFailedToRefreshCache = errors.New("go-cache: failed to refresh cache")
)
