// Copyright (c) 2020 Uber Technologies, Inc.
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

package serializer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"go.uber.org/cadence/v2/.gen/go/shared"
	apiv1 "go.uber.org/cadence/v2/.gen/proto/api/v1"
	"go.uber.org/cadence/v2/internal/api"
	"go.uber.org/cadence/v2/internal/api/thrift"
	"go.uber.org/thriftrw/protocol"
	"go.uber.org/thriftrw/wire"
)

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
	MissingBinaryEncodingVersion = &api.BadRequestError{Message: "Missing binary encoding version."}
	// InvalidBinaryEncodingVersion indicate that the encoding version is incorrect
	InvalidBinaryEncodingVersion = &api.BadRequestError{Message: "Invalid binary encoding version."}
	// MsgPayloadNotThriftEncoded indicate message is not thrift encoded
	MsgPayloadNotThriftEncoded = &api.BadRequestError{Message: "Message payload is not thrift encoded."}
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

// SerializeBatchEvents will serialize history event data to blob data
func SerializeBatchEvents(events []*apiv1.HistoryEvent, encodingType apiv1.EncodingType) (*apiv1.DataBlob, error) {
	return serialize(events, encodingType)
}

// DeserializeBatchEvents will deserialize blob data to history event data
func DeserializeBatchEvents(data *apiv1.DataBlob) ([]*apiv1.HistoryEvent, error) {
	if data == nil {
		return nil, nil
	}
	var events []*apiv1.HistoryEvent
	if data != nil && len(data.Data) == 0 {
		return events, nil
	}
	events, err := deserialize(data)
	if err != nil {
		return nil, fmt.Errorf("DeserializeBatchEvents encoding: \"%v\", error: %v", data.EncodingType, err.Error())
	}
	return events, err
}

func serialize(events []*apiv1.HistoryEvent, encodingType apiv1.EncodingType) (*apiv1.DataBlob, error) {
	if events == nil {
		return nil, nil
	}

	var data []byte
	var err error

	switch encodingType {
	case apiv1.EncodingType_ENCODING_TYPE_THRIFTRW:
		data, err = thriftrwEncode(events)
	case apiv1.EncodingType_ENCODING_TYPE_JSON: // For backward-compatibility
		encodingType = apiv1.EncodingType_ENCODING_TYPE_JSON
		data, err = json.Marshal(events)
	case apiv1.EncodingType_ENCODING_TYPE_PROTO3:
		data, err = proto.Marshal(&apiv1.History{Events: events})
	default:
		return nil, fmt.Errorf("unknown or unsupported encoding type %v", encodingType)
	}

	if err != nil {
		return nil, fmt.Errorf("cadence serialization error: %v", err.Error())
	}
	return NewDataBlob(data, encodingType), nil
}

func thriftrwEncode(input interface{}) ([]byte, error) {
	switch input.(type) {
	case []*shared.HistoryEvent:
		return Encode(&shared.History{Events: input.([]*shared.HistoryEvent)})
	case *shared.HistoryEvent:
		return Encode(input.(*shared.HistoryEvent))
	case *shared.Memo:
		return Encode(input.(*shared.Memo))
	case *shared.ResetPoints:
		return Encode(input.(*shared.ResetPoints))
	case *shared.BadBinaries:
		return Encode(input.(*shared.BadBinaries))
	case *shared.VersionHistories:
		return Encode(input.(*shared.VersionHistories))
	default:
		return nil, nil
	}
}

func deserializeThrift(blob []byte) ([]*apiv1.HistoryEvent, error) {
	var target []*shared.HistoryEvent
	if err := thriftrwDecode(blob, target); err != nil {
		return nil, err
	}
	return thrift.ToHistoryEvents(target), nil
}

func deserializeThriftBasedJson(blob []byte) ([]*apiv1.HistoryEvent, error) {
	var target []*shared.HistoryEvent
	if err := json.Unmarshal(blob, target); err != nil {
		return nil, err
	}
	return thrift.ToHistoryEvents(target), nil
}

func deserializeProto(blob []byte) ([]*apiv1.HistoryEvent, error) {
	var target apiv1.History
	if err := proto.Unmarshal(blob, &target); err != nil {
		return nil, err
	}
	return target.Events, nil
}

func deserialize(data *apiv1.DataBlob) ([]*apiv1.HistoryEvent, error) {
	if data == nil {
		return nil, nil
	}
	if len(data.Data) == 0 {
		return nil, errors.New("DeserializeEvent empty data")
	}

	switch data.EncodingType {
	case apiv1.EncodingType_ENCODING_TYPE_THRIFTRW:
		return deserializeThrift(data.Data)
	case apiv1.EncodingType_ENCODING_TYPE_JSON: // For backward-compatibility
		return deserializeThriftBasedJson(data.Data)
	case apiv1.EncodingType_ENCODING_TYPE_PROTO3:
		return deserializeProto(data.Data)
	}

	return nil, nil
}

func thriftrwDecode(data []byte, target interface{}) error {
	switch target := target.(type) {
	case *[]*shared.HistoryEvent:
		history := shared.History{Events: *target}
		if err := Decode(data, &history); err != nil {
			return err
		}
		*target = history.Events
		return nil
	case *shared.HistoryEvent:
		return Decode(data, target)
	case *shared.Memo:
		return Decode(data, target)
	case *shared.ResetPoints:
		return Decode(data, target)
	case *shared.BadBinaries:
		return Decode(data, target)
	case *shared.VersionHistories:
		return Decode(data, target)
	default:
		return nil
	}
}

// NewDataBlob creates new blob data
func NewDataBlob(data []byte, encodingType apiv1.EncodingType) *apiv1.DataBlob {
	if data == nil || len(data) == 0 {
		return nil
	}

	return &apiv1.DataBlob{
		Data:         data,
		EncodingType: encodingType,
	}
}

// DeserializeBlobDataToHistoryEvents deserialize the blob data to history event data
func DeserializeBlobDataToHistoryEvents(
	dataBlobs []*apiv1.DataBlob, filterType apiv1.EventFilterType,
) (*apiv1.History, error) {

	var historyEvents []*apiv1.HistoryEvent

	for _, batch := range dataBlobs {
		events, err := DeserializeBatchEvents(batch)
		if err != nil {
			return nil, err
		}
		if len(events) == 0 {
			return nil, &api.InternalServiceError{
				Message: fmt.Sprintf("corrupted history event batch, empty events"),
			}
		}

		historyEvents = append(historyEvents, events...)
	}

	if filterType == apiv1.EventFilterType_EVENT_FILTER_TYPE_CLOSE_EVENT {
		historyEvents = []*apiv1.HistoryEvent{historyEvents[len(historyEvents)-1]}
	}
	return &apiv1.History{Events: historyEvents}, nil
}
