package concurrentroomtable

import (
	"chatExtensionServer/internal/types"
	"hash/fnv"
)

// SizeType used for size properties
type SizeType = uint64

// IndexType used for indexing the map
type IndexType = uint64

// RatioType used for fractional properties
type RatioType = float32

// KeyType used for the key's type
type KeyType = string

// ValueType used for the value's type
type ValueType = map[types.UIDType]bool

// CreateNewRoomTable creates a new room table
func CreateNewRoomTable() ConcurrentHashMap {
	hs := func(key KeyType) IndexType {

		h := fnv.New64a()
		h.Write([]byte(key))
		return h.Sum64()
	}

	numShards := uint64(100)
	return CreateNewConcurrentHashMap(0.7, 100, hs, numShards)
}
