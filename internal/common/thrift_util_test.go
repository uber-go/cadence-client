// Copyright (c) 2017-2021 Uber Technologies Inc.
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

package common

import (
	"context"
	"testing"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/stretchr/testify/assert"
)

func TestTListSerialize(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		data, err := TListSerialize(nil)
		assert.NoError(t, err)
		assert.Nil(t, data)
	})
	t.Run("normal", func(t *testing.T) {
		ts := []thrift.TStruct{
			&mockThriftStruct{Field1: "value1", Field2: 1},
			&mockThriftStruct{Field1: "value2", Field2: 2},
		}

		_, err := TListSerialize(ts)
		assert.NoError(t, err)
	})
}

func TestTListDeserialize(t *testing.T) {
	ts := []thrift.TStruct{
		&mockThriftStruct{},
		&mockThriftStruct{},
	}

	data, err := TListSerialize(ts)
	assert.NoError(t, err)

	err = TListDeserialize(ts, data)
	assert.NoError(t, err)
}

func TestIsUseThriftEncoding(t *testing.T) {
	ts := []interface{}{
		&mockThriftStruct{},
		&mockThriftStruct{},
	}

	result := IsUseThriftEncoding(ts)
	assert.True(t, result)

	ts = []interface{}{
		&mockThriftStruct{},
		"string",
	}

	result = IsUseThriftEncoding(ts)
	assert.False(t, result)
}

func TestIsUseThriftDecoding(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		str1 := &mockThriftStruct{}
		str2 := &mockThriftStruct{}
		ts := []interface{}{
			&str1,
			&str2,
		}

		assert.True(t, IsUseThriftDecoding(ts))
	})

	t.Run("fail", func(t *testing.T) {
		ts := []interface{}{
			&mockThriftStruct{},
			"string",
		}

		assert.False(t, IsUseThriftDecoding(ts))
	})
}

func TestIsThriftType(t *testing.T) {
	assert.True(t, IsThriftType(&mockThriftStruct{}))

	assert.False(t, IsThriftType(mockThriftStruct{}))
}

type mockThriftStruct struct {
	Field1 string
	Field2 int
}

func (m *mockThriftStruct) Read(ctx context.Context, iprot thrift.TProtocol) error {
	return nil
}

func (m *mockThriftStruct) Write(ctx context.Context, oprot thrift.TProtocol) error {
	return nil
}

func (m *mockThriftStruct) String() string {
	return ""
}
