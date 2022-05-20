package rtmp

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"
	"time"
)

var fmsKey [68]byte = [68]byte{
	0x47, 0x65, 0x6e, 0x75, 0x69, 0x6e, 0x65, 0x20,
	0x41, 0x64, 0x6f, 0x62, 0x65, 0x20, 0x46, 0x6c,
	0x61, 0x73, 0x68, 0x20, 0x4d, 0x65, 0x64, 0x69,
	0x61, 0x20, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x20, 0x30, 0x30, 0x31,
	0xf0, 0xee, 0xc2, 0x4a, 0x80, 0x68, 0xbe, 0xe8,
	0x2e, 0x00, 0xd0, 0xd1, 0x02, 0x9e, 0x7e, 0x57,
	0x6e, 0xec, 0x5d, 0x2d, 0x29, 0x80, 0x6f, 0xab,
	0x93, 0xb8, 0xe6, 0x36, 0xcf, 0xeb, 0x31, 0xae,
}

var fpKey [62]byte = [62]byte{
	0x47, 0x65, 0x6E, 0x75, 0x69, 0x6E, 0x65, 0x20,
	0x41, 0x64, 0x6F, 0x62, 0x65, 0x20, 0x46, 0x6C,
	0x61, 0x73, 0x68, 0x20, 0x50, 0x6C, 0x61, 0x79,
	0x65, 0x72, 0x20, 0x30, 0x30, 0x31,
	0xF0, 0xEE, 0xC2, 0x4A, 0x80, 0x68, 0xBE, 0xE8,
	0x2E, 0x00, 0xD0, 0xD1, 0x02, 0x9E, 0x7E, 0x57,
	0x6E, 0xEC, 0x5D, 0x2D, 0x29, 0x80, 0x6F, 0xAB,
	0x93, 0xB8, 0xE6, 0x36, 0xCF, 0xEB, 0x31, 0xAE,
}

var clientVersion [4]byte = [4]byte{0x80, 0x00, 0x07, 0x02}
var serverVersion [4]byte = [4]byte{0x04, 0x05, 0x00, 0x01}

func init() {
	rand.Seed(time.Now().Unix())
}

// https://blog.csdn.net/win_lin/article/details/13006803

// schema == 0 1536 bytes
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |     time      |    version    |   	 key     |     digest      |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

// schema == 1 1536 bytes
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |     time      |    version    |   	digest   |       key       |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

// key format 764 bytes

// |<- offset1 byte  ->|<-  128 bytes  ->|<-  764-offset-128-4 bytes  ->|<-   4 bytes ->|
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-++-+-+-+-+-+-+-+-+-+-+
// |      random       |       key       |        random                |     offset    |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-++-+-+-+-+-+-+-+-+-+-+

// digest format 764 bytes

// |<- 4 byte  ->|<- offset bytes  ->|<-  32 bytes ->|<-   (764-4-offset-32) bytes  ->|
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-++-+-+-+-+-+-+-+-+-+
// |    offset   |      random       |    digest     |              random            |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-++-+-+-+-+-+-+-+-+-+

func getOffset(data []byte, schema int) uint32 {
	var offset uint32 = 0
	if schema == HANDSHAKE_COMPLEX_SCHEMA0 {
		offset = uint32(data[HANDSHAKE_SCHEMA0_OFFSET-HANDSHAKE_OFFSET_SIZE])
		offset += uint32(data[HANDSHAKE_SCHEMA0_OFFSET-HANDSHAKE_OFFSET_SIZE+1])
		offset += uint32(data[HANDSHAKE_SCHEMA0_OFFSET-HANDSHAKE_OFFSET_SIZE+2])
		offset += uint32(data[HANDSHAKE_SCHEMA0_OFFSET-HANDSHAKE_OFFSET_SIZE+3])
		offset = HANDSHAKE_SCHEMA0_OFFSET + offset
	} else {
		offset = uint32(data[HANDSHAKE_FIX_SIZE])
		offset += uint32(data[HANDSHAKE_FIX_SIZE+1])
		offset += uint32(data[HANDSHAKE_FIX_SIZE+2])
		offset += uint32(data[HANDSHAKE_FIX_SIZE+3])
		offset = HANDSHAKE_FIX_SIZE + HANDSHAKE_OFFSET_SIZE + offset
	}
	return offset
}

