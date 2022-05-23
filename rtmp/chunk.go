package rtmp

import (
    "encoding/binary"
)

var ChunkType [4]byte = [4]byte{11, 7, 3, 0}

type basicHead struct {
    fmt  uint8
    csid uint32
}

func (bh *basicHead) encode() []byte {
    hdr := make([]byte, 3)
    hdr[0] = bh.fmt << 6
    if bh.csid < 64 {
        hdr[0] |= uint8(bh.csid)
        return hdr[:1]
    } else if bh.csid < 320 {
        hdr[1] = byte(bh.csid - 64)
        return hdr[:2]
    } else if bh.csid < 65600 {
        hdr[0] |= 1
        binary.BigEndian.PutUint16(hdr[1:], uint16(bh.csid-64))
        return hdr
    } else {
        panic("invaild csid")
    }
}

func (bh *basicHead) decode(data []byte) {
    bh.fmt = data[0] >> 6
    bh.csid = uint32(data[0] & 0x3F)
    if bh.csid == 0 {
        bh.csid = uint32(data[1]) + 64
    } else if bh.csid == 1 {
        bh.csid = uint32(data[2])*256 + uint32(data[1]) + 64
    }
}

type chunkMsgHead struct {
    timestamp   uint32
    msgLen      uint32
    msgTypeId   uint8
    msgStreamId uint32
}

func (cmh *chunkMsgHead) encode(fmt uint8) []byte {
    hdr := make([]byte, 11)
    switch fmt {
    case 0:
        binary.LittleEndian.PutUint32(hdr[7:], cmh.msgStreamId)
        fallthrough
    case 1:
        hdr[3] = byte(cmh.msgLen >> 16)
        hdr[4] = byte(cmh.msgLen >> 8)
        hdr[5] = byte(cmh.msgLen)
        hdr[6] = cmh.msgTypeId
        fallthrough
    case 2:
        if cmh.timestamp > 0x00ffffff {
            hdr[0] = 0xff
            hdr[1] = 0xff
            hdr[2] = 0xff
        } else {
            hdr[0] = byte(cmh.timestamp >> 16)
            hdr[1] = byte(cmh.timestamp >> 8)
            hdr[2] = byte(cmh.timestamp)
        }
    case 3:
    default:
        panic("unkow fmt")
    }
    return hdr[:ChunkType[fmt]]
}

func (cmh *chunkMsgHead) decode(fmt uint8, data []byte) {
    switch fmt {
    case 0:
        cmh.msgStreamId = uint32(data[7])<<24 | uint32(data[8])<<16 | uint32(data[9])<<8 | uint32(data[10])
        fallthrough
    case 1:
        cmh.msgLen = uint32(data[3])<<16 | uint32(data[4])<<8 | uint32(data[5])
        cmh.msgTypeId = data[6]
        fallthrough
    case 2:
        cmh.timestamp = uint32(data[0])<<16 | uint32(data[1])<<8 | uint32(data[2])
    case 3:
    default:
        panic("unkown fmt")
    }

}

func clacBasicHeadLen(data []byte) int {

    length := 1
    csid := data[0] & 0x3F

    if csid == 0 {
        length += 1
    } else if csid == 1 {
        length += 2
    }

    return length
}

type chunkPacket struct {
    basic  basicHead
    msgHdr chunkMsgHead
    data   []byte
}

func (chk *chunkPacket) decodeHead(data []byte) {

    chk.basic.fmt = data[0] >> 6
    chk.basic.csid = uint32(data[0] & 0x3F)
    if chk.basic.csid == 0 {
        chk.basic.csid = uint32(data[1]) + 64
        data = data[2:]
    } else if chk.basic.csid == 1 {
        chk.basic.csid = uint32(data[2])*256 + uint32(data[1]) + 64
        data = data[3:]
    } else {
        data = data[1:]
    }

    switch chk.basic.fmt {
    case 0:
        chk.msgHdr.msgStreamId = uint32(data[7])<<24 | uint32(data[8])<<16 | uint32(data[9])<<8 | uint32(data[10])
        fallthrough
    case 1:
        chk.msgHdr.msgLen = uint32(data[3])<<16 | uint32(data[4])<<8 | uint32(data[5])
        chk.msgHdr.msgTypeId = data[6]
        fallthrough
    case 2:
        chk.msgHdr.timestamp = uint32(data[0])<<16 | uint32(data[1])<<8 | uint32(data[2])
    case 3:
    default:
        panic("unkown fmt")
    }
}

