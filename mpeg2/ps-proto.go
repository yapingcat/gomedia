package mpeg2

import "../mpeg"

// Table 2-33 – Program Stream pack header
// pack_header() {
// 	pack_start_code 									32      bslbf
// 	'01'            									2 		bslbf
// 	system_clock_reference_base [32..30] 				3 		bslbf
// 	marker_bit                           				1 		bslbf
// 	system_clock_reference_base [29..15] 				15 		bslbf
// 	marker_bit                           				1 		bslbf
// 	system_clock_reference_base [14..0]  				15 		bslbf
// 	marker_bit                           				1 		bslbf
// 	system_clock_reference_extension     				9 		uimsbf
// 	marker_bit                           				1 		bslbf
// 	program_mux_rate                     				22		uimsbf
// 	marker_bit                           				1		bslbf
// 	marker_bit                           				1		bslbf
// 	reserved                             				5		bslbf
// 	pack_stuffing_length                 				3		uimsbf
// 	for (i = 0; i < pack_stuffing_length; i++) {
// 			stuffing_byte                               8       bslbf
// 	}
// 	if (nextbits() == system_header_start_code) {
// 			system_header ()
// 	}
// }

type PSPackHeader struct {
	System_clock_reference_base      uint64 //33 bits
	System_clock_reference_extension uint16 //9 bits
	Program_mux_rate                 uint32 //22 bits
	Pack_stuffing_length             uint8  //3 bits
	Sys_Header                       *System_header
}

func (ps_pkg_hdr *PSPackHeader) Decode(bs *mpeg.BitStream) {
	bs.SkipBits(34)
	ps_pkg_hdr.System_clock_reference_base = bs.GetBits(3)
	bs.SkipBits(1)
	ps_pkg_hdr.System_clock_reference_base = ps_pkg_hdr.System_clock_reference_base<<15 | bs.GetBits(15)
	bs.SkipBits(1)
	ps_pkg_hdr.System_clock_reference_base = ps_pkg_hdr.System_clock_reference_base<<15 | bs.GetBits(15)
	ps_pkg_hdr.System_clock_reference_extension = bs.Uint16(9)
	bs.SkipBits(1)
	ps_pkg_hdr.Program_mux_rate = bs.Uint32(22)
	bs.SkipBits(1)
	bs.SkipBits(1)
	bs.SkipBits(5)
	ps_pkg_hdr.Pack_stuffing_length = bs.Uint8(3)
	bs.SkipBits(int(ps_pkg_hdr.Pack_stuffing_length))

}

func (ps_pkg_hdr *PSPackHeader) Encode(bsw *mpeg.BitStreamWriter) {

}

type Elementary_Stream struct {
	Stream_id                uint8
	P_STD_buffer_bound_scale uint8
	P_STD_buffer_size_bound  uint16
}

type System_header struct {
	Header_length                uint16
	Rate_bound                   uint32
	Audio_bound                  uint8
	Fixed_flag                   uint8
	CSPS_flag                    uint8
	System_audio_lock_flag       uint8
	System_video_lock_flag       uint8
	Video_bound                  uint8
	Packet_rate_restriction_flag uint8
	Streams                      []*Elementary_Stream
}

type Elementary_stream_map struct {
	Stream_type                   uint8
	Elementary_stream_id          uint8
	Elementary_stream_info_length uint16
}

type Program_stream_map struct {
	Map_stream_id                uint8
	Program_stream_map_length    uint16
	Current_next_indicator       uint8
	Program_stream_map_version   uint8
	Program_stream_info_length   uint16
	Elementary_stream_map_length uint16
	Stream_map                   []*Elementary_stream_map
}
