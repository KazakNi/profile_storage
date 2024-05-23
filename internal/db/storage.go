package storage

import "sync"

type InMemoryStorage struct {
	sync.RWMutex
	Storage map[string][]byte
}

func (i *InMemoryStorage) Get(key string) (value []byte, ok bool) {
	i.RLock()
	value, ok = i.Storage[key]
	i.RUnlock()

	return value, ok
}

func (i *InMemoryStorage) Set(key string, value []byte) {
	i.Lock()
	i.Storage[key] = value
	i.Unlock()
}

func (i *InMemoryStorage) GetUsers() (users [][]byte) {
	i.RLock()
	for _, v := range i.Storage {
		users = append(users, v)
	}
	i.RUnlock()

	return users
}

func (i *InMemoryStorage) Delete(key string) {
	i.Lock()
	delete(i.Storage, key)
	i.Unlock()
}
