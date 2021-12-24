package flv

import (
    "encoding/binary"
    "errors"
    "io"

    "github.com/yapingcat/gomedia/mpeg"
)

//  FLV File
//  TheFLVheader
//  An FLV file shall begin with the FLV header:
//  -------------------------------------------------------------------------------------------------------------
//  FLV header
//  Field                       Type                            Comment
//  -------------------------------------------------------------------------------------------------------------
//  Signature                   UI8                            Signature byte always 'F' (0x46)
//
//  Signature                   UI8                            Signature byte always 'L' (0x4C)
//
//  Signature                   UI8                         Signature byte always 'V' (0x56)
//
//  Version                     UI8                         File version (for example, 0x01 for FLV version 1)
//
//  TypeFlagsReserved           UB[5]                       Shall be 0
//
//  TypeFlagsAudio              UB[1]                       1 = Audio tags are present
//
//  TypeFlagsReserved           UB[1]                       Shall be 0
//
//  TypeFlagsVideo              UB[1]                       1 = Video tags are present
//
//  DataOffset                  UI32                        The length of this header in bytes
//  -------------------------------------------------------------------------------------------------------------
//
//  TheFLVFileBody
//  After the FLV header, the remainder of an FLV file shall consist of alternating back-pointers and tags.
//  They interleave as shown in the following table:
//  -------------------------------------------------------------------------------------------------------------
//  FLV File Body
//  Field                       Type                        Comment
//  -------------------------------------------------------------------------------------------------------------
//  PreviousTagSize0            UI32                        Always 0
//  Tag1                        FLVTAG                      First tag
//  PreviousTagSize1            UI32                        Size of previous tag, including its header, in bytes. For FLV version 1,
//                                                          this value is 11 plus the DataSize of the previous tag
//  Tag2                        FLVTAG                      Second tag
//
//  ....
//
//  PreviousTagSizeN-1          UI32                        Size of second-to-last tag, including its header, in bytes.
//  ---------------------------------------------------------------------------------------------------------------

type FlvReader struct {
    reader      io.Reader
    timeout     uint32
    hasDeadline bool
    asc         []byte
    spss        map[uint64][]byte
    ppss        map[uint64][]byte
    OnFrame     func(mpeg.CodecID, []byte, uint32, uint32)
    OnTag       func(ftag FlvTag, tag interface{})
}

func CreateFlvReader(reader io.Reader) *FlvReader {
    flvFile := &FlvReader{
        reader:  reader,
        timeout: 0,
        asc:     make([]byte, 512),
        spss:    make(map[uint64][]byte),
        ppss:    make(map[uint64][]byte),
        OnFrame: nil,
        OnTag:   nil,
    }
    return flvFile
}

//设置读超时，只有对实现了SetReadDeadline接口的 io.Reader 有效
//如果没有实现SetReadDeadline，则默认都是阻塞读
func (f *FlvReader) SetReadTimeout(timeout uint32) {
    if _, ok := f.reader.(setReadDeadline); ok {
        f.hasDeadline = true
        f.timeout = timeout
    }
}

