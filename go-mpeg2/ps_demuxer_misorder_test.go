package mpeg2

import "testing"

// PSDemuxer.Input panicked when a system header (0x000001BB) arrived before a
// pack header; it must return an error instead of crashing.
func TestPSSystemHeaderBeforePackHeaderNoPanic(t *testing.T) {
	data := []byte{0x00, 0x00, 0x01, 0xBB}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Input panicked on a misordered PS stream: %v", r)
		}
	}()
	_ = NewPSDemuxer().Input(data)
}
