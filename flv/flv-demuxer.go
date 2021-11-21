package flv

import (
    "bufio"
    "io"
)

func ReadAudioTag(reader *bufio.Reader) (tag AudioTag, err error) {
    buf := make([]byte, 2)
    if _, err := io.ReadFull(reader, buf[:1]); err != nil {
        return AudioTag{}, err
    }

    if (buf[0]&0xF0)>>4 == byte(FLV_AAC) {
        if _, err := io.ReadFull(reader, buf[1:2]); err != nil {
            return AudioTag{}, err
        }
        tag.Decode(buf[:2])
    } else {
        tag.Decode(buf[:1])
    }

    return
}

func ReadVideoTag(reader *bufio.Reader) (tag VideoTag, err error) {
    buf := make([]byte, 6)
    if _, err := io.ReadFull(reader, buf[:1]); err != nil {
        return VideoTag{}, err
    }

    codecid := buf[1] & 0x0F
    if codecid == byte(FLV_AVC) || codecid == byte(FLV_HEVC) {
        if _, err := io.ReadFull(reader, buf[2:6]); err != nil {
            return VideoTag{}, err
        }
        tag.Decode(buf[:6])
    } else {
        tag.Decode(buf[:1])
    }
    return
}
