// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package context

import (
	"net/http"
	"sync"
	"time"
)

var (
	mutex sync.RWMutex
	data  = make(map[*http.Request]map[interface{}]interface{})
	datat = make(map[*http.Request]int64)
)

// Set stores a value for a given key in a given request.
func Set(r *http.Request, key, val interface{}) {
	mutex.Lock()
	if data[r] == nil {
		data[r] = make(map[interface{}]interface{})
		datat[r] = time.Now().Unix()
	}
	data[r][key] = val
	mutex.Unlock()
}

// Get returns a value stored for a given key in a given request.
func Get(r *http.Request, key interface{}) interface{} {
	mutex.RLock()
	defer mutex.RUnlock()
	if ctx := data[r]; ctx != nil {
		value := ctx[key]
		return value
	}
	return nil
}

// GetOk returns stored value and presence state like multi-value return of map access.
func GetOk(r *http.Request, key interface{}) (interface{}, bool) {
	mutex.RLock()
	defer mutex.RUnlock()
	if _, ok := data[r]; ok {
		value, ok := data[r][key]
		return value, ok
	}
	return nil, false
}

// GetAll returns all stored values for the request as a map. Nil is returned for invalid requests.
func GetAll(r *http.Request) map[interface{}]interface{} {
	mutex.RLock()
	defer mutex.RUnlock()
	if context, ok := data[r]; ok {
		result := make(map[interface{}]interface{}, len(context))
		for k, v := range context {
			result[k] = v
		}
		return result
	}
	return nil
}

// GetAllOk returns all stored values for the request as a map and a boolean value that indicates if
// the request was registered.
func GetAllOk(r *http.Request) (map[interface{}]interface{}, bool) {
	mutex.RLock()
	defer mutex.RUnlock()
	context, ok := data[r]
	result := make(map[interface{}]interface{}, len(context))
	for k, v := range context {
		result[k] = v
	}
	return result, ok
}

// Delete removes a value stored for a given key in a given request.
func Delete(r *http.Request, key interface{}) {
	mutex.Lock()
	if data[r] != nil {
		delete(data[r], key)
	}
	mutex.Unlock()
}

// Clear removes all values stored for a given request.
//
// This is usually called by a handler wrapper to clean up request
// variables at the end of a request lifetime. See ClearHandler().
func Clear(r *http.Request) {
	mutex.Lock()
	clear(r)
	mutex.Unlock()
}

// clear is Clear without the lock.
func clear(r *http.Request) {
	delete(data, r)
	delete(datat, r)
}

// Purge removes request data stored for longer than maxAge, in seconds.
// It returns the amount of requests removed.
//
// If maxAge <= 0, all request data is removed.
//
// This is only used for sanity check: in case context cleaning was not
// properly set some request data can be kept forever, consuming an increasing
// amount of memory. In case this is detected, Purge() must be called
// periodically until the problem is fixed.
func Purge(maxAge int) int {
	mutex.Lock()
	count := 0
	if maxAge <= 0 {
		count = len(data)
		data = make(map[*http.Request]map[interface{}]interface{})
		datat = make(map[*http.Request]int64)
	} else {
		min := time.Now().Unix() - int64(maxAge)
		for r := range data {
			if datat[r] < min {
				clear(r)
				count++
			}
		}
	}
	mutex.Unlock()
	return count
}

// ClearHandler wraps an http.Handler and clears request values at the end
// of a request lifetime.
func ClearHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer Clear(r)
		h.ServeHTTP(w, r)
	})
}
