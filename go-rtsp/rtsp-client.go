package rtsp

import (
    "bytes"
    "encoding/binary"
    "errors"
    "fmt"
    "net/url"
    "strings"
    "time"

    "github.com/yapingcat/gomedia/go-rtsp/sdp"
)

// The client can assume the following states:

// Init:
//        SETUP has been sent, waiting for reply.

// Ready:
//        SETUP reply received or PAUSE reply received while in Playing
//        state.

// Playing:
//        PLAY reply received

// Recording:
//        RECORD reply received

// state       message sent     next state after response
// Init        SETUP            Ready
//             TEARDOWN         Init
// Ready       PLAY             Playing
//             RECORD           Recording
//             TEARDOWN         Init
//             SETUP            Ready
// Playing     PAUSE            Ready
//             TEARDOWN         Init
//             PLAY             Playing
//             SETUP            Playing (changed transport)
// Recording   PAUSE            Ready
//             TEARDOWN         Init
//             RECORD           Recording
//             SETUP            Recording (changed transport)

const (
    STATE_Init = iota
    STATE_Ready
    STATE_Playing
    STATE_Recording
)

type OutPutCallBack func([]byte) error
type TrackCallBack func(track *RtspTrack)
type RtspClient struct {
    uri              string
    usrName          string
    passwd           string
    isRecord         bool
    cseq             int
    auth             authenticate
    output           OutPutCallBack
    onTrack          TrackCallBack
    lastRequest      *RtspRequest
    tracks           map[string]*RtspTrack
    cache            []byte
    reponseHandler   func(res *RtspResponse) error
    serverCapability []string
    state            int
    sdpContext       *sdp.Sdp
    setupStep        int
}

type ClientOption func(cli *RtspClient)

func WithEnableRecord() ClientOption {
    return func(cli *RtspClient) {
        cli.isRecord = true
    }
}

func NewRtspClient(uri string, opt ...ClientOption) (*RtspClient, error) {
    cli := &RtspClient{
        cseq:             1,
        state:            STATE_Init,
        serverCapability: []string{OPTIONS, DESCRIBE, SETUP, PLAY, TEARDOWN, ANNOUNCE, RECORD, PAUSE, SET_PARAMETER, GET_PARAMETER, REDIRECT},
        setupStep:        0,
    }
    u, err := url.Parse(uri)
    if err != nil {
        return nil, err
    }
    if u.User != nil {
        cli.usrName = u.User.Username()
        if _, ok := u.User.Password(); ok {
            cli.passwd, _ = u.User.Password()
        }
    }
    u.User = nil
    cli.uri = u.String()
    return cli, nil
}

func (client *RtspClient) AddTrack(track *RtspTrack) {
    client.tracks[track.TrackName] = track
}

func (client *RtspClient) OnTrack(onTrack func(track *RtspTrack)) {
    client.onTrack = onTrack
}

func (client *RtspClient) SetOutput(output OutPutCallBack) {
    client.output = output
}

func (client *RtspClient) Start() error {
    req := makeOptions(client.uri, client.cseq)
    client.reponseHandler = client.handleOption
    return client.sendRtspRequest(&req)
}

func (client *RtspClient) Stop() {

}

func (client *RtspClient) Input(data []byte) (err error) {
    buf := client.cache
    if len(buf) > 0 {
        buf = append(buf, data...)
    } else {
        buf = data
    }

    for len(buf) > 0 {
        ret := 0
        if buf[0] == '$' {
            ret, err = client.handlRtpOverRtsp(buf)
        } else {
            ret, err = client.handleRtspMessage(buf)
        }
        if err != nil {
            break
        }
        buf = buf[ret:]
    }

    if len(buf) == 0 || (err != nil && !errors.Is(err, errNeedMore)) {
        return
    }

    if cap(client.cache) >= len(buf) {
        client.cache = client.cache[:len(buf)]
    } else {
        client.cache = make([]byte, len(buf))
    }
    copy(client.cache, buf)
    return nil
}

func (client *RtspClient) sendRtspRequest(req *RtspRequest) error {
    client.lastRequest = req
    client.cseq++
    if client.auth != nil {
        req.Fileds[Authorization] = client.auth.authenticateInfo()
    }
    return client.sendToServer([]byte(req.Encode()))
}

func (client *RtspClient) sendToServer(data []byte) error {
    if client.output != nil {
        return client.output(data)
    }
    return nil
}

func (client *RtspClient) handlRtpOverRtsp(packet []byte) (ret int, err error) {
    if len(packet) < 4 {
        return 0, errNeedMore
    }
    channel := packet[1]
    length := binary.BigEndian.Uint16(packet[2:])
    if len(packet)-4 < int(length) {
        return 0, errNeedMore
    }

    for _, track := range client.tracks {
        if track.transport.interleaved[0] == int(channel) ||
            track.transport.interleaved[1] == int(channel) {
            return 4 + int(length), track.input(packet[4 : 4+length])
        }
    }
    //improve compatibility
    return 4 + int(length), nil
}

func (client *RtspClient) handleRtspMessage(msg []byte) (int, error) {

    idx := bytes.IndexFunc(msg, func(r rune) bool {
        if r == ' ' {
            return false
        } else {
            return true
        }
    })

    msg = msg[idx:]
    if bytes.HasPrefix(msg, []byte{'R', 'T', 'S', 'P'}) {
        return client.handleResponse(msg)
    } else {
        return client.handleRequest(msg)
    }
}

