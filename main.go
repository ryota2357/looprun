package main

import (
	"fmt"

	"github.com/ryota2357/looprun/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Println(err)
	}
}