func (f *FlvReader) LoopRead() error {
    if err := f.readFileHeader(); err != nil {
        return nil
    }
    data := make([]byte, 4096)
    for {
        var tag [16]byte
        if _, err := readAtLeastWithTimeout(f.reader, tag[0:11], 11, f.timeout); err != nil {
            return err
        }

        var ftag FlvTag
        ftag.Decode(tag[:11])
        dts := uint32(ftag.TimestampExtended)<<24 | ftag.Timestamp
        pts := dts
        var taglen int = 0
        var acid FLV_SOUND_FORMAT
        var vcid FLV_VIDEO_CODEC_ID
        var packetType uint8 = 0
        if ftag.TagType == uint8(AUDIO_TAG) {
            atag, err := ReadAudioTagWithTimeout(f.reader, f.timeout)
            if err != nil {
                return err
            }
            acid = FLV_SOUND_FORMAT(atag.SoundFormat)
            taglen = GetTagLenByAudioCodec(acid)
            packetType = atag.AACPacketType
            if f.OnTag != nil {
                f.OnTag(ftag, atag)
            }
        } else if ftag.TagType == uint8(VIDEO_TAG) {
            vtag, err := ReadVideoTagWithTimeout(f.reader, f.timeout)
            if err != nil {
                return err
            }
            pts = dts + uint32(vtag.CompositionTime)
            packetType = vtag.AVCPacketType
            vcid = FLV_VIDEO_CODEC_ID(vtag.CodecId)
            taglen = GetTagLenByVideoCodec(vcid)
            if f.OnTag != nil {
                f.OnTag(ftag, vtag)
            }
        } else if ftag.TagType == uint8(SCRIPT_TAG) {
            if f.OnTag != nil {
                f.OnTag(ftag, nil)
            }
        }

        if cap(data) < int(ftag.DataSize)-taglen {
            data = make([]byte, ftag.DataSize)
        }
        data = data[:int(ftag.DataSize)-taglen]
        if _, err := readAtLeastWithTimeout(f.reader, data, len(data), f.timeout); err != nil {
            return err
        }

        if ftag.TagType == uint8(AUDIO_TAG) {
            f.demuxAudio(acid, packetType, data, pts, dts)
        } else if ftag.TagType == uint8(VIDEO_TAG) {
            f.demuxVideo(vcid, packetType, data, pts, dts)
        }

        if _, err := readAtLeastWithTimeout(f.reader, data[:4], 4, f.timeout); err != nil {
            return err
        }
    }
}

func (f *FlvReader) readFileHeader() error {

    var flvhdr [9]byte
    if _, err := readAtLeastWithTimeout(f.reader, flvhdr[0:9], 9, f.timeout); err != nil {
        return err
    }
    if flvhdr[0] != 'F' || flvhdr[1] != 'L' || flvhdr[2] != 'V' {
        return errors.New("this file Is Not FLV File")
    }

    if _, err := readAtLeastWithTimeout(f.reader, flvhdr[0:4], 4, f.timeout); err != nil {
        return err
    }

    return nil
}

func (f *FlvReader) demuxAudio(cid FLV_SOUND_FORMAT, packetType uint8, data []byte, pts uint32, dts uint32) {
    var audioFrame []byte
    if cid == FLV_AAC {
        if packetType == AAC_SEQUENCE_HEADER {
            if len(f.asc) < len(data) {
                panic("asc too large!!!")
            }
            copy(f.asc, data)
            f.asc = f.asc[:len(data)]
            return
        } else {
            audioFrame = mpeg.ConvertASCToADTS(f.asc, len(data)+7)
            audioFrame = append(audioFrame, data...)
        }
    } else {
        audioFrame = data
    }
    if f.OnFrame != nil {
        f.OnFrame(CovertFlvAudioCodecId2MpegCodecId(cid), audioFrame, pts, dts)
    }
}

