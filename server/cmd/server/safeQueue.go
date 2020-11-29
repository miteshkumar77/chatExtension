package main

import (
	"chatExtensionServer/internal/types"
	"container/list"
)

type SafeQueue struct {
	queue    *list.List
	headLock chan bool
	tailLock chan bool
}

func (this *SafeQueue) Init() {
	this.queue = list.New()
	this.headLock = make(chan bool, 1)
	this.headLock <- true
	this.tailLock = make(chan bool, 1)
}

func (this *SafeQueue) Pop() *types.Message {
	this.headLock <- true
	var ret *types.Message = this.queue.Remove(this.queue.Back()).(*types.Message)
	if this.queue.Len() > 0 {
		<-this.headLock
	}
	return ret
}

func (this *SafeQueue) Push(item *types.Message) {
	this.tailLock <- true
	this.queue.PushBack(item)
	<-this.tailLock
	if this.queue.Len() == 1 {
		<-this.headLock
	}
}
