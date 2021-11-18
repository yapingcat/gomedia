package mpeg

type H265NaluHdr struct {
    Forbidden_zero_bit uint8
    Nal_ref_idc        uint8
    Nal_unit_type      uint8
}

func (hdr *H265NaluHdr) Decode(bs *BitStream) {
    hdr.Forbidden_zero_bit = bs.GetBit()
    hdr.Nal_ref_idc = bs.Uint8(2)
    hdr.Nal_unit_type = bs.Uint8(5)
}
