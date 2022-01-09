package mp4

import "encoding/binary"

// aligned(8) class TrackFragmentHeaderBox extends FullBox(‘tfhd’, 0, tf_flags){
//     unsigned int(32) track_ID;
//     // all the following are optional fields
//     unsigned int(64) base_data_offset;
//     unsigned int(32) sample_description_index;
//     unsigned int(32) default_sample_duration;
//     unsigned int(32) default_sample_size;
//     unsigned int(32) default_sample_flags
// }

type MP4_THFD_FLAG uint32

const (
    TF_FLAG_BASE_DATA_OFFSET                 MP4_THFD_FLAG = 0x000001
    TF_FLAG_SAMPLE_DESCRIPTION_INDEX_PRESENT MP4_THFD_FLAG = 0x000002
    TF_FLAG_DEFAULT_SAMPLE_DURATION_PRESENT  MP4_THFD_FLAG = 0x000008
    TF_FLAG_DEFAULT_SAMPLE_SIZE_PRESENT      MP4_THFD_FLAG = 0x000010
    TF_FLAG_DEAAULT_SAMPLE_FLAGS_PRESENT     MP4_THFD_FLAG = 0x000020
    TF_FLAG_DURATION_IS_EMPTY                MP4_THFD_FLAG = 0x010000
    TF_FLAG_DEAAULT_BASE_IS_MOOF             MP4_THFD_FLAG = 0x020000
)

type TrackFragmentHeaderBox struct {
    Box                    *FullBox
    Track_ID               uint32
    BaseDataOffset         uint64
    SampleDescriptionIndex uint32
    DefaultSampleDuration  uint32
    DefaultSampleSize      uint32
    DefaultSampleFlags     uint32
}

func NewTrackFragmentHeaderBox(trackid uint32, tfFlags uint32) *TrackFragmentHeaderBox {
    return &TrackFragmentHeaderBox{
        Box:                    NewFullBox([4]byte{'t', 'f', 'h', 'd'}, 0),
        Track_ID:               trackid,
        SampleDescriptionIndex: 1,
    }
}

func (tfhd *TrackFragmentHeaderBox) Size() uint64 {
    n := tfhd.Box.Size()
    thfdFlags := uint32(tfhd.Box.Flags[0])<<16 | uint32(tfhd.Box.Flags[1])<<8 | uint32(tfhd.Box.Flags[2])
    if thfdFlags&uint32(TF_FLAG_BASE_DATA_OFFSET) > 0 {
        n += 8
    }
    if thfdFlags&uint32(TF_FLAG_SAMPLE_DESCRIPTION_INDEX_PRESENT) > 0 {
        n += 4
    }
    if thfdFlags&uint32(TF_FLAG_DEFAULT_SAMPLE_DURATION_PRESENT) > 0 {
        n += 4
    }
    if thfdFlags&uint32(TF_FLAG_DEFAULT_SAMPLE_SIZE_PRESENT) > 0 {
        n += 4
    }
    if thfdFlags&uint32(TF_FLAG_DEAAULT_SAMPLE_FLAGS_PRESENT) > 0 {
        n += 4
    }
    return n
}

func (tfhd *TrackFragmentHeaderBox) Decode(rh Reader) (offset int, err error) {
    if offset, err = tfhd.Box.Decode(rh); err != nil {
        return
    }

    needSize := tfhd.Box.Box.Size - 12
    buf := make([]byte, needSize)
    if _, err = rh.ReadAtLeast(buf); err != nil {
        return 0, err
    }
    n := 0
    tfhd.Track_ID = binary.BigEndian.Uint32(buf[n:])
    n += 4
    thfdFlags := uint32(tfhd.Box.Flags[0])<<16 | uint32(tfhd.Box.Flags[1])<<8 | uint32(tfhd.Box.Flags[2])
    if thfdFlags&uint32(TF_FLAG_BASE_DATA_OFFSET) > 0 {
        tfhd.BaseDataOffset = binary.BigEndian.Uint64(buf[n:])
        n += 8
    }
    if thfdFlags&uint32(TF_FLAG_SAMPLE_DESCRIPTION_INDEX_PRESENT) > 0 {
        tfhd.SampleDescriptionIndex = binary.BigEndian.Uint32(buf[n:])
        n += 4
    }
    if thfdFlags&uint32(TF_FLAG_DEFAULT_SAMPLE_DURATION_PRESENT) > 0 {
        tfhd.DefaultSampleDuration = binary.BigEndian.Uint32(buf[n:])
        n += 4
    }
    if thfdFlags&uint32(TF_FLAG_DEFAULT_SAMPLE_SIZE_PRESENT) > 0 {
        tfhd.DefaultSampleSize = binary.BigEndian.Uint32(buf[n:])
        n += 4
    }
    if thfdFlags&uint32(TF_FLAG_DEAAULT_SAMPLE_FLAGS_PRESENT) > 0 {
        tfhd.DefaultSampleFlags = binary.BigEndian.Uint32(buf[n:])
        n += 4
    }
    offset += n
    return
}

func (tfhd *TrackFragmentHeaderBox) Encode() (int, []byte) {
    tfhd.Box.Box.Size = tfhd.Size()
    offset, buf := tfhd.Box.Encode()
    binary.BigEndian.PutUint32(buf[offset:], tfhd.Track_ID)
    offset += 4
    thfdFlags := uint32(tfhd.Box.Flags[0])<<16 | uint32(tfhd.Box.Flags[1])<<8 | uint32(tfhd.Box.Flags[2])
    if thfdFlags&uint32(TF_FLAG_BASE_DATA_OFFSET) > 0 {
        binary.BigEndian.PutUint64(buf[offset:], tfhd.BaseDataOffset)
        offset += 8
    }
    if thfdFlags&uint32(TF_FLAG_SAMPLE_DESCRIPTION_INDEX_PRESENT) > 0 {
        binary.BigEndian.PutUint32(buf[offset:], tfhd.SampleDescriptionIndex)
        offset += 4
    }
    if thfdFlags&uint32(TF_FLAG_DEFAULT_SAMPLE_DURATION_PRESENT) > 0 {
        binary.BigEndian.PutUint32(buf[offset:], tfhd.DefaultSampleDuration)
        offset += 4
    }
    if thfdFlags&uint32(TF_FLAG_DEFAULT_SAMPLE_SIZE_PRESENT) > 0 {
        binary.BigEndian.PutUint32(buf[offset:], tfhd.DefaultSampleSize)
        offset += 4
    }
    if thfdFlags&uint32(TF_FLAG_DEAAULT_SAMPLE_FLAGS_PRESENT) > 0 {
        binary.BigEndian.PutUint32(buf[offset:], tfhd.DefaultSampleFlags)
        offset += 4
    }
    return offset, buf
}
