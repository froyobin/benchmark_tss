package remote

import "testing"

func TestRunCommand(t *testing.T) {
	type args struct {
		ip           string
		pemLocation  string
		ins          string
		digitalOcean bool
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"test run remote command",
			args{
				"128.199.106.202",
				"./",
				"date",
				true,
			},
			"",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RunCommand(tt.args.ip, tt.args.pemLocation, tt.args.ins, tt.args.digitalOcean)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RunCommand() got = %v, want %v", got, tt.want)
			}
		})
	}
}
