package concurrentratelimittable

import (
	"chatExtensionServer/internal/types"
	"encoding/binary"
	"hash/fnv"
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
type ValueType = uint16

// CreateNewRateLimitTable creates a new rate limit table
func CreateNewRateLimitTable() ConcurrentHashMap {
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
