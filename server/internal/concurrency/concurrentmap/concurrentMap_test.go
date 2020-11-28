package concurrentmap

import (
	"encoding/binary"
	"hash/fnv"
	"sync"
	"testing"
)

func CreateNewMapWithNShards(N uint64) ConcurrentHashMap {
	hs := func(i int) uint64 {

		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(i))
		h := fnv.New64()
		h.Write([]byte(b))
		return uint64(h.Sum64())
	}

	return CreateNewConcurrentHashMap(0.7, 100, hs, N)
}

func TestMapCreation(t *testing.T) {
	var numShards uint64 = 10
	m := CreateNewMapWithNShards(numShards)

	if m.Size() != 0 {
		t.Error("New map size should be 0")
	}
}

func TestSet(t *testing.T) {
	var numShards uint64 = 10
	m := CreateNewMapWithNShards(numShards)

	i1 := 15765
	k1 := 45

	i2 := 2442
	k2 := 34

	m.Set(k1, i1)
	m.Set(k2, i2)

	if m.Size() != 2 {
		t.Error("map should contain exactly 2 users.")
	}

}

func TestGetVal(t *testing.T) {
	var numShards uint64 = 10
	m := CreateNewMapWithNShards(numShards)

	exists, _ := m.GetVal(1)

	if exists == true {
		t.Error("GetVal for a non-existing value should return (false, _), but got (true, _)")
	}

	i1 := 15765
	k1 := 45

	i2 := 2442
	k2 := 34

	m.Set(k1, i1)
	m.Set(k2, i2)

	exists1, val1 := m.GetVal(k1)
	exists2, val2 := m.GetVal(k2)

	if exists1 == false || exists2 == false {
		t.Error("GetVal retuns exists==false for an existing key")
	}

	if val1 != i1 || val2 != i2 {
		t.Error("GetVal does not return same as original value storeds")
	}
}

func TestContains(t *testing.T) {
	var numShards uint64 = 10
	m := CreateNewMapWithNShards(numShards)

	if m.Contains(1) != false {
		t.Error("Contains returned true for a non-existent key")
	}

	i1 := 15765
	k1 := 45

	m.Set(k1, i1)

	if m.Contains(k1) != true {
		t.Error("Contains returned false for an existing key")
	}
}

func TestErase(t *testing.T) {
	var numShards uint64 = 10
	m := CreateNewMapWithNShards(numShards)

	i1 := 15765
	k1 := 45

	m.Set(k1, i1)

	wasErased := m.Erase(k1)

	if wasErased != true {
		t.Error("Erase returned false after call to existing key")
	}

	if m.Size() != 0 {
		t.Error("Map did not decrease in size after erasing an existing key")
	}

	if m.Contains(k1) != false {
		t.Error("Key still exists after erasing it")
	}

	wasErased = m.Erase(k1)

	if wasErased != false {
		t.Error("Erase returned true after call to non-existent key")
	}

}

func TestSetExistingKey(t *testing.T) {
	var numShards uint64 = 10
	m := CreateNewMapWithNShards(numShards)

	i1 := 15765
	k1 := 45

	i2 := 2442

	m.Set(k1, i1)
	m.Set(k1, i2)

	if m.Size() != 1 {
		t.Error("Incorrect size after setting 1 key twice")
	}

	exists, value := m.GetVal(k1)

	if exists != true {
		t.Error("GetVal returned exists == false for an existing key")
	}

	if value != i2 {
		t.Error("GetVal returned incorrect value")
	}

}

func TestCallBackUpdate(t *testing.T) {
	var numShards uint64 = 10
	m := CreateNewMapWithNShards(numShards)

	i1 := 15
	k1 := 45

	updateFunc := func(original int) int {
		return original*3 + 51
	}

	expectedValue := updateFunc(i1)

	m.Set(k1, i1)

	m.CallBackUpdate(k1, updateFunc)

	exists, value := m.GetVal(k1)

	if exists != true {
		t.Error("GetVal returns exist == false for an existing key")
	}

	if value != expectedValue {
		t.Error("GetVal didn't return correct value after applying update function")
	}

}

func TestSize(t *testing.T) {
	var numShards uint64 = 10
	m := CreateNewMapWithNShards(numShards)

	numElements := 109

	for i := 0; i < numElements; i++ {
		m.Set(i+1, i+1)
	}

	if m.Size() != uint64(numElements) {
		t.Error("Size not equal to number of elements added")
	}
}

func TestConcurrentCallBackUpdates(t *testing.T) {
	const numIterations = 1000
	const numGoRoutines = 6

	var wg sync.WaitGroup
	wg.Add(numGoRoutines)

	var numShards uint64 = 10
	m := CreateNewMapWithNShards(numShards)

	const key = 1
	const val = 0

	incr := func(original int) int {
		return original + 1
	}

	m.Set(key, val)
	for goRoutine := 0; goRoutine < numGoRoutines; goRoutine++ {
		go func() {
			for iteration := 0; iteration < numIterations; iteration++ {
				m.CallBackUpdate(key, incr)
			}
			wg.Done()
		}()
	}

	wg.Wait()

	exists, value := m.GetVal(key)

	if exists != true {
		t.Error("GetVal returned exists == false for an existing key")
	}

	if value != int(numIterations*numGoRoutines) {
		t.Errorf("Final val expected: %d, but got %d", int(numIterations*numGoRoutines), value)
	}

}

func TestConcurrentErases(t *testing.T) {
	const additions = 10000

	const removalsPerRoutine = 30
	const numGoRoutines = 35
	const remaining = additions - removalsPerRoutine*numGoRoutines

	if remaining < 0 {
		t.Error("Cannot have more total removals than additions")
	}

	var wg sync.WaitGroup
	wg.Add(numGoRoutines)

	var numShards uint64 = 10
	m := CreateNewMapWithNShards(numShards)

	for id := 0; id < additions; id++ {
		m.Set(id, id)
	}

	for routine := 0; routine < numGoRoutines; routine++ {
		go func(start int, incr int) {
			for index, elem := start, 0; elem < removalsPerRoutine; index, elem = index+incr, elem+1 {
				m.Erase(index)
			}
			wg.Done()
		}(routine, numGoRoutines)
	}

	wg.Wait()

	result := m.Size()

	if result != remaining {
		t.Errorf("Expected %d but got %d for size after concurrent deletes", remaining, result)
	}
}

func TestCallBackIterator(t *testing.T) {
	numElements := 100
	values := make([]int, numElements)
	expectedValues := make([]int, numElements)
	valueFunc := func(index int) int {
		return index*5 + 2
	}
	for i := 0; i < numElements; i++ {
		values[i] = valueFunc(i)
		expectedValues[i] = -1
	}

	var numShards uint64 = 10
	m := CreateNewMapWithNShards(numShards)

	for index, value := range values {
		m.Set(index, value)
	}

	iterFunc := func(cBKey int, cBVal int) {
		expectedValues[cBKey] = cBVal
	}

	m.CallBackIterator(iterFunc)

	for index := range expectedValues {
		if expectedValues[index] != values[index] {
			t.Errorf("Value for index %d expected to be %d but was %d", index, values[index], expectedValues[index])
		}
	}

}
