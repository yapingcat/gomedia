package mp4

import (
    "encoding/binary"
    "errors"
	"io"

    "github.com/yapingcat/gomedia/codec"
)

type sampleCache struct {
    pts    uint64
    dts    uint64
    hasVcl bool
    cache  []byte
}

type sampleEntry struct {
    pts                    uint64
    dts                    uint64
    offset                 uint64
    size                   uint64
	SampleDescriptionIndex uint32 //always should be 1
}

type movchunk struct {
    chunknum    uint32
    samplenum   uint32
    chunkoffset uint64
}

type mp4track struct {
    cid         MP4_CODEC_TYPE
    trackId     uint32
    stbltable   *movstbl
    duration    uint32
    timescale   uint32
    width       uint32
    height      uint32
    sampleRate  uint32
    sampleBits  uint8
    chanelCount uint8
    samplelist  []sampleEntry
    elst        *movelst
    extra       extraData
    lastSample  *sampleCache
}

func (track *mp4track) addSampleEntry(entry sampleEntry) {
    if len(track.samplelist) <= 1 {
        track.duration = 0
    } else {
        delta := int64(entry.dts - track.samplelist[len(track.samplelist)-1].dts)
        if delta < 0 {
            track.duration += 1
        } else {
            track.duration += uint32(delta)
        }
    }
    track.samplelist = append(track.samplelist, entry)
}

func newmp4track(cid MP4_CODEC_TYPE) *mp4track {
    track := &mp4track{
        cid:        cid,
        timescale:  1000,
        stbltable:  nil,
        samplelist: make([]sampleEntry, 0),
        lastSample: &sampleCache{
            hasVcl: false,
            cache:  make([]byte, 0, 128),
        },
    }
    if cid == MP4_CODEC_H264 {
        track.extra = new(h264ExtraData)
    } else if cid == MP4_CODEC_H265 {
        track.extra = newh265ExtraData()
    } else if cid == MP4_CODEC_AAC {
        track.extra = new(aacExtraData)
    }
    return track
}

func (track *mp4track) makeStblTable() {
    if track.stbltable == nil {
        track.stbltable = new(movstbl)
    }
    sameSize := true
    stts := new(movstts)
    stts.entrys = make([]sttsEntry, 0)
    movchunks := make([]movchunk, 0)
    ctts := new(movctts)
    ctts.entrys = make([]cttsEntry, 0)
    ckn := uint32(0)
    for i, sample := range track.samplelist {
        sttsEntry := sttsEntry{sampleCount: 1, sampleDelta: 0}
        cttsEntry := cttsEntry{sampleCount: 1, sampleOffset: uint32(sample.pts) - uint32(sample.dts)}
        if i == len(track.samplelist)-1 {
            stts.entrys = append(stts.entrys, sttsEntry)
            stts.entryCount++
        } else {
            delta := track.samplelist[i+1].dts - sample.dts
            if len(stts.entrys) > 0 && delta == uint64(stts.entrys[len(stts.entrys)-1].sampleDelta) {
                stts.entrys[len(stts.entrys)-1].sampleCount++
            } else {
                sttsEntry.sampleDelta = uint32(delta)
                stts.entrys = append(stts.entrys, sttsEntry)
                stts.entryCount++
            }
        }

        if len(ctts.entrys) == 0 {
            ctts.entrys = append(ctts.entrys, cttsEntry)
        } else {
            if ctts.entrys[len(ctts.entrys)-1].sampleOffset == cttsEntry.sampleOffset {
                ctts.entrys[len(ctts.entrys)-1].sampleCount++
            } else {
                ctts.entrys = append(ctts.entrys, cttsEntry)
                ctts.entryCount++
            }
        }

        if sameSize && track.samplelist[i+1].size != track.samplelist[i].size {
            sameSize = false
        }
        if i > 0 && sample.offset == track.samplelist[i-1].offset+track.samplelist[i-1].size {
            movchunks[ckn-1].samplenum++
        } else {
            ck := movchunk{chunknum: ckn, samplenum: 1, chunkoffset: sample.offset}
            movchunks = append(movchunks, ck)
            ckn++
        }
    }
    stsz := &movstsz{
        sampleSize:  0,
        sampleCount: uint32(len(track.samplelist)),
    }
    if sameSize {
        stsz.sampleSize = uint32(track.samplelist[0].size)
    } else {
        stsz.entrySizelist = make([]uint32, stsz.sampleCount)
        for i := 0; i < len(stsz.entrySizelist); i++ {
            stsz.entrySizelist[i] = uint32(track.samplelist[i].size)
        }
    }

    stsc := &movstsc{
        entrys:     make([]stscEntry, len(movchunks)),
        entryCount: 0,
    }
    for i, chunk := range movchunks {
        if i == 0 || chunk.samplenum != movchunks[i-1].samplenum {
            stsc.entrys[stsc.entryCount].firstChunk = chunk.chunknum + 1
            stsc.entrys[stsc.entryCount].sampleDescriptionIndex = 1
            stsc.entrys[stsc.entryCount].samplesPerChunk = chunk.samplenum
            stsc.entryCount++
        }
    }
    stco := &movstco{entryCount: ckn, chunkOffsetlist: make([]uint64, ckn)}
    for i := 0; i < int(stco.entryCount); i++ {
        stco.chunkOffsetlist[i] = movchunks[i].chunkoffset
    }
    track.stbltable.stts = stts
    track.stbltable.stsc = stsc
    track.stbltable.stco = stco
    track.stbltable.stsz = stsz
    if track.cid == MP4_CODEC_H264 || track.cid == MP4_CODEC_H265 {
        track.stbltable.ctts = ctts
    }
}

