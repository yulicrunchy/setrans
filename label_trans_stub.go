// +build !selinux !linux

package setrans

func new() (*Conn, error) {
	return nil, nil
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (c *Conn) close() error {
	return nil
}

// TransToRaw accepts a translated SELinux label and returns
// the translation into the raw context
func (c *Conn) transToRaw(trans string) (string, error) {
	return "", nil
}

// RawToTrans accepts a raw SELinux label and returns
// the translation of the context from mcstransd
func (c *Conn) rawToTrans(raw string) (string, error) {
	return "", nil
}

// RawToColor accepts a raw SELinux label and returns
// the color of the context from mcstransd
func (c *Conn) rawToColor(raw string) (string, error) {
	return "", nil
}
