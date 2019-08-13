// +build integration

package yod

import (
	"bytes"
	"io"
	"log"
	"net"
	"sync"
	"syscall"
	"testing"
	"time"
)

var wg = sync.WaitGroup{}

type mocker struct {
	data []byte
}

func (m *mocker) handler(c Context) {
	b, err := c.Read()
	if err != nil {
		log.Fatal(err)
	}
	m.data = b
	wg.Done()
}


func TestFramework(t *testing.T) {
	wg.Add(1)
	y := New()
	h := mocker{data: []byte{}}
	y.Add("10000", h.handler)
	go y.Start()

	time.Sleep(50 * time.Millisecond)
	conn, err := net.Dial("tcp", ":10000")
	if err != nil {
		t.Fatal(err)
	}

	want := "test data"
	io.Copy(conn, bytes.NewBufferString(want))
	conn.Close()

	wg.Wait()
	if want != string(h.data) {
		t.Errorf("we sent %q to the server but got %q\n", want, h.data)
	}
	y.interrupt <- syscall.SIGINT
}

