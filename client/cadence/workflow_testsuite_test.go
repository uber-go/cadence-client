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
	"go.uber.org/zap"
)

type WorkflowTestSuiteUnitTest struct {
	suite.Suite
	logger *zap.Logger
	clock  *clock.Mock
}

func (s *WorkflowTestSuiteUnitTest) SetupSuite() {
	logger, _ := zap.NewDevelopment()
	s.logger = logger
	s.clock = clock.NewMock()
}

func (s *WorkflowTestSuiteUnitTest) SetupTest() {
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowTestSuiteUnitTest))
}

func (s *WorkflowTestSuiteUnitTest) Test_ActivityOverride() {
	t := NewWorkflowTestSuite().SetLogger(s.logger).SetClock(s.clock)
	t.RegisterWorkflow(testWorkflowHello)
	t.RegisterActivity(testActivityHello, DefaultTestTaskList)

	fakeActivity := func(ctx context.Context, msg string) (string, error) {
		return "fake_" + msg, nil
	}
	t.Override(testActivityHello, fakeActivity)

	env := t.TestWorkflow(testWorkflowHello)
	env.StartDispatcherLoop(time.Second * 2)

	s.True(env.IsTestCompleted())
	s.NoError(env.GetTestError())
	var result string
	err := env.GetTestResult().GetResult(&result)
	s.NoError(err)
	s.Equal("fake_world", result)
}

func (s *WorkflowTestSuiteUnitTest) Test_OnActivityStartedListener() {
	testTaskList := "test-on-activity-started-listener"
	workflowPart := func(ctx Context) error {
		ctx = WithActivityOptions(ctx, NewActivityOptions().
			WithTaskList(testTaskList).
			WithScheduleToCloseTimeout(time.Minute).
			WithScheduleToStartTimeout(time.Minute).
			WithStartToCloseTimeout(time.Minute).
			WithHeartbeatTimeout(time.Second*20))

		err := ExecuteActivity(ctx, testActivityHello, "msg1").Get(ctx, nil)
		if err != nil {
			return err
		}
		err = ExecuteActivity(ctx, testActivityHello, "msg2").Get(ctx, nil)
		if err != nil {
			return err
		}
		err = ExecuteActivity(ctx, testActivityHello, "msg3").Get(ctx, nil)
		if err != nil {
			return err
		}
		return nil
	} // END of workflow code

	t := NewWorkflowTestSuite().SetLogger(s.logger).SetClock(s.clock)
	t.RegisterActivity(testActivityHello, testTaskList)
	env := t.TestWorkflowPart(workflowPart)

	var activityCalls []string
	env.SetOnActivityTaskStartedListener(func(ctx context.Context, args BlobArgs, activityType string) {
		var input string
		s.NoError(args.GetArgs(&input))
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

func (s *WorkflowTestSuiteUnitTest) Test_WorkflowWithTimer() {
	t := NewWorkflowTestSuite().SetLogger(s.logger).SetClock(s.clock)
	t.RegisterWorkflow(testWorkflowWithTimer)

	env := t.TestWorkflow(testWorkflowWithTimer, time.Second*2)
	env.SetOnTimerScheduledListener(func(timerID string, duration time.Duration) {
		s.clock.Add(duration)
	})
	env.StartDispatcherLoop(time.Second * 2)

	s.True(env.IsTestCompleted())
	s.NoError(env.GetTestError())
}

func (s *WorkflowTestSuiteUnitTest) Test_WorkflowPartWithControlledDecisionTask() {
	// lambda method with cadence.Context as its first parameter, tested as partial workflow
	workflowPart := func(ctx Context) (string, error) {
		f1 := NewTimer(ctx, time.Second*2)
		ctx = WithActivityOptions(ctx, NewActivityOptions().
			WithTaskList(DefaultTestTaskList).
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

	t := NewWorkflowTestSuite().SetLogger(s.logger).SetClock(s.clock)
	t.RegisterActivity(testActivityHello, DefaultTestTaskList)
	env := t.TestWorkflowPart(workflowPart)
	env.EnableAutoStartDecisionTask(false) // manually control the execution

	wg := &sync.WaitGroup{}
	wg.Add(2)
	env.SetOnActivityTaskEndedListener(func(BlobResult, error, string) {
		wg.Done()
	})
	env.SetOnTimerScheduledListener(func(timerID string, duration time.Duration) {
		s.clock.Add(duration)
		wg.Done()
	})

	go func() {
		wg.Wait()
		env.StartDecisionTask() // start a new decision task after the activity is done the timer is fired.
	}()

	env.StartDispatcherLoop(time.Second * 2)

	s.True(env.IsTestCompleted())
	s.NoError(env.GetTestError())
	s.NotNil(env.GetTestResult())
	var result string
	err := env.GetTestResult().GetResult(&result)
	s.NoError(err)
	s.Equal("hello_controlled_execution", result)
}

func (s *WorkflowTestSuiteUnitTest) Test_WorkflowActivityCancellation() {
	t := NewWorkflowTestSuite().SetLogger(s.logger).SetClock(s.clock)
	t.RegisterWorkflow(testWorkflowActivityCancellation)
	t.RegisterActivity(testActivityWithSleep, DefaultTestTaskList)

	env := t.TestWorkflow(testWorkflowActivityCancellation)
	env.StartDispatcherLoop(time.Second * 2)

	s.True(env.IsTestCompleted())
	s.NoError(env.GetTestError())
}

func (s *WorkflowTestSuiteUnitTest) Test_ActivityWithUserContext() {
	testKey, testValue := "test_key", "test_value"
	userCtx := context.WithValue(context.Background(), testKey, testValue)
	workerOptions := NewWorkerOptions().WithActivityContext(userCtx)
	t := NewWorkflowTestSuite().SetLogger(s.logger).SetWorkerOption(workerOptions)
	env := t.NewTestWorkflowEnvironment()

	// inline activity using value passing through user context.
	activityWithUserContext := func(ctx context.Context, keyName string) (string, error) {
		value := ctx.Value(keyName)
		if value != nil {
			return value.(string), nil
		}
		return "", errors.New("value not found from ctx")
	}

	blob, err := env.TestActivity(activityWithUserContext, testKey)
	s.NoError(err)
	var value string
	blob.GetResult(&value)
	s.Equal(testValue, value)
}

func testWorkflowHello(ctx Context) (string, error) {
	ctx = WithActivityOptions(ctx, NewActivityOptions().
		WithTaskList(DefaultTestTaskList).
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

func testWorkflowWithTimer(ctx Context, timerDuration time.Duration) error {
	GetLogger(ctx).Info("testTimerWorkflow start", zap.Duration("TimerDuration", timerDuration))
	timerFuture := NewTimer(ctx, timerDuration)

	selector := NewSelector(ctx)
	selector.AddFuture(timerFuture, func(f Future) {
		GetLogger(ctx).Info("timer fired")
	})
	selector.Select(ctx)
	return nil
}

func testWorkflowActivityCancellation(ctx Context) (string, error) {
	ctx = WithActivityOptions(ctx, NewActivityOptions().
		WithTaskList(DefaultTestTaskList).
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
