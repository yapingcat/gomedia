package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"sync"

	"github.com/yapingcat/gomedia/go-codec"
	"github.com/yapingcat/gomedia/go-rtmp"
)

var port = flag.String("port", "1935", "rtmp server listen port")

type MediaCenter map[string]*MediaProducer

var center MediaCenter
var mtx sync.Mutex

func init() {
    center = make(map[string]*MediaProducer)
}

func (c *MediaCenter) register(name string, p *MediaProducer) {
    mtx.Lock()
    defer mtx.Unlock()
    (*c)[name] = p
}

func (c *MediaCenter) unRegister(name string) {
    mtx.Lock()
    defer mtx.Unlock()
    delete(*c, name)
}

func (c *MediaCenter) find(name string) *MediaProducer {
    mtx.Lock()
    defer mtx.Unlock()
    if p, found := (*c)[name]; found {
        return p
    } else {
        return nil
    }
}

type MediaFrame struct {
    cid   codec.CodecID
    frame []byte
    pts   uint32
    dts   uint32
}

type MediaProducer struct {
    name      string
    session   *MediaSession
    mtx       sync.Mutex
    consumers []*MediaSession
    quit      chan struct{}
    die       sync.Once
}

func newMediaProducer(name string, sess *MediaSession) *MediaProducer {
    return &MediaProducer{
        name:      name,
        session:   sess,
        consumers: make([]*MediaSession, 0, 10),
        quit:      make(chan struct{}),
    }
}

func (producer *MediaProducer) start() {
    go producer.dispatch()
}

func (producer *MediaProducer) stop() {
    producer.die.Do(func() {
        close(producer.quit)
        center.unRegister(producer.name)
    })
}

func (producer *MediaProducer) dispatch() {
    defer func() {
        fmt.Println("quit dispatch")
        producer.stop()
    }()
    for {
        select {
        case frame := <-producer.session.C:
            if frame == nil {
                continue
            }
            producer.mtx.Lock()
            tmp := make([]*MediaSession, len(producer.consumers))
            copy(tmp, producer.consumers)
            producer.mtx.Unlock()
            for _, c := range tmp {
                if c.ready() {
                    c.play(frame)
                }
            }
        case <-producer.session.quit:
            return
        case <-producer.quit:
            return
        }
    }
}

func (producer *MediaProducer) addConsumer(consumer *MediaSession) {
    producer.mtx.Lock()
    defer producer.mtx.Unlock()
    producer.consumers = append(producer.consumers, consumer)
}

func (producer *MediaProducer) removeConsumer(id string) {
    producer.mtx.Lock()
    defer producer.mtx.Unlock()
    for i, consume := range producer.consumers {
        if consume.id == id {
            producer.consumers = append(producer.consumers[i:], producer.consumers[i+1:]...)
        }
    }
}

type MediaSession struct {
    handle    *rtmp.RtmpServerHandle
    conn      net.Conn
    lists     []*MediaFrame
    mtx       sync.Mutex
    id        string
    isReady   bool
    frameCome chan struct{}
    quit      chan struct{}
    source    *MediaProducer
    die       sync.Once
    C         chan *MediaFrame
}

func newMediaSession(conn net.Conn) *MediaSession {
    id := fmt.Sprintf("%d", rand.Uint64())
    return &MediaSession{
        id:        id,
        conn:      conn,
        handle:    rtmp.NewRtmpServerHandle(),
        quit:      make(chan struct{}),
        frameCome: make(chan struct{}, 1),
        C:         make(chan *MediaFrame, 30),
    }
}

func (sess *MediaSession) init() {

    sess.handle.OnPlay(func(app, streamName string, start, duration float64, reset bool) rtmp.StatusCode {
        if source := center.find(streamName); source == nil {
            return rtmp.NETSTREAM_PLAY_NOTFOUND
        }
        return rtmp.NETSTREAM_PLAY_START
    })

    sess.handle.OnPublish(func(app, streamName string) rtmp.StatusCode {
        return rtmp.NETSTREAM_PUBLISH_START
    })

    sess.handle.SetOutput(func(b []byte) error {
        _, err := sess.conn.Write(b)
        return err
    })

    sess.handle.OnStateChange(func(newState rtmp.RtmpState) {
        if newState == rtmp.STATE_RTMP_PLAY_START {
            fmt.Println("play start")
            name := sess.handle.GetStreamName()
            source := center.find(name)
            sess.source = source
            if source != nil {
                source.addConsumer(sess)
                fmt.Println("ready to play")
                sess.isReady = true
                go sess.sendToClient()
            }
        } else if newState == rtmp.STATE_RTMP_PUBLISH_START {
            fmt.Println("publish start")
            sess.handle.OnFrame(func(cid codec.CodecID, pts, dts uint32, frame []byte) {
                f := &MediaFrame{
                    cid:   cid,
                    frame: frame, //make([]byte, len(frame)),
                    pts:   pts,
                    dts:   dts,
                }
                //copy(f.frame, frame)
                sess.C <- f
            })
            name := sess.handle.GetStreamName()
            p := newMediaProducer(name, sess)
            go p.dispatch()
            center.register(name, p)
        }
    })
}

func (sess *MediaSession) start() {
    defer sess.stop()
    for {
        buf := make([]byte, 65536)
        n, err := sess.conn.Read(buf)
        if err != nil {
            fmt.Println(err)
            return
        }
        err = sess.handle.Input(buf[:n])
        if err != nil {
            fmt.Println(err)
            return
        }
    }
}

func (sess *MediaSession) stop() {
    sess.die.Do(func() {
        close(sess.quit)
        if sess.source != nil {
            sess.source.removeConsumer(sess.id)
            sess.source = nil
        }
        sess.conn.Close()
    })
}

func (sess *MediaSession) ready() bool {
    return sess.isReady
}

func (sess *MediaSession) play(frame *MediaFrame) {
    sess.mtx.Lock()
    sess.lists = append(sess.lists, frame)
    sess.mtx.Unlock()
    select {
    case sess.frameCome <- struct{}{}:
    default:
    }
}

func (sess *MediaSession) sendToClient() {
    firstVideo := true
    for {
        select {
        case <-sess.frameCome:
            sess.mtx.Lock()
            frames := sess.lists
            sess.lists = nil
            sess.mtx.Unlock()
            for _, frame := range frames {
                if firstVideo { //wait for I frame
                    if frame.cid == codec.CODECID_VIDEO_H264 {
                        if !codec.IsH264IDRFrame(frame.frame) {
                            continue
                        }
                        firstVideo = false
                    } else {
                        continue
                    }
                }
                err := sess.handle.WriteFrame(frame.cid, frame.frame, frame.pts, frame.dts)
                if err != nil {
                    sess.stop()
                    return
                }
            }
        case <-sess.quit:
            return
        }
    }
}

func startRtmpServer() {
    addr := "0.0.0.0:" + *port
    listen, _ := net.Listen("tcp4", addr)
    for {
        conn, _ := listen.Accept()
        sess := newMediaSession(conn)
        sess.init()
        go sess.start()
    }
}

func main() {
    flag.Parse()
    go startRtmpServer()
    select {}
}
