package main

import (
	"fmt"

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

	for i := 0; i < 3; i++ {
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

}
