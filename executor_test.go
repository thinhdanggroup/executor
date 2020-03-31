package executor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidateConfig(t *testing.T) {
	assert := assert.New(t)

	executor, err := NewExecutor(ExecutorConfig{})

	assert.NotNil(err)
	assert.Nil(executor)
}

func TestPublishJobSuccess(t *testing.T) {
	var (
		value int
		as    *assert.Assertions
		err   error
	)

	as = assert.New(t)

	executor, err := NewExecutor(DefaultExecutorConfig())

	as.Nil(err)

	value = 1

	err = executor.Publish(func(input int) {
		as.Equal(value, input)
	}, value)

	as.Nil(err)
	executor.Wait()
}

func TestPublishJobFail(t *testing.T) {
	var (
		as  *assert.Assertions
		err error
	)

	as = assert.New(t)

	executor, err := NewExecutor(DefaultExecutorConfig())

	as.Nil(err)

	err = executor.Publish(func(input int) {
		as.Equal(1, input)
	})

	as.NotNil(err)
	as.Equal(err.Error(), "Call with too few input arguments")
	executor.Wait()

	err = executor.Publish(func(input int) {
		as.Equal(1, input)
	}, 1, 1)

	as.NotNil(err)
	as.Equal(err.Error(), "Call with too many input arguments")
	executor.Wait()

}

func TestRateLimiter(t *testing.T) {
	var (
		as  *assert.Assertions
		err error
	)

	as = assert.New(t)

	executor, err := NewExecutor(ExecutorConfig{
		ReqPerSeconds: 2,
		NumWorkers:    2,
		QueueSize:     10,
	})
	as.Nil(err)

	startTime := time.Now().Unix()
	for i := 0; i < 8; i++ {
		err = executor.Publish(func(input int) {
			as.Equal(1, input)
		}, 1)
		as.Nil(err)
	}
	executor.Close()
	endTime := time.Now().Unix()
	as.InDelta(endTime-startTime, 4, 1)

}
