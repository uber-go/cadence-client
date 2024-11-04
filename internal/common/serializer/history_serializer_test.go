package serializer

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
	"reflect"
	"testing"
)

func TestSerializationRoundup(t *testing.T) {
	for _, encoding := range []shared.EncodingType{shared.EncodingTypeJSON, shared.EncodingTypeThriftRW} {
		t.Run(encoding.String(), func(t *testing.T) {

			events := []*shared.HistoryEvent{
				{
					EventId:   common.Int64Ptr(1),
					Timestamp: common.Int64Ptr(1),
					EventType: common.EventTypePtr(shared.EventTypeActivityTaskCompleted),
					Version:   common.Int64Ptr(1),
					ActivityTaskCompletedEventAttributes: &shared.ActivityTaskCompletedEventAttributes{
						Result: []byte("result"),
					},
				},
			}

			serialized, err := SerializeBatchEvents(events, encoding)
			require.NoError(t, err)

			deserialized, err := DeserializeBatchEvents(serialized)
			require.NoError(t, err)

			assert.Equal(t, events, deserialized)
		})
	}
}

func TestDeserializeBlobDataToHistoryEvents(t *testing.T) {
	events := []*shared.HistoryEvent{
		{
			EventId:   common.Int64Ptr(1),
			Timestamp: common.Int64Ptr(1),
			EventType: common.EventTypePtr(shared.EventTypeDecisionTaskStarted),
			Version:   common.Int64Ptr(1),
			DecisionTaskStartedEventAttributes: &shared.DecisionTaskStartedEventAttributes{
				ScheduledEventId: common.Int64Ptr(1),
			},
		},
		{
			EventId:   common.Int64Ptr(1),
			Timestamp: common.Int64Ptr(1),
			EventType: common.EventTypePtr(shared.EventTypeActivityTaskCompleted),
			Version:   common.Int64Ptr(1),
			ActivityTaskCompletedEventAttributes: &shared.ActivityTaskCompletedEventAttributes{
				Result: []byte("result"),
			},
		},
	}

	serialized, err := SerializeBatchEvents(events, shared.EncodingTypeThriftRW)
	require.NoError(t, err)

	deserialized, err := DeserializeBlobDataToHistoryEvents([]*shared.DataBlob{serialized}, shared.HistoryEventFilterTypeCloseEvent)
	require.NoError(t, err)

	assert.Equal(t, events[1], deserialized.Events[0])
}

func TestDeserializeBlobDataToHistoryEvents_failure(t *testing.T) {
	for _, tc := range []struct {
		name              string
		serialized        *shared.DataBlob
		expectedErrString string
	}{
		{
			name:              "empty blob",
			serialized:        &shared.DataBlob{},
			expectedErrString: "corrupted history event batch, empty events",
		},
		{
			name:              "corrupted blob",
			serialized:        &shared.DataBlob{Data: []byte("corrupted"), EncodingType: shared.EncodingTypeThriftRW.Ptr()},
			expectedErrString: "BadRequestError{Message: Invalid binary encoding version.}",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := DeserializeBlobDataToHistoryEvents([]*shared.DataBlob{tc.serialized}, shared.HistoryEventFilterTypeCloseEvent)
			assert.ErrorContains(t, err, tc.expectedErrString)
		})
	}
}

func TestThriftEncodingRoundtrip(t *testing.T) {
	for _, tc := range []struct {
		input interface{}
	}{
		{
			input: &shared.HistoryEvent{
				EventId:   common.Int64Ptr(1),
				EventType: shared.EventTypeDecisionTaskStarted.Ptr(),
			},
		},
		{
			input: &shared.Memo{
				Fields: map[string][]byte{"key": []byte("value")},
			},
		},
		{
			input: &shared.ResetPoints{
				Points: []*shared.ResetPointInfo{
					{
						BinaryChecksum: common.StringPtr("checksum"),
					},
				},
			},
		},
		{
			input: &shared.BadBinaries{
				Binaries: map[string]*shared.BadBinaryInfo{
					"key": {
						Reason: common.StringPtr("reason"),
					},
				},
			},
		},
		{
			input: &shared.VersionHistories{
				CurrentVersionHistoryIndex: common.Int32Ptr(1),
			},
		},
		{
			input: nil,
		},
	} {
		name := "nil"
		if tc.input != nil {
			name = reflect.TypeOf(tc.input).String()
		}
		t.Run(name, func(t *testing.T) {
			serialized, err := thriftrwEncode(tc.input)
			require.NoError(t, err)

			var deserialized interface{}
			if tc.input != nil {
				deserialized = createEmptyPointer(tc.input)
				err = thriftrwDecode(serialized, deserialized)
				require.NoError(t, err)
			}

			assert.Equal(t, tc.input, deserialized)
		})
	}
}

func TestSerialization_corner_cases(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		res, err := serialize(nil, shared.EncodingTypeThriftRW)
		assert.Nil(t, res)
		assert.NoError(t, err)
	})
	t.Run("unsupported encoding", func(t *testing.T) {
		_, err := SerializeBatchEvents(nil, -1)
		assert.ErrorContains(t, err, "unknown or unsupported encoding type")
	})
	t.Run("serialization error", func(t *testing.T) {
		res, err := Encode(nil)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, MsgPayloadNotThriftEncoded)
	})
}

func createEmptyPointer(input interface{}) interface{} {
	inputType := reflect.TypeOf(input)
	if inputType.Kind() != reflect.Ptr {
		panic("input must be a pointer to a struct")
	}
	elemType := inputType.Elem()
	newInstance := reflect.New(elemType)
	return newInstance.Interface()
}
