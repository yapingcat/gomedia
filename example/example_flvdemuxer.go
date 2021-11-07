package main

import (
	"fmt"
	"os"

	"../flv"
)

func exampleForReader(filename string) {
	fr := flv.CreateFlvFileReader()
	fr.SetOnTag(func(ftag flv.FlvTag, tag interface{}) {
		var infostr string = "Tag:"
		if ftag.TagType == 8 {
			infostr += fmt.Sprintf("[%8s]", "Audio")
		} else if ftag.TagType == 9 {
			infostr += fmt.Sprintf("[%8s]", "Video")
		} else if ftag.TagType == 18 {
			infostr += fmt.Sprintf("[%8s]", "MetaData")
		}
		infostr += fmt.Sprintf("[Size:%8d]", int(ftag.DataSize))
		infostr += fmt.Sprintf("[TimeStamp:%8d]", int(ftag.Timestamp))
		if ftag.TimestampExtended != 0 {
			infostr += fmt.Sprintf("[TimestampExtended:%10d]", int(ftag.TimestampExtended))
		}
		if ftag.TagType == 8 {
			atag := tag.(flv.AudioTag)
			infostr += fmt.Sprintf("[SoundFormat:%2d]", int(atag.SoundFormat))
			infostr += fmt.Sprintf("[SoundRate:%d]", int(atag.SoundRate))
			infostr += fmt.Sprintf("[SoundSize:%d]", int(atag.SoundSize))
			infostr += fmt.Sprintf("[SoundType:%d]", int(atag.SoundType))
			if atag.SoundFormat == 10 {
				infostr += fmt.Sprintf("[AACPacketType:%d]", int(atag.AACPacketType))
			}
		} else if ftag.TagType == 9 {
			vtag := tag.(flv.VideoTag)
			infostr += fmt.Sprintf("[FrameType:%d]", int(vtag.FrameType))
			infostr += fmt.Sprintf("[CodecId:%d]", int(vtag.CodecId))
			if vtag.CodecId == 7 {
				infostr += fmt.Sprintf("[AVCPacketType:%d]", int(vtag.AVCPacketType))
				infostr += fmt.Sprintf("[CompositionTime:%5d]", int(vtag.CompositionTime))
			}
		}
		fmt.Println(infostr)
	})
	fr.Open(filename)
	if err := fr.DeMuxFile(); err != nil {
		fmt.Println(err)
	}

}

func main() {
	fn := os.Args[1]
	exampleForReader(fn)
}
