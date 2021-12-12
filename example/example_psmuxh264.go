package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "strings"

    "github.com/yapingcat/gomedia/mpeg"
    "github.com/yapingcat/gomedia/mpeg2"
)

func main() {
    f, err := os.Open(os.Args[1])
    if err != nil {
        fmt.Println(err)
        return
    }
    defer f.Close()
    psfilename := os.Args[1] + ".ps"

    ps, err := os.OpenFile(psfilename, os.O_CREATE|os.O_RDWR, 666)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer ps.Close()

    muxer := mpeg2.NewPsMuxer()
    muxer.OnPacket = func(pkg []byte) {
        ps.Write(pkg)
    }
    var pid uint8 = 0
    FileName := strings.ToUpper(os.Args[1])
    if strings.HasSuffix(FileName, "H264") ||
        strings.HasSuffix(FileName, "264") {
        pid = muxer.AddStream(mpeg2.PS_STREAM_H264)
    } else if strings.HasSuffix(FileName, "H265") ||
        strings.HasSuffix(FileName, "265") ||
        strings.HasSuffix(FileName, "HEVC") {
        pid = muxer.AddStream(mpeg2.PS_STREAM_H265)
    }

    buf, _ := ioutil.ReadAll(f)
    pts := uint64(0)
    dts := uint64(0)
    mpeg.SplitFrameWithStartCode(buf, func(nalu []byte) bool {
        muxer.Write(pid, nalu, pts, dts)
        if mpeg.H264NaluType(nalu) <= mpeg.H264_NAL_I_SLICE {
            pts += 40
            dts += 40
        }
        return true
    })

}
