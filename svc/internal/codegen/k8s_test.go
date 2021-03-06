package codegen

import (
	"fmt"
	"github.com/unionj-cloud/go-doudou/pathutils"
	"testing"
)

func TestModifyVersion(t *testing.T) {
	type args struct {
		yfile string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "",
			args: args{
				yfile: pathutils.Abs("./testfiles/k8s.yaml"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := modifyVersion(tt.args.yfile, "v1.0.0")
			fmt.Println(string(result))
		})
	}
}

func TestGenK8s(t *testing.T) {
	type args struct {
		dir     string
		svcname string
		image   string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "",
			args: args{
				dir:     pathutils.Abs("./testfiles"),
				svcname: "corpus",
				image:   "google.com/corpus:v2.0.0",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			GenK8s(tt.args.dir, tt.args.svcname, tt.args.image)
		})
	}
}
