package executor

import (
	"fmt"
	"reflect"
	"sync"

	"go.uber.org/ratelimit"
)

type Executor struct {
	RateLimit  ratelimit.Limiter
	WaitGroup  *sync.WaitGroup
	Channel    chan *Job
	NumWorkers int
}

type Job struct {
	Handler interface{}
	Args    []reflect.Value
}

type ExecutorConfig struct {
	ReqPerSeconds int
	QueueSize     int
	NumWorkers    int
}

func NewExecutor(limit int, queueSize int, numWorker int) *Executor {
	pipeline := &Executor{
		RateLimit:  ratelimit.New(limit),
		WaitGroup:  &sync.WaitGroup{},
		Channel:    make(chan *Job, queueSize),
		NumWorkers: numWorker,
	}

	pipeline.initWorker(numWorker)
	return pipeline
}

func NewJob(handler interface{}, inputArgs ...interface{}) (*Job, error) {
	var (
		err  error
		args []reflect.Value
	)

	parsedHandler, err := getFunc(handler)

	if err != nil {
		return nil, err
	}

	nArgs := len(inputArgs)
	for i := 0; i < nArgs; i++ {
		args = append(args, reflect.ValueOf(inputArgs[i]))
	}

	return &Job{
		Handler: parsedHandler,
		Args:    args,
	}, nil
}

func (pipeline *Executor) Publish(handler interface{}, inputArgs ...interface{}) error {
	job, err := NewJob(handler, inputArgs...)

	if err != nil {
		return err
	}

	pipeline.PublishJob(job)
	return nil
}

func (pipeline *Executor) PublishJob(job *Job) {
	if pipeline.RateLimit != nil {
		pipeline.RateLimit.Take()
	}

	pipeline.WaitGroup.Add(1)
	pipeline.Channel <- job
}

func (pipeline *Executor) initWorker(numWorker int) {
	for i := 0; i < numWorker; i++ {
		go pipeline.runWorker()
	}
}

func (pipeline *Executor) runWorker() {
	for {
		job, ok := <-pipeline.Channel

		if !ok {
			break
		}

		fn := job.Handler.(reflect.Value)
		_ = fn.Call(job.Args)

		pipeline.WaitGroup.Done()
	}
}

func (pipeline *Executor) Wait() {
	pipeline.WaitGroup.Wait()
}

func (pipeline *Executor) Close() {
	pipeline.WaitGroup.Wait()
	close(pipeline.Channel)
}

func getFunc(handler interface{}) (interface{}, error) {
	f := reflect.Indirect(reflect.ValueOf(handler))

	if f.Kind() != reflect.Func {
		return f, fmt.Errorf("%T must be a Function ", f)
	}

	return f, nil
}
