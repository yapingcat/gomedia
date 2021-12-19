package mp4

import "github.com/yapingcat/gomedia/mpeg"

//List of Class Tags for Descriptors
const (
    Forbidden                           = 0x00
    ObjectDescrTag                      = 0x01
    InitialObjectDescrTag               = 0x02
    ES_DescrTag                         = 0x03
    DecoderConfigDescrTag               = 0x04
    DecSpecificInfoTag                  = 0x05
    SLConfigDescrTag                    = 0x06
    ContentIdentDescrTag                = 0x07
    SupplContentIdentDescrTag           = 0x08
    IPI_DescrPointerTag                 = 0x09
    IPMP_DescrPointerTag                = 0x0A
    IPMP_DescrTag                       = 0x0B
    QoS_DescrTag                        = 0x0C
    RegistrationDescrTag                = 0x0D
    ES_ID_IncTag                        = 0x0E
    ES_ID_RefTag                        = 0x0F
    MP4_IOD_Tag                         = 0x10
    MP4_OD_Tag                          = 0x11
    IPL_DescrPointerRefTag              = 0x12
    ExtensionProfileLevelDescrTag       = 0x13
    profileLevelIndicationIndexDescrTag = 0x14
)

// abstract aligned(8) expandable(228-1) class BaseDescriptor : bit(8) tag=0 {
// 	// empty. To be filled by classes extending this class.
// }

//  int sizeOfInstance = 0;
// 	bit(1) nextByte;
// 	bit(7) sizeOfInstance;
// 	while(nextByte) {
// 		bit(1) nextByte;
// 		bit(7) sizeByte;
// 		sizeOfInstance = sizeOfInstance<<7 | sizeByte;
// }

type BaseDescriptor struct {
    tag            uint8
    sizeOfInstance uint32
}

func (base *BaseDescriptor) Decode(data []byte) {
    bs := mpeg.NewBitStream(data)
    base.tag = bs.Uint8(8)
    nextbit := uint8(1)
    for nextbit == 1 {
        nextbit = bs.GetBit()
        base.sizeOfInstance = base.sizeOfInstance<<7 | bs.Uint32(7)
    }
}

func (base *BaseDescriptor) Encode() []byte {
    bsw := mpeg.NewBitStreamWriter(5)
    bsw.PutByte(base.tag)
    size := base.sizeOfInstance
    bsw.PutUint8(1, 1)
    bsw.PutUint8(uint8(size>>21), 7)
    bsw.PutUint8(1, 1)
    bsw.PutUint8(uint8(size>>14), 7)
    bsw.PutUint8(1, 1)
    bsw.PutUint8(uint8(size>>7), 7)
    bsw.PutUint8(1, 0)
    bsw.PutUint8(uint8(size), 7)
    return bsw.Bits()
}

func makeBaseDescriptor(tag uint8, size uint32) []byte {
    base := BaseDescriptor{
        tag:            tag,
        sizeOfInstance: size,
    }
    return base.Encode()
}
