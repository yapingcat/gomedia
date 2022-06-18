# gomedia
 mpeg-ts,mpeg-ps,flv,mp4 muxer/demuxer
 
## Installation
```
go get github.com/yapingcat/gomedia
```
 
## H264/H265/AAC/VP8/OPUS
 [USAGE](https://github.com/yapingcat/gomedia/blob/main/codec/README.md)
  - decode sps/pps/vps/slice header
  - decode HEVCDecoderConfigurationRecord/AVCDecoderConfigurationRecord/AAC-ADTS/AudioSpecificConfiguration
  - encode HEVCDecoderConfigurationRecord/AVCDecoderConfigurationRecord/AAC-ADTS/AudioSpecificConfiguration
  - decode OPUS Extradata(ID Head "OpusHead") /OPUS Packet(TOC...)
  - encode OPUS Extradata
  - decode VP8 Frame Tag/Key Frame Head

## mpeg-ts
  - mux
    - H264
    - H265
    - AAC
  - demux
    - H264
    - H265
    - AAC

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
  - demux 
    - H264
    - H265
    - AAC
    - G711A
    - G711U
  
## mp4
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
  
  [USAGE](https://github.com/yapingcat/gomedia/blob/main/rtmp/README.md)
  
  - support client/server
  - support play/publish
  - support h264/h265/aac/g711a/g711u
 

## fmp4
  on the way...





  