type extraData interface {
    export() []byte
    load(data []byte)
}

type h264ExtraData struct {
    spss [][]byte
    ppss [][]byte
}

func (extra *h264ExtraData) export() []byte {
    return codec.CreateH264AVCCExtradata(extra.spss, extra.ppss)
}

func (extra *h264ExtraData) load(data []byte) {
    extra.spss, extra.ppss = codec.CovertExtradata(data)
}

type h265ExtraData struct {
    hvccExtra *codec.HEVCRecordConfiguration
}

func newh265ExtraData() *h265ExtraData {
    return &h265ExtraData{
        hvccExtra: codec.NewHEVCRecordConfiguration(),
    }
}

func (extra *h265ExtraData) export() []byte {
    if extra.hvccExtra != nil {
        return extra.hvccExtra.Encode()
    }
    panic("extra.hvccExtra must init")
}

func (extra *h265ExtraData) load(data []byte) {
    if extra.hvccExtra != nil {
        extra.hvccExtra.Decode(data)
    }
    panic("extra.hvccExtra must init")
}

type aacExtraData struct {
    asc []byte
}

func (extra *aacExtraData) export() []byte {
    return extra.asc
}

func (extra *aacExtraData) load(data []byte) {
    extra.asc = make([]byte, len(data))
    copy(extra.asc, data)
}

type Movmuxer struct {
	writer        io.WriteSeeker
    nextTrackId   uint32
    mdatOffset    uint32
    tracks        map[uint32]*mp4track
    duration      uint32
    width         uint32
    height        uint32
}

