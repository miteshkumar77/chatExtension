package main

import "container/list"

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

func (this *SafeQueue) Pop() *Message {
	this.readLock <- true
	var ret *Message = this.queue.Remove(this.queue.Back()).(*Message)
	if this.queue.Len() > 0 {
		<-this.readLock
	}
	return ret
}

func (this *SafeQueue) Push(item *Message) {
	this.writeLock <- true
	this.queue.PushBack(item)
	<-this.writeLock
	if this.queue.Len() == 1 {
		<-this.readLock
	}
}
