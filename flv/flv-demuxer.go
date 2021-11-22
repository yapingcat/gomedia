package flv

import (
    "io"
)

func ReadAudioTag(reader io.Reader) (tag AudioTag, err error) {
    return ReadAudioTagWithTimeout(reader, 0)
}

func ReadAudioTagWithTimeout(reader io.Reader, timeout uint32) (tag AudioTag, err error) {
    buf := make([]byte, 2)
    if _, err := readAtLeastWithTimeout(reader, buf[:1], 1, timeout); err != nil {
        return AudioTag{}, err
    }

    if (buf[0]&0xF0)>>4 == byte(FLV_AAC) {
        if _, err := readAtLeastWithTimeout(reader, buf[1:2], 1, timeout); err != nil {
            return AudioTag{}, err
        }
        tag.Decode(buf[:2])
    } else {
        tag.Decode(buf[:1])
    }
    return
}

func ReadVideoTag(reader io.Reader) (tag VideoTag, err error) {
    return ReadVideoTagWithTimeout(reader, 0)
}

func ReadVideoTagWithTimeout(reader io.Reader, timeout uint32) (tag VideoTag, err error) {
    buf := make([]byte, 5)
    if _, err := readAtLeastWithTimeout(reader, buf[:1], 1, timeout); err != nil {
        return VideoTag{}, err
    }

    codecid := buf[0] & 0x0F
    if codecid == byte(FLV_AVC) || codecid == byte(FLV_HEVC) {
        if _, err := readAtLeastWithTimeout(reader, buf[1:5], 4, timeout); err != nil {
            return VideoTag{}, err
        }
        tag.Decode(buf[:5])
    } else {
        tag.Decode(buf[:1])
    }
    return
}
