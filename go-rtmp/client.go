package rtmp

import (
	"encoding/binary"
	"errors"
	"strings"

	"github.com/yapingcat/gomedia/go-codec"
	"github.com/yapingcat/gomedia/go-flv"
)

type RtmpConnectCmd int

const (
    CONNECT RtmpConnectCmd = iota
    CLOSE
    CREATE_STREAM
    GET_STREAM_LENGTH
)

type RtmpClient struct {
    tcurl          string
    app            string
    streamName     string
    cmdChan        *chunkStreamWriter
    userCtrlChan   *chunkStreamWriter
    sourceChan     *chunkStreamWriter
    audioChan      *chunkStreamWriter
    videoChan      *chunkStreamWriter
    reader         *chunkStreamReader
    wndAckSize     uint32
    state          RtmpParserState
    streamState    RtmpState
    hs             *clientHandShake
    output         OutputCB
    onframe        OnFrame
    onstatus       OnStatus
    onerror        OnError
    onstateChange  OnStateChange
    videoDemuxer   flv.VideoTagDemuxer
    audioDemuxer   flv.AudioTagDemuxer
    videoMuxer     flv.AVTagMuxer
    audioMuxer     flv.AVTagMuxer
    timestamp      uint32
    lastMethod     RtmpConnectCmd
    lastMethodTid  int
    tid            uint32
    streamId       uint32
    writeChunkSize uint32
    isPublish      bool
}

func NewRtmpClient(options ...func(*RtmpClient)) *RtmpClient {
    cli := &RtmpClient{
        hs:             newClientHandShake(),
        cmdChan:        newChunkStreamWriter(CHUNK_CHANNEL_CMD),
        userCtrlChan:   newChunkStreamWriter(CHUNK_CHANNEL_USE_CTRL),
        sourceChan:     newChunkStreamWriter(CHUNK_CHANNEL_NET_STREAM),
        reader:         newChunkStreamReader(FIX_CHUNK_SIZE),
        tid:            4,
        wndAckSize:     DEFAULT_ACK_SIZE,
        writeChunkSize: DEFAULT_CHUNK_SIZE,
        isPublish:      false,
    }

    for _, o := range options {
        o(cli)
    }
    return cli
}

func WithChunkSize(chunkSize uint32) func(*RtmpClient) {
    return func(rc *RtmpClient) {
        if rc != nil {
            rc.writeChunkSize = chunkSize
        }
    }
}

func WithComplexHandshake() func(*RtmpClient) {
    return func(rc *RtmpClient) {
        if rc != nil {
            rc.hs.simpleHs = false
        }
    }
}

func WithComplexHandshakeSchema(schema int) func(*RtmpClient) {
    return func(rc *RtmpClient) {
        if rc != nil {
            rc.hs.schema = schema
        }
    }
}

func WithWndAckSize(ackSize uint32) func(*RtmpClient) {
    return func(rc *RtmpClient) {
        if rc != nil {
            rc.wndAckSize = ackSize
        }
    }
}

func WithEnablePublish() func(*RtmpClient) {
    return func(rc *RtmpClient) {
        if rc != nil {
            rc.isPublish = true
        }
    }
}

func WithAudioMuxer(muxer flv.AVTagMuxer) func(*RtmpClient) {
    return func(rc *RtmpClient) {
        if rc != nil {
            rc.audioMuxer = muxer
        }
    }
}

func (cli *RtmpClient) SetOutput(output OutputCB) {
    cli.output = output
    cli.hs.output = output
}

func (cli *RtmpClient) OnFrame(onframe OnFrame) {
    cli.onframe = onframe
}

func (cli *RtmpClient) OnError(onerror OnError) {
    cli.onerror = onerror
}

func (cli *RtmpClient) OnStatus(onstatus OnStatus) {
    cli.onstatus = onstatus
}

func (cli *RtmpClient) OnStateChange(stateChange OnStateChange) {
    cli.onstateChange = stateChange
}

