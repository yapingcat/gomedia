package rtsp

import (
    "errors"
    "fmt"
    "strconv"
    "strings"
    "time"
)

type RangeType int

const (
    RANGE_NPT RangeType = iota
    RANGE_UTC
)

type RangeTime struct {
    rangeType RangeType
    begin     int64 // -1 means now
    end       int64 // -1 meas has no end
}

func (rt RangeTime) EncodeString() string {
    switch rt.rangeType {
    case RANGE_NPT:
        npt := "npt="
        if rt.begin == -1 {
            npt += "now-"
        } else {
            npt += fmt.Sprintf("%d.%d-", rt.begin/1000, rt.begin%1000)
            if rt.end != -1 {
                npt += fmt.Sprintf("%d.%d", rt.end/1000, rt.end%1000)
            }
        }
        return npt
    case RANGE_UTC:
        clock := "clock="
        beg := time.Unix(rt.begin/1000, rt.begin%1000*1000000)
        clock += beg.Format("20060102T150405.999Z-")
        if rt.end != -1 {
            end := time.Unix(rt.end/1000, rt.end%1000*1000000)
            clock += end.Format("20060102T150405.999Z")
        }
        return clock
    default:
        panic("unsupport range time type")
    }
}

func parseNPT(npt string) int64 {
    if strings.Contains(npt, ":") {
        var h, m, s, mill int
        r, _ := fmt.Sscanf(npt, "%d:%d:%d.%d", &h, &m, &s, &mill)
        timeInMilliseconds := (h*3600 + m*60 + s) * 1000
        if r == 4 {
            timeInMilliseconds += mill
        }
        return int64(timeInMilliseconds)
    } else {
        t, _ := strconv.ParseFloat(npt, 32)
        return int64(t * 1000)
    }
}

func parseRange(str string) (*RangeTime, error) {
    strs := strings.Split(str, ";")
    rt := &RangeTime{}
    timestr := strings.Split(strs[0], "=")
    switch timestr[0] {
    case "npt":
        rt.rangeType = RANGE_NPT
        tp := strings.Split(timestr[1], "-")
        if tp[0] == "now" {
            rt.begin = -1
        }
        rt.begin = parseNPT(tp[0])
        rt.end = -1
        if len(tp) > 1 {
            rt.end = parseNPT(tp[1])
        }
        return rt, nil
    case "clock":
        rt.rangeType = RANGE_UTC
        layout := "20060102T150405Z"
        tp := strings.Split(timestr[1], "-")
        t, _ := time.Parse(layout, tp[0])
        rt.begin = t.UTC().UnixNano() / 1000000
        if len(tp) > 1 {
            t, _ = time.Parse(layout, tp[0])
            rt.end = t.UTC().UnixNano() / 1000000
        }
        return rt, nil
    default:
        return rt, errors.New("unsupport" + timestr[0] + " Range type")
    }
}
