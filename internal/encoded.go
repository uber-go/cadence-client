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
	"reflect"

	"go.uber.org/cadence/encoded"
)

// defaultDataConverter uses thrift encoder/decoder when possible, for everything else use json.
type defaultDataConverter struct{}

var defaultJsonDataConverter = &defaultDataConverter{}

// DefaultDataConverter is default data converter used by Cadence worker
var DefaultDataConverter = getDefaultDataConverter()

// getDefaultDataConverter return default data converter used by Cadence worker
func getDefaultDataConverter() encoded.DataConverter {
	return defaultJsonDataConverter
}

func (dc *defaultDataConverter) ToData(r ...interface{}) ([]byte, error) {
	if len(r) == 1 && isTypeByteSlice(reflect.TypeOf(r[0])) {
		return r[0].([]byte), nil
	}

	var encoder encoding
	if isUseThriftEncoding(r) {
		encoder = &thriftEncoding{}
	} else {
		encoder = &jsonEncoding{}
	}

	data, err := encoder.Marshal(r)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (dc *defaultDataConverter) FromData(data []byte, to ...interface{}) error {
	if len(to) == 1 && isTypeByteSlice(reflect.TypeOf(to[0])) {
		reflect.ValueOf(to[0]).Elem().SetBytes(data)
		return nil
	}

	var encoder encoding
	if isUseThriftDecoding(to) {
		encoder = &thriftEncoding{}
	} else {
		encoder = &jsonEncoding{}
	}

	return encoder.Unmarshal(data, to)
}
