package sdp

import (
	"errors"
	"strconv"
	"strings"
)

//c=<nettype> <addrtype> <connection-address>
//c=IN IP4 224.2.36.42/127
type Connection struct {
    Nettype  string
    Addrtype string
    Address  string
}

func (c *Connection) Decode(connectionData string) error {
    items := strings.Split(connectionData, " ")
    if len(items) < 3 {
        return errors.New("parser \"c=\" field failed")
    }
    c.Nettype = items[0]
    c.Addrtype = items[1]
    c.Address = items[2]
    return nil
}

type RtpMap struct {
    PayloadType int
    EncodeName  string
    ClockRate   int
    EncodParam  string
}

func (r *RtpMap) Decode(rtpmap string) error {
    items := strings.SplitN(rtpmap, " ", 2)
    r.PayloadType, _ = strconv.Atoi(items[0])
    if len(items) == 1 {
        return nil
    }
    param := strings.Split(items[1], "/")
    r.EncodeName = param[0]
    r.ClockRate, _ = strconv.Atoi(param[1])
    if len(param) > 2 {
        r.EncodParam = param[2]
    }
    return nil
}

type Media struct {
    MediaType    string
    Ports        []uint16
    Proto        string
    Fmts         []uint8
    PayloadType  int
    EncodeName   string
    ClockRate    int
    ChannelCount int
    ControlUrl   string
    Attrs        map[string]string
}

func (m *Media) Encode() string {
    mediaTxt := "m=" + m.MediaType + " " + strconv.Itoa(int(m.Ports[0]))
    if len(m.Ports) > 1 {
        mediaTxt += "/" + strconv.Itoa(len(m.Ports))
    }
    mediaTxt += " "
    mediaTxt += m.Proto
    for _, pt := range m.Fmts {
        mediaTxt += " " + strconv.Itoa(int(pt))
    }
    mediaTxt += "\r\n"

    for attrKey, attrValue := range m.Attrs {
        mediaTxt += "a=" + attrKey
        if len(attrValue) > 0 {
            mediaTxt += ":" + attrValue
        }
        mediaTxt += "\r\n"
    }
    return mediaTxt
}

// func (m *Media) Decode(mediaDes string) error {
// 	lines := strings.FieldsFunc(mediaDes, func(r rune) bool {
// 		if r == '\r' || r == '\n' {
// 			return true
// 		} else {
// 			return false
// 		}
// 	})
// 	m.ParseMLine(string(lines[0]))
// 	for _, line := range lines[1:] {
// 		nameValue := strings.SplitN(line, "=", 2)
// 		if len(nameValue) < 2 {
// 			return errors.New("parser sdp line failed")
// 		}
// 		name := nameValue[0]
// 		value := nameValue[1]
// 		if name != "a" {
// 			continue
// 		}

// 	}

// }

func (m *Media) ParseMLine(mediaLine string) error {
    strs := strings.SplitN(mediaLine, " ", 4)
    m.MediaType = strs[0]
    pn := strings.SplitN(strs[1], "/", 2)
    p, _ := strconv.Atoi(pn[0])
    m.Ports = append(m.Ports, uint16(p))
    if len(pn) > 1 {
        numberOfPort, _ := strconv.Atoi(pn[1])
        for i := 1; i < numberOfPort; i++ {
            m.Ports = append(m.Ports, uint16(p)+1)
        }
    }
    m.Proto = strs[2]
    fmts := strings.Split(strs[3], " ")
    for _, fmt := range fmts {
        f, _ := strconv.Atoi(fmt)
        m.Fmts = append(m.Fmts, uint8(f))
    }
    return nil
}

type Sdp struct {
    SessionName    string
    SessionInfo    string
    ControlUrl     string
    ConnectionData Connection
    Attrs          map[string]string
    Medias         []*Media
}

func (sdp *Sdp) Encode() string {
    sdptxt := "v=0\r\n"
    sdptxt += "o=- 0 0 IN IP4 0.0.0.0\r\n"
    sdptxt += "s=gomedia rtsp\r\n"
    sdptxt += "c=IN IP4 \r\n"
    sdptxt += "t=0 0\r\n"
    for attrName, attrValue := range sdp.Attrs {
        sdptxt += "a=" + attrName
        if len(attrValue) > 0 {
            sdptxt += ":" + attrValue
        }
        sdptxt += "\r\n"
    }

    for _, m := range sdp.Medias {
        sdptxt += m.Encode()
    }
    return sdptxt
}

