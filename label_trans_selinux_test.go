// +build selinux,linux

package setrans

import (
	"fmt"
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
				orig: "dbclient_u:dbclient_r:dbclient_t:T REL TO USA",
				want: "staff_u:staff_r:staff_t:s0-s15:c0.c1023",
				name: "test a valid TransToRaw",
			},
			{
				orig: "dbclient_u:dbclient_r:dbclient_t:T REL TO NATO",
				want: "staff_u:staff_r:staff_t:s0-s15:c0.c1023",
				name: "test a valid TransToRaw",
			},
		}

		conn, err := New()
		if err != nil {
			t.Fatalf("failed to connect to mcstrans: %q", err)
		}
		defer conn.Close()

		for _, tt := range tests {
			t, _ := conn.TransToRaw(tt.orig)
			fmt.Println(t)
			//check(t, err, tt.err, got, tt.want)
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
