package rtmp

import (
    "encoding/binary"
    "errors"

    "github.com/yapingcat/gomedia/codec"
    "github.com/yapingcat/gomedia/flv"
)

//example
//1. rtmp 推流服务端
//
//listen, _ := net.Listen("tcp4", "0.0.0.0:1935")
//conn, _ := listen.Accept()
//
// handle := NewRtmpServerHandle()
// handle.OnPublish(func(app, streamName string) StatusCode {
//     return NETSTREAM_PUBLISH_START
// })
//
// handle.SetOutput(func(b []byte) error {
//     _, err := conn.Write(b)
//     return err
// })

// handle.OnFrame(func(cid codec.CodecID, pts, dts uint32, frame []byte) {
//     if cid == codec.CODECID_VIDEO_H264 {
//        //do something
//     }
//     ........
// })

//
// 把从网络中接收到的数据，input到rtmp句柄当中
// buf := make([]byte, 60000)
// for {
//     n, err := conn.Read(buf)
//     if err != nil {
//         fmt.Println(err)
//         break
//     }
//     err = handle.Input(buf[0:n])
//     if err != nil {
//         fmt.Println(err)
//         break
//     }
// }

// rtmp播放服务端
// listen, _ := net.Listen("tcp4", "0.0.0.0:1935")
// conn, _ := listen.Accept()

// ready := make(chan struct{})
// handle := NewRtmpServerHandle()
// handle.onPlay = func(app, streamName string, start, duration float64, reset bool) StatusCode {
//        return NETSTREAM_PLAY_START
//  }
//
// handle.OnStateChange(func(newstate RtmpState) {
//    if newstate == STATE_RTMP_PLAY_START {
//        close(ready) //关闭这个通道，通知推流协程可以向客户端推流了
//    }
//  })
//
//  handle.SetOutput(func(b []byte) error {
//       _, err := conn.Write(b)
//      return err
//  })
//
//  go func() {
//
//      等待推流
//      <-ready
//
//      开始推流
//      handle.WriteVideo(cid, frame, pts, dts)
//      handle.WriteAudio(cid, frame, pts, dts)
//
//  }()
//
//  把从网络中接收到的数据，input到rtmp句柄当中
//  buf := make([]byte, 60000)
//  for {
//      n, err := conn.Read(buf)
//      if err != nil {
//          fmt.Println(err)
//          break
//      }
//      err = handle.Input(buf[0:n])
//      if err != nil {
//          fmt.Println(err)
//          break
//      }
//  }
//  conn.Close()

type RtmpServerHandle struct {
    app            string
    streamName     string
    tcUrl          string
    state          RtmpParserState
    streamState    RtmpState
    cmdChan        *chunkStreamWriter
    userCtrlChan   *chunkStreamWriter
    audioChan      *chunkStreamWriter
    videoChan      *chunkStreamWriter
    reader         *chunkStreamReader
    writeChunkSize uint32
    hs             *serverHandShake
    wndAckSize     uint32
    peerWndAckSize uint32
    videoDemuxer   flv.VideoTagDemuxer
    audioDemuxer   flv.AudioTagDemuxer
    videoMuxer     flv.AVTagMuxer
    audioMuxer     flv.AVTagMuxer
    onframe        OnFrame
    output         OutputCB
    onRelease      OnReleaseStream
    onChangeState  OnStateChange
    onPlay         OnPlay
    onPublish      OnPublish
    timestamp      uint32
    streamId       uint32
}

func NewRtmpServerHandle(options ...func(*RtmpServerHandle)) *RtmpServerHandle {
    server := &RtmpServerHandle{
        hs:             newServerHandShake(),
        cmdChan:        newChunkStreamWriter(CHUNK_CHANNEL_CMD),
        userCtrlChan:   newChunkStreamWriter(CHUNK_CHANNEL_USE_CTRL),
        reader:         newChunkStreamReader(FIX_CHUNK_SIZE),
        wndAckSize:     DEFAULT_ACK_SIZE,
        writeChunkSize: DEFAULT_CHUNK_SIZE,
        streamId:       1,
    }

    for _, o := range options {
        o(server)
    }

    return server
}

