package main

import "chatExtensionServer/internal/types"

// RateLimiter rate limit manager class
type RateLimiter struct {
	counter map[types.UIDType]uint16
	limit   uint16
}

func (this *RateLimiter) Init(limit uint16) {
	this.counter = make(map[types.UIDType]uint16)
	this.limit = limit
}

func (this *RateLimiter) Resolve(userID types.UIDType) {
	this.counter[userID]--
}

func (this *RateLimiter) Add(userID types.UIDType) bool {
	if this.counter[userID] > this.limit {
		return false
	}
	this.counter[userID]++
	return true
}

func (this *RateLimiter) Timeout(userID types.UIDType) {
	delete(this.counter, userID)
}
