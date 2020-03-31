package main

import (
	"github.com/sirupsen/logrus"
	"github.com/thinhdanggroup/executor"
)

func main() {
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

}
