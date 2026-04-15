package ownmap

import (
	"sync"
	"time"
)

// TODO: надо сделать интерфейс мапы, интерфейс KeyValue
// надо подумать как сделать KeyValue без ttl
// интерфейс должен совпадать с базовой мапой go, с мапой
// го для параллельных вычислений
// в .env нужно сделать параметр выбора типа мапы
type OwnMap struct {
	mu            sync.RWMutex
	keys          []int
	values        [][]KeyValue
	capacity      int
	cleaner       *time.Ticker
	default_value string
}

type KeyValue struct {
	key       string
	value     string
	expiresAt time.Time
}

func NewKeyValue(key string, value string, expires_at time.Time) KeyValue {
	return KeyValue{
		key:       key,   // len=0, cap=capacity
		value:     value, // len=0, cap=capacity
		expiresAt: expires_at,
	}
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

func (om *OwnMap) Set(key string, value string, expires_at time.Time) {
	om.mu.Lock()
	defer om.mu.Unlock()
	inner_key := hashString(key, om.capacity)
	om.keys[inner_key] = inner_key
	for index, k_v := range om.values[inner_key] {
		if key == k_v.key {
			om.values[inner_key][index].value = value
			om.values[inner_key][index].expiresAt = expires_at
			return
		}
	}
	om.values[inner_key] = append(om.values[inner_key], NewKeyValue(key, value, expires_at))
}

func (om *OwnMap) Get(key string) string {
	om.mu.RLock()
	defer om.mu.RUnlock()
	inner_key := hashString(key, om.capacity)
	for _, v := range om.values[inner_key] {
		if v.key == key {
			return v.value
		}
	}
	return om.default_value
}

func (om *OwnMap) Remove(key string) {
	om.mu.Lock()
	defer om.mu.Unlock()
	inner_key := hashString(key, om.capacity)
	for i, v := range om.values[inner_key] {
		if v.key == key {
			om.values[inner_key][i] = KeyValue{}
		}
	}
}

func (om *OwnMap) Keys() []string {
	om.mu.RLock()
	defer om.mu.RUnlock()
	keys := make([]string, 0, om.capacity)
	for _, v := range om.values {
		for _, k_v := range v {
			keys = append(keys, k_v.key)
		}
	}
	return keys
}

func (om *OwnMap) Values() []string {
	om.mu.RLock()
	defer om.mu.RUnlock()
	values := make([]string, 0, om.capacity)
	for _, v := range om.values {
		for _, k_v := range v {
			values = append(values, k_v.value)
		}
	}
	return values
}

func (om *OwnMap) Items() []KeyValue {
	om.mu.RLock()
	defer om.mu.RUnlock()
	values := make([]KeyValue, 0, om.capacity)
	for _, v := range om.values {
		for _, k_v := range v {
			values = append(values, k_v)
		}
	}
	return values
}

func hashString(s string, capacity int) int {
	var h uint = 5381 // djb2 алгоритм

	for i := 0; i < len(s); i++ {
		h = h*33 + uint(s[i])
	}

	return int(h % uint(capacity))
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
	for _, k_v := range om.Items() {
		if !k_v.expiresAt.IsZero() && k_v.expiresAt.Before(now) {
			om.Remove(k_v.key)
		}
	}
}
