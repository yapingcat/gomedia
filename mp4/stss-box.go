package mp4

import (
	"encoding/binary"
)

// aligned(8) class SyncSampleBox extends FullBox(‘stss’, version = 0, 0) {
//  	unsigned int(32) entry_count;
//  	int i;
//  	for (i=0; i < entry_count; i++) {
//  		unsigned int(32) sample_number;
//  	}
//  }

type SyncSampleBox struct {
	box    *FullBox
	entrys []uint32
}

func NewSyncSampleBox() *SyncSampleBox {
	return &SyncSampleBox{
		box: NewFullBox([4]byte{'s', 't', 's', 's'}, 0),
	}
}

func (stss *SyncSampleBox) Size() uint64 {
	if len(stss.entrys) == 0 {
		return stss.box.Size()
	} else {
		return stss.box.Size() + 4 + 4*uint64(len(stss.entrys))
	}
}

func (stss *SyncSampleBox) Encode() (int, []byte) {
	stss.box.Box.Size = stss.Size()
	offset, buf := stss.box.Encode()
	binary.BigEndian.PutUint32(buf[offset:], uint32(len(stss.entrys)))
	offset += 4
	for _, sampleNumber := range stss.entrys {
		binary.BigEndian.PutUint32(buf[offset:], sampleNumber)
		offset += 4
	}
	return offset, buf
}

func makeStss(track *mp4track) (boxdata []byte) {
	stss := NewSyncSampleBox()
	for i, sample := range track.samplelist {
		if sample.isKeyFrame {
			stss.entrys = append(stss.entrys, uint32(i+1))
		}
	}
	_, boxdata = stss.Encode()
	return
}
