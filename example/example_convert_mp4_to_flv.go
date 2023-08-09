package main

import (
    "fmt"
    "io"
    "os"

    "github.com/yapingcat/gomedia/go-flv"
    "github.com/yapingcat/gomedia/go-mp4"
)

func main() {

    mp4filename := os.Args[1]

    f, err := os.Open(mp4filename)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer f.Close()
    demuxer := mp4.CreateMp4Demuxer(f)
    if infos, err := demuxer.ReadHead(); err != nil && err != io.EOF {
        fmt.Println(err)
    } else {
        fmt.Printf("%+v\n", infos)
    }
    flvFile, _ := os.OpenFile(os.Args[2], os.O_CREATE|os.O_RDWR, 0666)
    defer flvFile.Close()
    fw := flv.CreateFlvWriter(flvFile)
    fw.WriteFlvHeader()
    for {
        pkg, err := demuxer.ReadPacket()
        if err != nil {
            fmt.Println(err)
            break
        }
        fmt.Printf("track:%d,cid:%+v,pts:%d dts:%d\n", pkg.TrackId, pkg.Cid, pkg.Pts, pkg.Dts)
        if pkg.Cid == mp4.MP4_CODEC_H264 {
            fw.WriteH264(pkg.Data, uint32(pkg.Pts), uint32(pkg.Dts))
        } else if pkg.Cid == mp4.MP4_CODEC_AAC {
            fw.WriteAAC(pkg.Data, uint32(pkg.Pts), uint32(pkg.Dts))
        } else if pkg.Cid == mp4.MP4_CODEC_H265 {
            fw.WriteH265(pkg.Data, uint32(pkg.Pts), uint32(pkg.Dts))
        }
    }
}
