package tools

import (
	"fmt"
	"testing"
)

func TestCreateNewConfigure(t *testing.T) {
	seretkey := "YzQ1NjI5Zjc2MmVkNTBjY2M2ODFjYzExODNhNDhjYmMyOGUzMjkxZmE0M2QyZTY5ZTczMGIxMGJkZjAyZmM1OA=="

	raw, _ := getP2PIDFromPrivKey(seretkey)
	fmt.Println(raw)
}

//func TestCreateNewConfigure(t *testing.T) {
//	type args struct {
//		start int
//		num   int
//	}
//	tests := []struct {
//		name    string
//		args    args
//		wantErr bool
//	}{
//		{
//			"test create new configuration",
//			args{
//				0,
//				9,
//			},
//			false,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if _, err := CreateNewConfigure(tt.args.start, tt.args.num, "../storage"); (err != nil) != tt.wantErr {
//				t.Errorf("CreateNewConfigure() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}
