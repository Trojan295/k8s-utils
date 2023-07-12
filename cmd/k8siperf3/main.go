package main

import (
	"fmt"
	"os"

	"github.com/trojan295/k8s-utils/cmd/k8siperf3/command"
)

func main() {
	if err := command.NewRoot().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
