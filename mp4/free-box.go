package mp4

type FreeBox struct {
    Box  BasicBox
    Data []byte
}

func NewFreeBox() *FreeBox {
    return &FreeBox{
        Box: BasicBox{
            Type: [4]byte{'f', 'r', 'e', 'e'},
        },
    }
}

func (free *FreeBox) Size() uint64 {
    return 8 + uint64(len(free.Data))
}

func (free *FreeBox) Decode(buf []byte) (int, error) {
    if offset, err := free.Box.Decode(buf); err != nil {
        return 0, err
    } else {
        if uint64(offset) < free.Box.Size {
            free.Data = append(free.Data, buf[offset:free.Box.Size]...)
        }
        return int(free.Box.Size), nil
    }
}

func (free *FreeBox) Encode() (int, []byte) {
    free.Box.Size = free.Size()
    offset, buf := free.Box.Encode()
    copy(buf[offset:], free.Data)
    return int(free.Box.Size), buf
}
