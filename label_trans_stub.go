// +build !selinux !linux

package setrans

func new() (*Conn, error) {
	return nil, nil
}

func (c *Conn) close() error {
	return nil
}

func (c *Conn) transToRaw(trans string) (string, error) {
	return "", nil
}

func (c *Conn) rawToTrans(raw string) (string, error) {
	return "", nil
}

func (c *Conn) rawToColor(raw string) (string, error) {
	return "", nil
}