func (sdp *Sdp) ParserSdp(sdpContent string) error {
    lines := strings.FieldsFunc(sdpContent, func(r rune) bool {
        if r == '\r' || r == '\n' {
            return true
        } else {
            return false
        }
    })
    for _, line := range lines {
        nameValue := strings.SplitN(line, "=", 2)
        if len(nameValue) < 2 {
            return errors.New("parser sdp line failed")
        }
        name := nameValue[0]
        value := nameValue[1]
        switch name[0] {
        case 's':
            sdp.SessionName = string(value)
        case 'i':
            sdp.SessionInfo = string(value)
        case 'c':
            if err := sdp.ConnectionData.Decode(string(value)); err != nil {
                return err
            }
        case 'a':
            attribute := strings.SplitN(value, ":", 2)
            var attrName string = string(attribute[0])
            var attrValue string = ""
            if len(attribute) > 1 {
                attrValue = string(attribute[1])
            }
            if len(sdp.Medias) == 0 {
                if sdp.Attrs == nil {
                    sdp.Attrs = make(map[string]string)
                }
                sdp.Attrs[attrName] = attrValue
            } else {
                if sdp.Medias[len(sdp.Medias)-1].Attrs == nil {
                    sdp.Medias[len(sdp.Medias)-1].Attrs = make(map[string]string)
                }
                sdp.Medias[len(sdp.Medias)-1].Attrs[attrName] = attrValue
            }
            switch attrName {
            case "rtpmap":
                rtpMap := &RtpMap{}
                rtpMap.Decode(attrValue)
                sdp.Medias[len(sdp.Medias)-1].EncodeName = rtpMap.EncodeName
                sdp.Medias[len(sdp.Medias)-1].ClockRate = rtpMap.ClockRate
                if len(sdp.Medias[len(sdp.Medias)-1].Fmts) > 0 &&
                    sdp.Medias[len(sdp.Medias)-1].Fmts[0] == uint8(rtpMap.PayloadType) {
                    sdp.Medias[len(sdp.Medias)-1].PayloadType = rtpMap.PayloadType
                }
                if rtpMap.EncodParam != "" && sdp.Medias[len(sdp.Medias)-1].MediaType == "audio" {
                    sdp.Medias[len(sdp.Medias)-1].ChannelCount, _ = strconv.Atoi(rtpMap.EncodParam)
                }
            case "control":
                if len(sdp.Medias) == 0 {
                    sdp.ControlUrl = attrValue
                } else {
                    sdp.Medias[len(sdp.Medias)-1].ControlUrl = attrValue
                }
            }
        case 'm':
            m := &Media{}
            if err := m.ParseMLine(string(value)); err != nil {
                return err
            }
            sdp.Medias = append(sdp.Medias, m)
        }
    }

    //https://datatracker.ietf.org/doc/html/rfc3551
    for i := 0; i < len(sdp.Medias); i++ {
        if _, found := sdp.Medias[i].Attrs["rtpmap"]; !found {
            if len(sdp.Medias[i].Fmts) == 0 || sdp.Medias[i].Fmts[0] >= 96 {
                continue
            }
            switch sdp.Medias[i].Fmts[0] {
            case 0:
                sdp.Medias[i].PayloadType = 0
                sdp.Medias[i].EncodeName = "PCMU"
                sdp.Medias[i].ClockRate = 8000
                sdp.Medias[i].ChannelCount = 1
            case 8:
                sdp.Medias[i].PayloadType = 8
                sdp.Medias[i].EncodeName = "PCMA"
                sdp.Medias[i].ClockRate = 8000
                sdp.Medias[i].ChannelCount = 1
            case 26:
                sdp.Medias[i].PayloadType = 26
                sdp.Medias[i].EncodeName = "JPEG"
                sdp.Medias[i].ClockRate = 90000
            case 33:
                sdp.Medias[i].PayloadType = 33
                sdp.Medias[i].EncodeName = "MP2T"
                sdp.Medias[i].ClockRate = 90000
            }
        }
    }
    return nil
}
