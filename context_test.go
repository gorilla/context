// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package context

import (
	"net/http"
	"testing"
)

type keyType int

const (
	key1 keyType = iota
	key2
)

func TestContext(t *testing.T) {
	assertEqual := func(val interface{}, exp interface{}) {
		if val != exp {
			t.Errorf("Expected %v, got %v.", exp, val)
		}
	}

	r, _ := http.NewRequest("GET", "http://localhost:8080/", nil)
	emptyR, _ := http.NewRequest("GET", "http://localhost:8080/", nil)

	// Get()
	assertEqual(Get(r, key1), nil)

	// Set()
	Set(r, key1, "1")
	assertEqual(Get(r, key1), "1")
	assertEqual(len(data[r]), 1)

	Set(r, key2, "2")
	assertEqual(Get(r, key2), "2")
	assertEqual(len(data[r]), 2)

	//GetOk
	value, ok := GetOk(r, key1)
	assertEqual(value, "1")
	assertEqual(ok, true)

	value, ok = GetOk(r, "not exists")
	assertEqual(value, nil)
	assertEqual(ok, false)

	Set(r, "nil value", nil)
	value, ok = GetOk(r, "nil value")
	assertEqual(value, nil)
	assertEqual(ok, true)

	// GetAll()
	values := GetAll(r)
	assertEqual(len(values), 3)

	// GetAll() for empty request
	values = GetAll(emptyR)
	if values != nil {
		t.Error("GetAll didn't return nil value for invalid request")
	}

	// GetAllOk()
	values, ok = GetAllOk(r)
	assertEqual(len(values), 3)
	assertEqual(ok, true)

	// GetAllOk() for empty request
	values, ok = GetAllOk(emptyR)
	assertEqual(value, nil)
	assertEqual(ok, false)

	// Delete()
	Delete(r, key1)
	assertEqual(Get(r, key1), nil)
	assertEqual(len(data[r]), 2)

	Delete(r, key2)
	assertEqual(Get(r, key2), nil)
	assertEqual(len(data[r]), 1)

	// Clear()
	Clear(r)
	assertEqual(len(data), 0)
}

func parallelReader(r *http.Request, key string, iterations int, wait, done chan struct{}) {
	<-wait
	for i := 0; i < iterations; i++ {
		Get(r, key)
	}
	done <- struct{}{}

}

func parallelWriter(r *http.Request, key, value string, iterations int, wait, done chan struct{}) {
	<-wait
	for i := 0; i < iterations; i++ {
		Get(r, key)
	}
	done <- struct{}{}

}

func BenchmarkMutex(b *testing.B) {

	b.StopTimer()
	r, _ := http.NewRequest("GET", "http://localhost:8080/", nil)
	done := make(chan struct{})
	numWriters := 64
	numReaders := numWriters * 8
	iterations := 128
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		wait := make(chan struct{})

		for i := 0; i < numReaders; i++ {
			go parallelReader(r, "test", iterations, wait, done)
		}

		for i := 0; i < numWriters; i++ {
			go parallelWriter(r, "test", "123", iterations, wait, done)
		}

		close(wait)

		for i := 0; i < numReaders+numWriters; i++ {
			<-done
		}

	}

}
