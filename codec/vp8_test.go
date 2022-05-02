package codec

import (
    "fmt"
    "testing"
)

func TestGetResloution(t *testing.T) {
    type args struct {
        frame []byte
    }
    tests := []struct {
        name       string
        args       args
        wantWidth  int
        wantHeight int
        wantErr    bool
    }{
        {name: "test1", wantWidth: 768, wantHeight: 320, wantErr: false, args: args{frame: []byte{0xB0, 0xF0, 0x00, 0x9D, 0x01, 0x2A, 0x00, 0x03, 0x40, 0x01}}},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            gotWidth, gotHeight, err := GetResloution(tt.args.frame)
            fmt.Printf("w:%d,h:%d\n", gotWidth, gotHeight)
            if (err != nil) != tt.wantErr {
                t.Errorf("GetResloution() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if gotWidth != tt.wantWidth {
                t.Errorf("GetResloution() gotWidth = %v, want %v", gotWidth, tt.wantWidth)
            }
            if gotHeight != tt.wantHeight {
                t.Errorf("GetResloution() gotHeight = %v, want %v", gotHeight, tt.wantHeight)
            }
        })
    }
}
