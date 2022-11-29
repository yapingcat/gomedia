package rtcp

import (
    "encoding/binary"
    "errors"
)

type Comm struct {
    Padding     bool
    PT          uint8
    Length      uint16
    PayloadLen  uint16
    PaddingData []byte
}

func (pkt *Comm) Decode(data []byte) error {
    if len(data) < 4 {
        return errors.New("length of app rtcp < 4")
    }
    v := data[0] >> 6
    if v != 2 {
        return errors.New("unsupport rtcp version")
    }
    if (data[0] >> 5) > 0 {
        pkt.Padding = true
    } else {
        pkt.Padding = false
    }
    pkt.PT = data[1]
    pkt.Length = binary.BigEndian.Uint16(data[2:])
    pkt.PayloadLen = pkt.Length * 4
    if pkt.Padding {
        paddingLen := data[pkt.Length*4-1]
        pkt.PayloadLen = pkt.Length*4 - uint16(paddingLen)
        pkt.PaddingData = data[pkt.Length*4-uint16(paddingLen):]
    }
    return nil
}

func (pkt *Comm) Encode() []byte {
    data := make([]byte, pkt.Length*4+4)
    data[0] = 0x80
    if pkt.Padding {
        data[0] |= 0x20
        data[len(data)-1] = byte(len(pkt.PaddingData)) + 1
        copy(data[len(data)-1-len(pkt.PaddingData):], pkt.PaddingData)
    }
    data[1] = pkt.PT
    binary.BigEndian.PutUint16(data[2:], pkt.Length)
    return data
}
