## usage
rtmp库只负责rtmp协议层面的解析，网络链接(一般为tcp)需要调用方自己去管理和收发数据。调用方只需要把从网络中收到的rtmp数据**input**到rtmp库，把rtmp库**output**出来的数据发送到网络即可

rtmp拉流客户端为例子
```golang
//step1 连接远端rtmp服务器
conn,err := net.Dial("tcp4", host)

//step2 创建rtmp客户端句柄，你可以指定chunk大小,握手模式(简单复杂)
client := rtmp.NewRtmpClient(rtmp.WithChunkSize(6000), rtmp.WithComplexHandshake())

//step 3 设置一些必须的回调函数，对于拉流客户端OnFrame 和 SetOutput是必须的
client.OnFrame(func(cid codec.CodecID, pts, dts uint32, frame []byte) {
    //接收到的音视频数据回调
})

client.SetOutput(func(b []byte) error {
    //在output回调函数里面，把数据发送给对端
    _, err := conn.Write(b)
    return err
})

//step 4 调用Start，参数是rtmp url 
client.Start(rtmpUrl)

//step 5 接收网络数据，送入到rtmp库中
for {
    n, err = c.Read(buf)
    if err != nil {
        break
    }
    err = client.Input(buf[:n])
    if err != nil {
        break
    }
}

```