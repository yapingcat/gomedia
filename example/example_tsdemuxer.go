package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"../mpeg"
	"../mpeg2"
)

func main() {

	tsfile := os.Args[1]
	rfd, _ := os.Open(tsfile)
	buf, _ := ioutil.ReadAll(rfd)
	fmt.Printf("read %d size\n", len(buf))
	fd, err := os.OpenFile("1.h264", os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}

	fd2, err := os.OpenFile("4.aac", os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	fd3, err := os.OpenFile("5.log", os.O_CREATE|os.O_RDWR, 0666)
	demuxer := mpeg2.NewTSDemuxer()
	demuxer.OnFrame = func(cid mpeg2.TS_STREAM_TYPE, frame []byte, pts uint64, dts uint64) {
		if cid == mpeg2.TS_STREAM_H264 {
			if mpeg.H264NaluType(frame) == 9 {
				return
			}
			//fmt.Println(len(frame))
			n, err := fd.Write(frame)
			if err != nil || n != len(frame) {
				fmt.Println(err)
			}
		} else if cid == mpeg2.TS_STREAM_AAC {
			n, err := fd2.Write(frame)
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
	fmt.Println(demuxer.Input(buf))
	fd.Close()
}