func (client *RtspClient) handleResponse(res []byte) (ret int, err error) {
    response := RtspResponse{}
    ret, err = response.parse(string(res))
    if err != nil {
        return
    }
    if response.StatusCode == 401 {
        return ret, client.handleUnAuth(response)
    }
    return ret, client.reponseHandler(&response)
}

func (client *RtspClient) handleRequest(req []byte) (int, error) {

}

func (client *RtspClient) handleUnAuth(response RtspResponse) error {

    if _, found := response.Fileds[WWWAuthenticate]; !found {
        return errors.New("need WWW-Authenticate")
    }

    if client.auth == nil {
        client.auth = createAuthByAuthenticate(response.Fileds[WWWAuthenticate])
        client.auth.setUserInfo(client.usrName, client.passwd)
    }
    client.auth.setMethod(client.lastRequest.Method)
    client.auth.setUri(client.lastRequest.Uri)
    client.auth.decode(response.Fileds[WWWAuthenticate])
    client.lastRequest.Fileds.Add(CSeq, client.cseq)
    client.lastRequest.Fileds[Date] = time.Now().UTC().Format("02 Jan 06 15:04:05 GMT")
    client.lastRequest.Fileds[Authorization] = client.auth.authenticateInfo()
    client.cseq++
    return client.sendToServer([]byte(client.lastRequest.Encode()))
}

func (client *RtspClient) handleOption(res *RtspResponse) error {
    if client.state == STATE_Init {
        if res.Fileds.Has(Public) {
            client.serverCapability = strings.Split(res.Fileds[Public], ",")
            for i := 0; i < len(client.serverCapability); i++ {
                client.serverCapability[i] = strings.TrimSpace(client.serverCapability[i])
            }
        }
    }

    if client.isRecord {
        if !hasRecordAbility(client.serverCapability) {
            return errors.New(fmt.Sprintf("server capability:%s ,unsupport Record mode ", res.Fileds[Public]))
        }

        return nil
    } else {
        if !hasPlayAbility(client.serverCapability) {
            return errors.New(fmt.Sprintf("server capability:%s ,unsupport Play mode ", res.Fileds[Public]))
        }
        describe := makeDescribe(client.uri, client.cseq)
        client.reponseHandler = client.handleDescribe
        return client.sendRtspRequest(&describe)
    }
}

// 1.The RTSP Content-Base field
// 2.The RTSP Content-Location field
// 3.The RTSP request URL
func (client *RtspClient) handleDescribe(res *RtspResponse) (err error) {
    client.sdpContext, err = sdp.ParserSdp([]byte(res.Body))
    if err != nil {
        return err
    }

    baseUrl := client.uri
    if res.Fileds.Has(ContentBase) {
        baseUrl = res.Fileds[ContentBase]
    } else if res.Fileds.Has(ContentLocation) {
        baseUrl = res.Fileds[ContentLocation]
    }
    if !strings.HasSuffix(baseUrl, "/") {
        baseUrl += "/"
    }

    getControlUrl := func(url string) string {
        if url == "*" {
            return baseUrl
        } else if !strings.HasPrefix(url, "rtsp://") {
            if strings.HasPrefix(url, "/") {
                return baseUrl + client.sdpContext.ControlUrl[1:]
            } else {
                return baseUrl + client.sdpContext.ControlUrl
            }
        } else {
            return url
        }
    }

    client.sdpContext.ControlUrl = getControlUrl(client.sdpContext.ControlUrl)

    for _, media := range client.sdpContext.Medias {
        var track *RtspTrack = nil
        if media.MediaType == "audio" {
            track = NewTrack(media.MediaType, NewAudioCodec(media.EncodeName, uint8(media.PayloadType), uint32(media.ClockRate), media.ChannelCount))
        } else {
            track = NewTrack(media.MediaType, NewVideoCodec(media.EncodeName, uint8(media.PayloadType), uint32(media.ClockRate)))
        }
        if client.onTrack != nil {
            client.onTrack(track)
        }
        client.tracks[media.MediaType] = track
        media.ControlUrl = getControlUrl(media.ControlUrl)
    }
    for i := client.setupStep; i < len(client.sdpContext.Medias); i++ {
        if !client.tracks[client.sdpContext.Medias[client.setupStep].MediaType].isOpen {
            continue
        }
        client.setupStep = i + 1
        req := makeSetup(client.sdpContext.Medias[client.setupStep].ControlUrl, client.cseq)
        return client.sendRtspRequest(&req)
    }
    return nil
}

func (client *RtspClient) handleSetup(res *RtspResponse) error {

}

func (client *RtspClient) handlePlay(res *RtspResponse) error {

}

func (client *RtspClient) handleTeardown(res *RtspResponse) error {

}

func (client *RtspClient) handlePause(res *RtspResponse) error {

}

func (client *RtspClient) handleAnnounce(res *RtspResponse) error {

}

func (client *RtspClient) handleRecord(res *RtspResponse) error {

}

func (client *RtspClient) handleGetParameter(res *RtspResponse) error {

}

func (client *RtspClient) handleSetParameter(res *RtspResponse) error {

}

func (client *RtspClient) handleRedirect(res *RtspResponse) error {

}