func (chk *chunkPacket) encode() []byte {
    pkt := chk.basic.encode()
    pkt = append(pkt, chk.msgHdr.encode(chk.basic.fmt)...)
    if chk.msgHdr.timestamp > 0x00ffffff {
        tmp := make([]byte, 4)
        binary.BigEndian.PutUint32(tmp, chk.msgHdr.timestamp)
        pkt = append(pkt, tmp...)
    }
    pkt = append(pkt, chk.data...)
    return pkt
}

type ParserState int

const (
    S_BASIC_HEAD ParserState = iota
    S_MSG_HEAD
    S_EXTEND_TS
    S_PAYLOAD
)

type chunkStreamWriter struct {
    csid      uint32
    timestamp uint32
    current   *chunkPacket
    chunkSize uint32
}

func newChunkStreamWriter(csid uint32) *chunkStreamWriter {
    return &chunkStreamWriter{
        csid:      csid,
        chunkSize: FIX_CHUNK_SIZE,
    }
}

func (cs *chunkStreamWriter) writeData(data []byte, msgType MessageType, streamId uint32, ts uint32) []byte {

    lastChunk := cs.current
    format := 0
    delta := ts
    if lastChunk != nil && streamId == lastChunk.msgHdr.msgStreamId && ts >= cs.timestamp {
        format = 1
        delta = ts - cs.timestamp
        if msgType == MessageType(lastChunk.msgHdr.msgTypeId) && int(lastChunk.msgHdr.msgLen) == len(data) {
            format = 2
            if delta == lastChunk.msgHdr.timestamp {
                format = 3
            }
        }
    }

    if lastChunk == nil {
        cs.current = &chunkPacket{
            basic: basicHead{
                fmt:  uint8(format),
                csid: cs.csid,
            },
            msgHdr: chunkMsgHead{
                timestamp:   delta,
                msgLen:      uint32(len(data)),
                msgTypeId:   uint8(msgType),
                msgStreamId: streamId,
            },
        }
        lastChunk = cs.current
    }
    lastChunk.basic.fmt = uint8(format)
    lastChunk.msgHdr.timestamp = delta
    lastChunk.msgHdr.msgLen = uint32(len(data))
    lastChunk.msgHdr.msgTypeId = uint8(msgType)
    lastChunk.msgHdr.msgStreamId = streamId

    chks := make([]byte, 0, cs.chunkSize)
    for len(data) > 0 {
        if len(data) > int(cs.chunkSize) {
            lastChunk.data = data[:cs.chunkSize]
            data = data[cs.chunkSize:]
        } else {
            lastChunk.data = data
            data = data[:0]
        }
        chks = append(chks, lastChunk.encode()...)
        lastChunk.basic.fmt = 3
    }
    cs.timestamp = ts
    return chks
}

type chunkStream struct {
    timestamp uint32
    pkt       *chunkPacket
    hdr       []byte
    message   []byte
}

func newChunkStream() *chunkStream {
    return &chunkStream{
        timestamp: 0,
        pkt:       &chunkPacket{},
        hdr:       make([]byte, 0, 14),
        message:   make([]byte, 0, FIX_CHUNK_SIZE),
    }
}

type chunkStreamReader struct {
    current   *chunkStream
    cks       map[uint32]*chunkStream
    chunkSize uint32
    state     ParserState
    headCache []byte
}

func newChunkStreamReader(chunkSize uint32) *chunkStreamReader {
    return &chunkStreamReader{
        current:   &chunkStream{},
        cks:       make(map[uint32]*chunkStream),
        state:     S_BASIC_HEAD,
        chunkSize: chunkSize,
        headCache: make([]byte, 0, 14),
    }
}

