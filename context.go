// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package context

import (
	"io"
	"net/http"
)

type attributeStore map[interface{}]interface{}

type decoratedReadCloser struct {
	body       io.ReadCloser
	attributes *attributeStore
}

func decorate(r *http.Request) *decoratedReadCloser {
	result := decoratedReadCloser{body: r.Body}
	r.Body = &result
	return &result
}

func (w *decoratedReadCloser) Read(p []byte) (n int, err error) {
	return w.body.Read(p)
}

func (w *decoratedReadCloser) Close() error {
	return w.body.Close()
}

func attributes(r *http.Request) attributeStore {
	if atts, ok := peekAttributes(r); ok {
		return *atts
	}

	atts := make(attributeStore)

	// decorate the request
	decorator := decorate(r)
	decorator.attributes = &atts

	return atts
}

func peekAttributes(r *http.Request) (*attributeStore, bool) {
	if v, ok := r.Body.(*decoratedReadCloser); ok {
		return v.attributes, true
	}
	return nil, false
}

// Set stores a value for a given key in a given request.
func Set(r *http.Request, key, val interface{}) {
	data := attributes(r)
	data[key] = val
}

// Get returns a value stored for a given key in a given request.
func Get(r *http.Request, key interface{}) interface{} {
	data := attributes(r)
	if v, ok := data[key]; ok {
		return v
	}
	return nil
}

// GetOk returns stored value and presence state like multi-value return of map access.
func GetOk(r *http.Request, key interface{}) (interface{}, bool) {
	data := attributes(r)
	v, ok := data[key]
	return v, ok
}

// GetAll returns all stored values for the request as a map. Nil is returned for invalid requests.
func GetAll(r *http.Request) map[interface{}]interface{} {
	if atts, ok := peekAttributes(r); ok {
		return *atts
	}
	return nil
}

// GetAllOk returns all stored values for the request as a map and a boolean value that indicates if
// the request was registered.
func GetAllOk(r *http.Request) (map[interface{}]interface{}, bool) {
	if atts, ok := peekAttributes(r); ok {
		return *atts, ok
	}
	return nil, false
}

// Delete removes a value stored for a given key in a given request.
func Delete(r *http.Request, key interface{}) {
	delete(attributes(r), key)
}

// Clear removes all values stored for a given request.
//
// This is usually called by a handler wrapper to clean up request
// variables at the end of a request lifetime. See ClearHandler().
func Clear(r *http.Request) {
	clear(r)
}

// clear is Clear without the lock.
func clear(r *http.Request) {
	data := attributes(r)
	for k := range data {
		delete(data, k)
	}
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
	return 0
}

// ClearHandler wraps an http.Handler and clears request values at the end
// of a request lifetime.
func ClearHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer Clear(r)
		h.ServeHTTP(w, r)
	})
}
