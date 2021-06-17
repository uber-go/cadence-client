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

package api

import (
	"github.com/gogo/protobuf/types"
	"time"
)

func DurationToProto(d time.Duration) *types.Duration {
	return types.DurationProto(d)
}

func DurationFromProto(d *types.Duration) time.Duration {
	if d == nil {
		return time.Duration(0)
	}
	duration, err := types.DurationFromProto(d)
	if err != nil {
		panic(err)
	}
	return duration
}

func SecondsFromProto(d *types.Duration) int32 {
	if d == nil {
		return 0
	}
	duration, err := types.DurationFromProto(d)
	if err != nil {
		panic(err)
	}
	return int32(duration / time.Second)
}

func SecondsToProto(s int32) *types.Duration {
	return DurationToProto(time.Duration(s) * time.Second)
}

func SecondsPtrToProto(s *int32) *types.Duration {
	if s == nil {
		return nil
	}
	return DurationToProto(time.Duration(*s) * time.Second)
}

func TimeFromProto(t *types.Timestamp) time.Time {
	if t == nil {
		return time.Time{}
	}
	timestamp, err := types.TimestampFromProto(t)
	if err != nil {
		panic(err)
	}
	return timestamp
}

func TimeToProto(t time.Time) *types.Timestamp {
	timestamp, err := types.TimestampProto(t)
	if err != nil {
		panic(err)
	}
	return timestamp
}
