package ownmap

import (
	"sync"
	"time"
)

type Map interface {
	Set(key, value string, expiresAt time.Time)
	Get(key string) string
	Remove(key string)
	Keys() []string
	Values() []string
}

type KeyValueItem interface {
	Key() string
	Value() string
}

type KeyValue struct {
	key       string
	value     string
	expiresAt time.Time
}

func (kv KeyValue) Key() string   { return kv.key }
func (kv KeyValue) Value() string { return kv.value }

func NewKeyValue(key, value string, expiresAt time.Time) KeyValue {
	return KeyValue{key: key, value: value, expiresAt: expiresAt}
}

// OwnMap — хэш-мапа с TTL
type OwnMap struct {
	mu       sync.RWMutex
	keys     []int
	values   [][]KeyValue
	capacity int
	cleaner  *time.Ticker
}

func NewOwnMap(capacity int) *OwnMap {
	om := &OwnMap{
		keys:     make([]int, capacity),
		values:   make([][]KeyValue, capacity),
		capacity: capacity,
	}
	om.RunCleaner(100 * time.Millisecond)
	return om
}

func (om *OwnMap) Set(key, value string, expiresAt time.Time) {
	om.mu.Lock()
	defer om.mu.Unlock()
	idx := hashString(key, om.capacity)
	om.keys[idx] = idx
	for i, kv := range om.values[idx] {
		if kv.key == key {
			om.values[idx][i].value = value
			om.values[idx][i].expiresAt = expiresAt
			return
		}
	}
	om.values[idx] = append(om.values[idx], NewKeyValue(key, value, expiresAt))
}

func (om *OwnMap) Get(key string) string {
	om.mu.RLock()
	defer om.mu.RUnlock()
	for _, kv := range om.values[hashString(key, om.capacity)] {
		if kv.key == key {
			return kv.value
		}
	}
	return ""
}

func (om *OwnMap) Remove(key string) {
	om.mu.Lock()
	defer om.mu.Unlock()
	idx := hashString(key, om.capacity)
	for i, kv := range om.values[idx] {
		if kv.key == key {
			om.values[idx][i] = KeyValue{}
		}
	}
}

func (om *OwnMap) Keys() []string {
	om.mu.RLock()
	defer om.mu.RUnlock()
	keys := make([]string, 0, om.capacity)
	for _, bucket := range om.values {
		for _, kv := range bucket {
			if kv.key != "" {
				keys = append(keys, kv.key)
			}
		}
	}
	return keys
}

func (om *OwnMap) Values() []string {
	om.mu.RLock()
	defer om.mu.RUnlock()
	vals := make([]string, 0, om.capacity)
	for _, bucket := range om.values {
		for _, kv := range bucket {
			if kv.key != "" {
				vals = append(vals, kv.value)
			}
		}
	}
	return vals
}

func (om *OwnMap) Items() []KeyValue {
	om.mu.RLock()
	defer om.mu.RUnlock()
	items := make([]KeyValue, 0, om.capacity)
	for _, bucket := range om.values {
		items = append(items, bucket...)
	}
	return items
}

func (om *OwnMap) RunCleaner(interval time.Duration) {
	om.cleaner = time.NewTicker(interval)
	go func() {
		for range om.cleaner.C {
			om.cleanupExpired()
		}
	}()
}

func (om *OwnMap) cleanupExpired() {
	now := time.Now()
	for _, kv := range om.Items() {
		if !kv.expiresAt.IsZero() && kv.expiresAt.Before(now) {
			om.Remove(kv.key)
		}
	}
}

// StdMap — обёртка над sync.Map без TTL (expiresAt игнорируется)
type StdMap struct {
	m sync.Map
}

func NewStdMap() *StdMap { return &StdMap{} }

func (s *StdMap) Set(key, value string, _ time.Time) {
	s.m.Store(key, value)
}

func (s *StdMap) Get(key string) string {
	v, ok := s.m.Load(key)
	if !ok {
		return ""
	}
	return v.(string)
}

func (s *StdMap) Remove(key string) { s.m.Delete(key) }

func (s *StdMap) Keys() []string {
	var keys []string
	s.m.Range(func(k, _ any) bool {
		keys = append(keys, k.(string))
		return true
	})
	return keys
}

func (s *StdMap) Values() []string {
	var vals []string
	s.m.Range(func(_, v any) bool {
		vals = append(vals, v.(string))
		return true
	})
	return vals
}

func hashString(s string, capacity int) int {
	var h uint = 5381
	for i := 0; i < len(s); i++ {
		h = h*33 + uint(s[i])
	}
	return int(h % uint(capacity))
}
