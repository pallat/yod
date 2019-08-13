package yod

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/net/html/charset"
)

type Yod struct {
	middlewares []MiddlewareFunc
	handlers    map[string]Handler
	interrupt   chan os.Signal
}

func New() *Yod {
	return &Yod{
		middlewares: make([]MiddlewareFunc, 0, 0),
		handlers:    make(map[string]Handler),
		interrupt:   make(chan os.Signal),
	}
}

type Handler func(Context)

type MiddlewareFunc func(Handler) Handler

func (y *Yod) Start() {
	signal.Notify(y.interrupt, syscall.SIGINT, syscall.SIGKILL)

	for port, handler := range y.handlers {
		for _, middleware := range y.middlewares {
			handler = middleware(y.handlers[port])
		}
		go y.listen(port, handler)
		fmt.Printf("listen to :%s\n", port)
	}

	<-y.interrupt
}

func (y *Yod) Use(middleware MiddlewareFunc) {
	y.middlewares = append(y.middlewares, middleware)
}

func (y *Yod) Add(port string, h Handler) {
	y.handlers[port] = h
}

func (y *Yod) listen(port string, h Handler) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("connect listener", r)
		}
	}()

	for {
		l, err := net.Listen("tcp", ":"+port)
		if err != nil {
			log.Fatal(err)
		}

		for {
			conn, err := l.Accept()
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					return
				}
				log.Println(err)
				continue
			}

			ctx := &context{
				conn: conn,
			}

			go middleware(h)(ctx)
		}
	}
}

type Context interface {
	Read() ([]byte, error)
	ReadCharset(label string) ([]byte, error)
	Write([]byte) (int64, error)
}

type context struct {
	conn net.Conn
	Context
}

func (c *context) Read() ([]byte, error) {
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, c.conn)
	if err != nil {
		return nil, fmt.Errorf("error read data from conn %s", err)
	}
	return buf.Bytes(), nil
}

func (c *context) ReadCharset(label string) ([]byte, error) {
	reader, err := charset.NewReaderLabel(label, c.conn)
	if err != nil {
		return nil, fmt.Errorf("error read data from conn %s", err)
	}

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, reader)
	if err != nil {
		return nil, fmt.Errorf("error read data from conn %s", err)
	}
	return buf.Bytes(), nil
}

func (c *context) Write(b []byte) (int64, error) {
	reader := bytes.NewBuffer(b)
	return io.Copy(c.conn, reader)
}

func (c *context) close() {
	one := []byte{}
	c.conn.SetReadDeadline(time.Now())
	if _, err := c.conn.Read(one); err == io.EOF {
		c.conn.Close()
		c.conn = nil
	}
}

func middleware(next Handler) Handler {
	return func(c Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Println(r)
			}
		}()

		defer c.(*context).close()
		next(c)
	}
}
