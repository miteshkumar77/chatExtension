package ratelimiting

import (
	"chatExtensionServer/internal/concurrency/concurrentratelimittable"
	"chatExtensionServer/internal/types"
)

// RateLimiter rate limit manager class
type RateLimiter struct {
	counter concurrentratelimittable.ConcurrentHashMap
	limit   uint16
}

// Init creates a rate limiting table with a specified
// message limit
func (limiter *RateLimiter) Init(limit uint16) {
	limiter.counter = concurrentratelimittable.CreateNewRateLimitTable()
	limiter.limit = limit
}

// Resolve signifies that a user generated task has been resolved
// and no longer contributes to throttling the user
func (limiter *RateLimiter) Resolve(userID types.UIDType) {
	limiter.counter.CallBackUpdate(userID, func(original uint16) uint16 {
		println(original)
		return original - 1
	})
}

// Add signifies that a user has generated a request to complete a task
// if the user has already generated the maximum number of tasks
// (equivalent to limit) then Add will return true to indicate that
// the user should be throttled, and the user's limit is reset
func (limiter *RateLimiter) Add(userID types.UIDType) bool {

	var willThrottleUser bool
	limiter.counter.CallBackUpdateInsertOrDelete(userID, func(exists bool, original uint16) (bool, uint16) {
		if exists == false {
			willThrottleUser = false
			return false, 1
		}

		if original > limiter.limit {
			willThrottleUser = true
			return true, 0
		}

		willThrottleUser = false
		return false, original + 1
	})

	return willThrottleUser
}
