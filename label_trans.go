// Package setrans provides a mechanism for translating contexts using mcstransd.
package setrans

import (
	"errors"
	"net"
)

type requestType uint32

const (
	reqRawToTrans requestType = 2
	reqTransToRaw requestType = 3
	reqRawToColor requestType = 4
)

var (
	// ErrInvalidLevel is returned if a context cannot be translated because it has an invalid level.
	ErrInvalidLevel = errors.New("invalid level provided")
)

type setransMsg struct {
	label   string
	reqType requestType
}

// Conn is used to keep an setrans connection open to the
// mcstrans socket
type Conn struct {
	conn       net.Conn
	mcstransch chan setransMsg
	errch      chan error
}

// New creates a new connection to mcstransd
func New() (*Conn, error) {
	return new()
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (c *Conn) Close() error {
	return c.close()
}

// TransToRaw accepts a translated SELinux label and returns
// the translation into the raw context
func (c *Conn) TransToRaw(trans string) (string, error) {
	return c.transToRaw(trans)
}

// RawToTrans accepts a raw SELinux label and returns
// the translation of the context from mcstransd
func (c *Conn) RawToTrans(raw string) (string, error) {
	return c.rawToTrans(raw)
}

// RawToColor accepts a raw SELinux label and returns
// the color of the context from mcstransd
func (c *Conn) RawToColor(raw string) (string, error) {
	return c.rawToColor(raw)
}
