package mp4

import (
    "encoding/binary"
    "io"
)

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

const (
    TR_FLAG_DATA_OFFSET                  uint32 = 0x000001
    TR_FLAG_DATA_FIRST_SAMPLE_FLAGS      uint32 = 0x000004
    TR_FLAG_DATA_SAMPLE_DURATION         uint32 = 0x000100
    TR_FLAG_DATA_SAMPLE_SIZE             uint32 = 0x000200
    TR_FLAG_DATA_SAMPLE_FLAGS            uint32 = 0x000400
    TR_FLAG_DATA_SAMPLE_COMPOSITION_TIME uint32 = 0x000800
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
        n += 4
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

func (trun *TrackRunBox) Decode(r io.Reader) (offset int, err error) {
    if offset, err = trun.Box.Decode(r); err != nil {
        return
    }
    needSize := trun.Box.Box.Size - 12
    buf := make([]byte, needSize)
    if _, err = io.ReadFull(r, buf); err != nil {
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
        binary.BigEndian.PutUint32(buf[offset:], uint32(trun.Dataoffset))
        offset += 4
    }
    if trunFlags&uint32(TR_FLAG_DATA_FIRST_SAMPLE_FLAGS) > 0 {
        binary.BigEndian.PutUint32(buf[offset:], trun.FirstSampleFlags)
        offset += 4
    }

    for i := 0; i < int(trun.SampleCount); i++ {
        if trunFlags&uint32(TR_FLAG_DATA_SAMPLE_DURATION) != 0 {
            binary.BigEndian.PutUint32(buf[offset:], trun.EntryList.entrys[i].sampleDuration)
            offset += 4
        }
        if trunFlags&uint32(TR_FLAG_DATA_SAMPLE_SIZE) != 0 {
            binary.BigEndian.PutUint32(buf[offset:], trun.EntryList.entrys[i].sampleSize)
            offset += 4
        }
        if trunFlags&uint32(TR_FLAG_DATA_SAMPLE_FLAGS) != 0 {
            binary.BigEndian.PutUint32(buf[offset:], trun.EntryList.entrys[i].sampleSize)
            trun.EntryList.entrys[i].sampleFlags = binary.BigEndian.Uint32(buf[offset:])
            offset += 4
        }
        if trunFlags&uint32(TR_FLAG_DATA_SAMPLE_COMPOSITION_TIME) != 0 {
            binary.BigEndian.PutUint32(buf[offset:], trun.EntryList.entrys[i].sampleCompositionTimeOffset)
            offset += 4
        }
    }
    return offset, buf
}

func makeTrunBoxes(track *mp4track) []byte {
    boxes := make([]byte, 0, 128)
    start := 0
    end := 0
    for i := 1; i < len(track.samplelist); i++ {
        if track.samplelist[i].offset == track.samplelist[i-1].offset+track.samplelist[i-1].size {
            continue
        }
        end = i
        boxes = append(boxes, makeTurnBox(track, start, end)...)
        start = end
    }

    if start < len(track.samplelist) {
        boxes = append(boxes, makeTurnBox(track, start, len(track.samplelist))...)
    }
    return boxes
}

func makeTurnBox(track *mp4track, start, end int) []byte {
    flag := TR_FLAG_DATA_OFFSET
    if isVideo(track.cid) && track.samplelist[start].isKeyFrame {
        flag |= TR_FLAG_DATA_FIRST_SAMPLE_FLAGS
    }

    for j := start; j < end; j++ {
        if track.samplelist[j].size != uint64(track.defaultSize) {
            flag |= TR_FLAG_DATA_SAMPLE_SIZE
        }
        if j+1 < end {
            if track.samplelist[j+1].dts-track.samplelist[j].dts != uint64(track.defaultDuration) {
                flag |= TR_FLAG_DATA_SAMPLE_DURATION
            }
        } else {
            if track.lastSample.dts-track.samplelist[j].dts != uint64(track.defaultDuration) {
                flag |= TR_FLAG_DATA_SAMPLE_DURATION
            }
        }
        if track.samplelist[j].pts != track.samplelist[j].dts {
            flag |= TR_FLAG_DATA_SAMPLE_COMPOSITION_TIME
        }
    }

    trun := NewTrackRunBox()
    trun.Box.Flags[2] = uint8(flag >> 16)
    trun.Box.Flags[1] = uint8(flag >> 8)
    trun.Box.Flags[0] = uint8(flag)
    trun.SampleCount = uint32(end - start)
    trun.Dataoffset = int32(track.samplelist[start].offset)
    trun.FirstSampleFlags = MOV_FRAG_SAMPLE_FLAG_DEPENDS_NO
    trun.EntryList = new(movtrun)
    for i := start; i < end; i++ {
        sampleDuration := uint32(0)
        if i == end-1 {
            sampleDuration = uint32(track.lastSample.dts - track.samplelist[i].dts)
        } else {
            sampleDuration = uint32(track.samplelist[i+1].dts - track.samplelist[i].dts)
        }

        entry := trunEntry{
            sampleDuration:              sampleDuration,
            sampleSize:                  uint32(track.samplelist[i].size),
            sampleCompositionTimeOffset: uint32(track.samplelist[i].pts - track.samplelist[i].dts),
        }
        trun.EntryList.entrys = append(trun.EntryList.entrys, entry)
    }
    _, boxData := trun.Encode()
    return boxData
}
