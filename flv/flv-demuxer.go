package flv

import (
    "encoding/binary"
    "errors"
    "fmt"

    "github.com/yapingcat/gomedia/mpeg"
)

type OnVideoFrameCallBack func(codecid mpeg.CodecID, frame []byte, cts int)
type VideoTagDemuxer interface {
    Decode(data []byte) error
    OnFrame(onframe OnVideoFrameCallBack)
}

type AVCTagDemuxer struct {
    spss    map[uint64][]byte
    ppss    map[uint64][]byte
    onframe OnVideoFrameCallBack
}

func newAVCTagDemuxer() *AVCTagDemuxer {
    return &AVCTagDemuxer{
        spss:    make(map[uint64][]byte),
        ppss:    make(map[uint64][]byte),
        onframe: nil,
    }
}

func (demuxer *AVCTagDemuxer) OnFrame(onframe OnVideoFrameCallBack) {
    demuxer.onframe = onframe
}

func (demuxer *AVCTagDemuxer) Decode(data []byte) error {

    if len(data) < 5 {
        return errors.New("avc tag size < 5")
    }

    vtag := VideoTag{}
    vtag.Decode(data[0:5])
    data = data[5:]
    if vtag.AVCPacketType == AVC_SEQUENCE_HEADER {
        tmpspss, tmpppss := mpeg.CovertExtradata(data)
        for _, sps := range tmpspss {
            spsid := mpeg.GetSPSId(sps)
            tmpsps := make([]byte, len(sps))
            copy(tmpsps, sps)
            demuxer.spss[spsid] = tmpsps
        }
        for _, pps := range tmpppss {
            ppsid := mpeg.GetPPSId(pps)
            tmppps := make([]byte, len(pps))
            copy(tmppps, pps)
            demuxer.spss[ppsid] = tmppps
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
            if naluType == mpeg.H264_NAL_I_SLICE {
                idr = true
            } else if naluType == mpeg.H264_NAL_SPS {
                hassps = true
            } else if naluType == mpeg.H264_NAL_PPS {
                haspps = true
            }
            tmpdata = tmpdata[4+naluSize:]
        }

        if idr && (!hassps || !haspps) {
            var nalus []byte = make([]byte, 0, 2048)
            for _, sps := range demuxer.spss {
                nalus = append(nalus, sps...)
            }
            for _, pps := range demuxer.ppss {
                nalus = append(nalus, pps...)
            }
            nalus = append(nalus, data...)
            if demuxer.onframe != nil {
                demuxer.onframe(mpeg.CODECID_VIDEO_H264, nalus, int(vtag.CompositionTime))
            }
        } else {
            if demuxer.onframe != nil {
                demuxer.onframe(mpeg.CODECID_VIDEO_H264, data, int(vtag.CompositionTime))
            }
        }
    }
    return nil
}

type HevcTagDemuxer struct {
    SpsPpsVps []byte
    onframe   OnVideoFrameCallBack
}

func newHevcTagDemuxer() *HevcTagDemuxer {
    return &HevcTagDemuxer{
        SpsPpsVps: make([]byte, 0),
        onframe:   nil,
    }
}

func (demuxer *HevcTagDemuxer) OnFrame(onframe OnVideoFrameCallBack) {
    demuxer.onframe = onframe
}

func (demuxer *HevcTagDemuxer) Decode(data []byte) error {

    if len(data) < 5 {
        return errors.New("hevc tag size < 5")
    }

    vtag := VideoTag{}
    vtag.Decode(data[0:5])
    fmt.Println(vtag.CodecId)
    fmt.Println(vtag.AVCPacketType)
    data = data[5:]
    if vtag.AVCPacketType == AVC_SEQUENCE_HEADER {
        hvcc := mpeg.NewHEVCRecordConfiguration()
        fmt.Printf("sequence %d\n", len(data))
        hvcc.Decode(data)
        demuxer.SpsPpsVps = hvcc.ToNalus()
    } else {
        var hassps bool
        var haspps bool
        var hasvps bool
        var idr bool
        tmpdata := data
        for len(tmpdata) > 0 {
            naluSize := binary.BigEndian.Uint32(tmpdata)
            mpeg.CovertAVCCToAnnexB(tmpdata)
            naluType := mpeg.H265NaluType(tmpdata)
            if naluType >= 16 && naluType <= 21 {
                idr = true
            } else if naluType == mpeg.H265_NAL_SPS {
                hassps = true
            } else if naluType == mpeg.H265_NAL_PPS {
                haspps = true
            } else if naluType == mpeg.H265_NAL_VPS {
                hasvps = true
            }
            tmpdata = tmpdata[4+naluSize:]
        }

        if idr && (!hassps || !haspps || !hasvps) {
            var nalus []byte = make([]byte, 0, 2048)
            nalus = append(demuxer.SpsPpsVps, data...)
            if demuxer.onframe != nil {
                demuxer.onframe(mpeg.CODECID_VIDEO_H265, nalus, int(vtag.CompositionTime))
            }
        } else {
            if demuxer.onframe != nil {
                demuxer.onframe(mpeg.CODECID_VIDEO_H265, data, int(vtag.CompositionTime))
            }
        }
    }
    return nil
}

type OnAudioFrameCallBack func(codecid mpeg.CodecID, frame []byte)

type AudioTagDemuxer interface {
    Decode(data []byte) error
    OnFrame(onframe OnAudioFrameCallBack)
}

type AACTagDemuxer struct {
    asc     []byte
    onframe OnAudioFrameCallBack
}

func newAACTagDemuxer() *AACTagDemuxer {
    return &AACTagDemuxer{
        asc:     make([]byte, 0, 2),
        onframe: nil,
    }
}

func (demuxer *AACTagDemuxer) OnFrame(onframe OnAudioFrameCallBack) {
    demuxer.onframe = onframe
}

func (demuxer *AACTagDemuxer) Decode(data []byte) error {

    if len(data) < 2 {
        return errors.New("aac tag size < 2")
    }

    atag := AudioTag{}
    atag.Decode(data[0:2])
    data = data[2:]
    if atag.AACPacketType == AAC_SEQUENCE_HEADER {
        demuxer.asc = data
    } else {
        adts := mpeg.ConvertASCToADTS(demuxer.asc, len(data)+7)
        adts = append(adts, data...)
        if demuxer.onframe != nil {
            demuxer.onframe(mpeg.CODECID_AUDIO_AAC, adts)
        }
    }
    return nil
}

type G711Demuxer struct {
    format  FLV_SOUND_FORMAT
    onframe OnAudioFrameCallBack
}

func newG711Demuxer(format FLV_SOUND_FORMAT) *G711Demuxer {
    return &G711Demuxer{
        format:  format,
        onframe: nil,
    }
}

func (demuxer *G711Demuxer) OnFrame(onframe OnAudioFrameCallBack) {
    demuxer.onframe = onframe
}

func (demuxer *G711Demuxer) Decode(data []byte) error {

    if len(data) < 1 {
        return errors.New("g711 tag size < 1")
    }

    atag := AudioTag{}
    atag.Decode(data[0:1])
    data = data[1:]

    if demuxer.onframe != nil {
        demuxer.onframe(demuxer.format.ToMpegCodecId(), data)
    }
    return nil
}