func (server *RtmpServerHandle) SetOutput(output OutputCB) {
    server.output = output
    server.hs.output = output
}

func (server *RtmpServerHandle) OnFrame(onframe OnFrame) {
    server.onframe = onframe
}

func (server *RtmpServerHandle) OnPlay(onPlay OnPlay) {
    server.onPlay = onPlay
}

func (server *RtmpServerHandle) OnPublish(onPub OnPublish) {
    server.onPublish = onPub
}

func (server *RtmpServerHandle) OnRelease(onRelease OnReleaseStream) {
    server.onRelease = onRelease
}

//状态变更，回调函数，
//服务端在STATE_RTMP_PLAY_START状态下，开始发流
//客户端在STATE_RTMP_PUBLISH_START状态，开始推流
func (server *RtmpServerHandle) OnStateChange(stateChange OnStateChange) {
    server.onChangeState = stateChange
}

func (server *RtmpServerHandle) GetStreamName() string {
    return server.streamName
}

func (server *RtmpServerHandle) GetApp() string {
    return server.app
}

func (server *RtmpServerHandle) GetState() RtmpState {
    return server.streamState
}

func (server *RtmpServerHandle) Input(data []byte) error {
    for len(data) > 0 {
        switch server.state {
        case HandShake:
            server.changeState(STATE_HANDSHAKEING)
            r := server.hs.input(data)
            if server.hs.getState() == HANDSHAKE_DONE {
                server.changeState(STATE_HANDSHAKE_DONE)
                server.state = ReadChunk
            }
            data = data[r:]
        case ReadChunk:

            err := server.reader.readRtmpMessage(data, func(msg *rtmpMessage) error {
                server.timestamp = msg.timestamp
                return server.handleMessage(msg)
            })
            return err
        }
    }
    return nil
}

func (server *RtmpServerHandle) WriteFrame(cid codec.CodecID, frame []byte, pts, dts uint32) error {
    if cid == codec.CODECID_AUDIO_AAC || cid == codec.CODECID_AUDIO_G711A || cid == codec.CODECID_AUDIO_G711U {
        return server.WriteAudio(cid, frame, pts, dts)
    } else if cid == codec.CODECID_VIDEO_H264 || cid == codec.CODECID_VIDEO_H265 {
        return server.WriteVideo(cid, frame, pts, dts)
    } else {
        return errors.New("unsupport codec id")
    }
}

func (server *RtmpServerHandle) WriteAudio(cid codec.CodecID, frame []byte, pts, dts uint32) error {

    if server.audioMuxer == nil {
        server.audioMuxer = flv.CreateAudioMuxer(flv.CovertCodecId2SoundFromat(cid))
    }
    if server.audioChan == nil {
        server.audioChan = newChunkStreamWriter(CHUNK_CHANNEL_AUDIO)
        server.audioChan.chunkSize = server.writeChunkSize
    }
    tags := server.audioMuxer.Write(frame, pts, dts)
    for _, tag := range tags {
        pkt := server.audioChan.writeData(tag, AUDIO, server.streamId, dts)
        if len(pkt) > 0 {
            if err := server.output(pkt); err != nil {
                return err
            }
        }
    }
    return nil
}

func (server *RtmpServerHandle) WriteVideo(cid codec.CodecID, frame []byte, pts, dts uint32) error {
    if server.videoMuxer == nil {
        server.videoMuxer = flv.CreateVideoMuxer(flv.CovertCodecId2FlvVideoCodecId(cid))
    }
    if server.videoChan == nil {
        server.videoChan = newChunkStreamWriter(CHUNK_CHANNEL_VIDEO)
        server.videoChan.chunkSize = server.writeChunkSize
    }
    tags := server.videoMuxer.Write(frame, pts, dts)
    for _, tag := range tags {
        pkt := server.videoChan.writeData(tag, VIDEO, server.streamId, dts)
        if len(pkt) > 0 {
            if err := server.output(pkt); err != nil {
                return err
            }
        }
    }
    return nil
}

