package gcache

import (
	"bytes"
	"sync"
	"testing"
)

// TestNewCache 测试创建缓存
func TestNewCache(t *testing.T) {
	cache := NewCache(1024 * 1024) // 1MB
	if cache == nil {
		t.Fatal("NewCache returned nil")
	}
	defer cache.Close()
}

// TestCache_SetAndGet 测试基本的 Set 和 Get 操作
func TestCache_SetAndGet(t *testing.T) {
	cache := NewCache(1024 * 1024)
	defer cache.Close()

	key := "test-key"
	value := []byte("test-value")

	// 测试 Set
	err := cache.Set(key, value)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// 测试 Get
	got := cache.Get(key)
	if got == nil {
		t.Fatal("Get returned nil")
	}

	if !bytes.Equal(got, value) {
		t.Errorf("Get returned %v, want %v", got, value)
	}
}

// TestCache_GetNonExistent 测试获取不存在的 key
func TestCache_GetNonExistent(t *testing.T) {
	cache := NewCache(1024 * 1024)
	defer cache.Close()

	got := cache.Get("non-existent-key")
	if got != nil {
		t.Errorf("Get returned %v, want nil", got)
	}
}

// TestCache_Has 测试 Has 方法
func TestCache_Has(t *testing.T) {
	cache := NewCache(1024 * 1024)
	defer cache.Close()

	key := "test-key"
	value := []byte("test-value")

	// 测试不存在的 key
	if cache.Has(key) {
		t.Error("Has returned true for non-existent key")
	}

	// 设置 key
	err := cache.Set(key, value)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// 测试存在的 key
	if !cache.Has(key) {
		t.Error("Has returned false for existing key")
	}
}

// TestCache_Delete 测试 Delete 方法
func TestCache_Delete(t *testing.T) {
	cache := NewCache(1024 * 1024)
	defer cache.Close()

	key := "test-key"
	value := []byte("test-value")

	// 设置 key
	err := cache.Set(key, value)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// 验证存在
	if !cache.Has(key) {
		t.Error("Key should exist before delete")
	}

	// 删除 key
	err = cache.Delete(key)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// 验证已删除
	if cache.Has(key) {
		t.Error("Key should not exist after delete")
	}

	got := cache.Get(key)
	if got != nil {
		t.Errorf("Get returned %v after delete, want nil", got)
	}
}

// TestCache_Update 测试更新已存在的 key
func TestCache_Update(t *testing.T) {
	cache := NewCache(1024 * 1024)
	defer cache.Close()

	key := "test-key"
	value1 := []byte("value1")
	value2 := []byte("value2")

	// 设置第一个值
	err := cache.Set(key, value1)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// 验证第一个值
	got := cache.Get(key)
	if !bytes.Equal(got, value1) {
		t.Errorf("Get returned %v, want %v", got, value1)
	}

	// 更新为第二个值
	err = cache.Set(key, value2)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// 验证第二个值
	got = cache.Get(key)
	if !bytes.Equal(got, value2) {
		t.Errorf("Get returned %v, want %v", got, value2)
	}
}

// TestCache_EmptyValue 测试空值
func TestCache_EmptyValue(t *testing.T) {
	cache := NewCache(1024 * 1024)
	defer cache.Close()

	key := "test-key"
	value := []byte{}

	err := cache.Set(key, value)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	got := cache.Get(key)
	if got == nil {
		t.Error("Get returned nil for empty value")
	}
	if len(got) != 0 {
		t.Errorf("Get returned %v, want empty slice", got)
	}
}

// TestCache_LargeValue 测试大值
func TestCache_LargeValue(t *testing.T) {
	cache := NewCache(1024 * 1024)
	defer cache.Close()

	key := "test-key"
	value := make([]byte, 10*1024) // 10KB
	for i := range value {
		value[i] = byte(i % 256)
	}

	err := cache.Set(key, value)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	got := cache.Get(key)
	if !bytes.Equal(got, value) {
		t.Error("Large value mismatch")
	}
}

// TestCache_ConcurrentReadWrite 测试并发读写
func TestCache_ConcurrentReadWrite(t *testing.T) {
	cache := NewCache(10 * 1024 * 1024) // 10MB
	defer cache.Close()

	const numGoroutines = 100
	const numKeys = 1000

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2) // 读和写各一半

	// 并发写入
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numKeys; j++ {
				key := string(rune(id*numKeys + j))
				value := []byte{byte(id), byte(j)}
				err := cache.Set(key, value)
				if err != nil {
					t.Errorf("Set failed: %v", err)
				}
			}
		}(i)
	}

	// 并发读取
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numKeys; j++ {
				key := string(rune(id*numKeys + j))
				_ = cache.Get(key)
				_ = cache.Has(key)
			}
		}(i)
	}

	wg.Wait()
}

