// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package context

import (
	"net/http"
	"sync"
	"time"
)

var globalContext *Context = NewContext()

type Context struct {
	mutex sync.RWMutex
	data  map[*http.Request]map[interface{}]interface{}
	datat map[*http.Request]int64
}

func NewContext() *Context {
	data := make(map[*http.Request]map[interface{}]interface{})
	datat := make(map[*http.Request]int64)
	return &Context{data: data, datat: datat}
}

// Set stores a value for a given key in a given request.
func Set(r *http.Request, key, val interface{}) {
	globalContext.Set(r, key, val)
}

func (c *Context) Set(r *http.Request, key, val interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.data[r] == nil {
		c.data[r] = make(map[interface{}]interface{})
		c.datat[r] = time.Now().Unix()
	}
	c.data[r][key] = val
}

// Get returns a value stored for a given key in a given request.
func Get(r *http.Request, key interface{}) interface{} {
	return globalContext.Get(r, key)
}

func (c *Context) Get(r *http.Request, key interface{}) interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if c.data[r] != nil {
		return c.data[r][key]
	}
	return nil
}

// Delete removes a value stored for a given key in a given request.
func Delete(r *http.Request, key interface{}) {
	globalContext.Delete(r, key)
}

func (c *Context) Delete(r *http.Request, key interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.data[r] != nil {
		delete(c.data[r], key)
	}
}

// Clear removes all values stored for a given request.
//
// This is usually called by a handler wrapper to clean up request
// variables at the end of a request lifetime. See ClearHandler().
func Clear(r *http.Request) {
	globalContext.Clear(r)
}

func (c *Context) Clear(r *http.Request) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.clear(r)
}

// clear is Clear without the lock.
func (c *Context) clear(r *http.Request) {
	delete(c.data, r)
	delete(c.datat, r)
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
	return globalContext.Purge(maxAge)
}

func (c *Context) Purge(maxAge int) int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	count := 0
	if maxAge <= 0 {
		count = len(c.data)
		c.data = make(map[*http.Request]map[interface{}]interface{})
		c.datat = make(map[*http.Request]int64)
	} else {
		min := time.Now().Unix() - int64(maxAge)
		for r, _ := range c.data {
			if c.datat[r] < min {
				c.clear(r)
				count++
			}
		}
	}
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
