package main

import "tenant-crud/cmd/bootstrap"

func main() {
	boot, _ := bootstrap.New()
	boot.Start()
}