//url start with "rtmp://"
func (cli *RtmpClient) Start(url string) {
    loc := strings.Index(url, "rtmp://")
    cli.tcurl = "rtmp://"
    tmp := url[loc+7:]
    loc = strings.Index(tmp, "/")
    cli.tcurl += tmp[:loc]
    tmp = tmp[loc+1:]
    loc = strings.Index(tmp, "/")
    cli.app = tmp[:loc]
    cli.tcurl += "/" + cli.app
    cli.streamName = tmp[loc+1:]
    cli.hs.start()
}

func (cli *RtmpClient) GetState() RtmpState {
    return cli.streamState
}

func (cli *RtmpClient) Input(data []byte) error {

    switch cli.state {
    case HandShake:
        cli.changeState(STATE_HANDSHAKEING)
        cli.hs.input(data)
        if cli.hs.getState() != HANDSHAKE_DONE {
            return nil
        } else {
            cli.changeState(STATE_RTMP_CONNECTING)
            cli.state = ReadChunk
            cmd := makeConnect(cli.app, cli.tcurl)
            bufs := cli.cmdChan.writeData(cmd, Command_AMF0, 0, 0)
            if err := cli.output(bufs); err != nil {
                return err
            }
            cli.lastMethod = CONNECT
            cli.lastMethodTid = 1
        }
    case ReadChunk:

        err := cli.reader.readRtmpMessage(data, func(msg *rtmpMessage) error {
            cli.timestamp = msg.timestamp
            return cli.handleMessage(msg)
        })

        if err != nil {
            return err
        }
    default:
        panic("error state")
    }
    return nil
}

func (cli *RtmpClient) WriteFrame(cid codec.CodecID, frame []byte, pts, dts uint32) error {
    if cid == codec.CODECID_AUDIO_AAC || cid == codec.CODECID_AUDIO_G711A || cid == codec.CODECID_AUDIO_G711U {
        return cli.WriteAudio(cid, frame, pts, dts)
    } else if cid == codec.CODECID_VIDEO_H264 || cid == codec.CODECID_VIDEO_H265 {
        return cli.WriteVideo(cid, frame, pts, dts)
    } else {
        return errors.New("unsupport codec id")
    }
}

func (cli *RtmpClient) WriteAudio(cid codec.CodecID, frame []byte, pts, dts uint32) error {
    if cli.audioMuxer == nil {
        cli.audioMuxer = flv.CreateAudioMuxer(flv.CovertCodecId2SoundFromat(cid))
    }
    if cli.audioChan == nil {
        cli.audioChan = newChunkStreamWriter(CHUNK_CHANNEL_AUDIO)
        cli.audioChan.chunkSize = cli.writeChunkSize
    }
    tags := cli.audioMuxer.Write(frame, pts, dts)
    for _, tag := range tags {
        pkt := cli.audioChan.writeData(tag, AUDIO, cli.streamId, dts)
        if len(pkt) > 0 {
            if err := cli.output(pkt); err != nil {
                return err
            }
        }
    }
    return nil
}

func (cli *RtmpClient) WriteVideo(cid codec.CodecID, frame []byte, pts, dts uint32) error {
    if cli.videoMuxer == nil {
        cli.videoMuxer = flv.CreateVideoMuxer(flv.CovertCodecId2FlvVideoCodecId(cid))
    }
    if cli.videoChan == nil {
        cli.videoChan = newChunkStreamWriter(CHUNK_CHANNEL_VIDEO)
        cli.videoChan.chunkSize = cli.writeChunkSize
    }
    tags := cli.videoMuxer.Write(frame, pts, dts)
    for _, tag := range tags {
        pkt := cli.videoChan.writeData(tag, VIDEO, cli.streamId, dts)
        if len(pkt) > 0 {
            if err := cli.output(pkt); err != nil {
                return err
            }
        }
    }
    return nil
}

func (cli *RtmpClient) changeState(newState RtmpState) {
    if cli.streamState != newState {
        cli.streamState = newState
        if cli.onstateChange != nil {
            cli.onstateChange(newState)
        }
    }
}

