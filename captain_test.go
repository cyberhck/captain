package captain_test

import (
	"testing"
	"github.com/cyberhck/captain"
	"time"
	Assert "github.com/stretchr/testify/assert"
	"errors"
	"github.com/stretchr/testify/mock"
)

type LockProviderMock struct {
	mock.Mock
}

func (m *LockProviderMock) Acquire() error {
	m.Called()
	return nil
}

func (m *LockProviderMock) Release() error {
	args := m.Called(nil)
	return args.Error(1)
}

func TestNew(t *testing.T) {
	assert := Assert.New(t)

	testJob := captain.CreateJob()

	assert.Equal(testJob.RuntimeProcessingFrequency, 200*time.Millisecond)
	assert.Nil(testJob.ResultProcessor)
	assert.Nil(testJob.RuntimeProcessor)
	assert.Nil(testJob.LockProvider)
}

type LockProvider struct {
	acquire error
	release error
}

func (r LockProvider) Acquire() error { return r.acquire }
func (r LockProvider) Release() error { return r.release }

func getMockedLockProvider(acquire error, release error) captain.LockProvider {
	return LockProvider{
		acquire: acquire,
		release: release,
	}
}
func TestWithLockProvider(t *testing.T) {
	testJob := captain.CreateJob()
	testJob.WithLockProvider(getMockedLockProvider(nil, nil))
	Assert.NotNil(t, testJob.LockProvider)
}

func TestWithResultProcessor(t *testing.T) {
	testJob := captain.CreateJob()
	testJob.WithResultProcessor(func() {})
	Assert.NotNil(t, testJob.ResultProcessor)
}

func TestWithRuntimeProcessingFrequency(t *testing.T) {
	testJob := captain.CreateJob()
	testJob.WithRuntimeProcessingFrequency(1 * time.Second)
	Assert.Equal(t, testJob.RuntimeProcessingFrequency, 1*time.Second)
}

func TestWithRuntimeProcessor(t *testing.T) {
	testJob := captain.CreateJob()
	testJob.WithRuntimeProcessor(func(tick time.Time, message string, startTime time.Time) {})
	Assert.NotNil(t, testJob.RuntimeProcessor)
}

func TestRunPanicsIfLockNotAcquired(t *testing.T) {
	assert := Assert.New(t)
	testJob := captain.CreateJob()
	testJob.WithLockProvider(getMockedLockProvider(errors.New("couldn't acquire lock"), nil))
	testJob.SetWorker(func(channel chan string) {})
	assert.Panics(testJob.Run)
}

func TestRunDoesNotPanicIfLockAcquired(t *testing.T) {
	assert := Assert.New(t)
	testJob := captain.CreateJob()
	testJob.WithLockProvider(getMockedLockProvider(nil, nil))
	testJob.SetWorker(func(channel chan string) {})
	assert.NotPanics(testJob.Run)
}

func TestRunPanicsIfNoWorkerIsDefined(t *testing.T) {
	testJob := captain.CreateJob()
	mocked := new(LockProviderMock)
	mocked.On("Acquire").Return(nil)
	testJob.WithLockProvider(mocked)
	Assert.Panics(t, testJob.Run)
}

func TestRunCallsWorker(t *testing.T) {
	testJob := captain.CreateJob()
	mocked := new(LockProviderMock)
	mocked.On("Acquire").Return(nil)
	testJob.WithLockProvider(mocked)
	testJob.SetWorker(func(channel chan string) {})
	testJob.Run()
	mocked.AssertExpectations(t)
}

func TestRunWorksWithoutLockProvider(t *testing.T) {
	testJob := captain.CreateJob()
	testJob.SetWorker(func(channel chan string) {})
	Assert.NotPanics(t, testJob.Run)
}

func TestDoesNotPanicWhenNoRuntimeProcessorPresent(t *testing.T) {
	testJob := captain.CreateJob()
	testJob.SetWorker(func(channel chan string) {})
	Assert.NotPanics(t, testJob.Run)
}

func TestLongRunningProcessorWorksWithoutRuntimeProcessor(t *testing.T) {
	testJob := captain.CreateJob()
	testJob.SetWorker(func(channel chan string) {
		time.Sleep(10 * time.Millisecond)
		channel <- "Done..."
	})
	Assert.NotPanics(t, testJob.Run)
}
