package concurrentmap

import (
	"chatExtensionServer/internal/types"
	"encoding/binary"
	"hash/fnv"
	"sync"
)

// SizeType used for size properties
type SizeType = uint64

// IndexType used for indexing the map
type IndexType = uint64

// RatioType used for fractional properties
type RatioType = float32

// KeyType used for the key's type
type KeyType = types.UIDType

// ValueType used for the value's type
type ValueType = types.User

// CreateNewUserTable creates a new room table
func CreateNewUserTable() ConcurrentHashMap {
	hs := func(key KeyType) IndexType {

		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(key))
		h := fnv.New64()
		h.Write([]byte(b))
		return IndexType(h.Sum64())
	}

	numShards := uint64(100)
	return CreateNewConcurrentHashMap(0.7, 100, hs, numShards)
}

///---------------------------------------------------------------------------------------------------------
// hashElement class BEGIN

type hashElement struct {
	isOccupied bool
	isTomb     bool
	key        KeyType
	value      ValueType
}

func (element *hashElement) init() {
	element.isOccupied = false
	element.isTomb = false
}

// hashElement class END
///---------------------------------------------------------------------------------------------------------

///---------------------------------------------------------------------------------------------------------
// hashShard class BEGIN

type hashShard struct {
	occupancyRatio RatioType
	occupied       SizeType
	size           SizeType
	data           *[]hashElement
	hashFunc       *func(KeyType) IndexType
}

/// Private

// CallBackIterator calls a callback function cb for every element in the map
// in no particular order
func (shard *hashShard) shardCallBackIterator(cb *func(KeyType, ValueType)) {
	for i := 0; i < len(*shard.data); i++ {
		if (*shard.data)[i].isOccupied {
			(*cb)((*shard.data)[i].key, (*shard.data)[i].value)
		}
	}
}

func (shard *hashShard) getSize() SizeType {
	return shard.size
}

// Insert inserts a key value pair into the collection
func (shard *hashShard) shardInsert(key KeyType, hashValue IndexType, value ValueType) bool {
	shard.rehash()
	initIdx := shard.searchExists(key, hashValue)

	if (*shard.data)[initIdx].isOccupied && (*shard.data)[initIdx].key == key {
		return false
	}

	insertIdx := shard.searchToInsert(key, hashValue)
	if !(*shard.data)[insertIdx].isOccupied {
		shard.occupied++
	}

	(*shard.data)[insertIdx] = hashElement{true, true, key, value}
	shard.size++
	return true
}

// GetVal returns (true, value) if key exists, otherwise (false, nil)
func (shard *hashShard) shardGetVal(key KeyType, hashValue IndexType) (bool, ValueType) {
	initIdx := shard.searchExists(key, hashValue)
	if (*shard.data)[initIdx].isOccupied && (*shard.data)[initIdx].key == key {
		return true, (*shard.data)[initIdx].value
	}
	var def ValueType
	return false, def
}

// Set key's value to value, whether it exists or not
func (shard *hashShard) shardSet(key KeyType, hashValue IndexType, value ValueType) {
	shard.rehash()
	initIdx := shard.searchExists(key, hashValue)
	if (*shard.data)[initIdx].isOccupied && (*shard.data)[initIdx].key == key {
		(*shard.data)[initIdx].value = value
		return
	}
	shard.shardInsert(key, hashValue, value)
}

// Contains returns true iff the map contains key
func (shard *hashShard) shardContains(key KeyType, hashValue IndexType) bool {
	initIdx := shard.searchExists(key, hashValue)
	if (*shard.data)[initIdx].isOccupied && (*shard.data)[initIdx].key == key {
		return true
	}

	return false
}

// Erase deletes the key, value pair from the collection iff it exists, and return true
// otherwise it returns false
func (shard *hashShard) shardErase(key KeyType, hashValue IndexType) bool {
	firstOcc := shard.searchExists(key, hashValue)
	if !(*shard.data)[firstOcc].isTomb || !(*shard.data)[firstOcc].isOccupied {
		return false
	}

	if (*shard.data)[firstOcc].key == key {
		shard.size--
		(*shard.data)[firstOcc].isOccupied = false
		(*shard.data)[firstOcc].isTomb = true
		return true
	}
	return false
}

