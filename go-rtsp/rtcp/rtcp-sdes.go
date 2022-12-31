package rtcp

import "encoding/binary"

//           0                   1                   2                   3
//           0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
// 			+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//  header  |V=2|P|    SC   |  PT=SDES=202  |             length            |
// 			+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+
// chunk 1	|                          SSRC/CSRC_1                          |
// 	     	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// 			|                           SDES items                          |
// 			|                              ...                              |
// 			+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+
// chunk 2  |                          SSRC/CSRC_2                          |
//    		+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// 			|                           SDES items                          |
// 			|                              ...                              |
// 			+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+

const (
    SDES_CNAME = 1
    SDES_NAME  = 2
    SDES_EMAIL = 3
    SDES_PHONE = 4
    SDES_LOC   = 5
    SDES_TOOL  = 6
    SDES_NOTE  = 7
    SDES_PRIV  = 8
)

type ChunkItem struct {
    Type   uint8
    Length uint8
    Txt    []byte
}

func (item *ChunkItem) Encode() []byte {
    data := make([]byte, 2+len(item.Txt))
    data[0] = item.Type
    data[1] = item.Length
    copy(data, item.Txt)
    return data
}

func MakeCNameItem(name []byte) *ChunkItem {
    return &ChunkItem{
        Type:   SDES_CNAME,
        Length: uint8(len(name)),
        Txt:    make([]byte, len(name)),
    }
}

type SDESChunk struct {
    SSRC uint32
    Item *ChunkItem
}

type SourceDescription struct {
    Comm
    SC     uint8
    Chunks []SDESChunk
}

func NewSourceDescription() *SourceDescription {
    return &SourceDescription{
        Comm: Comm{PT: RTCP_SDES},
    }
}

func (pkt *SourceDescription) Decode(data []byte) error {
    if err := pkt.Comm.Decode(data); err != nil {
        return err
    }
    pkt.SC = data[0] & 0x1F
    offset := 4
    for i := 0; i < int(pkt.SC); i++ {
        chk := SDESChunk{}
        chk.SSRC = binary.BigEndian.Uint32(data[4:])
        offset += 4
        chk.Item = &ChunkItem{
            Type:   data[offset],
            Length: data[offset+1],
        }
        chk.Item.Txt = make([]byte, chk.Item.Length)
        copy(chk.Item.Txt, data[offset+2:offset+2+int(chk.Item.Length)])
    }
    return nil
}

func (pkt *SourceDescription) Encode() []byte {
    pkt.Comm.Length = pkt.calcLength()
    data := pkt.Comm.Encode()
    pkt.SC = uint8(len(pkt.Chunks))
    data[0] |= pkt.SC & 0x1f
    offset := 4
    for _, chk := range pkt.Chunks {
        binary.BigEndian.PutUint32(data[offset:], chk.SSRC)
        offset += 4
        data[offset] = chk.Item.Type
        data[offset+1] = chk.Item.Length
        copy(data[offset+2:], chk.Item.Txt)
        offset += 2 + len(chk.Item.Txt)
    }
    return data
}

func (pkt *SourceDescription) calcLength() uint16 {
    length := 0
    for _, chk := range pkt.Chunks {
        length += 4
        length += int(chk.Item.Length) + 2
    }
    length += 1
    if length%4 == 0 {
        return uint16(length) / 4
    } else {
        return uint16(length)/4 + 1
    }
}