// TestCache_ConcurrentDelete 测试并发删除
func TestCache_ConcurrentDelete(t *testing.T) {
	cache := NewCache(10 * 1024 * 1024)
	defer cache.Close()

	const numKeys = 1000

	// 先设置一些 key
	for i := 0; i < numKeys; i++ {
		key := string(rune(i))
		value := []byte{byte(i)}
		cache.Set(key, value)
	}

	// 并发删除
	const numGoroutines = 50
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numKeys/numGoroutines; j++ {
				key := string(rune(id*numKeys/numGoroutines + j))
				_ = cache.Delete(key)
			}
		}(i)
	}

	wg.Wait()
}

// TestCache_Close 测试 Close 方法
func TestCache_Close(t *testing.T) {
	cache := NewCache(1024 * 1024)

	key := "test-key"
	value := []byte("test-value")

	err := cache.Set(key, value)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// 关闭缓存
	err = cache.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// 关闭后应该无法获取数据（fastcache.Reset 会清空缓存）
	got := cache.Get(key)
	if got != nil {
		t.Error("Get returned value after Close, want nil")
	}
}

// TestCache_MultipleKeys 测试多个 key
func TestCache_MultipleKeys(t *testing.T) {
	cache := NewCache(10 * 1024 * 1024)
	defer cache.Close()

	const numKeys = 1000
	keys := make([]string, numKeys)
	values := make([][]byte, numKeys)

	// 设置多个 key
	for i := 0; i < numKeys; i++ {
		keys[i] = string(rune(i))
		values[i] = []byte{byte(i), byte(i >> 8), byte(i >> 16)}
		err := cache.Set(keys[i], values[i])
		if err != nil {
			t.Fatalf("Set failed for key %d: %v", i, err)
		}
	}

	// 验证所有 key
	for i := 0; i < numKeys; i++ {
		if !cache.Has(keys[i]) {
			t.Errorf("Key %d should exist", i)
		}
		got := cache.Get(keys[i])
		if !bytes.Equal(got, values[i]) {
			t.Errorf("Value mismatch for key %d", i)
		}
	}
}

// TestCache_SpecialCharacters 测试特殊字符 key
func TestCache_SpecialCharacters(t *testing.T) {
	cache := NewCache(1024 * 1024)
	defer cache.Close()

	testCases := []struct {
		name  string
		key   string
		value []byte
	}{
		{"empty key", "", []byte("value")},
		{"unicode key", "测试-key", []byte("测试-value")},
		{"special chars", "key!@#$%^&*()", []byte("value")},
		{"newline", "key\nvalue", []byte("value\nkey")},
		{"unicode value", "key", []byte("测试-value")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := cache.Set(tc.key, tc.value)
			if err != nil {
				t.Fatalf("Set failed: %v", err)
			}

			got := cache.Get(tc.key)
			if !bytes.Equal(got, tc.value) {
				t.Errorf("Get returned %v, want %v", got, tc.value)
			}

			if !cache.Has(tc.key) {
				t.Error("Has returned false for existing key")
			}
		})
	}
}

// TestCache_PoolReuse 测试 sync.Pool 的重用
func TestCache_PoolReuse(t *testing.T) {
	cache := NewCache(1024 * 1024)
	defer cache.Close()

	// 多次 Get 操作，验证 pool 正常工作
	for i := 0; i < 100; i++ {
		key := string(rune(i))
		value := []byte{byte(i)}
		cache.Set(key, value)
		got := cache.Get(key)
		if !bytes.Equal(got, value) {
			t.Errorf("Value mismatch at iteration %d", i)
		}
	}
}

// BenchmarkCache_Set 基准测试 Set 操作
func BenchmarkCache_Set(b *testing.B) {
	cache := NewCache(100 * 1024 * 1024)
	defer cache.Close()

	key := "bench-key"
	value := []byte("bench-value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(key, value)
	}
}

// BenchmarkCache_Get 基准测试 Get 操作
func BenchmarkCache_Get(b *testing.B) {
	cache := NewCache(100 * 1024 * 1024)
	defer cache.Close()

	key := "bench-key"
	value := []byte("bench-value")
	cache.Set(key, value)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.Get(key)
	}
}

// BenchmarkCache_Has 基准测试 Has 操作
func BenchmarkCache_Has(b *testing.B) {
	cache := NewCache(100 * 1024 * 1024)
	defer cache.Close()

	key := "bench-key"
	value := []byte("bench-value")
	cache.Set(key, value)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.Has(key)
	}
}

// BenchmarkCache_Concurrent 并发基准测试
func BenchmarkCache_Concurrent(b *testing.B) {
	cache := NewCache(100 * 1024 * 1024)
	defer cache.Close()

	b.RunParallel(func(pb *testing.PB) {
		key := "bench-key"
		value := []byte("bench-value")
		for pb.Next() {
			cache.Set(key, value)
			_ = cache.Get(key)
		}
	})
}