func (shard *hashShard) init(occupancyRatio RatioType, initialSize SizeType, hashFunc *func(KeyType) IndexType) {
	shard.occupancyRatio = occupancyRatio
	shard.occupied, shard.size = 0, 0
	tmp := make([]hashElement, initialSize)
	shard.data = &tmp
	shard.hashFunc = hashFunc
	for i := 0; i < len(*shard.data); i++ {
		(*shard.data)[i].init()
	}
}

// searchToInsert walks forward from a particular key's hash index until the next unoccupied index
func (shard *hashShard) searchToInsert(key KeyType, hashValue IndexType) IndexType {
	// hval := shard.mhash(key)
	var x IndexType = 0
	idx := (hashValue%IndexType(len(*shard.data)) + shard.probe(x)) % IndexType(len(*shard.data))

	for (*shard.data)[idx].isOccupied {
		x++
		idx = (hashValue%IndexType(len(*shard.data)) + shard.probe(x)) % IndexType(len(*shard.data))
	}
	return idx
}

// searchExists walks forward from a particular key's hash index until the next non-tomb index
func (shard *hashShard) searchExists(key KeyType, hashValue IndexType) IndexType {
	var x IndexType = 0
	var idx = (hashValue%IndexType(len(*shard.data)) + shard.probe(x)) % IndexType(len(*shard.data))
	for (*shard.data)[idx].isTomb {
		if (*shard.data)[idx].isOccupied && (*shard.data)[idx].key == key {
			return idx
		}
		x++
		idx = (hashValue%IndexType(len(*shard.data)) + shard.probe(x)) % IndexType(len(*shard.data))
	}

	return idx
}

func (shard *hashShard) mhash(key KeyType) IndexType {
	return (*shard.hashFunc)(key)
}

func (shard *hashShard) probe(x IndexType) IndexType {
	return x
}

func (shard *hashShard) rehash() {
	if RatioType(shard.occupied)/RatioType(len(*shard.data)) < shard.occupancyRatio {
		return
	}

	tmp := make([]hashElement, len(*shard.data)*2)
	nData := &tmp

	for i := 0; i < len(*nData); i++ {
		(*nData)[i].init()
	}

	nData, shard.data = shard.data, nData
	shard.size = 0
	shard.occupied = 0

	for i := 0; i < len(*nData); i++ {
		if (*nData)[i].isOccupied {
			shard.shardInsert((*nData)[i].key, shard.mhash((*nData)[i].key), (*nData)[i].value)
		}
	}
}

// hashShard class END
///---------------------------------------------------------------------------------------------------------

///---------------------------------------------------------------------------------------------------------
// ConcurrentHashMap class BEGIN

// ConcurrentHashMap is a thread safe key value O(1) look up data structure
type ConcurrentHashMap struct {
	hashFunc *func(KeyType) IndexType
	shards   []hashShard
	RWLocks  []sync.RWMutex
}

/// Public

// CreateNewConcurrentHashMap creates a new empty hashmap
func CreateNewConcurrentHashMap(occupancyRatio RatioType,
	initialSize SizeType,
	hashFunc func(KeyType) IndexType, shards SizeType) ConcurrentHashMap {
	ret := ConcurrentHashMap{}
	ret.init(occupancyRatio, initialSize, hashFunc, shards)
	return ret
}

// CallBackUpdate allows a callback to update a value associated with a key based on
// its original value if it exists
func (table *ConcurrentHashMap) CallBackUpdate(key KeyType, cb func(ValueType) ValueType) {
	hashValue := table.mhash(key)
	shard := table.getShard(hashValue)

	table.RWLocks[shard].Lock()

	exists, value := table.shards[shard].shardGetVal(key, hashValue)
	if exists {
		table.shards[shard].shardSet(key, hashValue, cb(value))
	}

	table.RWLocks[shard].Unlock()

}

