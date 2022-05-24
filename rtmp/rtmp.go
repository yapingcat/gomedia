package rtmp

import (
    "github.com/yapingcat/gomedia/codec"
)

const (
    CHUNK_CHANNEL_USE_CTRL   = 2
    CHUNK_CHANNEL_CMD        = 3
    CHUNK_CHANNEL_VIDEO      = 5
    CHUNK_CHANNEL_AUDIO      = 6
    CHUNK_CHANNEL_META       = 7
    CHUNK_CHANNEL_NET_STREAM = 8
)

const (
    FIX_CHUNK_SIZE     = 128
    DEFAULT_CHUNK_SIZE = 60000
    DEFAULT_ACK_SIZE   = 5000000
)

const (
    HANDSHAKE_SIZE           = 1536
    HANDSHAKE_FIX_SIZE       = 8
    HANDSHAKE_OFFSET_SIZE    = 4
    HANDSHAKE_DIGEST_SIZE    = 32
    HANDSHAKE_SCHEMA_SIZE    = 764
    HANDSHAKE_SCHEMA0_OFFSET = 776 // 8 + 764 + 4
    HANDSHAKE_SCHEMA1_OFFSET = 12  // 8 + 4
)

const (
    HANDSHAKE_COMPLEX_SCHEMA0 = 0
    HANDSHAKE_COMPLEX_SCHEMA1 = 1
)

const (
    PUBLISHING_LIVE   = "live"
    PUBLISHING_RECORD = "record"
    PUBLISHING_APPEND = "append"
)

const (
    LimitType_HARD    = 0
    LimitType_SOFT    = 1
    LimitType_DYNAMIC = 2
)

type RtmpParserState int

const (
    HandShake RtmpParserState = iota
    ReadChunk
)

type RtmpState int

const (
    STATE_HANDSHAKEING RtmpState = iota
    STATE_HANDSHAKE_DONE
    STATE_RTMP_CONNECTING
    STATE_RTMP_PLAY_START
    STATE_RTMP_PLAY_FAILED
    STATE_RTMP_PUBLISH_START
    STATE_RTMP_PUBLISH_FAILED
)

//https://blog.csdn.net/wq892373445/article/details/118387494

type StatusLevel string

const (
    LEVEL_STATUS StatusLevel = "status"
    LEVEL_ERROR  StatusLevel = "error"
    LEVEL_WARN   StatusLevel = "warning"
)

type StatusCode string

const (
    NETSTREAM_PUBLISH_START     StatusCode = "NetStream.Publish.Start"
    NETSTREAM_PLAY_START        StatusCode = "NetStream.Play.Start"
    NETSTREAM_PLAY_STOP         StatusCode = "NetStream.Play.Stop"
    NETSTREAM_PLAY_FAILED       StatusCode = "NetStream.Play.Failed"
    NETSTREAM_PLAY_NOTFOUND     StatusCode = "NetStream.Play.StreamNotFound"
    NETSTREAM_PLAY_RESET        StatusCode = "NetStream.Play.Reset"
    NETSTREAM_PAUSE_NOTIFY      StatusCode = "NetStream.Pause.Notify"
    NETSTREAM_UNPAUSE_NOTIFY    StatusCode = "NetStream.Unpause.Notify"
    NETSTREAM_RECORD_START      StatusCode = "NetStream.Record.Start"
    NETSTREAM_RECORD_STOP       StatusCode = "NetStream.Record.Stop"
    NETSTREAM_RECORD_FAILED     StatusCode = "NetStream.Record.Failed"
    NETSTREAM_SEEK_FAILED       StatusCode = "NetStream.Seek.Failed"
    NETSTREAM_SEEK_NOTIFY       StatusCode = "NetStream.Seek.Notify"
    NETCONNECT_CONNECT_CLOSED   StatusCode = "NetConnection.Connect.Closed"
    NETCONNECT_CONNECT_FAILED   StatusCode = "NetConnection.Connect.Failed"
    NETCONNECT_CONNECT_SUCCESS  StatusCode = "NetConnection.Connect.Success"
    NETCONNECT_CONNECT_REJECTED StatusCode = "NetConnection.Connect.Rejected"
    NETSTREAM_CONNECT_CLOSED    StatusCode = "NetStream.Connect.Closed"
    NETSTREAM_CONNECT_FAILED    StatusCode = "NetStream.Connect.Failed"
    NETSTREAM_CONNECT_SUCCESSS  StatusCode = "NetStream.Connect.Success"
    NETSTREAM_CONNECT_REJECTED  StatusCode = "NetStream.Connect.Rejected"
)

