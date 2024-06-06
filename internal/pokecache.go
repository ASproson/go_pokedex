package internal

import (
	"sync"
	"time"
)

type Cache struct {
	mu    sync.Mutex
	cache map[string]cacheEntry
}

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}
