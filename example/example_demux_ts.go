package main

import (
    "bytes"
    "fmt"
    "io/ioutil"
    "os"

    "github.com/yapingcat/gomedia/codec"
    "github.com/yapingcat/gomedia/mpeg2"
)

func main() {

    tsfile := os.Args[1]
    tsFd, err := os.Open(tsfile)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer tsFd.Close()
    h264FileFd, err := os.OpenFile("video.h264", os.O_CREATE|os.O_RDWR, 0666)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer h264FileFd.Close()

    aacFileFd, err := os.OpenFile("audio.aac", os.O_CREATE|os.O_RDWR, 0666)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer aacFileFd.Close()

    fd3, err := os.OpenFile("ts_debug.log", os.O_CREATE|os.O_RDWR, 0666)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer fd3.Close()

    foundAudio := false
    demuxer := mpeg2.NewTSDemuxer()
    demuxer.OnFrame = func(cid mpeg2.TS_STREAM_TYPE, frame []byte, pts uint64, dts uint64) {
        if cid == mpeg2.TS_STREAM_H264 {
            naluType := codec.H264NaluType(frame)
            fmt.Println("frame:", "naluType", naluType, "pts:", pts, "dts:", dts, "size:", len(frame))
            n, err := h264FileFd.Write(frame)
            if err != nil || n != len(frame) {
                fmt.Println(err)
            }
        } else if cid == mpeg2.TS_STREAM_AAC {
            if !foundAudio {
                foundAudio = true
            }
            n, err := aacFileFd.Write(frame)
            if err != nil || n != len(frame) {
                fmt.Println(err)
            }
        }
    }
    demuxer.OnTSPacket = func(pkg *mpeg2.TSPacket) {
        if pkg == nil {
            return
        }
        fd3.WriteString("\n***************TS Packet******************\n")
        fd3.WriteString("---------------TS Header------------------\n")
        pkg.PrettyPrint(fd3)
        if pkg.Field != nil {
            fd3.WriteString("\n--------------Adaptation Field-----------------\n")
            pkg.Field.PrettyPrint(fd3)
        }
        switch value := pkg.Payload.(type) {
        case *mpeg2.Pat:
            fd3.WriteString("\n----------------PAT------------------\n")
            value.PrettyPrint(fd3)
        case *mpeg2.Pmt:
            fd3.WriteString("\n----------------PMT------------------\n")
            value.PrettyPrint(fd3)
        case *mpeg2.PesPacket:
            fd3.WriteString("\n----------------PES------------------\n")
            value.PrettyPrint(fd3)
        case []byte:
            fd3.WriteString("\n----------------Raw Data------------------\n")
            fd3.WriteString(fmt.Sprintf("Size: %d\n", len(value)))
            fd3.WriteString("Raw Data:")
            for i := 0; i < 12 && i < len(value); i++ {
                if i%4 == 0 {
                    fd3.WriteString("\n")
                    fd3.WriteString("    ")
                }
                fd3.WriteString(fmt.Sprintf("0x%02x ", value[i]))
            }
        }
    }

    buf, _ := ioutil.ReadAll(tsFd)
    fmt.Printf("read %d size\n", len(buf))
    fmt.Println(demuxer.Input(bytes.NewReader(buf)))
    /*
       if ts file is large,please use bufio.NewReader
       demuxer.Input(bufio.NewReader(tsFd))
    */
}
