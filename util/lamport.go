package util

import (
	"sync"
)

// LamportClock implements a Lamport logical clock
type LamportClock struct {
	time int32
	mu   sync.Mutex
}

// NewLamportClock creates a new Lamport clock initialized to 0
func NewLamportClock() *LamportClock {
	return &LamportClock{
		time: 0,
	}
}

// Tick increments the clock (for local events)
func (lc *LamportClock) Tick() int32 {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.time++
	return lc.time
}

// Update updates the clock based on a received timestamp (for received events)
// Implements: L = max(L, T) + 1
func (lc *LamportClock) Update(receivedTime int32) int32 {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	if receivedTime > lc.time {
		lc.time = receivedTime
	}
	lc.time++
	return lc.time
}

// GetTime returns the current clock value without incrementing
func (lc *LamportClock) GetTime() int32 {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	return lc.time
}