package main

import (
	"log"

	"tenant-crud/cmd/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		log.Fatal(err)
	}
}
