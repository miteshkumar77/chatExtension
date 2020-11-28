package main

// RateLimiter rate limit manager class
type RateLimiter struct {
	counter map[uidType]uint16
	limit   uint16
}

func (this *RateLimiter) Init(limit uint16) {
	this.counter = make(map[uidType]uint16)
	this.limit = limit
}

func (this *RateLimiter) Resolve(userID uidType) {
	this.counter[userID]--
}

func (this *RateLimiter) Add(userID uidType) bool {
	if this.counter[userID] > this.limit {
		return false
	}
	this.counter[userID]++
	return true
}

func (this *RateLimiter) Timeout(userID uidType) {
	delete(this.counter, userID)
}