func CreateMp4Muxer(w io.WriteSeeker) *Movmuxer {
    muxer := &Movmuxer{
		writer:        w,
        nextTrackId:   1,
        tracks:        make(map[uint32]*mp4track),
    }
    ftyp := NewFileTypeBox()
    ftyp.Major_brand = mov_tag(isom)
    ftyp.Minor_version = 0x200
    ftyp.Compatible_brands = make([]uint32, 4)
    ftyp.Compatible_brands[0] = mov_tag(isom)
    ftyp.Compatible_brands[1] = mov_tag(iso2)
    ftyp.Compatible_brands[2] = mov_tag(avc1)
    ftyp.Compatible_brands[3] = mov_tag(mp41)
    len, boxdata := ftyp.Encode()
	muxer.writer.Write(boxdata[0:len])
    free := NewFreeBox()
    freelen, freeboxdata := free.Encode()
    muxer.writer.Write(freeboxdata[0:freelen])
    currentOffset, _ := muxer.writer.Seek(0, io.SeekCurrent)
    muxer.mdatOffset = uint32(currentOffset)
    mdat := BasicBox{Type: [4]byte{'m', 'd', 'a', 't'}}
    mdat.Size = 8
    mdatlen, mdatBox := mdat.Encode()
    muxer.writer.Write(mdatBox[0:mdatlen])
    return muxer
}

func (muxer *Movmuxer) AddAudioTrack(cid MP4_CODEC_TYPE, channelcount uint8, sampleBits uint8, sampleRate uint) uint32 {
    track := newmp4track(cid)
    track.trackId = muxer.nextTrackId
    track.sampleRate = uint32(sampleRate)
    track.sampleBits = sampleBits
    track.chanelCount = channelcount
    muxer.tracks[muxer.nextTrackId] = track
    muxer.nextTrackId++
    return track.trackId
}

func (muxer *Movmuxer) AddVideoTrack(cid MP4_CODEC_TYPE) uint32 {
    track := newmp4track(cid)
    track.trackId = muxer.nextTrackId
    muxer.tracks[muxer.nextTrackId] = track
    muxer.nextTrackId++
    return track.trackId
}

func (muxer *Movmuxer) Write(track uint32, data []byte, pts uint64, dts uint64) error {
    mp4track := muxer.tracks[track]

    var err error
    if mp4track.cid == MP4_CODEC_H264 {
        err = muxer.writeH264(mp4track, data, pts, dts)
    } else if mp4track.cid == MP4_CODEC_H265 {
        err = muxer.writeH265(mp4track, data, pts, dts)
    } else if mp4track.cid == MP4_CODEC_AAC {
        err = muxer.writeAAC(mp4track, data, pts, dts)
    } else if mp4track.cid == MP4_CODEC_G711A || mp4track.cid == MP4_CODEC_G711U {
        err = muxer.writeG711(mp4track, data, pts, dts)
    } else {
        return errors.New("UnSupport Codec")
    }
    if err != nil {
        return err
    }
    return nil
}

func (muxer *Movmuxer) WriteTrailer() (err error) {

    var currentOffset int64
    for _, track := range muxer.tracks {
        if track.lastSample != nil && len(track.lastSample.cache) > 0 {
            if currentOffset, err = muxer.writer.Seek(0, io.SeekCurrent); err != nil {
                return err
            }
            entry := sampleEntry{
                pts:                    track.lastSample.pts,
                dts:                    track.lastSample.dts,
                size:                   0,
                SampleDescriptionIndex: 1,
				offset:                 uint64(currentOffset),
            }
            n := 0
			if n, err = muxer.writer.Write(track.lastSample.cache); err != nil {
                return err
            }
            entry.size = uint64(n)
            track.addSampleEntry(entry)
        }
    }

    if currentOffset, err = muxer.writer.Seek(0, io.SeekCurrent); err != nil {
        return err
    }
	datalen := currentOffset - int64(muxer.mdatOffset)
    if datalen > 0xFFFFFFFF {
        mdat := BasicBox{Type: [4]byte{'m', 'd', 'a', 't'}}
        mdat.Size = uint64(datalen)
        mdatBoxLen, mdatBox := mdat.Encode()
        if _, err = muxer.writer.Seek(int64(muxer.mdatOffset)-8, io.SeekStart); err != nil {
            return
        }
        if _, err = muxer.writer.Write(mdatBox[0:mdatBoxLen]); err != nil {
            return
        }
		if _, err = muxer.writer.Seek(currentOffset, io.SeekStart); err != nil {
            return
        }
    } else {
		if _, err = muxer.writer.Seek(int64(muxer.mdatOffset), io.SeekStart); err != nil {
            return
        }
        tmpdata := make([]byte, 4)
        binary.BigEndian.PutUint32(tmpdata, uint32(datalen))
		if _, err = muxer.writer.Write(tmpdata); err != nil {
            return
        }
		if _, err = muxer.writer.Seek(currentOffset, io.SeekStart); err != nil {
            return
        }
    }

    maxdurtaion := uint32(0)
    for _, track := range muxer.tracks {
        if maxdurtaion < track.duration {
            maxdurtaion = track.duration
        }
    }

    muxer.duration = maxdurtaion
    mvhd := makeMvhdBox(muxer.nextTrackId, muxer.duration)
    moovsize := len(mvhd)
    traks := make([][]byte, len(muxer.tracks))
    idx := 0
    for _, track := range muxer.tracks {
        traks[idx] = makeTrak(track)
        moovsize += len(traks[idx])
        idx++
    }
    moov := BasicBox{Type: [4]byte{'m', 'o', 'o', 'v'}}
    moov.Size = 8 + uint64(moovsize)
    offset, moovBox := moov.Encode()
    copy(moovBox[offset:], mvhd)
    offset += len(mvhd)
    for _, trak := range traks {
        copy(moovBox[offset:], trak)
        offset += len(trak)
    }
    if _, err = muxer.writer.Write(moovBox); err != nil {
        return
    }
    return
}

