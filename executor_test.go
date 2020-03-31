package executor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateConfig(t *testing.T) {
	assert := assert.New(t)

	executor, err := NewExecutor(ExecutorConfig{})

	assert.NotNil(err)
	assert.Nil(executor)
}

func TestRunExecutor(t *testing.T) {
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
