package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/yapingcat/gomedia/mp4"
)

type cacheReaderSeeker struct {
	buf    []byte
	offset int
}

func newCacheReaderSeeker(buf []byte) *cacheReaderSeeker {
	return &cacheReaderSeeker{
		buf:    buf,
		offset: 0,
	}
}

func (rs *cacheReaderSeeker) Read(p []byte) (n int, err error) {
	if rs.offset == len(rs.buf) {
		return 0, io.EOF
	}
	if rs.offset > len(rs.buf) {
		return -1, errors.New(fmt.Sprint("out of range", rs.offset, len(p)))
	}

	if len(rs.buf)-rs.offset < len(p) {
		copy(p, rs.buf[rs.offset:])
		rs.offset = len(rs.buf)
		return len(rs.buf) - int(rs.offset), nil
	}
	copy(p, rs.buf[rs.offset:])
	rs.offset = rs.offset + len(p)
	return len(p), nil
}

func (rs *cacheReaderSeeker) Seek(offset int64, whence int) (int64, error) {
	if whence == io.SeekCurrent {
		if rs.offset+int(offset) > len(rs.buf) {
			return -1, errors.New(fmt.Sprint("SeekCurrent out of range", len(rs.buf), offset, rs.offset))
		}
		rs.offset += int(offset)
		return int64(rs.offset), nil
	} else if whence == io.SeekStart {
		if offset > int64(len(rs.buf)) {
			return -1, errors.New(fmt.Sprint("SeekStart out of range", len(rs.buf), offset, rs.offset))
		}
		rs.offset = int(offset)
		return offset, nil
	} else {
		if offset > 0 {
			return -1, errors.New(fmt.Sprint("SeekEnd out of range", len(rs.buf), offset, rs.offset))
		}
		rs.offset = len(rs.buf) + int(offset)
		return int64(rs.offset), nil
	}
}

var mp4FileName = flag.String("mp4file", "test.mp4", "mp4 file you want to decode")
var rawVideo = flag.String("videofile", "v.h264", "export raw video data to the videofile")
var rawAudio = flag.String("audiofile", "a.aac", "export raw audio data to the audiofile")

func main() {
	flag.Parse()
	f, err := os.Open(*mp4FileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	vfile, err := os.OpenFile(*rawVideo, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer vfile.Close()
	afile, err := os.OpenFile(*rawAudio, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer afile.Close()

	b, _ := io.ReadAll(f)

	/*
		If there are no specific requirements,please use bytes.Reader
		crs := bytes.NewReader(b)
	*/
	crs := newCacheReaderSeeker(b)

	demuxer := mp4.CreateMp4Demuxer(crs)
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
		if pkg.Cid == mp4.MP4_CODEC_H264 {
			vfile.Write(pkg.Data)
		} else if pkg.Cid == mp4.MP4_CODEC_AAC {
			afile.Write(pkg.Data)
		}
	}

}