func clacDigest(data []byte, key []byte, schema int) (digest [32]byte, offset uint32) {
	ctx := hmac.New(sha256.New, key)
	offset = rand.Uint32() % (HANDSHAKE_SCHEMA_SIZE - HANDSHAKE_OFFSET_SIZE - HANDSHAKE_DIGEST_SIZE)

	if schema == HANDSHAKE_COMPLEX_SCHEMA0 {
		data[HANDSHAKE_SCHEMA0_OFFSET-4] = byte(offset / 4)
		data[HANDSHAKE_SCHEMA0_OFFSET-3] = byte(offset / 4)
		data[HANDSHAKE_SCHEMA0_OFFSET-2] = byte(offset / 4)
		data[HANDSHAKE_SCHEMA0_OFFSET-1] = byte(offset - offset/4*3)
		ctx.Write(data[:HANDSHAKE_SCHEMA0_OFFSET+offset])
		ctx.Write(data[HANDSHAKE_SCHEMA0_OFFSET+offset+HANDSHAKE_DIGEST_SIZE : HANDSHAKE_SIZE])
		copy(digest[:], ctx.Sum(nil))
		offset += HANDSHAKE_SCHEMA0_OFFSET
	} else {
		data[HANDSHAKE_FIX_SIZE] = byte(offset / 4)
		data[HANDSHAKE_FIX_SIZE+1] = byte(offset / 4)
		data[HANDSHAKE_FIX_SIZE+2] = byte(offset / 4)
		data[HANDSHAKE_FIX_SIZE+3] = byte(offset - offset/4*3)
		ctx.Write(data[:HANDSHAKE_SCHEMA1_OFFSET+offset])
		ctx.Write(data[HANDSHAKE_SCHEMA1_OFFSET+offset+HANDSHAKE_DIGEST_SIZE:])
		copy(digest[:], ctx.Sum(nil))
		offset += HANDSHAKE_SCHEMA1_OFFSET
	}
	return
}

func getDigest(data []byte, schema int) []byte {
	offset := getOffset(data, schema)
	return data[offset : offset+32]
}

func makeC0() []byte {
	return []byte{3}
}

func makeC1() []byte {
	ts := uint32(time.Now().Unix())
	c1 := make([]byte, 1536)
	binary.BigEndian.PutUint32(c1, ts)

	for i := 8; i < len(c1); i++ {
		c1[i] = byte(rand.Uint32())
	}
	return c1
}

func makeC2(s1 []byte) []byte {
	c2 := make([]byte, 1536)
	copy(c2, s1[:4])
	binary.BigEndian.PutUint32(c2[4:], uint32(time.Now().Unix()))
	copy(c2[8:], s1[8:])
	return c2
}

func makeS0() []byte {
	return []byte{3}
}

func makeS1() []byte {
	ts := uint32(time.Now().Unix())
	s1 := make([]byte, 1536)
	binary.BigEndian.PutUint32(s1, ts)

	for i := 8; i < len(s1); i++ {
		s1[i] = byte(rand.Uint32())
	}
	return s1
}

func makeS2(c1 []byte) []byte {
	s2 := make([]byte, 1536)
	copy(s2, c1[:4])
	binary.BigEndian.PutUint32(s2[4:], uint32(time.Now().Unix()))
	copy(s2[8:], c1[8:])
	return s2
}

func makeComplexC0() []byte {
	return makeC0()
}

func makeComplexC1(schema int) []byte {
	c1 := make([]byte, 1536)
	for i := 8; i < len(c1); i++ {
		c1[i] = byte(rand.Uint32())
	}
	binary.BigEndian.PutUint32(c1, uint32(time.Now().Unix()))
	copy(c1[4:], clientVersion[:])
	digest, offset := clacDigest(c1, fpKey[:30], schema)
	copy(c1[offset:], digest[:])
	return c1
}

