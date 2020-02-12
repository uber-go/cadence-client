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

package persistence

import (
	"encoding/json"
	"fmt"
	s "go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
)

type (
	// PayloadSerializer is used by persistence to serialize/deserialize history event(s) and others
	// It will only be used inside persistence, so that serialize/deserialize is transparent for application
	PayloadSerializer interface {
		// deserialize history events
		DeserializeBatchEvents(data *s.DataBlob) ([]*s.HistoryEvent, error)
	}

	// CadenceSerializationError is an error type for cadence serialization
	CadenceSerializationError struct {
		msg string
	}

	// CadenceDeserializationError is an error type for cadence deserialization
	CadenceDeserializationError struct {
		msg string
	}

	// UnknownEncodingTypeError is an error type for unknown or unsupported encoding type
	UnknownEncodingTypeError struct {
		encodingType s.EncodingType
	}
)

// DeserializeBatchEvents will deserialize blob data to history event data
func DeserializeBatchEvents(data *s.DataBlob) ([]*s.HistoryEvent, error) {
	if data == nil {
		return nil, nil
	}
	var events []*s.HistoryEvent
	if data != nil && len(data.Data) == 0 {
		return events, nil
	}
	err := deserialize(data, &events)
	return events, err
}

func deserialize(data *s.DataBlob, target interface{}) error {
	if data == nil {
		return nil
	}
	if len(data.Data) == 0 {
		return NewCadenceDeserializationError("DeserializeEvent empty data")
	}
	var err error

	switch *(data.EncodingType) {
	case s.EncodingTypeThriftRW:
		err = thriftrwDecode(data.Data, target)
	case s.EncodingTypeJSON: // For backward-compatibility
		err = json.Unmarshal(data.Data, target)

	}

	if err != nil {
		return NewCadenceDeserializationError(fmt.Sprintf("DeserializeBatchEvents encoding: \"%v\", error: %v", data.EncodingType, err.Error()))
	}
	return nil
}

func thriftrwDecode(data []byte, target interface{}) error {
	switch target := target.(type) {
	case *[]*s.HistoryEvent:
		history := s.History{Events: *target}
		if err := common.Decode(data, &history); err != nil {
			return err
		}
		*target = history.GetEvents()
		return nil
	case *s.HistoryEvent:
		return common.Decode(data, target)
	case *s.Memo:
		return common.Decode(data, target)
	case *s.ResetPoints:
		return common.Decode(data, target)
	case *s.BadBinaries:
		return common.Decode(data, target)
	case *s.VersionHistories:
		return common.Decode(data, target)
	default:
		return nil
	}
}

// NewUnknownEncodingTypeError returns a new instance of encoding type error
func NewUnknownEncodingTypeError(encodingType s.EncodingType) error {
	return &UnknownEncodingTypeError{encodingType: encodingType}
}

func (e *UnknownEncodingTypeError) Error() string {
	return fmt.Sprintf("unknown or unsupported encoding type %v", e.encodingType)
}

func (e *CadenceSerializationError) Error() string {
	return fmt.Sprintf("cadence serialization error: %v", e.msg)
}

// NewCadenceDeserializationError returns a CadenceDeserializationError
func NewCadenceDeserializationError(msg string) *CadenceDeserializationError {
	return &CadenceDeserializationError{msg: msg}
}

func (e *CadenceDeserializationError) Error() string {
	return fmt.Sprintf("cadence deserialization error: %v", e.msg)
}