func (c StatusCode) Level() StatusLevel {
    switch c {
    case NETSTREAM_PUBLISH_START:
        return "status"
    case NETSTREAM_PLAY_START:
        return "status"
    case NETSTREAM_PLAY_STOP:
        return "status"
    case NETSTREAM_PLAY_FAILED:
        return "error"
    case NETSTREAM_PLAY_NOTFOUND:
        return "error"
    case NETSTREAM_PLAY_RESET:
        return "status"
    case NETSTREAM_PAUSE_NOTIFY:
        return "status"
    case NETSTREAM_UNPAUSE_NOTIFY:
        return "status"
    case NETSTREAM_RECORD_START:
        return "status"
    case NETSTREAM_RECORD_STOP:
        return "status"
    case NETSTREAM_RECORD_FAILED:
        return "error"
    case NETSTREAM_SEEK_FAILED:
        return "error"
    case NETSTREAM_SEEK_NOTIFY:
        return "status"
    case NETCONNECT_CONNECT_CLOSED:
        return "status"
    case NETCONNECT_CONNECT_FAILED:
        return "error"
    case NETCONNECT_CONNECT_SUCCESS:
        return "status"
    case NETCONNECT_CONNECT_REJECTED:
        return "error"
    case NETSTREAM_CONNECT_CLOSED:
        return "status"
    case NETSTREAM_CONNECT_FAILED:
        return "error"
    case NETSTREAM_CONNECT_SUCCESSS:
        return "status"
    case NETSTREAM_CONNECT_REJECTED:
        return "error"
    }
    return ""
}

func (c StatusCode) Description() StatusLevel {
    switch c {
    case NETSTREAM_PUBLISH_START:
        return "Start publishing stream"
    case NETSTREAM_PLAY_START:
        return "Start play stream "
    case NETSTREAM_PLAY_STOP:
        return "Stop play stream"
    case NETSTREAM_PLAY_FAILED:
        return "Play stream failed"
    case NETSTREAM_PLAY_NOTFOUND:
        return "Stream not found"
    case NETSTREAM_PLAY_RESET:
        return "Reset stream"
    case NETSTREAM_PAUSE_NOTIFY:
        return "Pause stream"
    case NETSTREAM_UNPAUSE_NOTIFY:
        return "Unpause stream"
    case NETSTREAM_RECORD_START:
        return "Start record stream"
    case NETSTREAM_RECORD_STOP:
        return "Stop record stream"
    case NETSTREAM_RECORD_FAILED:
        return "Record stream failed"
    case NETSTREAM_SEEK_FAILED:
        return "Seek stream failed"
    case NETSTREAM_SEEK_NOTIFY:
        return "Seek stream"
    case NETCONNECT_CONNECT_CLOSED:
        return "Close connection"
    case NETCONNECT_CONNECT_FAILED:
        return "Connect failed"
    case NETCONNECT_CONNECT_SUCCESS:
        return "Connection succeeded"
    case NETCONNECT_CONNECT_REJECTED:
        return "Connection rejected"
    case NETSTREAM_CONNECT_CLOSED:
        return "Connection closed"
    case NETSTREAM_CONNECT_FAILED:
        return "Connection failed"
    case NETSTREAM_CONNECT_SUCCESSS:
        return "Connect Stream suceessed"
    case NETSTREAM_CONNECT_REJECTED:
        return "Reject connect stram"
    }
    return ""
}

type OutputCB func([]byte) error
type OnFrame func(cid codec.CodecID, pts, dts uint32, frame []byte)
type OnStatus func(code, level, describe string)
type OnError func(code, describe string)
type OnReleaseStream func(app, streamName string)
type OnPlay func(app, streamName string, start, duration float64, reset bool) StatusCode
type OnPublish func(app, streamName string) StatusCode
type OnStateChange func(newState RtmpState)
