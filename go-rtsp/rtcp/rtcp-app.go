package rtcp

import (
	"encoding/binary"
	"errors"
)

//  0                   1                   2                   3
//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |V=2|P| subtype |   PT=APP=204  |             length            |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                           SSRC/CSRC                           |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                          name (ASCII)                         |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                   application-dependent data                ...
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

type App struct {
	Comm
	SubType uint8
	SSRC    uint32
	Name    []byte
	AppData []byte
}

func (pkt *App) Decode(data []byte) error {
	if err := pkt.Comm.Decode(data); err != nil {
		return err
	}

	pkt.SubType = data[0] & 0x1F
	if pkt.Length > uint16(len(data)-4) {
		return errors.New("app rtcp packet need more data")
	}
	pkt.SSRC = binary.BigEndian.Uint32(data[4:])
	pkt.Name = data[8:12]
	pkt.AppData = data[12 : 12+pkt.PayloadLen]
	return nil
}

func (pkt *App) Encode() []byte {
	pkt.Comm.Length = pkt.calcLength()
	data := pkt.Comm.Encode()
	data[1] |= (0x1F & pkt.SubType)
	offset := 4
	binary.BigEndian.PutUint32(data[offset:], pkt.SSRC)
	offset += 4
	copy(data[offset:], pkt.Name)
	offset += 4
	copy(data[offset:], pkt.AppData)
	return data
}

func (pkt *App) calcLength() uint16 {
	return uint16((8 + len(pkt.AppData) + len(pkt.PaddingData) + 1) / 4)
}
