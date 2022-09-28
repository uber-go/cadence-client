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

import "testing"

func Test_linearRecommender_Recommend(t *testing.T) {
	type fields struct {
		lower        ResourceUnit
		upper        ResourceUnit
		targetUsages Usages
	}
	type args struct {
		currentResource ResourceUnit
		currentUsages   Usages
	}

	defaultFields := fields{
		lower: 5,
		upper: 15,
		targetUsages: map[UsageType]MilliUsage{
			PollerUtilizationRate: 500,
		},
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   ResourceUnit
	}{
		{
			name:   "on target usage",
			fields: defaultFields,
			args: args{
				currentResource: 10,
				currentUsages: map[UsageType]MilliUsage{
					PollerUtilizationRate: 500,
				},
			},
			want: ResourceUnit(10),
		},
		{
			name:   "under utilized, scale down",
			fields: defaultFields,
			args: args{
				currentResource: 10,
				currentUsages: map[UsageType]MilliUsage{
					PollerUtilizationRate: 400,
				},
			},
			want: ResourceUnit(8),
		},
		{
			name:   "under utilized, scale down but bounded",
			fields: defaultFields,
			args: args{
				currentResource: 10,
				currentUsages: map[UsageType]MilliUsage{
					PollerUtilizationRate: 200,
				},
			},
			want: ResourceUnit(5),
		},
		{
			name:   "zero utilization, scale down to min",
			fields: defaultFields,
			args: args{
				currentResource: 10,
				currentUsages: map[UsageType]MilliUsage{
					PollerUtilizationRate: 0,
				},
			},
			want: ResourceUnit(5),
		},
		{
			name:   "over utilized, scale up",
			fields: defaultFields,
			args: args{
				currentResource: 10,
				currentUsages: map[UsageType]MilliUsage{
					PollerUtilizationRate: 600,
				},
			},
			want: ResourceUnit(12),
		},
		{
			name:   "over utilized, scale up but bounded",
			fields: defaultFields,
			args: args{
				currentResource: 10,
				currentUsages: map[UsageType]MilliUsage{
					PollerUtilizationRate: 1000,
				},
			},
			want: ResourceUnit(15),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLinearRecommender(tt.fields.lower, tt.fields.upper, tt.fields.targetUsages)
			if got := l.Recommend(tt.args.currentResource, tt.args.currentUsages); got != tt.want {
				t.Errorf("Recommend() = %v, want %v", got, tt.want)
			}
		})
	}
}
