# Executor

![version](https://img.shields.io/badge/version-0.1.0-red) [![contributors](https://img.shields.io/badge/contributors-1-blue)]() [![Build Status](https://travis-ci.org/thinhdanggroup/executor.svg?branch=master)](https://travis-ci.org/thinhdanggroup/executor) [![Coverage Status](https://coveralls.io/repos/github/thinhdanggroup/executor/badge.svg?branch=master)](https://coveralls.io/github/thinhdanggroup/executor?branch=master)

Executor is a simple worker pool implemented for Golang.

Features:

- Execute concurrent job by Goroutine
- Rate limiter
- Dynamic handler by reflect

More information at [Blog](https://medium.com/@thinhda/executor-worker-pool-in-go-86ef94ffb141).

### Install

```bash
$ go get github.com/thinhdanggroup/executor
```

### Usage

```go
executor, err := executor.NewExecutor(executor.DefaultExecutorConfig())

if err != nil {
  logrus.Error(err)
}

// close resource before quit
defer executor.Close()

for i := 0; i < 10; i++ {
  executor.Publish(func(input int) {
    fmt.Printf("2 * %d = %d \n", input, 2*input)
  }, i)

  executor.Publish(func(input int) {
    fmt.Printf("2 ^ %d = %d \n", input, input^2)
  }, i)

  executor.Publish(func(a int, b int) {
    fmt.Printf("%d + %d = %d \n", a, b, a+b)
  }, i, i+1)
}

// Output:
// 2 * 0 = 0 
// 2 ^ 0 = 2 
// 2 ^ 1 = 3 
// 1 + 2 = 3 
// 2 * 2 = 4 
// 2 ^ 2 = 0 
// 2 + 3 = 5 
// 0 + 1 = 1 
// 2 * 1 = 2
```