// CallBackAction calls a callback on the value of a key if the key exists
func (table *ConcurrentHashMap) CallBackAction(key KeyType, cb func(ValueType)) {
	hashValue := table.mhash(key)
	shard := table.getShard(hashValue)

	table.RWLocks[shard].RLock()

	exists, value := table.shards[shard].shardGetVal(key, hashValue)
	if exists {
		cb(value)
	}

	table.RWLocks[shard].RUnlock()

}

// CallBackIterator calls a callback function cb for every element in the map
// in no particular order
func (table *ConcurrentHashMap) CallBackIterator(cb func(KeyType, ValueType)) {

	for shardIndex, shard := range table.shards {

		table.RWLocks[shardIndex].RLock()

		shard.shardCallBackIterator(&cb)

		table.RWLocks[shardIndex].RUnlock()
	}
}

// Insert inserts a key value pair into the collection if it does not exist
// and returns true, otherwise just returns false
func (table *ConcurrentHashMap) Insert(key KeyType, value ValueType) bool {
	hashValue := table.mhash(key)
	shard := table.getShard(hashValue)

	table.RWLocks[shard].Lock()

	wasPairInserted := table.shards[shard].shardInsert(key, hashValue, value)

	table.RWLocks[shard].Unlock()

	return wasPairInserted
}

// GetVal returns (true, value) if key exists, otherwise (false, nil)
func (table *ConcurrentHashMap) GetVal(key KeyType) (bool, ValueType) {
	hashValue := table.mhash(key)
	shard := table.getShard(hashValue)

	table.RWLocks[shard].RLock()

	exists, value := table.shards[shard].shardGetVal(key, hashValue)

	table.RWLocks[shard].RUnlock()

	return exists, value
}

// Set key's value to value, whether it exists or not
func (table *ConcurrentHashMap) Set(key KeyType, value ValueType) {
	hashValue := table.mhash(key)
	shard := table.getShard(hashValue)

	table.RWLocks[shard].Lock()

	table.shards[shard].shardSet(key, hashValue, value)

	table.RWLocks[shard].Unlock()
}

// Contains returns true iff the map contains key
func (table *ConcurrentHashMap) Contains(key KeyType) bool {
	hashValue := table.mhash(key)
	shard := table.getShard(hashValue)

	table.RWLocks[shard].RLock()

	exists := table.shards[shard].shardContains(key, hashValue)

	table.RWLocks[shard].RUnlock()

	return exists
}

// Erase deletes the key, value pair from the collection iff it exists, and return true
// otherwise it returns false
func (table *ConcurrentHashMap) Erase(key KeyType) bool {
	hashValue := table.mhash(key)
	shard := table.getShard(hashValue)

	table.RWLocks[shard].Lock()

	wasErased := table.shards[shard].shardErase(key, hashValue)

	table.RWLocks[shard].Unlock()

	return wasErased
}

// NumShards returns the number of shards that the map has
func (table *ConcurrentHashMap) NumShards() SizeType {
	return SizeType(len(table.shards))
}

// Size returns the number of unique keys that have values within the collection
func (table *ConcurrentHashMap) Size() SizeType {
	var ans SizeType = 0
	for shardIndex, shard := range table.shards {

		table.RWLocks[shardIndex].RLock()

		ans += shard.getSize()

		table.RWLocks[shardIndex].RUnlock()
	}

	return ans
}

/// Private

// getShard gets the shard associated with a hash value of a particular key
func (table *ConcurrentHashMap) getShard(hash IndexType) IndexType {
	return hash % IndexType(len(table.shards))
}

func (table *ConcurrentHashMap) init(occupancyRatio RatioType, initialSize SizeType, hashFunc func(KeyType) IndexType, shards SizeType) {
	table.RWLocks = make([]sync.RWMutex, shards)
	table.shards = make([]hashShard, shards)
	table.hashFunc = &hashFunc
	for i := 0; i < len(table.shards); i++ {
		(table.shards)[i].init(occupancyRatio, initialSize, &hashFunc)
	}

}

func (table *ConcurrentHashMap) mhash(key KeyType) IndexType {
	return (*table.hashFunc)(key)
}

// ConcurrentHashMap class END
///---------------------------------------------------------------------------------------------------------
