package main

type SafeUserMap struct {
	users      [](map[uidType]*User)
	readLocks  [](chan bool)
	writeLocks [](chan bool)
	size       int
}

func (this *SafeUserMap) Init(shards uint32) {
	this.users = make([](map[uidType]*User), shards)
	this.readLocks = make([](chan bool), shards)
	this.writeLocks = make([](chan bool), shards)
	for i := 0; i < len(this.readLocks); i++ {
		this.users[i] = make(map[uidType]*User)
		this.readLocks[i] = make(chan bool, 1)
		this.writeLocks[i] = make(chan bool, 1)
	}
}

func (this *SafeUserMap) getShard(userID uidType) int {
	return int(userID) % len(this.readLocks)
}

func (this *SafeUserMap) Size() int {
	ret := 0
	for i, _ := range this.users {
		ret += len(this.users[i])
	}
	return ret
}

func (this *SafeUserMap) Shards() uint32 {
	return uint32(len(this.readLocks))
}

func (this *SafeUserMap) Set(userID uidType, user *User) {
	shard := this.getShard(userID)
	readLock := &this.readLocks[shard]
	writeLock := &this.writeLocks[shard]

	(*readLock) <- true
	(*writeLock) <- true

	this.users[shard][userID] = user

	<-*readLock
	<-*writeLock
}

func (this *SafeUserMap) AtomicUpdate(userID uidType, updateFunc func(*User) *User) {
	shard := this.getShard(userID)
	readLock := &this.readLocks[shard]
	writeLock := &this.writeLocks[shard]

	(*readLock) <- true
	(*writeLock) <- true

	this.users[shard][userID] = updateFunc(this.users[shard][userID])

	<-*readLock
	<-*writeLock

}

func (this *SafeUserMap) Get(userID uidType) *User {
	shard := this.getShard(userID)
	readLock := &this.readLocks[shard]
	writeLock := &this.writeLocks[shard]

	*readLock <- true
	*writeLock <- true
	<-*writeLock

	ret, _ := this.users[shard][userID]

	<-*readLock

	return ret
}

func (this *SafeUserMap) Exists(userID uidType) bool {
	shard := this.getShard(userID)
	readLock := &this.readLocks[shard]
	writeLock := &this.writeLocks[shard]
	*readLock <- true
	*writeLock <- true
	<-*writeLock

	_, exists := this.users[shard][userID]

	<-*readLock

	return exists
}

func (this *SafeUserMap) Delete(userID uidType) {
	shard := this.getShard(userID)
	readLock := &this.readLocks[shard]
	writeLock := &this.writeLocks[shard]

	*readLock <- true
	*writeLock <- true

	delete(this.users[shard], userID)

	<-*readLock
	<-*writeLock
}

type SafeRoomMap struct {
	users map[string]map[uidType]bool
	locks [](chan bool)
}

func (this *SafeRoomMap) Init(shards uint32) {
	this.users = make(map[string]map[uidType]bool)
	this.locks = make([](chan bool), shards)
}