func (reader *chunkStreamReader) readRtmpMessage(data []byte, onMsg func(*rtmpMessage) error) error {
    for len(data) > 0 {
        switch reader.state {
        case S_BASIC_HEAD:
            length := 0
            if len(reader.headCache) > 0 {
                length = clacBasicHeadLen(reader.headCache)
            } else {
                length = clacBasicHeadLen(data)
            }

            if length > len(reader.headCache)+len(data) {
                reader.headCache = append(reader.headCache, data...)
                return nil
            } else {
                appendLen := length - len(reader.headCache)
                reader.headCache = append(reader.headCache, data[:appendLen]...)
                data = data[appendLen:]
            }
            basic := basicHead{}
            basic.decode(reader.headCache)
            if stream, found := reader.cks[basic.csid]; !found {
                reader.current = newChunkStream()
                reader.cks[basic.csid] = reader.current
            } else {
                reader.current = stream
            }
            reader.current.pkt.basic = basic
            reader.headCache = reader.headCache[:0]
            reader.state = S_MSG_HEAD
            if basic.fmt == 3 {
                if reader.current.pkt.msgHdr.timestamp == 0x00ffffff {
                    reader.state = S_EXTEND_TS
                } else {
                    reader.state = S_PAYLOAD
                }
            }
        case S_MSG_HEAD:
            length := int(ChunkType[reader.current.pkt.basic.fmt])
            if len(data)+len(reader.current.hdr) < length {
                reader.current.hdr = append(reader.current.hdr, data...)
                return nil
            } else {
                appendLen := length - len(reader.current.hdr)
                reader.current.hdr = append(reader.current.hdr, data[:appendLen]...)
                data = data[appendLen:]
            }
            reader.current.pkt.msgHdr.decode(reader.current.pkt.basic.fmt, reader.current.hdr)
            if reader.current.pkt.msgHdr.timestamp == 0x00ffffff {
                reader.state = S_EXTEND_TS
            } else {
                reader.state = S_PAYLOAD
            }
            reader.current.hdr = reader.current.hdr[:0]
        case S_EXTEND_TS:
            if len(data)+len(reader.current.hdr) < 4 {
                reader.current.hdr = append(reader.current.hdr, data...)
                return nil
            } else {
                appendLen := 4 - len(reader.current.hdr)
                reader.current.hdr = append(reader.current.hdr, data[:appendLen]...)
                data = data[appendLen:]
            }
            reader.current.pkt.msgHdr.timestamp = binary.BigEndian.Uint32(reader.current.hdr)
            reader.current.hdr = reader.current.hdr[:0]
            reader.state = S_PAYLOAD
        case S_PAYLOAD:
            needLen := 0
            if int(reader.current.pkt.msgHdr.msgLen)-len(reader.current.message) < int(reader.chunkSize) {
                needLen = int(reader.current.pkt.msgHdr.msgLen) - len(reader.current.message)
            } else {
                needLen = int(reader.chunkSize)
            }
            if len(reader.current.pkt.data) < needLen {
                addlen := needLen - len(reader.current.pkt.data)
                if len(data) >= addlen {
                    reader.current.message = append(reader.current.message, reader.current.pkt.data...)
                    reader.current.message = append(reader.current.message, data[:addlen]...)
                    data = data[addlen:]
                    reader.current.pkt.data = reader.current.pkt.data[:0]
                    reader.state = S_BASIC_HEAD
                } else {
                    reader.current.pkt.data = append(reader.current.pkt.data, data...)
                    data = data[:0]
                    continue
                }
            }

            if int(reader.current.pkt.msgHdr.msgLen) <= len(reader.current.message) {
                if reader.current.pkt.basic.fmt == 0 {
                    reader.current.timestamp = reader.current.pkt.msgHdr.timestamp
                } else {
                    reader.current.timestamp += reader.current.pkt.msgHdr.timestamp
                }
                msg := &rtmpMessage{
                    timestamp: reader.current.timestamp,
                    msg:       make([]byte, int(reader.current.pkt.msgHdr.msgLen)),
                    msgtype:   MessageType(reader.current.pkt.msgHdr.msgTypeId),
                    streamid:  reader.current.pkt.msgHdr.msgStreamId,
                }
                copy(msg.msg, reader.current.message)
                if err := onMsg(msg); err != nil {
                    return err
                }
                reader.current.message = reader.current.message[:0]
            }
        default:
            panic("unkown state")
        }
    }
    return nil
}
