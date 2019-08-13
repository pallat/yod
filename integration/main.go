package main

import (
	"flag"
	"log"

	"github.com/pallat/yod"
)

var (
	port = flag.String("port", "10000", "10000")
)

func main() {
	y := yod.New()
	y.Add(*port, handler)
	y.Start()
}

func handler(c yod.Context) {
	b, err := c.Read()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(b))
}
