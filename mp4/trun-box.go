package mp4

import "encoding/binary"

// aligned(8) class TrackRunBox extends FullBox(‘trun’, version, tr_flags) {
//      unsigned int(32) sample_count;
//      // the following are optional fields
//      signed int(32) data_offset;
//       unsigned int(32) first_sample_flags;
//      // all fields in the following array are optional
//      {
//          unsigned int(32) sample_duration;
//          unsigned int(32) sample_size;
//          unsigned int(32) sample_flags
//          if (version == 0)
//          {
//              unsigned int(32) sample_composition_time_offset;
//          }
//          else
//          {
//              signed int(32) sample_composition_time_offset;
//          }
//      }[ sample_count ]
// }

type MP4_TRUN_FALG uint32

const (
    TR_FLAG_DATA_OFFSET                  MP4_TRUN_FALG = 0x000001
    TR_FLAG_DATA_FIRST_SAMPLE_FLAGS      MP4_TRUN_FALG = 0x000004
    TR_FLAG_DATA_SAMPLE_DURATION         MP4_TRUN_FALG = 0x000100
    TR_FLAG_DATA_SAMPLE_SIZE             MP4_TRUN_FALG = 0x000200
    TR_FLAG_DATA_SAMPLE_FLAGS            MP4_TRUN_FALG = 0x000400
    TR_FLAG_DATA_SAMPLE_COMPOSITION_TIME MP4_TRUN_FALG = 0x000800
)

type TrackRunBox struct {
    Box              *FullBox
    SampleCount      uint32
    Dataoffset       int32
    FirstSampleFlags uint32
    EntryList        *movtrun
}

func NewTrackRunBox() *TrackRunBox {
    return &TrackRunBox{
        Box: NewFullBox([4]byte{'t', 'r', 'u', 'n'}, 1),
    }
}

func (trun *TrackRunBox) Size() uint64 {
    n := trun.Box.Size()
    trunFlags := uint32(trun.Box.Flags[0])<<16 | uint32(trun.Box.Flags[1])<<8 | uint32(trun.Box.Flags[2])
    if trunFlags&uint32(TR_FLAG_DATA_OFFSET) > 0 {
        n += 8
    }
    if trunFlags&uint32(TR_FLAG_DATA_FIRST_SAMPLE_FLAGS) > 0 {
        n += 4
    }
    if trunFlags&uint32(TR_FLAG_DATA_SAMPLE_DURATION) > 0 {
        n += 4 * uint64(trun.SampleCount)
    }
    if trunFlags&uint32(TR_FLAG_DATA_SAMPLE_SIZE) > 0 {
        n += 4 * uint64(trun.SampleCount)
    }
    if trunFlags&uint32(TR_FLAG_DATA_SAMPLE_FLAGS) > 0 {
        n += 4 * uint64(trun.SampleCount)
    }
    if trunFlags&uint32(TR_FLAG_DATA_SAMPLE_COMPOSITION_TIME) > 0 {
        n += 4 * uint64(trun.SampleCount)
    }
    return n
}

func (trun *TrackRunBox) Decode(rh Reader) (offset int, err error) {
    if offset, err = trun.Box.Decode(rh); err != nil {
        return
    }
    needSize := trun.Box.Box.Size - 12
    buf := make([]byte, needSize)
    if _, err = rh.ReadAtLeast(buf); err != nil {
        return 0, err
    }
    n := 0
    trun.SampleCount = binary.BigEndian.Uint32(buf[n:])
    n += 4
    trunFlags := uint32(trun.Box.Flags[0])<<16 | uint32(trun.Box.Flags[1])<<8 | uint32(trun.Box.Flags[2])
    if trunFlags&uint32(TR_FLAG_DATA_OFFSET) > 0 {
        trun.Dataoffset = int32(binary.BigEndian.Uint32(buf[n:]))
        n += 4
    }
    if trunFlags&uint32(TR_FLAG_DATA_FIRST_SAMPLE_FLAGS) > 0 {
        trun.FirstSampleFlags = binary.BigEndian.Uint32(buf[n:])
        n += 4
    }
    trun.EntryList = new(movtrun)
    trun.EntryList.entrys = make([]trunEntry, trun.SampleCount)
    for i := 0; i < int(trun.SampleCount); i++ {
        if trunFlags&uint32(TR_FLAG_DATA_SAMPLE_DURATION) > 0 {
            trun.EntryList.entrys[i].sampleDuration = binary.BigEndian.Uint32(buf[n:])
            n += 4
        }
        if trunFlags&uint32(TR_FLAG_DATA_SAMPLE_SIZE) > 0 {
            trun.EntryList.entrys[i].sampleSize = binary.BigEndian.Uint32(buf[n:])
            n += 4
        }
        if trunFlags&uint32(TR_FLAG_DATA_SAMPLE_FLAGS) > 0 {
            trun.EntryList.entrys[i].sampleFlags = binary.BigEndian.Uint32(buf[n:])
            n += 4
        }
        if trunFlags&uint32(TR_FLAG_DATA_SAMPLE_COMPOSITION_TIME) > 0 {
            trun.EntryList.entrys[i].sampleCompositionTimeOffset = binary.BigEndian.Uint32(buf[n:])
            n += 4
        }
    }
    offset += n
    return
}

func (trun *TrackRunBox) Encode() (int, []byte) {
    trun.Box.Box.Size = trun.Size()
    offset, buf := trun.Box.Encode()
    binary.BigEndian.PutUint32(buf[offset:], trun.SampleCount)
    offset += 4
    trunFlags := uint32(trun.Box.Flags[0])<<16 | uint32(trun.Box.Flags[1])<<8 | uint32(trun.Box.Flags[2])
    if trunFlags&uint32(TR_FLAG_DATA_OFFSET) > 0 {
        binary.BigEndian.PutUint32(buf[offset:], trun.SampleCount)
        offset += 4
    }
    if trunFlags&uint32(TR_FLAG_DATA_FIRST_SAMPLE_FLAGS) > 0 {
        binary.BigEndian.PutUint32(buf[offset:], trun.FirstSampleFlags)
        offset += 4
    }
    trun.EntryList = new(movtrun)
    trun.EntryList.entrys = make([]trunEntry, trun.SampleCount)
    for i := 0; i < int(trun.SampleCount); i++ {
        if trunFlags&uint32(TR_FLAG_DATA_SAMPLE_DURATION) > 0 {
            binary.BigEndian.PutUint32(buf[offset:], trun.EntryList.entrys[i].sampleDuration)
            offset += 4
        }
        if trunFlags&uint32(TR_FLAG_DATA_SAMPLE_SIZE) > 0 {
            binary.BigEndian.PutUint32(buf[offset:], trun.EntryList.entrys[i].sampleSize)
            offset += 4
        }
        if trunFlags&uint32(TR_FLAG_DATA_SAMPLE_FLAGS) > 0 {
            trun.EntryList.entrys[i].sampleFlags = binary.BigEndian.Uint32(buf[offset:])
            offset += 4
        }
        if trunFlags&uint32(TR_FLAG_DATA_SAMPLE_COMPOSITION_TIME) > 0 {
            trun.EntryList.entrys[i].sampleCompositionTimeOffset = binary.BigEndian.Uint32(buf[offset:])
            offset += 4
        }
    }
    return offset, buf
}
