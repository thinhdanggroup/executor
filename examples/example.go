package main

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/thinhdanggroup/executor"
)

func main() {
	executor, err := executor.New(executor.DefaultConfig())

	if err != nil {
		logrus.Error(err)
	}

	// close resource before quit
	defer executor.Close()

	for i := 0; i < 3; i++ {
		executor.Publish(mul, i)
		executor.Publish(pow, i)
		executor.Publish(sum, i, i+1)
	}

}

func mul(input int) {
	fmt.Printf("2 * %d = %d \n", input, 2*input)
}

func pow(input int) {
	fmt.Printf("2 ^ %d = %d \n", input, input^2)
}

func sum(a int, b int) {
	fmt.Printf("%d + %d = %d \n", a, b, a+b)
}
