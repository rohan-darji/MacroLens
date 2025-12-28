package cache

import (
	"context"
	"testing"
	"time"

	"github.com/macrolens/backend/internal/domain"
)

func TestMemoryCache_SetAndGet(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	tests := []struct {
		name    string
		key     string
		value   interface{}
		ttl     time.Duration
		wantErr bool
	}{
		{
			name:    "store and retrieve string",
			key:     "test-key-1",
			value:   "test-value",
			ttl:     1 * time.Minute,
			wantErr: false,
		},
		{
			name: "store and retrieve struct",
			key:  "test-key-2",
			value: map[string]interface{}{
				"fdcId":       "12345",
				"productName": "Milk",
			},
			ttl:     1 * time.Minute,
			wantErr: false,
		},
		{
			name:    "store with short TTL",
			key:     "test-key-3",
			value:   "expires-soon",
			ttl:     1 * time.Millisecond,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set value
			err := cache.Set(ctx, tt.key, tt.value, tt.ttl)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// For short TTL test, wait for expiration
			if tt.ttl < 10*time.Millisecond {
				time.Sleep(10 * time.Millisecond)
				// Should get cache miss after expiration
				_, err := cache.Get(ctx, tt.key)
				if err != domain.ErrCacheMiss {
					t.Errorf("Expected cache miss after expiration, got error = %v", err)
				}
				return
			}

			// Get value
			got, err := cache.Get(ctx, tt.key)
			if err != nil {
				t.Errorf("Get() error = %v", err)
				return
			}

			// For simple string comparison
			if tt.name == "store and retrieve string" {
				if got != tt.value {
					t.Errorf("Get() = %v, want %v", got, tt.value)
				}
			}
		})
	}
}

func TestMemoryCache_Get_CacheMiss(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	_, err := cache.Get(ctx, "non-existent-key")
	if err != domain.ErrCacheMiss {
		t.Errorf("Get() error = %v, want %v", err, domain.ErrCacheMiss)
	}
}

func TestMemoryCache_Delete(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	// Set a value
	key := "delete-test"
	err := cache.Set(ctx, key, "value", 1*time.Minute)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Verify it exists
	_, err = cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get() before delete error = %v", err)
	}

	// Delete it
	err = cache.Delete(ctx, key)
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// Verify it's gone
	_, err = cache.Get(ctx, key)
	if err != domain.ErrCacheMiss {
		t.Errorf("Get() after delete error = %v, want %v", err, domain.ErrCacheMiss)
	}
}

func TestMemoryCache_Exists(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	key := "exists-test"

	// Should not exist initially
	exists, err := cache.Exists(ctx, key)
	if err != nil {
		t.Errorf("Exists() error = %v", err)
	}
	if exists {
		t.Errorf("Exists() = true, want false for non-existent key")
	}

	// Set a value
	err = cache.Set(ctx, key, "value", 1*time.Minute)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Should exist now
	exists, err = cache.Exists(ctx, key)
	if err != nil {
		t.Errorf("Exists() error = %v", err)
	}
	if !exists {
		t.Errorf("Exists() = false, want true after setting value")
	}

	// Set with very short TTL
	shortKey := "short-ttl"
	err = cache.Set(ctx, shortKey, "value", 1*time.Millisecond)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	// Should not exist after expiration
	exists, err = cache.Exists(ctx, shortKey)
	if err != nil {
		t.Errorf("Exists() error = %v", err)
	}
	if exists {
		t.Errorf("Exists() = true, want false after expiration")
	}
}

func TestMemoryCache_Size(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	// Initial size should be 0
	if size := cache.Size(); size != 0 {
		t.Errorf("Size() = %d, want 0 for empty cache", size)
	}

	// Add some items
	for i := 0; i < 5; i++ {
		key := string(rune('a' + i))
		err := cache.Set(ctx, key, i, 1*time.Minute)
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}
	}

	// Size should be 5
	if size := cache.Size(); size != 5 {
		t.Errorf("Size() = %d, want 5", size)
	}

	// Delete one
	err := cache.Delete(ctx, "a")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Size should be 4
	if size := cache.Size(); size != 4 {
		t.Errorf("Size() = %d, want 4 after delete", size)
	}
}

func TestMemoryCache_Clear(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	// Add some items
	for i := 0; i < 5; i++ {
		key := string(rune('a' + i))
		err := cache.Set(ctx, key, i, 1*time.Minute)
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}
	}

	// Verify size
	if size := cache.Size(); size != 5 {
		t.Fatalf("Size() = %d, want 5 before clear", size)
	}

	// Clear cache
	cache.Clear()

	// Size should be 0
	if size := cache.Size(); size != 0 {
		t.Errorf("Size() = %d, want 0 after clear", size)
	}

	// All keys should be gone
	for i := 0; i < 5; i++ {
		key := string(rune('a' + i))
		_, err := cache.Get(ctx, key)
		if err != domain.ErrCacheMiss {
			t.Errorf("Get(%s) after clear error = %v, want %v", key, err, domain.ErrCacheMiss)
		}
	}
}

func TestMemoryCache_Concurrent(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	// Test concurrent access
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			key := string(rune('a' + id))
			// Set
			err := cache.Set(ctx, key, id, 1*time.Minute)
			if err != nil {
				t.Errorf("Concurrent Set() error = %v", err)
			}
			// Get
			_, err = cache.Get(ctx, key)
			if err != nil {
				t.Errorf("Concurrent Get() error = %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
