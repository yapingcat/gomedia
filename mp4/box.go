package mp4

import (
    "bytes"
    "encoding/binary"
)

type BoxEncoder interface {
    Encode(buf []byte) (int, error)
}

type BoxDecoder interface {
    Decode(buf []byte) (int, error)
}

type BasicBox struct {
    Size      uint32
    Type      [4]byte
    LargeSize uint64
    UserType  [16]byte
}

func (box *BasicBox) Decode(buf []byte) (int, error) {
    _ = buf[7]
    nn := 0
    box.Size = binary.BigEndian.Uint32(buf)
    copy(box.Type[:], buf[4:8])
    nn = 8
    if box.Size == 1 {
        _ = buf[nn+8]
        box.LargeSize = binary.BigEndian.Uint64(buf[nn:])
        nn += 8
    }
    if bytes.Equal(box.Type[:], []byte("uuid")) {
        _ = buf[nn+16]
        copy(box.UserType[:], buf[nn:])
        nn += 16
    }
    return nn, nil
}

func (box *BasicBox) Encode() []byte {
    buf := make([]byte, 8, 32)
    binary.BigEndian.PutUint32(buf, box.Size)
    copy(buf[4:], box.Type[:])
    nn := 8
    if box.Size == 1 {
        buf = buf[:16]
        nn += 8
        binary.BigEndian.PutUint32(buf[8:], uint32(box.LargeSize))
    }
    if bytes.Equal(box.Type[:], []byte("uuid")) {
        buf = buf[nn : nn+16]
        copy(buf[nn:nn+16], box.UserType[:])
    }
    return buf
}

func NewBasicBox(boxtype [4]byte) *BasicBox {
    return &BasicBox{
        Type: boxtype,
    }
}

type FullBox struct {
    Box     *BasicBox
    Version uint8
    Flags   [3]byte
}

func (box *FullBox) Decode(buf []byte) (int, error) {
    if declen, err := box.Box.Decode(buf); err != nil {
        return 0, err
    } else {
        box.Version = buf[declen]
        copy(box.Flags[:], buf[declen+1:declen+4])
        return declen + 4, nil
    }
}

func (box *FullBox) Encode() []byte {
    buf := box.Box.Encode()
    buf = append(buf, make([]byte, 4)...)
    buf[len(buf)-4] = box.Version
    copy(buf[len(buf)+1:], box.Flags[:])
    return buf
}
