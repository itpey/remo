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
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

var (
	ErrKeyNotFound = errors.New("key not found")
	ErrKeyExpired  = errors.New("key has expired")
	ErrEmptyKey    = errors.New("key cannot be empty")
	ErrNegativeTTL = errors.New("TTL cannot be negative")
)

// Storage represents an in-memory key-value storage with expiration.
type Storage struct {
	mu             sync.RWMutex
	data           map[string]*item
	cleanupRunning bool
	ctx            context.Context
	cancel         context.CancelFunc
}

// item represents a key-value pair with an expiration time.
type item struct {
	expiration time.Time
	value      interface{}
}

// New creates and returns a new instance of Storage.
func New() *Storage {
	store := &Storage{
		data:           make(map[string]*item),
		cleanupRunning: false,
	}
	return store
}

// Get retrieves a value from storage by key. Returns nil if the key does not exist or has expired.
func (s *Storage) Get(key string) (interface{}, error) {
	s.mu.RLock()
	item, exists := s.data[key]
	s.mu.RUnlock()

	if !exists {
		return nil, ErrKeyNotFound
	}

	if item.isExpired() {
		return nil, ErrKeyExpired
	}

	return item.value, nil
}

// Set sets a key-value pair in storage with an optional time-to-live (TTL) duration.
func (s *Storage) Set(key string, value interface{}, ttl time.Duration) error {
	if err := s.validateKeyAndTTL(key, ttl); err != nil {
		return err
	}

	expiration := s.calculateExpiration(ttl)
	s.mu.Lock()
	s.data[key] = newItem(value, expiration)
	s.mu.Unlock()
	return nil
}

// Delete removes an item from storage.
func (s *Storage) Delete(key string) {
	s.mu.Lock()
	_, exists := s.data[key]
	if exists {
		delete(s.data, key)
	}
	s.mu.Unlock()
}

// Reset clears all keys from storage.
func (s *Storage) Reset() {
	s.mu.Lock()
	s.data = make(map[string]*item)
	s.mu.Unlock()
}

// cleanup periodically removes expired items from storage.
func (s *Storage) cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.removeExpiredItems()
		case <-s.ctx.Done():
			return
		}
	}
}

// StartCleanup starts the automatic cleanup goroutine.
func (s *Storage) StartCleanup(interval time.Duration) {
	if !s.cleanupRunning {
		s.ctx, s.cancel = context.WithCancel(context.Background())
		s.cleanupRunning = true
		s.safeGo(func() {
			s.cleanup(interval)
		})
	}
}

// StopCleanup stops the automatic cleanup goroutine gracefully.
func (s *Storage) StopCleanup() {
	if s.cleanupRunning {
		s.cancel()
		s.cleanupRunning = false
	}
}

// removeExpiredItems removes items that have expired.
func (s *Storage) removeExpiredItems() {
	now := time.Now()
	s.mu.Lock()
	for key, item := range s.data {
		if item.isExpiredAt(now) {
			delete(s.data, key)
		}
	}
	s.mu.Unlock()
}

// validateKeyAndTTL checks if the key and TTL are valid.
func (s *Storage) validateKeyAndTTL(key string, ttl time.Duration) error {
	if key == "" {
		return ErrEmptyKey
	}
	if ttl < 0 {
		return ErrNegativeTTL
	}
	return nil
}

// calculateExpiration calculates the expiration time based on TTL.
func (s *Storage) calculateExpiration(ttl time.Duration) time.Time {
	if ttl <= 0 {
		return time.Time{}
	}
	return time.Now().Add(ttl)
}

// newItem creates a new item with the given value and expiration time.
func newItem(value interface{}, expiration time.Time) *item {
	return &item{
		expiration: expiration,
		value:      value,
	}
}

// isExpired checks if the item is expired.
func (i *item) isExpired() bool {
	return i.isExpiredAt(time.Now())
}

// isExpiredAt checks if the item is expired at a specific time.
func (i *item) isExpiredAt(now time.Time) bool {
	return !i.expiration.IsZero() && i.expiration.Before(now)
}

// safeGo runs a function in a goroutine and recovers from panics, logging them.
func (s *Storage) safeGo(f func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Remo: [Panic] %v", r)
			}
		}()
		f()
	}()
}
