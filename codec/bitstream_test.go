package codec

import (
    "testing"
)

var testbit []byte = []byte{0x01, 0x44, 0x55}

func Test_GetBits(t *testing.T) {
    bs := NewBitStream(testbit)
    t.Log(bs.GetBits(4))
    t.Log(bs.GetBits(4))
    t.Log(bs.GetBit())
    t.Log(bs.GetBits(4))
    t.Log(bs.GetBits(4))
    t.Log(bs.GetBits(4))
    t.Log(bs.GetBits(3))

}

func Test_UnRead(t *testing.T) {
    bs := NewBitStream(testbit)
    t.Log(bs.GetBits(4))
    t.Log(bs.GetBits(4))
    t.Log(bs.GetBit())
    t.Log(bs.GetBits(4))
    t.Log(bs.GetBits(4))
    t.Log(bs.GetBits(4))
    t.Log(bs.GetBits(3))
    bs.UnRead(3)
    t.Log(bs.GetBits(3))
    bs.UnRead(4)
    t.Log(bs.GetBits(4))
    bs.UnRead(5)
    t.Log(bs.GetBits(5))
    bs.UnRead(15)
    t.Log(bs.GetBits(2))
    t.Log(bs.GetBits(3))
}

func Test_SkipBits(t *testing.T) {
    bs := NewBitStream(testbit)
    bs.SkipBits(4)
    t.Log(bs.GetBits(4))
}

func Test_DistanceFromMarkDot(t *testing.T) {
    bs := NewBitStream(testbit)
    bs.SkipBits(4)
    bs.Markdot()
    t.Log(bs.GetBits(4))
    t.Log(bs.GetBits(4))
    t.Log(bs.GetBits(1))
    t.Log(bs.DistanceFromMarkDot())
}

func Test_BitStreamWriter(t *testing.T) {
    bsw := NewBitStreamWriter(4)
    bsw.PutByte(1)
    bsw.PutBytes([]byte{0xdd, 0xFF})
    bsw.PutUint8(3, 2)
    bsw.PutUint16(0x4c, 7)
    bsw.PutUint16(0xED, 6)
    t.Logf("%x", bsw.Bits())
}

func TestBitStream_RemainBits(t *testing.T) {
    type fields struct {
        bits        []byte
        bytesOffset int
        bitsOffset  int
        bitsmark    int
        bytemark    int
    }
    tests := []struct {
        name   string
        fields fields
        want   int
    }{
        {name: "test1", fields: fields{
            bits:        []byte{0x00, 0x01, 0x02, 0x03},
            bytesOffset: 0,
            bitsOffset:  0,
            bitsmark:    0,
            bytemark:    0,
        }, want: 32},
        {name: "test2", fields: fields{
            bits:        []byte{0x00, 0x01, 0x02, 0x03},
            bytesOffset: 0,
            bitsOffset:  1,
            bitsmark:    0,
            bytemark:    0,
        }, want: 31},
        {name: "test2", fields: fields{
            bits:        []byte{0x00, 0x01, 0x02, 0x03},
            bytesOffset: 1,
            bitsOffset:  1,
            bitsmark:    0,
            bytemark:    0,
        }, want: 23},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            bs := &BitStream{
                bits:        tt.fields.bits,
                bytesOffset: tt.fields.bytesOffset,
                bitsOffset:  tt.fields.bitsOffset,
                bitsmark:    tt.fields.bitsmark,
                bytemark:    tt.fields.bytemark,
            }
            if got := bs.RemainBits(); got != tt.want {
                t.Errorf("BitStream.RemainBits() = %v, want %v", got, tt.want)
            }
        })
    }
}

var bits []byte = []byte{0x80}
var bits1 []byte = []byte{0x40}
var bits2 []byte = []byte{0x60}
var bits3 []byte = []byte{0x20}

func TestBitStream_ReadUE(t *testing.T) {
    tests := []struct {
        name string
        bs   *BitStream
        want uint64
    }{
        {name: "test1", bs: NewBitStream(bits), want: 0},
        {name: "test1", bs: NewBitStream(bits1), want: 1},
        {name: "test1", bs: NewBitStream(bits2), want: 2},
        {name: "test1", bs: NewBitStream(bits3), want: 3},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := tt.bs.ReadUE(); got != tt.want {
                t.Errorf("BitStream.ReadUE() = %v, want %v", got, tt.want)
            }
        })
    }
}
