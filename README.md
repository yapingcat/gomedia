# gomedia
 mpeg-ts,mpeg-ps,flv,mp4,rtmp muxer/demuxer
 
## Installation
```
go get github.com/yapingcat/gomedia
```


## H264/H265/AAC/VP8/OPUS/MP3
 [USAGE](https://github.com/yapingcat/gomedia/blob/main/go-codec/README.md)
  - decode sps/pps/vps/slice header
  - decode HEVCDecoderConfigurationRecord/AVCDecoderConfigurationRecord/AAC-ADTS/AudioSpecificConfiguration
  - encode HEVCDecoderConfigurationRecord/AVCDecoderConfigurationRecord/AAC-ADTS/AudioSpecificConfiguration
  - decode OPUS Extradata(ID Head "OpusHead") /OPUS Packet(TOC...)
  - encode OPUS Extradata
  - decode VP8 Frame Tag/Key Frame Head
  - decode MP3 Frame head

## mpeg-ts
  - mux
    - H264
    - H265
    - AAC
    - MP3
  - demux
    - H264
    - H265
    - AAC
    - MP3

## mpeg-ps
  - mux 
    - H264
    - H265
    - AAC
    - G711A
    - G711U
  - demux 
    - H264
    - H265
    - AAC
    - G711A
    - G711U
   
## flv
  - mux 
    - H264
    - H265
    - AAC
    - G711A
    - G711U
    - MP3
  - demux 
    - H264
    - H265
    - AAC
    - G711A
    - G711U
    - MP3
  
## mp4
  - demux 
    - H264
    - H265
    - AAC
    - G711A
    - G711U
    - MP3
  - mux 
    - H264
    - H265
    - AAC
    - G711A
    - G711U
    - MP3
    - OPUS


## fmp4
  - demux 
    - H264
    - H265
    - AAC
    - G711A
    - G711U
  - mux 
    - H264
    - H265
    - AAC
    - G711A
    - G711U

## ogg
  - demux 
    - OPUS
    - VP8
  
## rtmp
  
  [USAGE](https://github.com/yapingcat/gomedia/blob/main/go-rtmp/README.md)
  
  - support client/server
  - support play/publish
  - support h264/h265/aac/g711a/g711u/mp3
  
  
## rtsp

  - support client/server(rfc2326)
  - support basic/digest
  - support rtp(rfc3550)
  - support g711/aac/h264/h265
 





  
