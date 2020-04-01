package executor

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"sync"

	"go.uber.org/ratelimit"
)

// Executor is a simple thread pool base on goroutine.
type Executor struct {
	RateLimit  ratelimit.Limiter
	WaitGroup  *sync.WaitGroup
	Channel    chan *Job
	NumWorkers int
}

// Job is a task will be executor execute.
type Job struct {
	Handler interface{}
	Args    []reflect.Value
}

// Config is a config of executor.
// ReqPerSeconds is request per seconds. If it is 0, no limit for requests.
// QueueSize is size of buffer. Executor use synchronize channel, publisher will waiting if channel is full.
// NumWorkers is a number of goroutine.
type Config struct {
	ReqPerSeconds int
	QueueSize     int
	NumWorkers    int
}

// DefaultConfig is a default config
func DefaultConfig() Config {
	return Config{
		ReqPerSeconds: 0,
		QueueSize:     2 * runtime.NumCPU(),
		NumWorkers:    runtime.NumCPU(),
	}
}

// New returns a Executors that will manage workers.
func New(config Config) (*Executor, error) {
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

// NewJob returns a Job that will be executed in workers.
func NewJob(handler interface{}, inputArgs ...interface{}) (*Job, error) {
	var (
		err  error
		args []reflect.Value
	)

	nArgs := len(inputArgs)
	parsedHandler, err := validateFunc(handler, nArgs)

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

// Publish to publish a handler and arguments
// Workers will run handler with provided arguments.
func (pipeline *Executor) Publish(handler interface{}, inputArgs ...interface{}) error {
	job, err := NewJob(handler, inputArgs...)

	if err != nil {
		return err
	}

	pipeline.PublishJob(job)
	return nil
}

// PublishJob publish a provided job.
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

// Wait for all worker done.
func (pipeline *Executor) Wait() {
	pipeline.WaitGroup.Wait()
}

// Close channel and wait all worker done.
func (pipeline *Executor) Close() {
	pipeline.WaitGroup.Wait()
	close(pipeline.Channel)
}

// validateFunc validate type of handler and number of arguments.
func validateFunc(handler interface{}, nArgs int) (interface{}, error) {
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

func (p *Config) validate() error {
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
