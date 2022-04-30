# gomedia
 mpeg-ts,mpeg-ps,flv,mp4 muxer/demuxer
 
## H264/H265/AAC
  - decode sps/pps/vps/slice header
  - decode HEVCDecoderConfigurationRecord/AVCDecoderConfigurationRecord/AAC-ADTS/AudioSpecificConfiguration
  - encode HEVCDecoderConfigurationRecord/AVCDecoderConfigurationRecord/AAC-ADTS/AudioSpecificConfiguration
  - decode OPUS Extradata(ID Head "OpusHead") /OPUS Packet(TOC...)
  - encode OPUS Extradata
## mpeg-ts
  - decode H264/H265/AAC
  - encode H264/H265/AAC
## mpeg-ps
  - decode H264/H265/AAC/G711A/G711U
  - encode H264/H265/AAC/G711A/G711U
## flv
  - decode H264/H265/AAC/G711A/G711U
  - encode H264/H265/AAC/G711A/G711U
  
## mp4
  - encode H264/H265/AAC/G711A/G711U
  - decode H264/H265/AAC/G711A/G711U

## ogg
  - decode OPUS

## fmp4
  on the way...





  