func makeComplexC2(s1 []byte, schema int) []byte {
	c2 := make([]byte, 1536)
	for i := 8; i < len(c2); i++ {
		c2[i] = byte(rand.Uint32())
	}
	s1digest := getDigest(s1, schema)
	ctx := hmac.New(sha256.New, fpKey[:])
	ctx.Write(s1digest)
	tmpKey := ctx.Sum(nil)
	ctx = hmac.New(sha256.New, tmpKey)
	ctx.Write(c2[:1504])
	c2digest := ctx.Sum(nil)
	copy(c2[1504:], c2digest)
	return c2
}

func makeComplexS0() []byte {
	return makeS0()
}

func makeComplexS1(schema int) []byte {

	s1 := make([]byte, 1536)
	for i := 8; i < len(s1); i++ {
		s1[i] = byte(rand.Uint32())
	}

	binary.BigEndian.PutUint32(s1, uint32(time.Now().Unix()))
	copy(s1[4:], serverVersion[:])
	digest, offset := clacDigest(s1, fmsKey[:36], schema)
	copy(s1[offset:], digest[:])
	return s1
}

func makeComplexS2(c1 []byte, schema int) []byte {
	s2 := make([]byte, 1536)
	for i := 8; i < len(s2); i++ {
		s2[i] = byte(rand.Uint32())
	}
	c1digest := getDigest(c1, schema)
	ctx := hmac.New(sha256.New, fmsKey[:])
	ctx.Write(c1digest)
	tmpKey := ctx.Sum(nil)
	ctx = hmac.New(sha256.New, tmpKey)
	ctx.Write(s2[:1504])
	s2digest := ctx.Sum(nil)
	copy(s2[1504:], s2digest)
	return s2
}

type HandShakeState int

const (
	CLIENT_S0 HandShakeState = iota
	CLIENT_S1
	CLIENT_S2

	SERVER_C0 HandShakeState = iota + 10
	SERVER_C1
	SERVER_C2

	HANDSHAKE_DONE HandShakeState = iota + 100
)

type clientHandShake struct {
	version  byte
	schema   int
	simpleHs bool
	cache    []byte
	state    HandShakeState
	output   OutputCB
}

func newClientHandShake() *clientHandShake {
	return &clientHandShake{
		simpleHs: true,
		cache:    make([]byte, 0, 1536),
		state:    CLIENT_S0,
	}
}

func (chs *clientHandShake) start() {
	var c0c1 []byte
	chs.state = CLIENT_S0
	if chs.simpleHs {
		c0c1 = makeC0()
		c0c1 = append(c0c1, makeC1()...)
	} else {
		c0c1 = makeComplexC0()
		c0c1 = append(c0c1, makeComplexC1(chs.schema)...)
		fmt.Println("make complex handshake", len(c0c1))
	}
	chs.output(c0c1)
}

func (chs *clientHandShake) input(data []byte) error {
	for len(data) > 0 {
		switch chs.state {
		case CLIENT_S0:
			chs.version = data[0]
			data = data[1:]
			chs.state = CLIENT_S1
		case CLIENT_S1:
			if len(data)+len(chs.cache) < 1536 {
				chs.cache = append(chs.cache, data...)
				return nil
			} else {
				length := 1536 - len(chs.cache)
				chs.cache = append(chs.cache, data[:length]...)
				data = data[length:]
			}
			var c2 []byte
			if chs.simpleHs {
				c2 = makeC2(chs.cache)
			} else {
				c2 = makeComplexC2(chs.cache, chs.schema)
			}
			chs.output(c2)
			chs.cache = chs.cache[:0]
			chs.state = CLIENT_S2
		case CLIENT_S2:
			if len(data)+len(chs.cache) < 1536 {
				chs.cache = append(chs.cache, data...)
				return nil
			} else {
				length := 1536 - len(chs.cache)
				chs.cache = append(chs.cache, data[:length]...)
				data = data[length:]
			}
			chs.state = HANDSHAKE_DONE
			chs.cache = nil
		default:
			panic("error state")
		}
	}

	return nil
}

