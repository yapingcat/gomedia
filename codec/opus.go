package codec

import (
	"encoding/binary"
	"errors"
)

//ffmpeg opus.h OpusPacket
type OpusPacket struct {
	Code       int
	Config     int
	Stereo     int
	Vbr        int
	FrameCount int
	FrameLen   []uint16
	Frame      []byte
}

func DecodeOpusPacket(packet []byte) *OpusPacket {
	pkt := &OpusPacket{}
	pkt.Code = int(packet[0] & 0x03)
	pkt.Stereo = int((packet[0] >> 2) & 0x01)
	pkt.Config = int(packet[0] >> 3)

	switch pkt.Code {
	case 0:
		pkt.FrameCount = 1
		pkt.FrameLen = make([]uint16, 1)
		pkt.FrameLen[0] = uint16(len(packet) - 1)
		pkt.Frame = packet[1:]
	case 1:
		pkt.FrameCount = 2
		pkt.FrameLen = make([]uint16, 1)
		pkt.FrameLen[0] = uint16(len(packet)-1) / 2
		pkt.Frame = packet[1:]
	case 2:
		pkt.FrameCount = 2
		hdr := 1
		N1 := int(packet[1])
		if N1 >= 252 {
			N1 = N1 + int(packet[2]*4)
			hdr = 2
		}
		pkt.FrameLen = make([]uint16, 2)
		pkt.FrameLen[0] = uint16(N1)
		pkt.FrameLen[1] = uint16(len(packet)-hdr) - uint16(N1)
	case 3:
		hdr := 2
		pkt.Vbr = int(packet[1] >> 7)
		padding := packet[1] >> 6
		pkt.FrameCount = int(packet[1] & 0x1F)
		paddingLen := 0
		if padding == 1 {
			for packet[hdr] == 255 {
				paddingLen += 254
				hdr++
			}
			paddingLen += int(packet[hdr])
		}

		if pkt.Vbr == 0 {
			pkt.FrameLen = make([]uint16, 1)
			pkt.FrameLen[0] = uint16(len(packet)-hdr-paddingLen) / uint16(pkt.FrameCount)
			pkt.Frame = packet[hdr : hdr+int(pkt.FrameLen[0]*uint16(pkt.FrameCount))]
		} else {
			n := 0
			for i := 0; i < int(pkt.FrameCount)-1; i++ {
				N1 := int(packet[hdr])
				hdr += 1
				if N1 >= 252 {
					N1 = N1 + int(packet[hdr]*4)
					hdr += 1
				}
				n += N1
				pkt.FrameLen = append(pkt.FrameLen, uint16(N1))
			}
			lastFrameLen := len(packet) - hdr - paddingLen - n
			pkt.FrameLen = append(pkt.FrameLen, uint16(lastFrameLen))
			pkt.Frame = packet[hdr : hdr+n+lastFrameLen]
		}
	default:
		panic("Error C must <= 3")
	}
	return pkt
}

const (
	LEFT_CHANNEL  = 0
	RIGHT_CHANNEL = 1
)

var (
	vorbisChanLayoutOffset [8][8]byte = [8][8]byte{
		{0},
		{0, 1},
		{0, 2, 1},
		{0, 1, 2, 3},
		{0, 2, 1, 3, 4},
		{0, 2, 1, 5, 3, 4},
		{0, 2, 1, 6, 5, 3, 4},
		{0, 2, 1, 7, 5, 6, 3, 4},
	}
)

type ChannelOrder func(channels int, idx int) int

func defalutOrder(channels int, idx int) int {
	return idx
}

func vorbisOrder(channels int, idx int) int {
	return int(vorbisChanLayoutOffset[channels-1][idx])
}

type ChannelMap struct {
	StreamIdx  int
	ChannelIdx int
	Silence    bool
	Copy       bool
	CopyFrom   int
}

type OpusContext struct {
	Preskip           int
	SampleRate        int
	ChannelCount      int
	StreamCount       int
	StereoStreamCount int
	OutputGain        uint16
	MapType           uint8
	ChannelMaps       []ChannelMap
}

func (ctx *OpusContext) ParseExtranData(extraData []byte) error {
	if string(extraData[0:8]) != "OpusHead" {
		return errors.New("magic signature must equal OpusHead")
	}

	_ = extraData[8] // version
	ctx.ChannelCount = int(extraData[9])
	ctx.Preskip = int(binary.LittleEndian.Uint16(extraData[10:]))
	ctx.SampleRate = int(binary.LittleEndian.Uint32(extraData[12:]))
	ctx.OutputGain = binary.LittleEndian.Uint16(extraData[16:])
	ctx.MapType = extraData[18]
	var channel []byte
	var order ChannelOrder
	if ctx.MapType == 0 {
		ctx.StreamCount = 1
		ctx.StereoStreamCount = ctx.ChannelCount - 1
		channel = []byte{0, 1}
		order = defalutOrder
	} else if ctx.MapType == 1 || ctx.MapType == 2 || ctx.MapType == 255 {
		ctx.StreamCount = int(extraData[19])
		ctx.StereoStreamCount = int(extraData[20])
		if ctx.MapType == 1 {
			channel = extraData[21 : 21+ctx.ChannelCount]
			order = vorbisOrder
		}
	} else {
		return errors.New("unsupport map type 255")
	}

	for i := 0; i < ctx.ChannelCount; i++ {
		cm := ChannelMap{}
		index := channel[order(ctx.ChannelCount, i)]
		if index == 255 {
			cm.Silence = true
			continue
		} else if index > byte(ctx.StereoStreamCount)+byte(ctx.StreamCount) {
			return errors.New("index must < (streamcount + stereo streamcount)")
		}

		for j := 0; j < i; j++ {
			if channel[order(ctx.ChannelCount, i)] == index {
				cm.Copy = true
				cm.CopyFrom = j
				break
			}
		}

		if int(index) < 2*ctx.StereoStreamCount {
			cm.StreamIdx = int(index) / 2
			if index&1 == 0 {
				cm.ChannelIdx = LEFT_CHANNEL
			} else {
				cm.ChannelIdx = RIGHT_CHANNEL
			}
		} else {
			cm.StreamIdx = int(index) - ctx.StereoStreamCount
			cm.ChannelIdx = 0
		}
		ctx.ChannelMaps = append(ctx.ChannelMaps, cm)
	}

	return nil
}

func (ctx *OpusContext) WriteOpusExtraData() []byte {
	extraData := make([]byte, 19)
	copy(extraData, string("OpusHead"))
	extraData[8] = 0x01
	extraData[9] = byte(ctx.ChannelCount)
	binary.LittleEndian.PutUint16(extraData[10:], uint16(ctx.Preskip))
	binary.LittleEndian.PutUint32(extraData[12:], uint32(ctx.SampleRate))
	return extraData
}

func WriteDefaultOpusExtraData() []byte {
	return []byte{
		'O', 'p', 'u', 's', 'H', 'e', 'a', 'd',
		1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
}
