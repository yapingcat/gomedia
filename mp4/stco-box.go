package mp4

import (
    "encoding/binary"
)

// aligned(8) class ChunkOffsetBox
//     extends FullBox(‘stco’, version = 0, 0) {
//         unsigned int(32) entry_count;
//         for (i=1; i <= entry_count; i++) {
//             unsigned int(32) chunk_offset;
//     }
// }
// aligned(8) class ChunkLargeOffsetBox
//     extends FullBox(‘co64’, version = 0, 0) {
//         unsigned int(32) entry_count;
//         for (i=1; i <= entry_count; i++) {
//             unsigned int(64) chunk_offset;
//         }
// }

type ChunkOffsetBox struct {
    box  *FullBox
    stco *movstco
}

func NewChunkOffsetBox() *ChunkOffsetBox {
    return &ChunkOffsetBox{
        box: NewFullBox([4]byte{'s', 't', 'c', 'o'}, 0),
    }
}

func (stco *ChunkOffsetBox) Size() uint64 {
    if stco.stco == nil {
        return stco.box.Size()
    } else {
        return stco.box.Size() + 4 + 4*uint64(stco.stco.entryCount)
    }
}

func (stco *ChunkOffsetBox) Encode() (int, []byte) {
    stco.box.Box.Size = stco.Size()
    offset, buf := stco.box.Encode()
    binary.BigEndian.PutUint32(buf[offset:], stco.stco.entryCount)
    offset += 4
    for i := 0; i < int(stco.stco.entryCount); i++ {
        binary.BigEndian.PutUint32(buf[offset:], uint32(stco.stco.chunkOffsetlist[i]))
        offset += 4
    }
    return offset, buf
}

type ChunkLargeOffsetBox struct {
    box  *FullBox
    stco *movstco
}

func NewChunkLargeOffsetBox() *ChunkLargeOffsetBox {
    return &ChunkLargeOffsetBox{
        box: NewFullBox([4]byte{'c', 'o', '6', '4'}, 4),
    }
}

func (co64 *ChunkLargeOffsetBox) Size() uint64 {
    if co64.stco == nil {
        return co64.box.Size()
    } else {
        return co64.box.Size() + 4 + 8*uint64(co64.stco.entryCount)
    }
}

func (co64 *ChunkLargeOffsetBox) Encode() (int, []byte) {
    co64.box.Box.Size = co64.Size()
    offset, buf := co64.box.Encode()
    binary.BigEndian.PutUint32(buf[offset:], co64.stco.entryCount)
    offset += 4
    for i := 0; i < int(co64.stco.entryCount); i++ {
        binary.BigEndian.PutUint64(buf[offset:], co64.stco.chunkOffsetlist[i])
        offset += 8
    }
    return offset, buf
}

func makeStco(stco *movstco) (boxdata []byte) {
    if stco.entryCount <= 0 {
        panic("stco chunkoffset list is empty")
    }

    if stco.chunkOffsetlist[stco.entryCount-1] > 0xFFFFFFFF {
        co64 := NewChunkLargeOffsetBox()
        co64.stco = stco
        _, boxdata = co64.Encode()
    } else {
        stcobox := NewChunkOffsetBox()
        stcobox.stco = stco
        _, boxdata = stcobox.Encode()
    }
    return
}
