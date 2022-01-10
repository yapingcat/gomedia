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

type elstEntry struct {
    segmentDuration   uint64
    mediaTime         int64
    mediaRateInteger  int16
    mediaRateFraction int16
}

type trunEntry struct {
    sampleDuration              uint32
    sampleSize                  uint32
    sampleFlags                 uint32
    sampleCompositionTimeOffset uint32
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

type movelst struct {
    entryCount uint32
    entrys     []elstEntry
}

type movtrun struct {
    entrys []trunEntry
}

type movstbl struct {
    stts *movstts
    ctts *movctts
    stsc *movstsc
    stsz *movstsz
    stco *movstco
}

type fragEntry struct {
    time       uint64
    moofOffset uint64
}

type movtfra struct {
    frags []fragEntry
}
