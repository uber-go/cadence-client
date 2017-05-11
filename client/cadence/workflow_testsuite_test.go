package cadence

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/facebookgo/clock"
	"github.com/stretchr/testify/suite"
)

const testTaskList = "test-task-list"

type WorkflowTestSuiteUnitTest struct {
	WorkflowTestSuite
	clock *clock.Mock
}

func (s *WorkflowTestSuiteUnitTest) SetupSuite() {
	s.clock = clock.NewMock()
	s.RegisterWorkflow(testWorkflowHello)
	s.RegisterActivity(testActivityHello, testTaskList)
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowTestSuiteUnitTest))
}

func (s *WorkflowTestSuiteUnitTest) Test_ActivityOverride() {
	fakeActivity := func(ctx context.Context, msg string) (string, error) {
		return "fake_" + msg, nil
	}

	env := s.StartWorkflow(testWorkflowHello)
	env.Override(testActivityHello, fakeActivity)
	env.StartDispatcherLoop(time.Second)

	s.True(env.IsTestCompleted())
	s.NoError(env.GetTestError())
	var result string
	err := env.GetTestResult().Get(&result)
	s.NoError(err)
	s.Equal("fake_world", result)
}

func (s *WorkflowTestSuiteUnitTest) Test_OnActivityStartedListener() {
	workflowPart := func(ctx Context) error {
		ctx = WithActivityOptions(ctx, NewActivityOptions().
			WithTaskList(testTaskList).
			WithScheduleToCloseTimeout(time.Minute).
			WithScheduleToStartTimeout(time.Minute).
			WithStartToCloseTimeout(time.Minute).
			WithHeartbeatTimeout(time.Second*20))

		for i := 1; i <= 3; i++ {
			err := ExecuteActivity(ctx, testActivityHello, fmt.Sprintf("msg%d", i)).Get(ctx, nil)
			if err != nil {
				return err
			}
		}
		return nil
	} // END of workflow code

	env := s.StartWorkflowPart(workflowPart)

	var activityCalls []string
	env.SetOnActivityStartedListener(func(ctx context.Context, args EncodedValues, activityType string) {
		var input string
		s.NoError(args.Get(&input))
		activityCalls = append(activityCalls, fmt.Sprintf("%s:%s", activityType, input))
	})
	expectedCalls := []string{
		"github.com/uber-go/cadence-client/client/cadence.testActivityHello:msg1",
		"github.com/uber-go/cadence-client/client/cadence.testActivityHello:msg2",
		"github.com/uber-go/cadence-client/client/cadence.testActivityHello:msg3",
	}

	env.StartDispatcherLoop(time.Second)
	s.True(env.IsTestCompleted())
	s.NoError(env.GetTestError())
	s.Equal(expectedCalls, activityCalls)
}

func (s *WorkflowTestSuiteUnitTest) Test_TimerWorkflow_ClockAutoFastForward() {
	var firedTimerRecord []string
	workflowFn := func(ctx Context) error {
		t1 := NewTimer(ctx, time.Second*5)
		t2 := NewTimer(ctx, time.Second*1)
		t3 := NewTimer(ctx, time.Second*2)
		t4 := NewTimer(ctx, time.Second*5)

		selector := NewSelector(ctx)
		selector.AddFuture(t1, func(f Future) {
			firedTimerRecord = append(firedTimerRecord, "t1")
		}).AddFuture(t2, func(f Future) {
			firedTimerRecord = append(firedTimerRecord, "t2")
		}).AddFuture(t3, func(f Future) {
			firedTimerRecord = append(firedTimerRecord, "t3")
		}).AddFuture(t4, func(f Future) {
			firedTimerRecord = append(firedTimerRecord, "t4")
		})

		selector.Select(ctx)
		selector.Select(ctx)
		selector.Select(ctx)
		selector.Select(ctx)

		return nil
	}

	env := s.StartWorkflowPart(workflowFn)
	env.StartDispatcherLoop(time.Second)

	s.True(env.IsTestCompleted())
	s.NoError(env.GetTestError())
	s.Equal([]string{"t2", "t3", "t1", "t4"}, firedTimerRecord)
}

func (s *WorkflowTestSuiteUnitTest) Test_WorkflowPartWithControlledDecisionTask() {
	workflowPart := func(ctx Context) (string, error) {
		f1 := NewTimer(ctx, time.Second*2)
		ctx = WithActivityOptions(ctx, NewActivityOptions().
			WithTaskList(testTaskList).
			WithScheduleToCloseTimeout(time.Minute).
			WithScheduleToStartTimeout(time.Minute).
			WithStartToCloseTimeout(time.Minute).
			WithHeartbeatTimeout(time.Second*20))

		f2 := ExecuteActivity(ctx, testActivityHello, "controlled_execution")

		timerErr := f1.Get(ctx, nil) // wait until timer fires
		if timerErr != nil {
			return "", timerErr
		}

		if !f2.IsReady() {
			return "", errors.New("activity is not ready yet")
		}

		var activityResult string
		activityErr := f2.Get(ctx, &activityResult)
		if activityErr != nil {
			return "", activityErr
		}

		return activityResult, nil
	} // END of workflow code

	env := s.StartWorkflowPart(workflowPart)
	env.SetClock(s.clock)
	env.EnableAutoStartDecisionTask(false)              // manually control the execution
	env.EnableClockFastForwardWhenBlockedByTimer(false) // disable auto clock fast forward

	wg := &sync.WaitGroup{}
	wg.Add(2)
	env.SetOnActivityEndedListener(func(EncodedValue, error, string) {
		wg.Done()
	})
	env.SetOnTimerScheduledListener(func(timerID string, duration time.Duration) {
		s.clock.Add(duration)
		wg.Done()
	})

	go func() {
		wg.Wait()
		env.StartDecisionTask() // start a new decision task after the activity is done and the timer is fired.
	}()

	env.StartDispatcherLoop(time.Second)

	s.True(env.IsTestCompleted())
	s.NoError(env.GetTestError())
	s.NotNil(env.GetTestResult())
	var result string
	err := env.GetTestResult().Get(&result)
	s.NoError(err)
	s.Equal("hello_controlled_execution", result)
}

