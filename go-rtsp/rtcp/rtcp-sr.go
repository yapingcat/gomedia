package rtcp

import (
    "encoding/binary"
)

// 0                   1                   2                   3
// 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
// 			+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// header 	|V=2|P|    RC   |   PT=SR=200   |             length            |
// 			+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// 			|                         SSRC of sender                        |
// 			+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+
// sender 	|              NTP timestamp, most significant word             |
// info   	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// 			|             NTP timestamp, least significant word             |
// 			+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// 			|                         RTP timestamp                         |
// 			+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// 			|                     sender's packet count                     |
// 			+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// 			|                      sender's octet count                     |
// 			+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+
// report 	|                 SSRC_1 (SSRC of first source)                 |
// block  	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// 1    	| fraction lost |       cumulative number of packets lost       |
// 			+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// 			|           extended highest sequence number received           |
// 			+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// 			|                      interarrival jitter                      |
// 			+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// 			|                         last SR (LSR)                         |
// 			+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// 			|                   delay since last SR (DLSR)                  |
// 			+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+
// report 	|                 SSRC_2 (SSRC of second source)                |
// block  	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// 2    	:                               ...                             :
// 			+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+
// 			|                  profile-specific extensions                  |
// 			+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

type SenderReport struct {
    Comm
    RC              uint8
    SSRC            uint32
    NTP             uint64
    RtpTimestamp    uint32
    SendPacketCount uint32
    SendOctetCount  uint32
    Blocks          []ReportBlock
}

func NewSenderReport() *SenderReport {
    return &SenderReport{
        Comm: Comm{PT: RTCP_SR},
    }
}

func (pkt *SenderReport) Decode(data []byte) error {
    if err := pkt.Comm.Decode(data); err != nil {
        return err
    }
    pkt.RC = data[0] & 0x1f
    pkt.SSRC = binary.BigEndian.Uint32(data[4:])
    pkt.NTP = binary.BigEndian.Uint64(data[8:])
    pkt.RtpTimestamp = binary.BigEndian.Uint32(data[16:])
    pkt.SendPacketCount = binary.BigEndian.Uint32(data[20:])
    pkt.SendOctetCount = binary.BigEndian.Uint32(data[24:])
    offset := 28
    for i := 0; i < int(pkt.RC); i++ {
        block := ReportBlock{}
        if err := block.Decode(data[offset:]); err != nil {
            return err
        }
        pkt.Blocks = append(pkt.Blocks, block)
    }
    return nil
}

func (pkt *SenderReport) Encode() []byte {
    pkt.Comm.Length = pkt.calcLength()
    data := pkt.Comm.Encode()
    data[0] |= pkt.RC & 0x1f
    binary.BigEndian.PutUint32(data[4:], pkt.SSRC)
    binary.BigEndian.PutUint64(data[8:], pkt.NTP)
    binary.BigEndian.PutUint32(data[16:], pkt.RtpTimestamp)
    binary.BigEndian.PutUint32(data[20:], pkt.SendPacketCount)
    binary.BigEndian.PutUint32(data[24:], pkt.SendOctetCount)
    offset := 28
    for _, block := range pkt.Blocks {
        copy(data[offset:], block.Encode())
        offset += 24
    }
    return data
}

func (pkt *SenderReport) calcLength() uint16 {
    return uint16(24+(24*len(pkt.Blocks))) / 4
}