func (muxer *Movmuxer) writeH264(track *mp4track, h264 []byte, pts, dts uint64) (err error) {
    h264extra, ok := track.extra.(*h264ExtraData)
    if !ok {
        panic("must init h264ExtraData first")
    }
    codec.SplitFrameWithStartCode(h264, func(nalu []byte) bool {
        nalu_type := codec.H264NaluType(nalu)
        switch nalu_type {
        case codec.H264_NAL_SPS:
            spsid := codec.GetSPSIdWithStartCode(nalu)
            for _, sps := range h264extra.spss {
                if spsid == codec.GetSPSIdWithStartCode(sps) {
                    return true
                }
            }
            tmp := make([]byte, len(nalu))
            copy(tmp, nalu)
            h264extra.spss = append(h264extra.spss, tmp)
            if muxer.width == 0 || muxer.height == 0 {
                muxer.width, muxer.height = codec.GetH264Resolution(h264extra.spss[0])
                track.width = muxer.width
                track.height = muxer.height
            }
        case codec.H264_NAL_PPS:
            ppsid := codec.GetPPSIdWithStartCode(nalu)
            for _, pps := range h264extra.ppss {
                if ppsid == codec.GetPPSIdWithStartCode(pps) {
                    return true
                }
            }
            tmp := make([]byte, len(nalu))
            copy(tmp, nalu)
            h264extra.ppss = append(h264extra.ppss, tmp)
        }
        //aud/sps/pps/sei 为帧间隔
        //通过first_slice_in_mb来判断，改nalu是否为一帧的开头
        if track.lastSample.hasVcl && isH264NewAccessUnit(nalu) {
            var currentOffset int64
            if currentOffset, err = muxer.writer.Seek(0, io.SeekCurrent); err != nil {
                return false
            }
            entry := sampleEntry{
                pts:                    track.lastSample.pts,
                dts:                    track.lastSample.dts,
                size:                   0,
                SampleDescriptionIndex: 1,
				offset:                 uint64(currentOffset),
            }
            n := 0
			if n, err = muxer.writer.Write(track.lastSample.cache); err != nil {
                return false
            }
            entry.size = uint64(n)
            track.addSampleEntry(entry)
            track.lastSample.cache = track.lastSample.cache[:0]
            track.lastSample.hasVcl = false
        }
        if codec.IsH264VCLNaluType(nalu_type) {
            track.lastSample.pts = pts
            track.lastSample.dts = dts
            track.lastSample.hasVcl = true
        }
        track.lastSample.cache = append(track.lastSample.cache, codec.ConvertAnnexBToAVCC(nalu)...)
        return true
    })
    return
}

