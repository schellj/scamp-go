package main

import "github.com/gudtech/scamp-go/scamp"

func main() {
	scamp.Initialize("/etc/SCAMP/soa.conf")
	scamp.NewConfig()
}