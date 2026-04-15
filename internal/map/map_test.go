package ownmap

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

var noExpiry = time.Time{}

func TestSetAndGet(t *testing.T) {
	om := NewOwnMap(16)
	om.Set("foo", "bar", noExpiry)

	if got := om.Get("foo"); got != "bar" {
		t.Errorf("Get(foo) = %q, want %q", got, "bar")
	}
}

func TestGetMissingKey(t *testing.T) {
	om := NewOwnMap(16)

	if got := om.Get("missing"); got != "" {
		t.Errorf("Get(missing) = %q, want %q", got, "")
	}
}

func TestSetOverwrite(t *testing.T) {
	om := NewOwnMap(16)
	om.Set("foo", "bar", noExpiry)
	om.Set("foo", "baz", noExpiry)

	if got := om.Get("foo"); got != "baz" {
		t.Errorf("Get(foo) after overwrite = %q, want %q", got, "baz")
	}
	count := 0
	for _, k := range om.Keys() {
		if k == "foo" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 1 entry for foo, got %d", count)
	}
}

func TestKeys(t *testing.T) {
	om := NewOwnMap(16)
	om.Set("a", "1", noExpiry)
	om.Set("b", "2", noExpiry)

	if got := len(om.Keys()); got != 2 {
		t.Errorf("Keys() len = %d, want 2", got)
	}
}

func TestValues(t *testing.T) {
	om := NewOwnMap(16)
	om.Set("a", "1", noExpiry)
	om.Set("b", "2", noExpiry)

	if got := len(om.Values()); got != 2 {
		t.Errorf("Values() len = %d, want 2", got)
	}
}

func TestItems(t *testing.T) {
	om := NewOwnMap(16)
	om.Set("a", "1", noExpiry)
	om.Set("b", "2", noExpiry)

	items := om.Items()
	if len(items) != 2 {
		t.Errorf("Items() len = %d, want 2", len(items))
	}
	for _, item := range items {
		if item.key == "a" && item.value != "1" {
			t.Errorf("item a = %q, want %q", item.value, "1")
		}
		if item.key == "b" && item.value != "2" {
			t.Errorf("item b = %q, want %q", item.value, "2")
		}
	}
}

func TestLargeMap(t *testing.T) {
	const n = 100_000
	om := NewOwnMap(n)

	expected := make(map[string]string, n)
	for i := 0; i < n; i++ {
		key := fmt.Sprintf("key-%d", rand.Int())
		value := fmt.Sprintf("val-%d", rand.Int())
		expected[key] = value
		om.Set(key, value, noExpiry)
	}

	for key, value := range expected {
		if got := om.Get(key); got != value {
			t.Errorf("Get(%q) = %q, want %q", key, got, value)
		}
	}
}

func TestRemove(t *testing.T) {
	om := NewOwnMap(16)
	om.Set("foo", "bar", noExpiry)
	om.Remove("foo")

	if got := om.Get("foo"); got != "" {
		t.Errorf("Get(foo) after Remove = %q, want %q", got, "")
	}
}

func TestRemoveMissingKey(t *testing.T) {
	om := NewOwnMap(16)
	om.Remove("nonexistent") // не должно паниковать
}

func TestCleanupExpired(t *testing.T) {
	om := NewOwnMap(16)
	om.Set("alive", "yes", noExpiry)
	om.Set("dead", "no", time.Now().Add(-1*time.Second))

	om.cleanupExpired()

	if got := om.Get("dead"); got != "" {
		t.Errorf("Get(dead) after cleanup = %q, want empty", got)
	}
	if got := om.Get("alive"); got != "yes" {
		t.Errorf("Get(alive) after cleanup = %q, want %q", got, "yes")
	}
}

func TestRunCleaner(t *testing.T) {
	om := NewOwnMap(16)
	om.Set("expiring", "val", time.Now().Add(50*time.Millisecond))

	time.Sleep(300 * time.Millisecond)

	if got := om.Get("expiring"); got != "" {
		t.Errorf("Get(expiring) after TTL = %q, want empty", got)
	}
}

func TestNewKeyValue(t *testing.T) {
	expiry := time.Now().Add(time.Hour)
	kv := NewKeyValue("k", "v", expiry)
	if kv.key != "k" || kv.value != "v" || !kv.expiresAt.Equal(expiry) {
		t.Errorf("NewKeyValue = {%q, %q, %v}, want {k, v, %v}", kv.key, kv.value, kv.expiresAt, expiry)
	}
}

func TestKeyValueInterface(t *testing.T) {
	kv := NewKeyValue("hello", "world", noExpiry)
	var item KeyValueItem = kv
	if item.Key() != "hello" {
		t.Errorf("Key() = %q, want %q", item.Key(), "hello")
	}
	if item.Value() != "world" {
		t.Errorf("Value() = %q, want %q", item.Value(), "world")
	}
}

// mapTestCase — общий набор тестов для любой реализации Map
func runMapTests(t *testing.T, m Map) {
	t.Helper()

	m.Set("foo", "bar", noExpiry)
	if got := m.Get("foo"); got != "bar" {
		t.Errorf("Get(foo) = %q, want %q", got, "bar")
	}

	// перезапись
	m.Set("foo", "baz", noExpiry)
	if got := m.Get("foo"); got != "baz" {
		t.Errorf("Get(foo) after overwrite = %q, want %q", got, "baz")
	}

	// отсутствующий ключ
	if got := m.Get("missing"); got != "" {
		t.Errorf("Get(missing) = %q, want empty", got)
	}

	// Keys / Values
	m.Set("a", "1", noExpiry)
	m.Set("b", "2", noExpiry)
	if l := len(m.Keys()); l < 2 {
		t.Errorf("Keys() len = %d, want >= 2", l)
	}
	if l := len(m.Values()); l < 2 {
		t.Errorf("Values() len = %d, want >= 2", l)
	}

	// Remove
	m.Set("del", "x", noExpiry)
	m.Remove("del")
	if got := m.Get("del"); got != "" {
		t.Errorf("Get(del) after Remove = %q, want empty", got)
	}

	// Remove несуществующего — не должно паниковать
	m.Remove("nonexistent")
}

func TestOwnMapInterface(t *testing.T) {
	runMapTests(t, NewOwnMap(16))
}

func TestStdMapInterface(t *testing.T) {
	runMapTests(t, NewStdMap())
}

func TestStdMapTTLIgnored(t *testing.T) {
	m := NewStdMap()
	m.Set("key", "val", time.Now().Add(-time.Hour)) // истёкший TTL игнорируется
	if got := m.Get("key"); got != "val" {
		t.Errorf("Get(key) = %q, want %q", got, "val")
	}
}

func TestStdMapLarge(t *testing.T) {
	const n = 100_000
	m := NewStdMap()
	expected := make(map[string]string, n)
	for i := 0; i < n; i++ {
		key := fmt.Sprintf("key-%d", rand.Int())
		val := fmt.Sprintf("val-%d", rand.Int())
		expected[key] = val
		m.Set(key, val, noExpiry)
	}
	for k, v := range expected {
		if got := m.Get(k); got != v {
			t.Errorf("Get(%q) = %q, want %q", k, got, v)
		}
	}
}
