package concurrentmap

import (
	"sync"
	"testing"
)

func BenchmarkCallBackIterator(b *testing.B) {
	var numShards uint64 = 10
	m := CreateNewMapWithNShards(numShards)

	numElements := 10000
	for i := 0; i < numElements; i++ {
		m.Set(i, i)
	}

	for i := 0; i < b.N; i++ {
		m.CallBackIterator(func(key int, val int) {
			tmp := key + val
			tmp++
		})
	}
}

func BenchmarkMoreGetLessSet(b *testing.B) {
	var numShards uint64 = 10
	m := CreateNewMapWithNShards(numShards)

	numSets := 10000
	numGetVals := 100000

	var wg sync.WaitGroup
	wg.Add(2)

	b.ResetTimer()
	go func() {
		for i := 0; i < numSets; i++ {
			m.Set(i, i)
		}
		wg.Done()
	}()

	go func() {
		for i := 0; i < numGetVals; i++ {
			m.GetVal(i / 10)
		}
		wg.Done()
	}()
	wg.Wait()
}

func BenchmarkMoreSetLessGet(b *testing.B) {
	var numShards uint64 = 10
	m := CreateNewMapWithNShards(numShards)

	numSets := 100000
	numGetVals := 10000

	var wg sync.WaitGroup
	wg.Add(2)

	b.ResetTimer()
	go func() {
		for i := 0; i < numSets; i++ {
			m.Set(i, i)
		}
		wg.Done()
	}()

	go func() {
		for i := 0; i < numGetVals; i++ {
			m.GetVal(i / 10)
		}
		wg.Done()
	}()
	wg.Wait()
}
