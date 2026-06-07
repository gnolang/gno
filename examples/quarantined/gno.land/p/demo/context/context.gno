// Package context provides a minimal implementation of Go context with support
// for Value and WithValue.
//
// Adapted from https://github.com/golang/go/tree/master/src/context/.
// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package context

type Context interface {
	// Value returns the value associated with this context for key, or nil
	// if no value is associated with key.
	Value(key any) any
}

// Empty returns a non-nil, empty context, similar with context.Background and
// context.TODO in Go.
func Empty() Context {
	return &emptyCtx{}
}

type emptyCtx struct{}

func (ctx emptyCtx) Value(key any) any {
	return nil
}

func (ctx emptyCtx) String() string {
	return "context.Empty"
}

type valueCtx struct {
	parent   Context
	key, val any
}

func (ctx *valueCtx) Value(key any) any {
	if ctx.key == key {
		return ctx.val
	}
	return ctx.parent.Value(key)
}

func stringify(v any) string {
	switch s := v.(type) {
	case stringer:
		return s.String()
	case string:
		return s
	}
	return "non-stringer"
}

type stringer interface {
	String() string
}

func (c *valueCtx) String() string {
	return stringify(c.parent) + ".WithValue(" +
		stringify(c.key) + ", " +
		stringify(c.val) + ")"
}

// WithValue returns a copy of parent in which the value associated with key is
// val.
func WithValue(parent Context, key, val any) Context {
	if key == nil {
		panic("nil key")
	}
	// XXX: if !reflect.TypeOf(key).Comparable() { panic("key is not comparable") }
	return &valueCtx{parent, key, val}
}
