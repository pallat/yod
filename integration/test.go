package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"net"
)

var (
	max  = flag.Int("max", 2, "100")
	port = flag.String("port", "10000", "10000")
)

func main() {
	flag.Parse()

	for i := 0; i < *max; i++ {
		conn, err := net.Dial("tcp", ":"+*port)
		if err != nil {
			log.Fatal(err)
		}

		if err != nil {
			log.Fatal(err)
		}

		want := "test data"
		io.Copy(conn, bytes.NewBufferString(want))
		conn.Close()
	}
}
