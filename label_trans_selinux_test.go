// +build selinux,linux

package setrans

import (
	"strings"
	"testing"
)

func check(t *testing.T, err error, wanterr, got, want string) {
	t.Helper()

	if err != nil {
		if wanterr == "" {
			t.Fatalf("got error %q when no error expected", err)
		}
		if !strings.Contains(err.Error(), wanterr) {
			t.Fatalf("got error %q does not match expected err %q", err, wanterr)
		}
		return
	}

	if got != want {
		t.Errorf("translated %q does not match expected %q", got, want)
	}
}

// These tests require an MLS policy and will not work with the default
// Redhat targeted policy
func TestTranslation(t *testing.T) {
	t.Run("Test TransToRaw", func(t *testing.T) {
		tests := []struct {
			orig, want, err, name string
		}{
			{
				orig: "staff_u:staff_r:staff_t:SystemLow-SystemHigh",
				want: "staff_u:staff_r:staff_t:s0-s15:c0.c1023",
				name: "test a valid TransToRaw",
			},
			{
				orig: "staff_u:staff_r:staff_t:FooLow-FooHigh",
				name: "invalid level not in setrans configuration to TransToRaw should fail",
				err:  ErrInvalidLevel.Error(),
			},
			{
				orig: "FooLow-FooHigh",
				name: "test an incomplete selinux context",
				err:  "failed to read from response",
			},
		}

		conn, err := New()
		if err != nil {
			t.Fatalf("failed to connect to mcstrans: %q", err)
		}
		defer conn.Close()

		for _, tt := range tests {
			got, err := conn.TransToRaw(tt.orig)
			check(t, err, tt.err, got, tt.want)
		}
	})

	t.Run("Test RawToColor", func(t *testing.T) {
		tests := []struct {
			orig, want, err, name string
		}{
			{
				orig: "staff_u:staff_r:staff_t:s0-s15:c0.c1023",
				want: "#000000 #ffffff #000000 #ffffff #000000 #ffffff #000000 #ffffff",
				name: "test a valid RawToColor",
			},
		}

		conn, err := New()
		if err != nil {
			t.Fatalf("failed to connect to mcstrans: %q", err)
		}
		defer conn.Close()

		for _, tt := range tests {
			got, err := conn.RawToColor(tt.orig)
			check(t, err, tt.err, got, tt.want)
		}
	})

	t.Run("Test RawToTrans", func(t *testing.T) {
		tests := []struct {
			orig, want, err, name string
		}{
			{
				orig: "staff_u:staff_r:staff_t:s0-s15:c0.c1023",
				want: "staff_u:staff_r:staff_t:SystemLow-SystemHigh",
				name: "test a valid RawToTrans",
			},
			{
				orig: "staff_u:staff_r:staff_t:x0-x15",
				name: "test invalid level",
				err:  ErrInvalidLevel.Error(),
			},
		}

		conn, err := New()
		if err != nil {
			t.Fatalf("failed to connect to mcstrans: %q", err)
		}
		defer conn.Close()

		for _, tt := range tests {
			got, err := conn.RawToTrans(tt.orig)
			check(t, err, tt.err, got, tt.want)
		}
	})
}

func BenchmarkTranslation(b *testing.B) {
	input := "staff_u:staff_r:staff_t:SystemLow-SystemHigh"

	conn, err := New()
	if err != nil {
		b.Fatalf("failed to connect to mcstrans: %q", err)
	}
	defer conn.Close()

	for n := 0; n < b.N; n++ {
		if _, err := conn.TransToRaw(input); err != nil {
			b.Fatalf("failed to translate to raw: %v", err)
		}
	}
}
