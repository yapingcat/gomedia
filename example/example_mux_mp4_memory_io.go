package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/yapingcat/gomedia/go-mp4"
	"github.com/yapingcat/gomedia/go-mpeg2"
)

type cacheWriterSeeker struct {
    buf    []byte
    offset int
}

func newCacheWriterSeeker(capacity int) *cacheWriterSeeker {
    return &cacheWriterSeeker{
        buf:    make([]byte, 0, capacity),
        offset: 0,
    }
}

func (ws *cacheWriterSeeker) Write(p []byte) (n int, err error) {
    if cap(ws.buf)-ws.offset >= len(p) {
        if len(ws.buf) < ws.offset+len(p) {
            ws.buf = ws.buf[:ws.offset+len(p)]
        }
        copy(ws.buf[ws.offset:], p)
        ws.offset += len(p)
        return len(p), nil
    }
    tmp := make([]byte, len(ws.buf), cap(ws.buf)+len(p)*2)
    copy(tmp, ws.buf)
    if len(ws.buf) < ws.offset+len(p) {
        tmp = tmp[:ws.offset+len(p)]
    }
    copy(tmp[ws.offset:], p)
    ws.buf = tmp
    ws.offset += len(p)
    return len(p), nil
}

func (ws *cacheWriterSeeker) Seek(offset int64, whence int) (int64, error) {
    if whence == io.SeekCurrent {
        if ws.offset+int(offset) > len(ws.buf) {
            return -1, errors.New(fmt.Sprint("SeekCurrent out of range", len(ws.buf), offset, ws.offset))
        }
        ws.offset += int(offset)
        return int64(ws.offset), nil
    } else if whence == io.SeekStart {
        if offset > int64(len(ws.buf)) {
            return -1, errors.New(fmt.Sprint("SeekStart out of range", len(ws.buf), offset, ws.offset))
        }
        ws.offset = int(offset)
        return offset, nil
    } else {
        return 0, errors.New("unsupport SeekEnd")
    }
}

func main() {
    tsfile := "demo.ts"
    mp4FileName := "cache.mp4"
    mp4File, err := os.OpenFile(mp4FileName, os.O_CREATE|os.O_RDWR, 0666)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer mp4File.Close()
    fmt.Println(mp4File.Seek(0, io.SeekCurrent))
    cws := newCacheWriterSeeker(4096)
    muxer, err := mp4.CreateMp4Muxer(cws)
    if err != nil {
        fmt.Println(err)
        return
    }
    vtid := muxer.AddVideoTrack(mp4.MP4_CODEC_H264)
    buf, err := ioutil.ReadFile(tsfile)
    if err != nil {
        panic(err)
    }
    fmt.Printf("read %d size\n", len(buf))
    demuxer := mpeg2.NewTSDemuxer()
    demuxer.OnFrame = func(cid mpeg2.TS_STREAM_TYPE, frame []byte, pts uint64, dts uint64) {
        if cid == mpeg2.TS_STREAM_H264 {
            err = muxer.Write(vtid, frame, uint64(pts), uint64(dts))
            if err != nil {
                panic(err)
            }
        }
    }
    err = demuxer.Input(bytes.NewReader(buf))
    if err != nil {
        panic(err)
    }
    fmt.Println("write trailer")
    err = muxer.WriteTrailer()
    if err != nil {
        panic(err)
    }
    _, err = mp4File.Write(cws.buf)
    if err != nil {
        panic(err)
    }
    fmt.Println(cws.offset, len(cws.buf))
}