func (chs *clientHandShake) getState() HandShakeState {
	return chs.state
}

type serverHandShake struct {
	version  byte
	schema   int
	simpleHs bool
	cache    []byte
	state    HandShakeState
	output   OutputCB
}

func newServerHandShake() *serverHandShake {
	return &serverHandShake{
		simpleHs: true,
		cache:    make([]byte, 0, 1536),
		state:    SERVER_C0,
	}
}

func (shs *serverHandShake) input(data []byte) (readBytes int) {
	for len(data) > 0 {
		switch shs.state {
		case SERVER_C0:
			shs.version = data[0]
			readBytes++
			data = data[1:]
			shs.state = SERVER_C1
		case SERVER_C1:
			if len(data)+len(shs.cache) < 1536 {
				shs.cache = append(shs.cache, data...)
				readBytes += len(data)
				return
			} else {
				length := 1536 - len(shs.cache)
				shs.cache = append(shs.cache, data[:length]...)
				data = data[length:]
				readBytes += length
			}
			var s0s1s2 []byte
			shs.checkC1(shs.cache)
			fmt.Println("handshake type:", shs.simpleHs, "schema:", shs.schema)
			if shs.simpleHs {
				s0s1s2 = makeS0()
				s0s1s2 = append(s0s1s2, makeS1()...)
				s0s1s2 = append(s0s1s2, makeS2(shs.cache)...)
			} else {
				s0s1s2 = makeComplexS0()
				s0s1s2 = append(s0s1s2, makeComplexS1(shs.schema)...)
				s0s1s2 = append(s0s1s2, makeComplexS2(shs.cache, shs.schema)...)
			}
			shs.output(s0s1s2)
			shs.cache = shs.cache[:0]
			shs.state = SERVER_C2
		case SERVER_C2:
			if len(data)+len(shs.cache) < 1536 {
				shs.cache = append(shs.cache, data...)
				readBytes += len(data)
				return
			}
			length := 1536 - len(shs.cache)
			readBytes += length
			shs.state = HANDSHAKE_DONE
			shs.cache = nil
			return readBytes
		default:
			panic("error state")
		}
	}

	return readBytes
}

func (shs *serverHandShake) getState() HandShakeState {
	return shs.state
}

func (shs *serverHandShake) checkC1(c1 []byte) {
	if c1[4] != 0 {
		shs.simpleHs = false
	} else {
		shs.simpleHs = true
		return
	}
	fmt.Println("check digest")
	digest := getDigest(c1, HANDSHAKE_COMPLEX_SCHEMA0)
	ctx := hmac.New(sha256.New, fpKey[:30])
	offset := getOffset(c1, HANDSHAKE_COMPLEX_SCHEMA0)
	ctx.Write(c1[:offset])
	ctx.Write(c1[offset+HANDSHAKE_DIGEST_SIZE:])
	expectDigest := ctx.Sum(nil)
	if bytes.Equal(digest, expectDigest[:]) {
		shs.schema = HANDSHAKE_COMPLEX_SCHEMA0
		return
	} else {
		digest = getDigest(c1, HANDSHAKE_COMPLEX_SCHEMA1)
		ctx := hmac.New(sha256.New, fpKey[:30])
		offset := getOffset(c1, HANDSHAKE_COMPLEX_SCHEMA1)
		ctx.Write(c1[:offset])
		ctx.Write(c1[offset+HANDSHAKE_DIGEST_SIZE:])
		expectDigest := ctx.Sum(nil)
		if bytes.Equal(digest, expectDigest[:]) {
			shs.schema = HANDSHAKE_COMPLEX_SCHEMA1
		} else {
			shs.simpleHs = true
			return
		}
	}
}
