# Executor

Executor is a simple thread pool implemented for Golang.

Features:

- Execute concurrent job by Goroutine
- Rate limiter
- Dynamic handler by reflect

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
    logrus.Infof("Idx=%d", input)
  }, i)
}

// INFO[0000] Idx=0                                        
// INFO[0000] Idx=1                                        
// INFO[0000] Idx=5                                        
// INFO[0000] Idx=6                                        
// INFO[0000] Idx=7                                        
// INFO[0000] Idx=8                                        
// INFO[0000] Idx=9                                        
// INFO[0000] Idx=3                                        
// INFO[0000] Idx=2                                        
// INFO[0000] Idx=4
```