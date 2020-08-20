package engine

import (
	"testing"
)

func TestByteCountSI(t *testing.T) {
	got := ByteCountSI(1000)
	if got != "1.0 kB" {
		t.Errorf("Got %s but wanted 1.0 kB", got)
	}

	got = ByteCountSI(10)
	if got != "10 B" {
		t.Errorf("Got %s but wanted 10 B", got)
	}
}
