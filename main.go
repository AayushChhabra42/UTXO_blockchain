package main

import (
	"os"

	"github.com/AayushChhabra42/Golang-Blockchain/cli"
)

func main() {
	defer os.Exit(0)
	cli := cli.CommandLine{}
	cli.Run()

}
