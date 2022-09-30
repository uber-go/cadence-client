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

package autoscaler

type (
	// ResourceUnit is the unit of scalable resources
	ResourceUnit uint

	// MilliUsage is the custom defined usage of ResourceUnit times 1000
	MilliUsage uint64

	// Usages are different measurements used by a Recommender to provide a recommended ResourceUnit
	Usages map[UsageType]MilliUsage

	// UsageType type of usage
	UsageType string
)

const (
	// PollerUtilizationRate is a scale from 0 to 1 to indicate poller usages
	PollerUtilizationRate UsageType = "pollerUtilizationRate"
)

// Value helper method for type conversion
func (r ResourceUnit) Value() float64 {
	return float64(r)
}

// Value helper method for type conversion
func (u MilliUsage) Value() float64 {
	return float64(u)
}