func (f *FlvReader) demuxVideo(cid FLV_VIDEO_CODEC_ID, packetType uint8, data []byte, pts uint32, dts uint32) {
    if cid == FLV_AVC {
        if packetType == AVC_SEQUENCE_HEADER {
            tmpspss, tmpppss := mpeg.CovertExtradata(data)
            for _, sps := range tmpspss {
                spsid := mpeg.GetSPSId(sps)
                tmpsps := make([]byte, len(sps))
                copy(tmpsps, sps)
                f.spss[spsid] = tmpsps
            }
            for _, pps := range tmpppss {
                ppsid := mpeg.GetPPSId(pps)
                tmppps := make([]byte, len(pps))
                copy(tmppps, pps)
                f.spss[ppsid] = tmppps
            }
        } else {
            var hassps bool
            var haspps bool
            var idr bool
            tmpdata := data
            for len(tmpdata) > 0 {
                naluSize := binary.BigEndian.Uint32(tmpdata)
                mpeg.CovertAVCCToAnnexB(tmpdata)
                naluType := mpeg.H264NaluType(tmpdata)
                if naluType == 5 {
                    idr = true
                } else if naluType == 7 {
                    hassps = true
                } else if naluType == 8 {
                    haspps = true
                }
                tmpdata = tmpdata[4+naluSize:]
            }
            if idr && (!hassps || !haspps) {
                var nalus []byte = make([]byte, 0, 2048)
                for _, sps := range f.spss {
                    nalus = append(nalus, sps...)
                }
                for _, pps := range f.ppss {
                    nalus = append(nalus, pps...)
                }
                nalus = append(nalus, data...)
                if f.OnFrame != nil {
                    f.OnFrame(mpeg.CODECID_VIDEO_H264, nalus, pts, dts)
                }
            } else {
                if f.OnFrame != nil {
                    f.OnFrame(mpeg.CODECID_VIDEO_H264, data, pts, dts)
                }
            }
        }
    } else {
        //TODO
        panic("not implement")
    }
}

type FlvWriter struct {
    writer io.Writer
    muxer  *FlvMuxer
}

func CreateFlvWriter(writer io.Writer) *FlvWriter {
    flvFile := &FlvWriter{
        writer: writer,
        muxer:  new(FlvMuxer),
    }
    return flvFile
}

func (f *FlvWriter) WriteFlvHeader() (err error) {

    var flvhdr [9]byte
    flvhdr[0] = 'F'
    flvhdr[1] = 'L'
    flvhdr[2] = 'V'
    flvhdr[3] = 0x01
    flvhdr[4] = 0x05
    flvhdr[5] = 0
    flvhdr[6] = 0
    flvhdr[7] = 0
    flvhdr[8] = 9

    if _, err = f.writer.Write(flvhdr[:9]); err != nil {
        return
    }
    var previousTagSize0 [4]byte
    previousTagSize0[0] = 0
    previousTagSize0[1] = 0
    previousTagSize0[2] = 0
    previousTagSize0[3] = 0
    if _, err = f.writer.Write(previousTagSize0[:4]); err != nil {
        return
    }
    return
}

//adts aac frame
func (f *FlvWriter) WriteAAC(data []byte, pts uint32, dts uint32) error {
    if f.muxer.audioMuxer == nil {
        f.muxer.SetAudioCodeId(FLV_AAC)
    } else {
        if _, ok := f.muxer.audioMuxer.(*AACMuxer); !ok {
            panic("audio codec change")
        }
    }
    if tags, err := f.muxer.WriteAudio(data, pts, dts); err != nil {
        return err
    } else {
        for _, tag := range tags {
            if _, err := f.writer.Write(tag); err != nil {
                return err
            }
            if err := f.writePreviousTagSize(uint32(len(tag))); err != nil {
                return err
            }
        }
    }
    return nil
}

//H264 Frame with startcode 0x0000001
func (f *FlvWriter) WriteH264(data []byte, pts uint32, dts uint32) error {
    if f.muxer.videoMuxer == nil {
        f.muxer.SetVideoCodeId(FLV_AVC)
    } else {
        if _, ok := f.muxer.videoMuxer.(*AVCMuxer); !ok {
            panic("video codec change")
        }
    }
    if tags, err := f.muxer.WriteVideo(data, pts, dts); err != nil {
        return err
    } else {
        for _, tag := range tags {
            if _, err := f.writer.Write(tag); err != nil {
                return err
            }
            if err := f.writePreviousTagSize(uint32(len(tag))); err != nil {
                return err
            }
        }
    }
    return nil
}

func (f *FlvWriter) writePreviousTagSize(preTagSize uint32) error {
    tagsize := make([]byte, 4)
    binary.BigEndian.PutUint32(tagsize, preTagSize)
    if _, err := f.writer.Write(tagsize); err != nil {
        return err
    }
    return nil
}
