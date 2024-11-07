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
	ts := []interface{}{
		&mockThriftStruct{},
		&mockThriftStruct{},
	}

	assert.True(t, IsUseThriftDecoding(ts))

	ts = []interface{}{
		&mockThriftStruct{},
		"string",
	}

	assert.False(t, IsUseThriftDecoding(ts))
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
