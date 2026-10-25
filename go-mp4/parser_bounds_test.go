package mp4

import (
	"bytes"
	"testing"
)

// SaioBox.Decode reads entryCount from the payload and loops reading buf[n:]
// with no bound on n; a 16-byte saio box with entryCount=1 and no entry used
// to slice past the buffer. ReadHead must not panic.
func TestSaioDecodeTruncatedNoPanic(t *testing.T) {
	input := []byte{
		0x00, 0x00, 0x00, 0x10, 's', 'a', 'i', 'o',
		0x00, 0x00, 0x00, 0x00, // FullBox version/flags
		0x00, 0x00, 0x00, 0x01, // entryCount=1, no entry follows
	}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("ReadHead panicked on a truncated saio box: %v", r)
		}
	}()
	_, _ = CreateMp4Demuxer(bytes.NewReader(input)).ReadHead()
}

// TrackRunBox.Decode reads SampleCount from the payload and loops over buf[n:]
// unbounded; a 16-byte trun box with a sample flag and SampleCount=1 but no
// sample entry used to slice past the buffer. ReadHead must not panic.
func TestTrunDecodeTruncatedNoPanic(t *testing.T) {
	input := []byte{
		0x00, 0x00, 0x00, 0x10, 't', 'r', 'u', 'n',
		0x00, 0x00, 0x02, 0x00, // flags 0x000200 (SAMPLE_SIZE)
		0x00, 0x00, 0x00, 0x01, // SampleCount=1, no sample entry follows
	}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("ReadHead panicked on a truncated trun box: %v", r)
		}
	}()
	_, _ = CreateMp4Demuxer(bytes.NewReader(input)).ReadHead()
}
