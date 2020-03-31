package executor

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"unsafe"

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

func DefaultExecutorConfig() ExecutorConfig {
	return ExecutorConfig{
		ReqPerSeconds: 0,
		QueueSize:     2 * runtime.NumCPU(),
		NumWorkers:    runtime.NumCPU(),
	}
}

func NewExecutor(config ExecutorConfig) (*Executor, error) {
	err := config.validate()

	if err != nil {
		return nil, err
	}

	var rl ratelimit.Limiter
	if config.ReqPerSeconds > 0 {
		rl = ratelimit.New(config.ReqPerSeconds)
	}

	pipeline := &Executor{
		RateLimit:  rl,
		WaitGroup:  &sync.WaitGroup{},
		Channel:    make(chan *Job, config.QueueSize),
		NumWorkers: config.NumWorkers,
	}

	pipeline.initWorker()
	return pipeline, nil
}

type rtype struct {
}
type funcType struct {
	inCount  uint16
	outCount uint16 // top bit is set if last input parameter is ...
}

func (t *rtype) NumIn() int {
	tt := (*funcType)(unsafe.Pointer(t))
	return int(tt.inCount)
}

func NewJob(handler interface{}, inputArgs ...interface{}) (*Job, error) {
	var (
		err  error
		args []reflect.Value
	)

	nArgs := len(inputArgs)
	parsedHandler, err := getFunc(handler, nArgs)

	if err != nil {
		return nil, err
	}

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

func (pipeline *Executor) initWorker() {
	for i := 0; i < pipeline.NumWorkers; i++ {
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

func getFunc(handler interface{}, nArgs int) (interface{}, error) {
	f := reflect.Indirect(reflect.ValueOf(handler))

	if f.Kind() != reflect.Func {
		return f, fmt.Errorf("%T must be a Function ", f)
	}

	method := reflect.ValueOf(handler)
	methodType := method.Type()
	numIn := methodType.NumIn()

	if nArgs < numIn {
		return nil, errors.New("Call with too few input arguments")
	} else if nArgs > numIn {
		return nil, errors.New("Call with too many input arguments")
	}
	return f, nil
}

func (p *ExecutorConfig) validate() error {
	if p.ReqPerSeconds < 0 {
		return fmt.Errorf("%T must non negative", "ReqPerSeconds")
	}
	if p.QueueSize <= 0 {
		return fmt.Errorf("%T must positive", "QueueSize")
	}

	if p.NumWorkers < 0 {
		return fmt.Errorf("%T must positive", "NumWorkers")
	}
	return nil
}
