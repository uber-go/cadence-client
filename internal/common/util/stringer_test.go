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

package util

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_anyToString(t *testing.T) {
	testCases := []struct {
		name             string
		thingToSerialize interface{}
		expected         string
	}{
		{
			name:             "nil",
			thingToSerialize: nil,
			expected:         "<nil>",
		},
		{
			name:             "int",
			thingToSerialize: 1,
			expected:         "1",
		},
		{
			name:             "string",
			thingToSerialize: "test",
			expected:         "test",
		},
		{
			name: "struct",
			thingToSerialize: struct {
				A string
				B int
				C bool
			}{A: "test", B: 1, C: true},
			expected: "(A:test, B:1, C:true)",
		},
		{
			name: "pointer",
			thingToSerialize: &struct {
				A string
				B int
				C bool
			}{A: "test", B: 1, C: true},
			expected: "(A:test, B:1, C:true)",
		},
		{
			name:             "slice",
			thingToSerialize: []int{1, 2, 3},
			expected:         "[1 2 3]",
		},
		{
			name:             "struct with slice",
			thingToSerialize: struct{ A []int }{A: []int{1, 2, 3}},
			expected:         "(A:[len=3])",
		},
		{
			name:             "struct with blob",
			thingToSerialize: struct{ A []byte }{A: []byte("blob-data")},
			expected:         "(A:[blob-data])",
		},
		{
			name:             "struct with struct",
			thingToSerialize: struct{ A struct{ B int } }{A: struct{ B int }{B: 1}},
			expected:         "(A:(B:1))",
		},
		{
			name:             "struct with pointer",
			thingToSerialize: struct{ A *int }{A: new(int)},
			expected:         "(A:0)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := anyToString(tc.thingToSerialize)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func Test_byteSliceToString(t *testing.T) {
	data := []byte("blob-data")
	v := reflect.ValueOf(data)
	strVal := valueToString(v)

	require.Equal(t, "[blob-data]", strVal)

	intBlob := []int32{1, 2, 3}
	v2 := reflect.ValueOf(intBlob)
	strVal2 := valueToString(v2)

	require.Equal(t, "[len=3]", strVal2)
}
