package mpeg

import "testing"

var testbit []byte = []byte{0x01, 0x44, 0x55}

func Test_GetBit(t *testing.T) {
	bs := NewBitStream(testbit)
	for i := 0; i < 24; i++ {
		t.Logf("Location:%d,Value:%d", i, bs.GetBit())
	}
	defer func() {
		if err := recover(); err != nil {
			t.Log(err)
		}
	}()
	bs.GetBit()
	t.Error("Except For panic")
}

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