func (cli *RtmpClient) handleMessage(msg *rtmpMessage) error {
    switch msg.msgtype {
    case SET_CHUNK_SIZE:
        if len(msg.msg) < 4 {
            return errors.New("bytes of \"set chunk size\"  < 4")
        }
        size := binary.BigEndian.Uint32(msg.msg)
        cli.reader.chunkSize = size
    case ABORT_MESSAGE:
        //TODO
    case ACKNOWLEDGEMENT:
        if len(msg.msg) < 4 {
            return errors.New("bytes of \"window acknowledgement size\"  < 4")
        }
        cli.wndAckSize = binary.BigEndian.Uint32(msg.msg)
    case USER_CONTROL:
        return cli.handleUserEvent(msg.msg)
    case WND_ACK_SIZE:
        //TODO
    case SET_PEER_BW:
        //TODO
    case AUDIO:
        return cli.handleAudioMessage(msg)
    case VIDEO:
        return cli.handleVideoMessage(msg)
    case Command_AMF0:
        return cli.handleCommandRes(msg.msg)
    case Command_AMF3:
    case Metadata_AMF0:
    case Metadata_AMF3:
    case SharedObject_AMF0:
    case SharedObject_AMF3:
    case Aggregate:
    default:
        return errors.New("unkow message type")
    }
    return nil
}

func (cli *RtmpClient) handleUserEvent(data []byte) error {
    event := decodeUserControlMsg(data)
    switch event.code {
    case StreamBegin:
    case StreamEOF:
    case StreamDry:
    case SetBufferLength:
    case StreamIsRecorded:
    case PingRequest:
    case PingResponse:
    default:
        panic("unkown event")
    }
    return nil
}

func (cli *RtmpClient) handleCommandRes(data []byte) error {
    item := amf0Item{}
    l := item.decode(data)
    data = data[l:]
    cmd := string(item.value.([]byte))
    switch cmd {
    case "_result":
        return cli.handleResult(data)
    case "_error":
        return cli.handleError(data)
    case "onStatus":
        return cli.handleStatus(data)
    default:
    }
    return nil
}

func (cli *RtmpClient) handleVideoMessage(msg *rtmpMessage) error {
    if cli.videoDemuxer == nil {
        cli.videoDemuxer = flv.CreateFlvVideoTagHandle(flv.FLV_VIDEO_CODEC_ID(msg.msg[0] & 0x0F))
        cli.videoDemuxer.OnFrame(func(codecid codec.CodecID, frame []byte, cts int) {
            dts := cli.timestamp
            pts := dts + uint32(cts)
            cli.onframe(codecid, pts, dts, frame)
        })
    }
    return cli.videoDemuxer.Decode(msg.msg)
}

func (cli *RtmpClient) handleAudioMessage(msg *rtmpMessage) error {
    if cli.audioDemuxer == nil {
        cli.audioDemuxer = flv.CreateAudioTagDemuxer(flv.FLV_SOUND_FORMAT((msg.msg[0] >> 4) & 0x0F))
        cli.audioDemuxer.OnFrame(func(codecid codec.CodecID, frame []byte) {
            dts := cli.timestamp
            pts := dts
            cli.onframe(codecid, pts, dts, frame)
        })
    }
    return cli.audioDemuxer.Decode(msg.msg)
}

func (cli *RtmpClient) handleResult(data []byte) error {
    switch cli.lastMethod {

    case CONNECT:
        return cli.handleConnectResponse(data)
    case CREATE_STREAM:
        return cli.handleCreateStreamResponse(data)
    case GET_STREAM_LENGTH:
        //TODO
    }
    return nil
}

