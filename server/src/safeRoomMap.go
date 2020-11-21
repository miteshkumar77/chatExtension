package main

import "hash/fnv"

type SafeRoomMap struct {
	rooms      [](map[string]map[uidType]bool)
	readLocks  [](chan bool)
	writeLocks [](chan bool)
}

func (this *SafeRoomMap) hash(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

func (this *SafeRoomMap) Init(shards uint32) {
	this.rooms = make([]map[string]map[int64]bool, shards)
	this.readLocks = make([]chan bool, shards)
	this.writeLocks = make([]chan bool, shards)
	for i := 0; i < len(this.readLocks); i++ {
		this.rooms[i] = make(map[string]map[int64]bool)
		this.readLocks[i] = make(chan bool)
		this.writeLocks[i] = make(chan bool)
	}
}

func (this *SafeRoomMap) getShard(roomID string) int {
	return int(this.hash(roomID)) % len(this.readLocks)
}

func (this *SafeRoomMap) Size() int {
	ret := 0
	for i, _ := range this.rooms {
		ret += len(this.rooms[i])
	}
	return ret
}

func (this *SafeRoomMap) Shards() uint32 {
	return uint32(len(this.readLocks))
}
