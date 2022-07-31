package codec

import (
	"reflect"
	"testing"
)

var nalu []byte = []byte{0x00, 0x00, 0x00, 0x01, 0x40, 0x01, 0x0C, 0x01,
	0xFF, 0xFF, 0x01, 0x60, 0x00, 0x00, 0x03, 0x00,
	0x90, 0x00, 0x00, 0x03, 0x00, 0x00, 0x03, 0x00,
	0x78, 0x99, 0x98, 0x09}

var result []byte = []byte{0x00, 0x00, 0x00, 0x01, 0x40, 0x01, 0x0C, 0x01,
	0xFF, 0xFF, 0x01, 0x60, 0x00, 0x00, 0x00, 0x90, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x78, 0x99, 0x98, 0x09}

func TestCovertRbspToSodb(t *testing.T) {
	type args struct {
		rbsp []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{name: "test", args: args{
			rbsp: nalu,
		}, want: result},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CovertRbspToSodb(tt.args.rbsp); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CovertRbspToSodb() = %x, want %x", got, tt.want)
			}
		})
	}
}

func TestFindStartCode(t *testing.T) {
	type args struct {
		nalu   []byte
		offset int
	}
	tests := []struct {
		name  string
		args  args
		want  int
		want1 START_CODE_TYPE
	}{
		{name: "test1", args: args{
			nalu:   []byte{0x00, 0x00, 0x00, 0x01, 0x67},
			offset: 0,
		}, want: 0, want1: START_CODE_4},
		{name: "test2", args: args{
			nalu:   []byte{0x00, 0x00, 0x01, 0x67},
			offset: 0,
		}, want: 0, want1: START_CODE_3},
		{name: "test3", args: args{
			nalu:   []byte{0x99, 0x00, 0x00, 0x01, 0x67},
			offset: 0,
		}, want: 1, want1: START_CODE_3},
		{name: "test4", args: args{
			nalu:   []byte{0x99, 0x00, 0x00, 0x00, 0x01, 0x67},
			offset: 0,
		}, want: 1, want1: START_CODE_4},
		{name: "test5", args: args{
			nalu:   []byte{0x99, 0x88, 0x77, 0x00, 0x01, 0x67},
			offset: 0,
		}, want: -1, want1: START_CODE_3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := FindStartCode(tt.args.nalu, tt.args.offset)
			if got != tt.want {
				t.Errorf("FindStartCode() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("FindStartCode() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