func (cli *RtmpClient) handleConnectResponse(data []byte) error {

    items, _ := decodeAmf0(data)
    if len(items) > 0 {
        if tid, ok := items[0].value.(float64); ok {
            if cli.lastMethodTid != int(tid) {
                return nil
            }
        }
    }

    cli.lastMethod = CREATE_STREAM
    cli.lastMethodTid = 2
    if !cli.isPublish {
        ack := makeAcknowledgementSize(cli.wndAckSize)
        bufs := cli.userCtrlChan.writeData(ack, WND_ACK_SIZE, 0, 0)
        cmd := makeCreateStream(cli.streamName, 2)
        bufs = append(bufs, cli.cmdChan.writeData(cmd, Command_AMF0, 0, 0)...)
        return cli.output(bufs)
    } else {
        buf := makeSetChunkSize(cli.writeChunkSize)
        bufs := cli.userCtrlChan.writeData(buf, SET_CHUNK_SIZE, 0, 0)
        cli.cmdChan.chunkSize = cli.writeChunkSize
        cli.userCtrlChan.chunkSize = cli.writeChunkSize
        cli.sourceChan.chunkSize = cli.writeChunkSize
        buf = makeReleaseStream(cli.streamName)
        bufs = append(bufs, cli.cmdChan.writeData(buf, Command_AMF0, 0, 0)...)
        buf = makeFcPublish(cli.streamName)
        bufs = append(bufs, cli.cmdChan.writeData(buf, Command_AMF0, 0, 0)...)
        buf = makeCreateStream(cli.streamName, 2)
        bufs = append(bufs, cli.cmdChan.writeData(buf, Command_AMF0, 0, 0)...)
        return cli.output(bufs)
    }
}

func (cli *RtmpClient) handleCreateStreamResponse(data []byte) error {

    items, _ := decodeAmf0(data)
    if len(items) > 0 {
        if tid, ok := items[0].value.(float64); ok {
            if cli.lastMethodTid != int(tid) {
                return nil
            }
        }
        if sid, ok := items[len(items)-1].value.(float64); ok {
            cli.streamId = uint32(sid)
        }
    }

    if !cli.isPublish {
        cli.lastMethod = GET_STREAM_LENGTH
        cli.lastMethodTid = 3
        cmd := makeGetStreamLength(3, cli.streamName)
        bufs := cli.cmdChan.writeData(cmd, Command_AMF0, cli.streamId, 0)
        req := makePlay(int(cli.tid), cli.streamName, -1, -1, true)
        bufs = append(bufs, cli.sourceChan.writeData(req, Command_AMF0, cli.streamId, 0)...)
        return cli.output(bufs)
    } else {
        data := makePublish(cli.streamName, PUBLISHING_LIVE)
        bufs := cli.cmdChan.writeData(data, Command_AMF0, cli.streamId, 0)
        return cli.output(bufs)
    }
}

func (cli *RtmpClient) handleError(data []byte) error {
    code := ""
    describe := ""
    _, objs := decodeAmf0(data)
    for _, obj := range objs {
        for _, item := range obj.items {
            if item.name == "code" {
                code = string(item.value.value.([]byte))
            } else if item.name == "describe" {
                describe = string(item.value.value.([]byte))
            }
            if cli.onerror != nil {
                cli.onerror(code, describe)
            }
        }
    }
    if cli.isPublish {
        cli.changeState(STATE_RTMP_PUBLISH_FAILED)
    } else {
        cli.changeState(STATE_RTMP_PLAY_FAILED)
    }
    return nil
}

func (cli *RtmpClient) handleStatus(data []byte) error {
    code := ""
    level := ""
    describe := ""

    foundInfoObj := false
    _, objs := decodeAmf0(data)
    for _, obj := range objs {
        for _, item := range obj.items {
            if item.name == "code" {
                foundInfoObj = true
                code = string(item.value.value.([]byte))
            } else if item.name == "level" {
                level = string(item.value.value.([]byte))
            } else if item.name == "description" {
                describe = string(item.value.value.([]byte))
            }
        }
    }

    if cli.onstatus != nil && foundInfoObj {
        cli.onstatus(code, level, describe)
    }

    if code == string(NETSTREAM_PUBLISH_START) {
        cli.changeState(STATE_RTMP_PUBLISH_START)
    } else if code == string(NETSTREAM_PLAY_START) {
        cli.changeState(STATE_RTMP_PLAY_START)
    } else if level == string(LEVEL_ERROR) {
        if cli.isPublish {
            cli.changeState(STATE_RTMP_PUBLISH_FAILED)
        } else {
            cli.changeState(STATE_RTMP_PLAY_FAILED)
        }
    }
    return nil
}
