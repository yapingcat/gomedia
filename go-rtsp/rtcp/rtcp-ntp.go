package rtcp

import "time"

func NTP2UtcClock(ntp uint64) time.Time {
    sec := ((ntp >> 32) - 0x83AA7E80) * 1000000
    us := ((ntp & 0xFFFFFFFF) * 15625) >> 26
    return time.Unix(int64(sec), int64(us))
}

func UtcClockToNTP(t time.Time) uint64 {
    ntp := (t.UnixNano()/1000000000 + 0x83AA7E80) << 32
    ntp = ntp | ((t.UnixNano() % 1000000000 / 1000) << 26 / 15625)
    return uint64(ntp)
}
