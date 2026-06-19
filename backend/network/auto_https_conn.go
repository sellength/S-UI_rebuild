package network

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
)

type AutoHttpsConn struct {
	net.Conn

	firstBuf     []byte
	bufStart     int
	isRedirected bool

	readRequestOnce sync.Once
}

func NewAutoHttpsConn(conn net.Conn) net.Conn {
	return &AutoHttpsConn{
		Conn: conn,
	}
}

func (c *AutoHttpsConn) readRequest() {
	c.firstBuf = make([]byte, 2048)
	n, err := c.Conn.Read(c.firstBuf)
	if err != nil {
		c.firstBuf = c.firstBuf[:0]
		return
	}
	c.firstBuf = c.firstBuf[:n]

	reader := bytes.NewReader(c.firstBuf)
	bufReader := bufio.NewReader(reader)
	request, err := http.ReadRequest(bufReader)
	if err != nil {
		// Non-HTTP request (likely TLS Client Hello), keep firstBuf for Read transparent transmission
		return
	}

	resp := http.Response{
		StatusCode: http.StatusTemporaryRedirect,
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{},
	}
	resp.Header.Set("Location", fmt.Sprintf("https://%v%v", request.Host, request.RequestURI))
	resp.Header.Set("Connection", "close")
	resp.ContentLength = 0
	_ = resp.Write(c.Conn)

	c.isRedirected = true
	c.firstBuf = nil
}

func (c *AutoHttpsConn) Read(buf []byte) (int, error) {
	c.readRequestOnce.Do(func() {
		c.readRequest()
	})

	if c.isRedirected {
		return 0, io.EOF
	}

	if c.firstBuf != nil {
		n := copy(buf, c.firstBuf[c.bufStart:])
		c.bufStart += n
		if c.bufStart >= len(c.firstBuf) {
			c.firstBuf = nil
		}
		return n, nil
	}

	return c.Conn.Read(buf)
}

