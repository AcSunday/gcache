package gcache

import (
	"bytes"
	"sync"
	"testing"
	"time"
)

// TestNewCacheWithTTL 测试创建带 TTL 的缓存
func TestNewCacheWithTTL(t *testing.T) {
	cache := NewCacheWithTTL(1024*1024, time.Second)
	if cache == nil {
		t.Fatal("NewCacheWithTTL returned nil")
	}
	defer cache.Close()
}

// TestCacheWithTTL_SetAndGet 测试基本的 Set 和 Get 操作
func TestCacheWithTTL_SetAndGet(t *testing.T) {
	cache := NewCacheWithTTL(1024*1024, time.Second)
	defer cache.Close()

	key := "test-key"
	value := []byte("test-value")

	// 测试 Set
	err := cache.Set(key, value, time.Second)
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

// TestCacheWithTTL_Expiration 测试过期功能
func TestCacheWithTTL_Expiration(t *testing.T) {
	cache := NewCacheWithTTL(1024*1024, time.Second)
	defer cache.Close()

	key := "test-key"
	value := []byte("test-value")

	// 设置一个很短的 TTL
	err := cache.Set(key, value, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// 立即获取应该成功
	got := cache.Get(key)
	if !bytes.Equal(got, value) {
		t.Errorf("Get returned %v, want %v", got, value)
	}

	// 等待过期
	time.Sleep(150 * time.Millisecond)

	// 过期后应该返回 nil
	got = cache.Get(key)
	if got != nil {
		t.Errorf("Get returned %v after expiration, want nil", got)
	}

	// Has 也应该返回 false
	if cache.Has(key) {
		t.Error("Has returned true for expired key")
	}
}

// TestCacheWithTTL_DifferentTTL 测试不同的 TTL 值
func TestCacheWithTTL_DifferentTTL(t *testing.T) {
	cache := NewCacheWithTTL(1024*1024, time.Second)
	defer cache.Close()

	key1 := "key1"
	key2 := "key2"
	value := []byte("test-value")

	// 设置不同的 TTL
	err := cache.Set(key1, value, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	err = cache.Set(key2, value, 200*time.Millisecond)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// 等待第一个 key 过期
	time.Sleep(100 * time.Millisecond)

	// key1 应该过期
	if cache.Get(key1) != nil {
		t.Error("key1 should be expired")
	}

	// key2 应该还存在
	got := cache.Get(key2)
	if !bytes.Equal(got, value) {
		t.Errorf("key2 Get returned %v, want %v", got, value)
	}

	// 等待 key2 也过期
	time.Sleep(150 * time.Millisecond)

	// key2 应该过期
	if cache.Get(key2) != nil {
		t.Error("key2 should be expired")
	}
}

// TestCacheWithTTL_UpdateTTL 测试更新 TTL
func TestCacheWithTTL_UpdateTTL(t *testing.T) {
	cache := NewCacheWithTTL(1024*1024, time.Second)
	defer cache.Close()

	key := "test-key"
	value := []byte("test-value")

	// 设置一个很短的 TTL
	err := cache.Set(key, value, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// 在过期前更新为更长的 TTL
	time.Sleep(30 * time.Millisecond)
	err = cache.Set(key, value, 200*time.Millisecond)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// 等待第一次 TTL 过期时间
	time.Sleep(50 * time.Millisecond)

	// 应该还存在（因为 TTL 被更新了）
	got := cache.Get(key)
	if !bytes.Equal(got, value) {
		t.Errorf("Get returned %v, want %v", got, value)
	}

	// 等待新的 TTL 过期
	time.Sleep(150 * time.Millisecond)

	// 现在应该过期
	if cache.Get(key) != nil {
		t.Error("Key should be expired after updated TTL")
	}
}

// TestCacheWithTTL_ZeroTTL 测试零 TTL
func TestCacheWithTTL_ZeroTTL(t *testing.T) {
	cache := NewCacheWithTTL(1024*1024, time.Second)
	defer cache.Close()

	key := "test-key"
	value := []byte("test-value")

	// 设置零 TTL（应该立即过期）
	err := cache.Set(key, value, 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// 应该立即过期
	got := cache.Get(key)
	if got != nil {
		t.Errorf("Get returned %v for zero TTL, want nil", got)
	}
}

// TestCacheWithTTL_NegativeTTL 测试负 TTL
func TestCacheWithTTL_NegativeTTL(t *testing.T) {
	cache := NewCacheWithTTL(1024*1024, time.Second)
	defer cache.Close()

	key := "test-key"
	value := []byte("test-value")

	// 设置负 TTL（应该立即过期）
	err := cache.Set(key, value, -time.Second)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// 应该立即过期
	got := cache.Get(key)
	if got != nil {
		t.Errorf("Get returned %v for negative TTL, want nil", got)
	}
}

// TestCacheWithTTL_Delete 测试删除操作
func TestCacheWithTTL_Delete(t *testing.T) {
	cache := NewCacheWithTTL(1024*1024, time.Second)
	defer cache.Close()

	key := "test-key"
	value := []byte("test-value")

	// 设置 key
	err := cache.Set(key, value, time.Second)
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

// TestCacheWithTTL_Concurrent 测试并发操作
func TestCacheWithTTL_Concurrent(t *testing.T) {
	cache := NewCacheWithTTL(10*1024*1024, time.Second)
	defer cache.Close()

	const numGoroutines = 50
	const numKeys = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2)

	// 并发写入
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numKeys; j++ {
				key := string(rune(id*numKeys + j))
				value := []byte{byte(id), byte(j)}
				err := cache.Set(key, value, time.Second)
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

// TestCacheWithTTL_ExpirationRace 测试过期竞态条件
func TestCacheWithTTL_ExpirationRace(t *testing.T) {
	cache := NewCacheWithTTL(1024*1024, time.Second)
	defer cache.Close()

	key := "test-key"
	value := []byte("test-value")

	// 设置一个很短的 TTL
	err := cache.Set(key, value, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// 并发读取，可能在过期过程中
	var wg sync.WaitGroup
	const numGoroutines = 10
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = cache.Get(key)
				_ = cache.Has(key)
				time.Sleep(time.Millisecond)
			}
		}()
	}

	wg.Wait()
}

// TestCacheWithTTL_LargeValue 测试大值
func TestCacheWithTTL_LargeValue(t *testing.T) {
	cache := NewCacheWithTTL(10*1024*1024, time.Second)
	defer cache.Close()

	key := "test-key"
	value := make([]byte, 10*1024) // 10KB
	for i := range value {
		value[i] = byte(i % 256)
	}

	err := cache.Set(key, value, time.Second)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	got := cache.Get(key)
	if !bytes.Equal(got, value) {
		t.Error("Large value mismatch")
	}
}

// TestCacheWithTTL_WrapUnwrap 测试包装和解包函数
func TestCacheWithTTL_WrapUnwrap(t *testing.T) {
	original := []byte("test-data")
	ttl := time.Second

	// 包装
	wrapped := wrapCacheWithTTL(original, ttl)
	if len(wrapped) != 8+len(original) {
		t.Errorf("Wrapped length %d, want %d", len(wrapped), 8+len(original))
	}

	// 立即解包应该成功
	unwrapped, ok := unwrapCacheWithTTL(wrapped)
	if !ok {
		t.Error("unwrapCacheWithTTL returned false")
	}
	if !bytes.Equal(unwrapped, original) {
		t.Errorf("Unwrapped %v, want %v", unwrapped, original)
	}
}

// TestCacheWithTTL_WrapUnwrapExpired 测试过期解包
func TestCacheWithTTL_WrapUnwrapExpired(t *testing.T) {
	original := []byte("test-data")
	ttl := -time.Second // 负 TTL，立即过期

	// 包装
	wrapped := wrapCacheWithTTL(original, ttl)

	// 解包应该失败（已过期）
	unwrapped, ok := unwrapCacheWithTTL(wrapped)
	if ok {
		t.Error("unwrapCacheWithTTL returned true for expired data")
	}
	if unwrapped != nil {
		t.Errorf("Unwrapped %v, want nil", unwrapped)
	}
}

// TestCacheWithTTL_WrapUnwrapInvalid 测试无效数据解包
func TestCacheWithTTL_WrapUnwrapInvalid(t *testing.T) {
	// 测试太短的数据
	shortData := []byte{1, 2, 3}
	unwrapped, ok := unwrapCacheWithTTL(shortData)
	if ok {
		t.Error("unwrapCacheWithTTL returned true for short data")
	}
	if unwrapped != nil {
		t.Errorf("Unwrapped %v, want nil", unwrapped)
	}

	// 测试 nil
	unwrapped, ok = unwrapCacheWithTTL(nil)
	if ok {
		t.Error("unwrapCacheWithTTL returned true for nil")
	}
	if unwrapped != nil {
		t.Errorf("Unwrapped %v, want nil", unwrapped)
	}
}

// BenchmarkCacheWithTTL_Set 基准测试 Set 操作
func BenchmarkCacheWithTTL_Set(b *testing.B) {
	cache := NewCacheWithTTL(100*1024*1024, time.Second)
	defer cache.Close()

	key := "bench-key"
	value := []byte("bench-value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(key, value, time.Second)
	}
}

// BenchmarkCacheWithTTL_Get 基准测试 Get 操作
func BenchmarkCacheWithTTL_Get(b *testing.B) {
	cache := NewCacheWithTTL(100*1024*1024, time.Second)
	defer cache.Close()

	key := "bench-key"
	value := []byte("bench-value")
	cache.Set(key, value, time.Hour) // 设置很长的 TTL 避免过期

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.Get(key)
	}
}

// BenchmarkCacheWithTTL_Concurrent 并发基准测试
func BenchmarkCacheWithTTL_Concurrent(b *testing.B) {
	cache := NewCacheWithTTL(100*1024*1024, time.Second)
	defer cache.Close()

	b.RunParallel(func(pb *testing.PB) {
		key := "bench-key"
		value := []byte("bench-value")
		for pb.Next() {
			cache.Set(key, value, time.Second)
			_ = cache.Get(key)
		}
	})
}
