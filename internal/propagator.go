// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package internal

import (
	"context"
	"fmt"
)

// ContextKey to be used for the StringMapPropagator
type ContextKey string

// StringMapPropagator propagates the list of keys across a workflow,
// interpreting the payloads as strings. This is provided as a sample for other
// context propagator implementations.
type StringMapPropagator struct {
	keys map[string]struct{}
}

// NewStringMapPropagator returns a context propagator that propagates a set of
// string key-value pairs across a workflow
func NewStringMapPropagator(keys []string) ContextPropagator {
	keyMap := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		keyMap[key] = struct{}{}
	}
	return &StringMapPropagator{keyMap}
}

// Inject injects values from context into headers for propagation
func (s *StringMapPropagator) Inject(ctx context.Context, writer HeaderWriter) error {
	for key := range s.keys {
		value, ok := ctx.Value(ContextKey(key)).(string)
		if !ok {
			return fmt.Errorf("unable to extract key from context %v", key)
		}
		writer.Set(key, []byte(value))
	}
	return nil
}

// InjectFromWorkflow injects values from context into headers for propagation
func (s *StringMapPropagator) InjectFromWorkflow(ctx Context, writer HeaderWriter) error {
	for key := range s.keys {
		value, ok := ctx.Value(ContextKey(key)).(string)
		if !ok {
			return fmt.Errorf("unable to extract key from context %v", key)
		}
		writer.Set(key, []byte(value))
	}
	return nil
}

// Extract extracts values from headers and puts them into context
func (s *StringMapPropagator) Extract(ctx context.Context, reader HeaderReader) (context.Context, error) {
	if err := reader.ForEachKey(func(key string, value []byte) error {
		if _, ok := s.keys[key]; ok {
			ctx = context.WithValue(ctx, ContextKey(key), string(value))
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return ctx, nil
}

// ExtractToWorkflow extracts values from headers and puts them into context
func (s *StringMapPropagator) ExtractToWorkflow(ctx Context, reader HeaderReader) (Context, error) {
	if err := reader.ForEachKey(func(key string, value []byte) error {
		if _, ok := s.keys[key]; ok {
			ctx = WithValue(ctx, ContextKey(key), string(value))
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return ctx, nil
}
