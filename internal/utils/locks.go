package utils

import (
	"sync"
	"sync/atomic"
	"time"
)

const HandlerTimeout = 5 * time.Second

type KeyedLocker struct {
	m sync.Map // string -> *entry
}

type entry struct {
	mu   sync.Mutex
	refs int64
}

func NewKeyedLocker() *KeyedLocker { return &KeyedLocker{} }

// Lock adquire o lock para a chave e retorna um unlock() para liberar.
func (k *KeyedLocker) Lock(key string) (unlock func()) {
	v, _ := k.m.LoadOrStore(key, &entry{})
	e := v.(*entry)
	atomic.AddInt64(&e.refs, 1)
	e.mu.Lock()

	return func() {
		e.mu.Unlock()
		if atomic.AddInt64(&e.refs, -1) == 0 {
			// melhor esforço: remove a entry quando ninguém mais referencia
			k.m.Delete(key)
		}
	}
}
