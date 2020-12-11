// +build selinux,linux

package setrans

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"time"
	"unsafe"
)

var sockpath = "/var/run/setrans/.setrans-unix"

const timeout = time.Second * 3

const nullChar = "\000"

/* This is a cursed function and I hate it.
 * The Go devlopers have decided that machine byteorder is an unimportant detail
 * and 1) force you to specify order when decoding byte arrays, and
 * 2) don't give you a simple way of figuring out the native endianness
 * so here we go. Create a pointer and then look at the byte order to set
 * the interface for reading off of the socket. */
var nativeEndian binary.ByteOrder

func setNativeEndian() {
	buf := [2]byte{}
	*(*uint16)(unsafe.Pointer(&buf[0])) = uint16(0xABCD)

	switch buf {
	case [2]byte{0xCD, 0xAB}:
		nativeEndian = binary.LittleEndian
	case [2]byte{0xAB, 0xCD}:
		nativeEndian = binary.BigEndian
	default:
		panic("Could not determine native endianness.")
	}
}

func sendRequest(conn net.Conn, t requestType, data string) (string, error) {
	c, ok := conn.(*net.UnixConn)
	if !ok {
		return "", fmt.Errorf("%T is not a unix connection", conn)
	}

	// mcstransd expects null terminated strings
	data += nullChar
	dataSize := uint32(len(data))

	data2Size := uint32(len(nullChar)) // unused by libselinux users

	d2 := []byte(nullChar)

	v := [][]byte{
		(*[4]byte)(unsafe.Pointer(&t))[:],
		(*[4]byte)(unsafe.Pointer(&dataSize))[:],
		(*[4]byte)(unsafe.Pointer(&data2Size))[:],
		[]byte(data),
		[]byte(d2),
	}

	f, err := c.File()
	if err != nil {
		return "", fmt.Errorf("failed to create file from UDS: %w", err)
	}
	defer f.Close()

	writer := net.Buffers(v)
	if _, err := writer.WriteTo(c); err != nil {
		return "", fmt.Errorf("failed to write to mcstransd: %w", err)
	}

	var uintsize uint32

	var hdr [unsafe.Sizeof(uintsize) * 3]byte
	_, err = c.Read(hdr[0:])
	if err != nil {
		return "", fmt.Errorf("failed to read from mcstransd: %w", err)
	}

	//function := nativeEndian.Uint32(hdr[0:4]) // unused
	responselen := nativeEndian.Uint32(hdr[4:8])
	//returncode := nativeEndian.Uint32(hdr[8:12]) // unused

	response := make([]byte, responselen)
	_, err = c.Read(response[0:])
	if err != nil {
		return "", fmt.Errorf("failed to read from mcstransd: %w", err)
	}

	str := strings.Trim(string(response), nullChar)
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

	setNativeEndian()

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
