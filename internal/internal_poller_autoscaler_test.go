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

package internal

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/atomic"
	"go.uber.org/cadence/internal/common/autoscaler"
	"go.uber.org/zap/zaptest"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func Test_pollerAutoscaler(t *testing.T) {
	type args struct {
		noTaskPoll, taskPoll, unrelated int
		initialPollerCount              int
		minPollerCount                  int
		maxPollerCount                  int
		targetUtilizationInMilli        uint64
		cooldownTime                    time.Duration
		isDryRun                        bool
	}

	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "dry run don't change anything",
			args: args{
				noTaskPoll:               100,
				taskPoll:                 0,
				unrelated:                0,
				initialPollerCount:       10,
				minPollerCount:           1,
				maxPollerCount:           10,
				targetUtilizationInMilli: 500,
				cooldownTime:             time.Millisecond * 200,
				isDryRun:                 true,
			},
			want: 10,
		},
		{
			name: "no utilization, scale to min",
			args: args{
				noTaskPoll:               100,
				taskPoll:                 0,
				unrelated:                0,
				initialPollerCount:       10,
				minPollerCount:           1,
				maxPollerCount:           10,
				targetUtilizationInMilli: 500,
				cooldownTime:             time.Millisecond * 200,
				isDryRun:                 false,
			},
			want: 1,
		},
		{
			name: "low utilization, scale down",
			args: args{
				noTaskPoll:               75,
				taskPoll:                 25,
				unrelated:                0,
				initialPollerCount:       10,
				minPollerCount:           1,
				maxPollerCount:           10,
				targetUtilizationInMilli: 500,
				cooldownTime:             time.Millisecond * 200,
				isDryRun:                 false,
			},
			want: 5,
		},
		{
			name: "over utilized, scale up",
			args: args{
				noTaskPoll:               0,
				taskPoll:                 100,
				unrelated:                0,
				initialPollerCount:       2,
				minPollerCount:           1,
				maxPollerCount:           10,
				targetUtilizationInMilli: 500,
				cooldownTime:             time.Millisecond * 200,
				isDryRun:                 false,
			},
			want: 4,
		},
		{
			name: "over utilized, scale up to max",
			args: args{
				noTaskPoll:               0,
				taskPoll:                 100,
				unrelated:                0,
				initialPollerCount:       6,
				minPollerCount:           1,
				maxPollerCount:           10,
				targetUtilizationInMilli: 500,
				cooldownTime:             time.Millisecond * 200,
				isDryRun:                 false,
			},
			want: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pollerScaler := newPollerScaler(
				tt.args.initialPollerCount,
				tt.args.minPollerCount,
				tt.args.maxPollerCount,
				tt.args.targetUtilizationInMilli,
				tt.args.isDryRun,
				tt.args.cooldownTime,
				zaptest.NewLogger(t))
			done := pollerScaler.Start()
			defer done()

			pollChan := generateRandomPollResults(tt.args.noTaskPoll, tt.args.taskPoll, tt.args.unrelated)
			wg := &sync.WaitGroup{}
			for i := 0; i < tt.args.maxPollerCount; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for pollResult := range pollChan {
						pollerScaler.Acquire(1)
						pollerScaler.CollectUsage(pollResult)
						pollerScaler.Release(1)
					}
				}()
			}
			wg.Wait()
			time.Sleep(time.Millisecond * 500)

			res := pollerScaler.GetCurrent()
			assert.Equal(t, tt.want, int(res))
		})
	}
}

func Test_pollerUsageEstimator(t *testing.T) {
	type args struct {
		noTaskPoll, taskPoll, unrelated int
		pollerCount                     int
	}
	tests := []struct {
		name    string
		args    args
		want    autoscaler.Usages
		wantErr bool
	}{
		{
			name: "400 no-task, 100 task, 100 unrelated",
			args: args{
				noTaskPoll:  400,
				taskPoll:    100,
				unrelated:   100,
				pollerCount: 5,
			},
			want: autoscaler.Usages{
				autoscaler.PollerUtilizationRate: 200,
			},
		},
		{
			name: "0 no-task, 0 task, 100 unrelated",
			args: args{
				noTaskPoll:  0,
				taskPoll:    0,
				unrelated:   100,
				pollerCount: 5,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			estimator := &pollerUsageEstimator{atomicBits: atomic.NewUint64(0)}
			pollChan := generateRandomPollResults(tt.args.noTaskPoll, tt.args.taskPoll, tt.args.unrelated)
			wg := &sync.WaitGroup{}
			for i := 0; i < tt.args.pollerCount; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for pollResult := range pollChan {
						estimator.CollectUsage(pollResult)
					}
				}()
			}
			wg.Wait()

			res, err := estimator.Estimate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, res)
			}
		})
	}
}

type unrelatedPolledTask struct{}

func generateRandomPollResults(noTaskPoll, taskPoll, unrelated int) <-chan interface{} {
	var result []interface{}
	for i := 0; i < noTaskPoll; i++ {
		result = append(result, (*polledTask)(nil))
	}
	for i := 0; i < taskPoll; i++ {
		result = append(result, &polledTask{})
	}
	for i := 0; i < unrelated; i++ {
		result = append(result, &unrelatedPolledTask{})
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})

	pollChan := make(chan interface{}, len(result))
	defer close(pollChan)
	for i := range result {
		pollChan <- result[i]
	}
	return pollChan
}
