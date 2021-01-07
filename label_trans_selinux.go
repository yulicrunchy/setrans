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

// getNativeEndian is a cursed function and I hate it.
// The Go developers have decided that machine byteorder is an unimportant
// detail and 1) force you to specify order when decoding byte arrays, and
// 2) don't give you a simple way of figuring out the native endianness
// so here we go. Create a pointer and then look at the byte order to set
// the interface for reading off of the socket.
func getNativeEndian() (binary.ByteOrder, error) {
	buf := [2]byte{}
	*(*uint16)(unsafe.Pointer(&buf[0])) = uint16(0xABCD)

	switch buf {
	case [2]byte{0xCD, 0xAB}:
		return binary.LittleEndian, nil
	case [2]byte{0xAB, 0xCD}:
		return binary.BigEndian, nil
	default:
		return nil, fmt.Errorf("could not determine native endianness")
	}
}

func (c *Conn) sendRequest(t requestType, data string) (string, error) {
	con, ok := c.conn.(*net.UnixConn)
	if !ok {
		return "", fmt.Errorf("%T is not a unix connection", c.conn)
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

	writer := net.Buffers(v)
	if _, err := writer.WriteTo(con); err != nil {
		return "", fmt.Errorf("failed to write to mcstransd: %w", err)
	}

	var uintsize uint32

	var hdr [unsafe.Sizeof(uintsize) * 3]byte
	_, err := con.Read(hdr[0:])
	if err != nil {
		return "", fmt.Errorf("failed to read from mcstransd: %w", err)
	}

	// the hdr buffer contains the following structure:
	// function: 0-4
	// response length: 4-8
	// return code: 8-12
	responselen := c.nativeEndian.Uint32(hdr[4:8])

	response := make([]byte, responselen)
	_, err = con.Read(response[0:])
	if err != nil {
		return "", fmt.Errorf("failed to read from mcstransd: %w", err)
	}

	str := strings.Trim(string(response), nullChar)
	return strings.TrimSpace(str), nil
}

func (c *Conn) makeRequest(label string, t requestType) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	response, err := c.sendRequest(t, label)
	if err != nil {
		return "", fmt.Errorf("failed to send initial request: %v: for label: %s: %w", t, label, err)
	}

	return response, nil
}

func new() (*Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var d net.Dialer
	conn, err := d.DialContext(ctx, "unix", sockpath)
	if err != nil {
		return nil, fmt.Errorf("failed to dial mcstransd: %w", err)
	}

	ne, err := getNativeEndian()
	if err != nil {
		return nil, err
	}

	return &Conn{
		conn:         conn,
		nativeEndian: ne,
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
