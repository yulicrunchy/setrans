// +build selinux,linux

package setrans

import (
	"context"
	"fmt"
	"net"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/jbrindle/vectorio"
)

var sockpath = "/var/run/setrans/.setrans-unix"

const timeout = time.Second * 3

const nullChar = "\000"

func sendRequest(conn net.Conn, t requestType, data string) (string, error) {
	c, ok := conn.(*net.UnixConn)
	if !ok {
		return "", fmt.Errorf("%T is not a unix connection", conn)
	}

	// mcstransd expects null terminated strings
	data += nullChar
	dataSize := uint32(len(data))

	data2Size := uint32(len(nullChar)) // unused by libselinux users

	d1 := []byte(data)
	d2 := []byte(nullChar)

	v := []syscall.Iovec{
		{Base: (*byte)(unsafe.Pointer(&t)), Len: uint64(unsafe.Sizeof(t))},
		{Base: (*byte)(unsafe.Pointer(&dataSize)), Len: uint64(unsafe.Sizeof(dataSize))},
		{Base: (*byte)(unsafe.Pointer(&data2Size)), Len: uint64(unsafe.Sizeof(data2Size))},
		{Base: (*byte)(unsafe.Pointer(&d1[0])), Len: uint64(dataSize)},
		{Base: (*byte)(unsafe.Pointer(&d2[0])), Len: uint64(data2Size)},
	}

	f, err := c.File()
	if err != nil {
		return "", fmt.Errorf("failed to create file from UDS: %w", err)
	}
	defer f.Close()

	if _, err := vectorio.WritevRaw(f.Fd(), v); err != nil {
		return "", fmt.Errorf("failed to write to mcstransd: %w", err)
	}

	var uintsize uint32
	elemsize := uint64(unsafe.Sizeof(uintsize))

	header := []syscall.Iovec{
		{Len: elemsize}, // function
		{Len: elemsize}, // response length
		{Len: elemsize}, // return value
	}

	if _, err := vectorio.ReadvRaw(f.Fd(), header); err != nil {
		return "", fmt.Errorf("failed to read from mcstransd: %w", err)
	}

	respvec := []syscall.Iovec{{Len: uint64(*header[1].Base)}}
	_, err = vectorio.ReadvRaw(f.Fd(), respvec)
	if err != nil {
		return "", fmt.Errorf("failed to read from response: %w", err)
	}

	b := *(*[]byte)(unsafe.Pointer(&respvec[0].Base))
	if len(b) <= len(nullChar) {
		return "", fmt.Errorf("failed to read from response")
	}

	str := strings.Trim(string(b), nullChar)
	return strings.TrimSpace(str), nil
}

func (c *Conn) makeRequest(con string, t requestType) (string, error) {
	go func(msg setransMsg) {
		response, err := sendRequest(c.conn, msg.reqType, msg.label)
		if err != nil {
			c.errch <- fmt.Errorf("failed to send initial request %v: %w", msg, err)
		}
		c.mcstransch <- setransMsg{label: response}
	}(setransMsg{reqType: t, label: con})

	select {
	case err := <-c.errch:
		return "", err
	case req := <-c.mcstransch:
		if con == req.label {
			return "", ErrInvalidLevel
		}

		return req.label, nil
	}
}

func new() (*Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var d net.Dialer
	conn, err := d.DialContext(ctx, "unix", sockpath)
	if err != nil {
		return nil, fmt.Errorf("failed to dial mcstransd: %w", err)
	}

	return &Conn{
		conn:       conn,
		mcstransch: make(chan setransMsg),
		errch:      make(chan error),
	}, nil
}

func (c *Conn) close() error {
	return c.conn.Close()
}

func (c *Conn) transToRaw(trans string) (string, error) {
	return c.makeRequest(trans, reqTransToRaw)
}

func (c *Conn) rawToTrans(raw string) (string, error) {
	return c.makeRequest(raw, reqRawToTrans)
}

func (c *Conn) rawToColor(raw string) (string, error) {
	return c.makeRequest(raw, reqRawToColor)
}
