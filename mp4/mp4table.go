package mp4

//based on ffmpeg

type sttsEntry struct {
    sampleCount uint32
    sampleDelta uint32
}

type movstts struct {
    entryCount uint32
    entrys     []sttsEntry
}

type cttsEntry struct {
    sampleCount  uint32
    sampleOffset uint32
}

type movctts struct {
    entryCount uint32
    entrys     []cttsEntry
}

type stscEntry struct {
    firstChunk             uint32
    samplesPerChunk        uint32
    sampleDescriptionIndex uint32
}

type movstsc struct {
    entryCount uint32
    entrys     []stscEntry
}

type movstsz struct {
    sampleSize    uint32
    sampleCount   uint32
    entrySizelist []uint32
}

type movstco struct {
    entryCount      uint32
    chunkOffsetlist []uint64
}

type movstbl struct {
    stts *movstts
    ctts *movctts
    stsc *movstsc
    stsz *movstsz
    stco *movstco
}
