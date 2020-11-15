package main

// RateLimiter rate limit helper api
type RateLimiter struct {
	counter map[uidType]uint16
	limit   uint16
}

func (this *RateLimiter) Init(limit uint16) {
	this.counter = make(map[uidType]uint16)
	this.limit = limit
}

func (this *RateLimiter) resolve(userID uidType) {
	this.counter[userID]--
}

func (this *RateLimiter) add(userID uidType) bool {
	if this.counter[userID] > this.limit {
		return false
	}
	this.counter[userID]++
	return true
}

func (this *RateLimiter) timeout(userID uidType) {
	delete(this.counter, userID)
}
