package main

import (
	"os"

	"github.com/vladlosev/node-relabeler/pkg/cmd"
)

func main() {
	cmd := cmd.NewWorkerCommand()
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