func (s *WorkflowTestSuiteUnitTest) Test_WorkflowActivityCancellation() {
	s.RegisterWorkflow(testWorkflowActivityCancellation)
	s.RegisterActivity(testActivityWithSleep, testTaskList)
	env := s.StartWorkflow(testWorkflowActivityCancellation)
	env.StartDispatcherLoop(time.Second)

	s.True(env.IsTestCompleted())
	s.NoError(env.GetTestError())
}

func (s *WorkflowTestSuiteUnitTest) Test_ActivityWithUserContext() {
	testKey, testValue := "test_key", "test_value"
	userCtx := context.WithValue(context.Background(), testKey, testValue)
	workerOptions := NewWorkerOptions().WithActivityContext(userCtx)

	// inline activity using value passing through user context.
	activityWithUserContext := func(ctx context.Context, keyName string) (string, error) {
		value := ctx.Value(keyName)
		if value != nil {
			return value.(string), nil
		}
		return "", errors.New("value not found from ctx")
	}

	env := s.NewTestWorkflowEnvironment()
	env.SetWorkerOption(workerOptions)
	blob, err := env.ExecuteActivity(activityWithUserContext, testKey)
	s.NoError(err)
	var value string
	blob.Get(&value)
	s.Equal(testValue, value)
}

func (s *WorkflowTestSuiteUnitTest) Test_CompleteActivity() {
	var activityInfo ActivityInfo
	fakeActivity := func(ctx context.Context, msg string) (string, error) {
		activityInfo = GetActivityInfo(ctx)
		return "", ErrActivityResultPending
	}

	env := s.StartWorkflow(testWorkflowHello)
	env.Override(testActivityHello, fakeActivity)
	env.StartDispatcherLoop(time.Millisecond)

	s.False(env.IsTestCompleted())
	s.NotNil(activityInfo)

	err := env.CompleteActivity(activityInfo.TaskToken, "async_complete", nil)
	s.NoError(err)
	env.StartDispatcherLoop(time.Millisecond * 10)

	s.True(env.IsTestCompleted())
	s.NoError(env.GetTestError())
	var result string
	err = env.GetTestResult().Get(&result)
	s.NoError(err)
	s.Equal("async_complete", result)
}

func testWorkflowHello(ctx Context) (string, error) {
	ctx = WithActivityOptions(ctx, NewActivityOptions().
		WithTaskList(testTaskList).
		WithScheduleToCloseTimeout(time.Minute).
		WithScheduleToStartTimeout(time.Minute).
		WithStartToCloseTimeout(time.Minute).
		WithHeartbeatTimeout(time.Second*20))

	var result string
	err := ExecuteActivity(ctx, testActivityHello, "world").Get(ctx, &result)
	if err != nil {
		return "", err
	}
	return result, nil
}

func testActivityHello(ctx context.Context, msg string) (string, error) {
	return "hello" + "_" + msg, nil
}

func testWorkflowActivityCancellation(ctx Context) (string, error) {
	ctx = WithActivityOptions(ctx, NewActivityOptions().
		WithTaskList(testTaskList).
		WithScheduleToCloseTimeout(time.Minute).
		WithScheduleToStartTimeout(time.Minute).
		WithStartToCloseTimeout(time.Minute).
		WithHeartbeatTimeout(time.Second*20))

	ctx, cancelHandler := WithCancel(ctx)
	f1 := ExecuteActivity(ctx, testActivityWithSleep, "msg1", 0)
	f2 := ExecuteActivity(ctx, testActivityWithSleep, "msg2", 5000)

	selector := NewSelector(ctx)
	selector.AddFuture(f1, func(f Future) {
		cancelHandler()
	}).AddFuture(f2, func(f Future) {
		cancelHandler()
	})

	selector.Select(ctx)

	GetLogger(ctx).Info("testWorkflowActivityCancellation completed.")
	return "result from testWorkflowActivityCancellation", nil
}

func testActivityWithSleep(ctx context.Context, msg string, sleepMs int) error {
	GetActivityLogger(ctx).Sugar().Infof("testActivityWithSleep %s, sleep %dms", msg, sleepMs)

	time.Sleep(time.Millisecond * time.Duration(sleepMs))

	RecordActivityHeartbeat(ctx)

	select {
	case <-ctx.Done():
		// We have been cancelled.
		return ctx.Err()
	default:
		// We are not cancelled yet.
	}

	return nil
}
