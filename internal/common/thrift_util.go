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

package common

import (
	"bytes"
	"github.com/apache/thrift/lib/go/thrift"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/thriftrw/protocol"
	"go.uber.org/thriftrw/wire"
)

// TSerialize is used to serialize thrift TStruct to []byte
func TSerialize(t thrift.TStruct) (b []byte, err error) {
	return thrift.NewTSerializer().Write(t)
}

// TListSerialize is used to serialize list of thrift TStruct to []byte
func TListSerialize(ts []thrift.TStruct) (b []byte, err error) {
	if ts == nil {
		return
	}

	t := thrift.NewTSerializer()
	t.Transport.Reset()

	// NOTE: we don't write any markers as thrift by design being a streaming protocol doesn't
	// recommend writing length.

	for _, v := range ts {
		if e := v.Write(t.Protocol); e != nil {
			err = thrift.PrependError("error writing TStruct: ", e)
			return
		}
	}

	if err = t.Protocol.Flush(); err != nil {
		return
	}

	if err = t.Transport.Flush(); err != nil {
		return
	}

	b = t.Transport.Bytes()
	return
}

// TDeserialize is used to deserialize []byte to thrift TStruct
func TDeserialize(t thrift.TStruct, b []byte) (err error) {
	return thrift.NewTDeserializer().Read(t, b)
}

// TListDeserialize is used to deserialize []byte to list of thrift TStruct
func TListDeserialize(ts []thrift.TStruct, b []byte) (err error) {
	t := thrift.NewTDeserializer()
	err = nil
	if _, err = t.Transport.Write(b); err != nil {
		return
	}

	for i := 0; i < len(ts); i++ {
		if e := ts[i].Read(t.Protocol); e != nil {
			err = thrift.PrependError("error reading TStruct: ", e)
			return
		}
	}

	return
}

type (
	// ThriftObject represents a thrift object
	ThriftObject interface {
		FromWire(w wire.Value) error
		ToWire() (wire.Value, error)
	}
)

const (
	// used by thriftrw binary codec
	preambleVersion0 byte = 0x59
)

var (
	// MissingBinaryEncodingVersion indicate that the encoding version is missing
	MissingBinaryEncodingVersion = &shared.BadRequestError{Message: "Missing binary encoding version."}
	// InvalidBinaryEncodingVersion indicate that the encoding version is incorrect
	InvalidBinaryEncodingVersion = &shared.BadRequestError{Message: "Invalid binary encoding version."}
	// MsgPayloadNotThriftEncoded indicate message is not thrift encoded
	MsgPayloadNotThriftEncoded = &shared.BadRequestError{Message: "Message payload is not thrift encoded."}
)

// Decode decode the object
func Decode(binary []byte, val ThriftObject) error {
	if len(binary) < 1 {
		return MissingBinaryEncodingVersion
	}

	version := binary[0]
	if version != preambleVersion0 {
		return InvalidBinaryEncodingVersion
	}

	reader := bytes.NewReader(binary[1:])
	wireVal, err := protocol.Binary.Decode(reader, wire.TStruct)
	if err != nil {
		return err
	}

	return val.FromWire(wireVal)
}

// Encode encode the object
func Encode(obj ThriftObject) ([]byte, error) {
	if obj == nil {
		return nil, MsgPayloadNotThriftEncoded
	}
	var writer bytes.Buffer
	// use the first byte to version the serialization
	err := writer.WriteByte(preambleVersion0)
	if err != nil {
		return nil, err
	}
	val, err := obj.ToWire()
	if err != nil {
		return nil, err
	}
	err = protocol.Binary.Encode(val, &writer)
	if err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}