func (server *RtmpServerHandle) changeState(newState RtmpState) {
    if server.streamState != newState {
        server.streamState = newState
        if server.onChangeState != nil {
            server.onChangeState(newState)
        }
    }
}

func (server *RtmpServerHandle) handleMessage(msg *rtmpMessage) error {
    switch msg.msgtype {
    case SET_CHUNK_SIZE:
        if len(msg.msg) < 4 {
            return errors.New("bytes of \"set chunk size\"  < 4")
        }
        size := binary.BigEndian.Uint32(msg.msg)
        server.reader.chunkSize = size
    case ABORT_MESSAGE:
        //TODO
    case ACKNOWLEDGEMENT:
        if len(msg.msg) < 4 {
            return errors.New("bytes of \"window acknowledgement size\"  < 4")
        }
        server.peerWndAckSize = binary.BigEndian.Uint32(msg.msg)
    case USER_CONTROL:
        //TODO
    case WND_ACK_SIZE:
        //TODO
    case SET_PEER_BW:
        //TODO
    case AUDIO:
        return server.handleAudioMessage(msg)
    case VIDEO:
        return server.handleVideoMessage(msg)
    case Command_AMF0:
        return server.handleCommand(msg.msg)
    case Command_AMF3:
    case Metadata_AMF0:
    case Metadata_AMF3:
    case SharedObject_AMF0:
    case SharedObject_AMF3:
    case Aggregate:
    default:
        return errors.New("unkown message type")
    }
    return nil
}

func (server *RtmpServerHandle) handleCommand(data []byte) error {
    item := amf0Item{}
    l := item.decode(data)
    data = data[l:]
    cmd := string(item.value.([]byte))
    switch cmd {
    case "connect":
        server.changeState(STATE_RTMP_CONNECTING)
        return server.handleConnect(data)
    case "releaseStream":
        server.handleReleaseStream(data)
    case "FCPublish":
    case "createStream":
        return server.handleCreateStream(data)
    case "play":
        return server.handlePlay(data)
    case "publish":
        return server.handlePublish(data)
    default:
    }
    return nil
}

func (server *RtmpServerHandle) handleConnect(data []byte) error {
    _, objs := decodeAmf0(data)
    if len(objs) > 0 {
        for _, item := range objs[0].items {
            if item.name == "app" {
                server.app = string(item.value.value.([]byte))
            } else if item.name == "tcUrl" {
                server.tcUrl = string(item.value.value.([]byte))
            }
        }
    }

    buf := makeSetChunkSize(server.writeChunkSize)
    bufs := server.userCtrlChan.writeData(buf, SET_CHUNK_SIZE, 0, 0)
    server.userCtrlChan.chunkSize = server.writeChunkSize
    server.cmdChan.chunkSize = server.writeChunkSize
    buf = makeAcknowledgementSize(server.wndAckSize)
    bufs = append(bufs, server.userCtrlChan.writeData(buf, WND_ACK_SIZE, 0, 0)...)
    buf = makeSetPeerBandwidth(server.wndAckSize, LimitType_DYNAMIC)
    bufs = append(bufs, server.userCtrlChan.writeData(buf, SET_PEER_BW, 0, 0)...)
    bufs = append(bufs, server.cmdChan.writeData(makeConnectRes(), Command_AMF0, 0, 0)...)
    return server.output(bufs)
}

func (server *RtmpServerHandle) handleReleaseStream(data []byte) {
    items, _ := decodeAmf0(data)
    if len(items) == 0 {
        return
    }
    streamName := string(items[len(items)-1].value.([]byte))
    if server.onRelease != nil {
        server.onRelease(server.app, streamName)
    }
}

func (server *RtmpServerHandle) handleCreateStream(data []byte) error {
    items, _ := decodeAmf0(data)
    if len(items) == 0 {
        return nil
    }
    tid := uint32(items[0].value.(float64))
    bufs := server.cmdChan.writeData(makeCreateStreamRes(tid, server.streamId), Command_AMF0, 0, 0)
    return server.output(bufs)
}

