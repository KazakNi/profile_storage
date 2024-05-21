package storage

import "sync"

type InMemoryStorage struct {
	sync.RWMutex
	Storage map[string][]byte
}
