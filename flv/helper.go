package flv

func PutUint24(b []byte, v uint32) {
	_ = b[2]
	b[0] = byte(v >> 16)
	b[1] = byte(v >> 8)
	b[2] = byte(v)
}

func GetUint24(b []byte) (v uint32) {
	_ = b[2]
	v = uint32(b[0])
	v = (v << 8) | uint32(b[1])
	v = (v << 8) | uint32(b[2])
	return v
}
