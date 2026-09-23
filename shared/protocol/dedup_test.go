package protocol

import "testing"

func TestDeduper(t *testing.T) {
	d := NewDeduper(2)
	if d.Seen("a") { t.Fatal("first a must be new") }
	if !d.Seen("a") { t.Fatal("second a must be duplicate") }
	if d.Seen("b") { t.Fatal("b must be new") }
	if d.Seen("c") { t.Fatal("c must be new") }
	if d.Seen("a") { t.Fatal("a evicted, must be new again") }
	if d.Seen("") { t.Fatal("empty id never duplicate") }
}
