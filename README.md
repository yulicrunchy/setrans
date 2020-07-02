# setrans

[![GoDoc](https://godoc.org/github.com/jbrindle/setrans?status.svg)](https://godoc.org/github.com/jbrindle/setrans)

Go SELinux label translation client.

This package connects to and interacts with mcstransd to get SELinux label and color translations.

## Example

```go
package main

import "github.com/jbrindle/setrans"

func main() {
    conn, _ := setrans.New()
    defer conn.Close()

    translated, _ := conn.TransToRaw("staff_u:staff_r:staff_t:SystemLow-SystemHigh")
    fmt.Println(translated) // staff_u:staff_r:staff_t:s0-s15:c0.c1023
}
```