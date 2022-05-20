package rtmp

import "encoding/binary"

const (
	StreamBegin      = 0
	StreamEOF        = 1
	StreamDry        = 2
	SetBufferLength  = 3
	StreamIsRecorded = 4
	PingRequest      = 6
	PingResponse     = 7
)

type UserEvent struct {
	code int
	data []uint32
}

func makeSetChunkSize(chunkSize uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, chunkSize)
	return b
}

func makeAcknowledgementSize(ackSize uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, ackSize)
	return b
}

func makeSetPeerBandwidth(size uint32, limitType int) []byte {
	b := make([]byte, 5)
	binary.BigEndian.PutUint32(b, size)
	b[4] = byte(limitType)
	return b
}

func makeUserControlMessage(event, value int) []byte {
	msg := make([]byte, 6)
	binary.BigEndian.PutUint16(msg, uint16(event))
	binary.BigEndian.PutUint32(msg[2:], uint32(value))
	return msg
}

func decodeUserControlMsg(data []byte) UserEvent {
	ue := UserEvent{}
	ue.code = int(binary.BigEndian.Uint16(data))
	ue.data = append(ue.data, binary.BigEndian.Uint32(data[2:]))
	if ue.code == SetBufferLength {
		ue.data = append(ue.data, binary.BigEndian.Uint32(data[6:]))
	}
	return ue
}