func (server *RtmpServerHandle) handlePlay(data []byte) error {
    items, _ := decodeAmf0(data)
    tid := int(items[0].value.(float64))
    streamName := string(items[2].value.([]byte))
    server.streamName = streamName
    start := float64(-2)
    duration := float64(-1)
    reset := false

    if len(items) > 3 {
        start = items[3].value.(float64)
    }
    if len(items) > 4 {
        duration = items[4].value.(float64)
    }

    if len(items) > 5 {
        reset = items[5].value.(bool)
    }

    code := NETSTREAM_PLAY_START
    if server.onPlay != nil {
        code = server.onPlay(server.app, streamName, start, duration, reset)
    }
    if code == NETSTREAM_PLAY_START {
        res := makeUserControlMessage(StreamBegin, int(server.streamId))
        bufs := server.userCtrlChan.writeData(res, USER_CONTROL, 0, 0)
        res = makeStatusRes(tid, NETSTREAM_PLAY_RESET, NETSTREAM_PLAY_RESET.Level(), string(NETSTREAM_PLAY_RESET.Description()))
        bufs = append(bufs, server.cmdChan.writeData(res, Command_AMF0, server.streamId, 0)...)
        res = makeStatusRes(tid, NETSTREAM_PLAY_START, NETSTREAM_PLAY_START.Level(), string(NETSTREAM_PLAY_START.Description()))
        bufs = append(bufs, server.cmdChan.writeData(res, Command_AMF0, server.streamId, 0)...)
        if err := server.output(bufs); err != nil {
            return err
        }
        server.changeState(STATE_RTMP_PLAY_START)
    } else {
        res := makeStatusRes(tid, code, code.Level(), string(code.Description()))
        if err := server.output(server.cmdChan.writeData(res, Command_AMF0, server.streamId, 0)); err != nil {
            return err
        }
        server.changeState(STATE_RTMP_PLAY_FAILED)
    }
    return nil
}

func (server *RtmpServerHandle) handlePublish(data []byte) error {
    items, _ := decodeAmf0(data)
    tid := int(items[0].value.(float64))
    streamName := string(items[2].value.([]byte))
    server.streamName = streamName
    code := NETSTREAM_PUBLISH_START
    if server.onPublish != nil {
        code = server.onPublish(server.app, streamName)
    }
    res := makeStatusRes(tid, code, code.Level(), string(code.Description()))
    if err := server.output(server.cmdChan.writeData(res, Command_AMF0, server.streamId, 0)); err != nil {
        return err
    }
    if code == NETSTREAM_PUBLISH_START {
        server.changeState(STATE_RTMP_PUBLISH_START)
    } else {
        server.changeState(STATE_RTMP_PUBLISH_FAILED)
    }
    return nil
}

func (server *RtmpServerHandle) handleVideoMessage(msg *rtmpMessage) error {
    if server.videoDemuxer == nil {
        server.videoDemuxer = flv.CreateFlvVideoTagHandle(flv.FLV_VIDEO_CODEC_ID(msg.msg[0] & 0x0F))
        server.videoDemuxer.OnFrame(func(codecid codec.CodecID, frame []byte, cts int) {
            dts := server.timestamp
            pts := dts + uint32(cts)
            server.onframe(codecid, pts, dts, frame)
        })
    }
    return server.videoDemuxer.Decode(msg.msg)
}

func (server *RtmpServerHandle) handleAudioMessage(msg *rtmpMessage) error {
    if server.audioDemuxer == nil {
        server.audioDemuxer = flv.CreateAudioTagDemuxer(flv.FLV_SOUND_FORMAT((msg.msg[0] >> 4) & 0x0F))
        server.audioDemuxer.OnFrame(func(codecid codec.CodecID, frame []byte) {
            dts := server.timestamp
            pts := dts
            server.onframe(codecid, pts, dts, frame)
        })
    }
    return server.audioDemuxer.Decode(msg.msg)
}
