package rtcp

import "encoding/binary"

//  	  0                   1                   2                   3
//  	  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
// 	     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// 	     |V=2|P|    SC   |   PT=BYE=203  |             length            |
// 	     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// 	     |                           SSRC/CSRC                           |
// 	     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// 	     :                              ...                              :
// 	     +=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+
// (opt) |     length    |            reason for leaving     ...
// 	     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

type Bye struct {
    Comm
    SC        uint8
    ReasonLen uint8
    Reason    string
    SSRCS     []uint32
}

func NewBye() *Bye {
    return &Bye{
        Comm: Comm{PT: RTCP_BYE},
    }
}

func (pkt *Bye) Decode(data []byte) error {

    if err := pkt.Comm.Decode(data); err != nil {
        return err
    }
    pkt.SC = data[0] & 0x1F
    offset := 4
    for i := 0; i < int(pkt.SC); i++ {
        pkt.SSRCS = append(pkt.SSRCS, binary.BigEndian.Uint32(data[offset+i*4:]))
        offset += i * 4
    }

    pkt.ReasonLen = data[offset]
    offset++
    pkt.Reason = string(data[offset : offset+int(pkt.ReasonLen)])
    return nil
}

func (pkt *Bye) Encode() []byte {
    pkt.Comm.Length = pkt.calcLength()
    data := pkt.Comm.Encode()
    data[1] |= (0x1F & pkt.SC)
    offset := 4
    for _, ssrc := range pkt.SSRCS {
        binary.BigEndian.PutUint32(data[offset:], ssrc)
        offset += 4
    }
    if len(pkt.Reason) > 0 {
        data[offset] = byte(len(pkt.Reason))
        copy(data[offset+1:], []byte(pkt.Reason))
    }
    return data
}

func (pkt *Bye) calcLength() uint16 {
    length := len(pkt.SSRCS) * 4
    if (len(pkt.Reason)+1)%4 == 0 {
        length += len(pkt.Reason) + 1
    } else {
        length += (len(pkt.Reason) + 4) / 4 * 4
    }
    return uint16(length) / 4
}
