package main

import (
	"chatExtensionServer/internal/types"
	"container/list"
)

type SafeQueue struct {
	queue     *list.List
	readLock  chan bool
	writeLock chan bool
}

func (this *SafeQueue) Init() {
	this.queue = list.New()
	this.readLock = make(chan bool, 1)
	this.readLock <- true
	this.writeLock = make(chan bool, 1)
}

func (this *SafeQueue) Pop() *types.Message {
	this.readLock <- true
	var ret *types.Message = this.queue.Remove(this.queue.Back()).(*types.Message)
	if this.queue.Len() > 0 {
		<-this.readLock
	}
	return ret
}

func (this *SafeQueue) Push(item *types.Message) {
	this.writeLock <- true
	this.queue.PushBack(item)
	<-this.writeLock
	if this.queue.Len() == 1 {
		<-this.readLock
	}
}
