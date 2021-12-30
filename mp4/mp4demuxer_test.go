package mp4

import (
    "fmt"
    "io"
    "os"
    "testing"
)

type mymp4reader struct {
    fp *os.File
}

func newmymp4reader(f *os.File) *mymp4reader {
    return &mymp4reader{
        fp: f,
    }
}

func (mp4w *mymp4reader) ReadAtLeast(p []byte) (n int, err error) {
    //fmt.Printf("read %d bytes\n", len(p))
    return io.ReadAtLeast(mp4w.fp, p, len(p))
}
func (mp4w *mymp4reader) Seek(offset int64, whence int) (int64, error) {
    //fmt.Printf("seek %d where %d\n", offset, whence)
    return mp4w.fp.Seek(offset, whence)
}
func (mp4w *mymp4reader) Tell() (offset int64) {
    offset, _ = mp4w.fp.Seek(0, 1)
    return
}

func TestCreateMovDemuxer(t *testing.T) {
    f, err := os.Open("source.200kbps.768x320.flv.mp4")
    if err != nil {
        fmt.Println(err)
        return
    }
    defer f.Close()
    vfile, err := os.OpenFile("v.h264", os.O_CREATE|os.O_RDWR, 0666)
    defer vfile.Close()
    afile, err := os.OpenFile("a.aac", os.O_CREATE|os.O_RDWR, 0666)
    defer afile.Close()
    demuxer := CreateMp4Demuxer(newmymp4reader(f))
    if infos, err := demuxer.ReadHead(); err != nil && err != io.EOF {
        fmt.Println(err)
    } else {
        fmt.Printf("%+v\n", infos)
    }
    mp4info := demuxer.GetMp4Info()
    fmt.Printf("%+v\n", mp4info)
    for {
        pkg, err := demuxer.ReadPacket()
        if err != nil {
            fmt.Println(err)
            break
        }
        fmt.Printf("track:%d,cid:%+v,pts:%d dts:%d\n", pkg.TrackId, pkg.Cid, pkg.Pts, pkg.Dts)
        if pkg.Cid == MP4_CODEC_H264 {
            vfile.Write(pkg.Data)
        } else if pkg.Cid == MP4_CODEC_AAC {
            afile.Write(pkg.Data)
        }
    }
}
