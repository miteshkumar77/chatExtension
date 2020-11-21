package main

import (
	"strconv"
	"sync"
	"testing"
)

func TestMapCreation(t *testing.T) {
	var m SafeUserMap
	m.Init(5)

	if m.Size() != 0 {
		t.Error("New map size should be 0")
	}

	if m.Shards() != 5 {
		t.Error("Incorrect number of shards")
	}
}

func TestSet(t *testing.T) {
	var m SafeUserMap
	m.Init(5)
	u1 := User{"albert", uidType(1), "a", nil}
	u2 := User{"john", uidType(2), "a", nil}
	m.Set(u1.UserID, &u1)
	m.Set(u2.UserID, &u2)
	if m.Size() != 2 {
		t.Error("map should contain exactly 2 users.")
	}
}

func TestGet(t *testing.T) {
	var m SafeUserMap
	m.Init(5)

	val := m.Get(uidType(1))

	if val != nil {
		t.Error("Getting missing key should return nil")
	}

	u1 := User{"albert", uidType(1), "a", nil}
	u2 := User{"john", uidType(2), "a", nil}
	m.Set(u1.UserID, &u1)
	m.Set(u2.UserID, &u2)

	if m.Get(u1.UserID) != &u1 || m.Get(u2.UserID) != &u2 {
		t.Error("map does not contain a user that was added")
	}
}

func TestExists(t *testing.T) {
	var m SafeUserMap
	m.Init(5)

	if m.Exists(uidType(1)) != false {
		t.Error("Exists returns true for a non-existing userID")
	}

	u1 := User{"albert", uidType(1), "a", nil}
	m.Set(u1.UserID, &u1)

	if m.Exists(uidType(1)) != true {
		t.Error("Exists returns false for an existing userID")
	}
}

func TestDelete(t *testing.T) {
	var m SafeUserMap
	m.Init(5)

	u1 := User{"albert", uidType(1), "a", nil}
	m.Set(u1.UserID, &u1)

	m.Delete(u1.UserID)

	if m.Size() != 0 {
		t.Error("Map did not decrease in size after deleting existing key")
	}

	if m.Exists(u1.UserID) != false {
		t.Error("Element still exists after delete")
	}

	m.Delete(u1.UserID)

}

func TestSetExistingKey(t *testing.T) {
	var m SafeUserMap
	m.Init(5)

	u1 := User{"albert", uidType(1), "a", nil}
	u2 := User{"john", uidType(1), "a", nil}
	m.Set(u1.UserID, &u1)
	m.Set(u2.UserID, &u2)

	if m.Size() != 1 {
		t.Error("Setting with the same keys twice didn't result in size of 1")
	}

	if m.Get(u2.UserID) != &u2 {
		t.Error("Key's value failed to overwrite to second value")
	}
}

func TestAtomicSet(t *testing.T) {
	var m SafeUserMap
	m.Init(5)

	u1 := User{"albert", uidType(1), "a", nil}
	const newName string = "john"
	m.Set(u1.UserID, &u1)

	m.AtomicUpdate(u1.UserID, func(oldUser *User) *User {
		return &User{newName, oldUser.UserID, oldUser.videoID, nil}
	})

	result := m.Get(u1.UserID)

	if result.UserName != "john" {
		t.Errorf("expected UserName: %s but got %s", newName, result.UserName)
	}
}

func TestCount(t *testing.T) {
	var m SafeUserMap
	m.Init(5)
	for i := 0; i < 10; i++ {
		m.Set(uidType(i+1), &User{"auser", uidType(i), "a", nil})
	}
	if m.Size() != 10 {
		t.Error("Size not 10 after 10 unique key inserts")
	}
}

func TestConcurrentUpdates(t *testing.T) {

	const iterations = 1000
	const goroutines = 6
	var wg sync.WaitGroup
	wg.Add(goroutines)

	var m SafeUserMap
	m.Init(5)

	const id = 1
	u1 := User{"albert", uidType(0), "a", nil}

	m.Set(id, &u1)
	for goroutine := 0; goroutine < goroutines; goroutine++ {
		go func() {
			for iter := 0; iter < iterations; iter++ {

				m.AtomicUpdate(id, func(oldUser *User) *User {
					return &User{oldUser.UserName, oldUser.UserID + 1, oldUser.videoID, nil}
				})
			}
			wg.Done()
		}()
	}

	wg.Wait()

	if m.Get(id).UserID != uidType(iterations*goroutines) {
		t.Errorf("Final userID expected: %d, but got %d", int(iterations*goroutines), int(u1.UserID))
	}
}

func TestConcurrentDeletes(t *testing.T) {

	const additions = 10000

	const removalsPerRoutine = 30
	const goroutines = 35
	const remaining = additions - removalsPerRoutine*goroutines
	if remaining < 0 {
		t.Error("Cannot have more total removals than additions")
	}

	var wg sync.WaitGroup
	wg.Add(goroutines)

	var m SafeUserMap
	m.Init(5)

	for id := 0; id < additions; id++ {
		m.Set(uidType(id), &User{strconv.Itoa(id), uidType(id), strconv.Itoa(id) + "vid", nil})
	}

	for routine := 0; routine < goroutines; routine++ {
		go func(start int, incr int) {
			for index, elem := start, 0; elem < removalsPerRoutine; index, elem = index+incr, elem+1 {
				m.Delete(uidType(index))
			}
			wg.Done()
		}(routine, goroutines)
	}

	wg.Wait()

	result := m.Size()

	if result != remaining {
		t.Errorf("Expected %d but got %d for size after concurrent deletes", remaining, result)
	}
}
