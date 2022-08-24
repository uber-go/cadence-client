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

import "math"

// Recommender a recommendation generator for Resource
type Recommender interface {
	Recommend(currentResource Resource, currentUsages Usages) Resource
}

type linearRecommender struct {
	lower, upper Resource
	targetUsages Usages
}

// NewLinearRecommender create a linear Recommender
func NewLinearRecommender(lower, upper Resource, targetUsages Usages) Recommender {
	return &linearRecommender{
		lower:        lower,
		upper:        upper,
		targetUsages: targetUsages,
	}
}

// Recommend recommends the new value
func (l *linearRecommender) Recommend(currentResource Resource, currentUsages Usages) Resource {
	var recommend float64

	// average recommendation over all UsageType
	for usageType := range currentUsages {
		var r float64
		if l.targetUsages[usageType] == 0 { // avoid division by zero
			r = math.MaxFloat64
		} else {
			r = currentResource.Value() * currentUsages[usageType].Value() / l.targetUsages[usageType].Value()
		}
		// boundary check
		r = math.Min(l.upper.Value(), math.Max(l.lower.Value(), r))
		recommend += r
	}
	recommend /= float64(len(currentUsages))
	return Resource(recommend)
}
