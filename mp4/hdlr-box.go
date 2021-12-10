package mp4

// Box Type: 'hdlr'
// Container: Media Box (‘mdia’) or Meta Box (‘meta’)
// Mandatory: Yes
// Quantity: Exactly one

// aligned(8) class HandlerBox extends FullBox(‘hdlr’, version = 0, 0) { unsigned int(32) pre_defined = 0;
// 	unsigned int(32) handler_type;
// 	const unsigned int(32)[3] reserved = 0;
// 	   string   name;
// 	}

type HandlerBox struct {
	Box          *FullBox
	Handler_type [4]byte
	Name         string
}

func NewHandlerBox(handlerType [4]byte, name string) *HandlerBox {
	return &HandlerBox{
		Box:          NewFullBox([4]byte{'h', 'd', 'l', 'r'}, 0),
		Handler_type: handlerType,
		Name:         name,
	}
}

func (hdlr *HandlerBox) Size() uint64 {
	return hdlr.Box.Size() + 16 + uint64(len(hdlr.Name))
}

func (hdlr *HandlerBox) Decode(buf []byte) (offset int, err error) {
	if offset, err = hdlr.Box.Decode(buf); err != nil {
		return 0, err
	}
	_ = buf[hdlr.Box.Box.Size]
	hdlr.Handler_type[0] = buf[offset]
	hdlr.Handler_type[1] = buf[offset+1]
	hdlr.Handler_type[2] = buf[offset+2]
	hdlr.Handler_type[3] = buf[offset+3]
	offset += 4
	hdlr.Name = string(buf[offset:int(hdlr.Box.Box.Size)])
	return int(hdlr.Box.Box.Size), nil
}

func (hdlr *HandlerBox) Encode() (int, []byte) {
	hdlr.Box.Box.Size = hdlr.Size()
	offset, buf := hdlr.Box.Encode()
	buf[offset] = hdlr.Handler_type[0]
	buf[offset+1] = hdlr.Handler_type[1]
	buf[offset+2] = hdlr.Handler_type[2]
	buf[offset+3] = hdlr.Handler_type[3]
	copy(buf[offset:], []byte(hdlr.Name))
	return offset + len(hdlr.Name), buf
}
