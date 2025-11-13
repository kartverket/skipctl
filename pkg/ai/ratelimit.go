package ai

import (
	"fmt"
	"sync"
	"time"
)

// RateLimiter limits the number of API calls
type RateLimiter struct {
	requests []time.Time
	mutex    sync.Mutex
	limit    int           // Max requests per window
	window   time.Duration // Time window
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make([]time.Time, 0),
		limit:    limit,
		window:   window,
	}
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow() error {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Remove old requests
	validRequests := make([]time.Time, 0)
	for _, req := range rl.requests {
		if req.After(cutoff) {
			validRequests = append(validRequests, req)
		}
	}
	rl.requests = validRequests

	// Check if we can make another request
	if len(rl.requests) >= rl.limit {
		return fmt.Errorf("rate limit exceeded: %d requests per %v", rl.limit, rl.window)
	}

	// Add current request
	rl.requests = append(rl.requests, now)
	return nil
}