func (muxer *Movmuxer) writeH265(track *mp4track, h265 []byte, pts, dts uint64) (err error) {
    h265extra, ok := track.extra.(*h265ExtraData)
    if !ok {
        panic("must init h265ExtraData first")
    }
    codec.SplitFrameWithStartCode(h265, func(nalu []byte) bool {
        nalu_type := codec.H265NaluType(nalu)
        switch nalu_type {
        case codec.H265_NAL_SPS:
            h265extra.hvccExtra.UpdateSPS(nalu)
            if muxer.width == 0 || muxer.height == 0 {
                muxer.width, muxer.height = codec.GetH265Resolution(nalu)
                track.width = muxer.width
                track.height = muxer.height
            }
        case codec.H265_NAL_PPS:
            h265extra.hvccExtra.UpdatePPS(nalu)
        case codec.H265_NAL_VPS:
            h265extra.hvccExtra.UpdateVPS(nalu)
        }

        if track.lastSample.hasVcl && isH265NewAccessUnit(nalu) {
            var currentOffset int64
			if currentOffset, err = muxer.writer.Seek(0, io.SeekCurrent); err != nil {
                return false
            }
            entry := sampleEntry{
                pts:                    track.lastSample.pts,
                dts:                    track.lastSample.dts,
                size:                   0,
                SampleDescriptionIndex: 1,
				offset:                 uint64(currentOffset),
            }
			n := 0
			if n, err = muxer.writer.Write(track.lastSample.cache); err != nil {
                return false
            }
            entry.size = uint64(n)
            track.addSampleEntry(entry)
            track.lastSample.cache = track.lastSample.cache[:0]
            track.lastSample.hasVcl = false
        }
        if codec.IsH265VCLNaluType(nalu_type) {
            track.lastSample.pts = pts
            track.lastSample.dts = dts
            track.lastSample.hasVcl = true
        }
        track.lastSample.cache = append(track.lastSample.cache, codec.ConvertAnnexBToAVCC(nalu)...)
        return true
    })
    return
}

func (muxer *Movmuxer) writeAAC(track *mp4track, aacframes []byte, pts, dts uint64) (err error) {
    aacextra, ok := track.extra.(*aacExtraData)
    if !ok {
        panic("must init aacExtraData first")
    }
    if aacextra.asc == nil || len(aacextra.asc) <= 0 {

        asc, err := codec.ConvertADTSToASC(aacframes)
        if err != nil {
            return err
        }
        aacextra.asc = make([]byte, len(asc))
        copy(aacextra.asc, asc)
    }

    //某些情况下，aacframes 可能由多个aac帧组成需要分帧，否则quicktime 貌似播放有问题
    codec.SplitAACFrame(aacframes, func(aac []byte) {
        var currentOffset int64
        if currentOffset, err = muxer.writer.Seek(0, io.SeekCurrent); err != nil {
            return
        }
        entry := sampleEntry{
            pts:                    pts,
            dts:                    dts,
            size:                   0,
            SampleDescriptionIndex: 1,
			offset:                 uint64(currentOffset),
        }
        n := 0
		n, err = muxer.writer.Write(aac[7:])
        entry.size = uint64(n)
        track.addSampleEntry(entry)
    })

    return
}

func (muxer *Movmuxer) writeG711(track *mp4track, g711 []byte, pts, dts uint64) (err error) {
    var currentOffset int64
    if currentOffset, err = muxer.writer.Seek(0, io.SeekCurrent); err != nil {
        return
    }
    entry := sampleEntry{
        pts:                    pts,
        dts:                    dts,
        size:                   0,
        SampleDescriptionIndex: 1,
        offset:                 uint64(currentOffset),
    }
    n := 0
    n, err = muxer.writer.Write(g711)
    entry.size = uint64(n)
    track.addSampleEntry(entry)
    return
}
