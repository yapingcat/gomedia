package rtp

import "testing"

// H265UnPacker.UnPack read pkg.Payload[0] with no length check (its H264 twin
// guards len==0). A valid 12-byte RTP header with no payload must not panic.
func TestH265UnPackEmptyPayloadNoPanic(t *testing.T) {
	pkt := []byte{
		0x80, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
	}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("UnPack panicked on an empty RTP payload: %v", r)
		}
	}()
	_ = NewH265UnPacker().UnPack(pkt)
}
