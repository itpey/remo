// Copyright 2023 itpey
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package remo

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestStorage_SetGetDelete(t *testing.T) {
	store := New()

	// Test setting a key-value pair and retrieving it.
	key := "testKey"
	value := "testValue"
	err := store.Set(key, value, time.Second)
	if err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	retrievedValue, err := store.Get(key)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}
	if retrievedValue != value {
		t.Errorf("Expected %v, but got %v", value, retrievedValue)
	}

	// Test setting a key with a negative TTL.
	err = store.Set("negativeTTLKey", "value", -time.Second)
	if err != ErrNegativeTTL {
		t.Errorf("Expected ErrNegativeTTL, but got %v", err)
	}

	// Test setting a key with an empty name.
	err = store.Set("", "value", time.Second)
	if err != ErrEmptyKey {
		t.Errorf("Expected ErrEmptyKey, but got %v", err)
	}

	// Test setting a key with a TTL of 0, which should not expire.
	keyZeroTTL := "keyZeroTTL"
	valueZeroTTL := "valueZeroTTL"
	err = store.Set(keyZeroTTL, valueZeroTTL, 0)
	if err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	time.Sleep(2 * time.Second)
	retrievedValue, err = store.Get(keyZeroTTL)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}
	if retrievedValue != valueZeroTTL {
		t.Errorf("Expected %v, but got %v", valueZeroTTL, retrievedValue)
	}

	// Test deleting a key.
	store.Delete(key)
	_, err = store.Get(key)
	if err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound, but got %v", err)
	}

	// Test deleting a non-existing key.
	nonExistingKey := "nonExistingKey"
	store.Delete(nonExistingKey)
}

func TestStorage_Cleanup(t *testing.T) {
	store := New()

	// Start the cleanup goroutine with a short cleanup interval.
	store.StartCleanup(100 * time.Millisecond)

	key := "cleanupKey"
	value := "cleanupValue"
	err := store.Set(key, value, time.Second)
	if err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	// Sleep for a longer time to ensure that the cleanup has run.
	time.Sleep(2 * time.Second)

	_, err = store.Get(key)
	if err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound after cleanup, but got %v", err)
	}

	// Stop the cleanup goroutine.
	store.StopCleanup()
}

func TestStorage_Reset(t *testing.T) {
	store := New()

	key := "resetKey"
	value := "resetValue"
	err := store.Set(key, value, time.Second)
	if err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	store.Reset()
	_, err = store.Get(key)
	if err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound after reset, but got %v", err)
	}
}

func TestStorage_ErrKeyExpired(t *testing.T) {
	store := New()

	key := "expiredKey"
	value := "expiredValue"
	err := store.Set(key, value, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	time.Sleep(200 * time.Millisecond)
	_, err = store.Get(key)
	if err != ErrKeyExpired {
		t.Errorf("Expected ErrKeyExpired, but got %v", err)
	}
}

func TestStorage_ConcurrentAccess(t *testing.T) {
	store := New()
	const key = "concurrentKey"
	const value = "concurrentValue"
	const numRoutines = 100
	const numOperationsPerRoutine = 100

	var wg sync.WaitGroup

	// Concurrently set keys.
	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numOperationsPerRoutine; j++ {
				err := store.Set(key, value, 2*time.Second)
				if err != nil {
					t.Errorf("Set() failed: %v", err)
				}
			}
		}()
	}

	// Concurrently get keys and check consistency.
	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numOperationsPerRoutine; j++ {
				retrievedValue, err := store.Get(key)
				if err != nil && err != ErrKeyNotFound && err != ErrKeyExpired {
					t.Errorf("Get() failed: %v", err)
				}
				if err == nil && retrievedValue != value {
					t.Errorf("Inconsistent value: Expected %v, but got %v", value, retrievedValue)
				}
			}
		}()
	}

	// Concurrently delete keys.
	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numOperationsPerRoutine; j++ {
				store.Delete(key)
			}
		}()
	}

	// Concurrently reset the storage.
	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numOperationsPerRoutine; j++ {
				store.Reset()
			}
		}()
	}

	wg.Wait()
}

// BenchmarkSet measures the performance of the Set operation.
func BenchmarkSet(b *testing.B) {
	store := New()
	key := "benchmarkKey"
	value := "benchmarkValue"

	for i := 0; i < b.N; i++ {
		err := store.Set(fmt.Sprintf("%s%d", key, i), value, 0)
		if err != nil {
			b.Fatalf("Set() failed: %v", err)
		}
	}
}

// BenchmarkGet measures the performance of the Get operation.
func BenchmarkGet(b *testing.B) {
	store := New()
	key := "benchmarkKey"
	value := "benchmarkValue"

	for i := 0; i < b.N; i++ {
		store.Set(fmt.Sprintf("%s%d", key, i), value, 0)
	}

	b.ResetTimer() // Reset timer to exclude setup time.

	for i := 0; i < b.N; i++ {
		_, err := store.Get(fmt.Sprintf("%s%d", key, i))
		if err != nil {
			b.Fatalf("Get() failed: %v", err)
		}
	}
}

// BenchmarkDelete measures the performance of the Delete operation.
func BenchmarkDelete(b *testing.B) {
	store := New()
	key := "benchmarkKey"
	value := "benchmarkValue"
	ttl := 2 * time.Second

	// Set up the storage with keys to delete.
	for i := 0; i < b.N; i++ {
		store.Set(fmt.Sprintf("%s%d", key, i), value, ttl)
	}

	b.ResetTimer() // Reset timer to exclude setup time.

	// Measure the performance of the Delete operation.
	for i := 0; i < b.N; i++ {
		store.Delete(fmt.Sprintf("%s%d", key, i))
	}
}

// BenchmarkReset measures the performance of the Reset operation.
func BenchmarkReset(b *testing.B) {
	store := New()
	key := "benchmarkKey"
	value := "benchmarkValue"

	// Set up the storage with keys to reset.
	for i := 0; i < b.N; i++ {
		store.Set(fmt.Sprintf("%s%d", key, i), value, 0)
	}

	b.ResetTimer() // Reset timer to exclude setup time.

	// Measure the performance of the Reset operation.
	for i := 0; i < b.N; i++ {
		store.Reset()
	}
}
