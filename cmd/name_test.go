package cmd

import (
	"github.com/unionj-cloud/go-doudou/astutils"
	"github.com/unionj-cloud/go-doudou/svc"
	"io/ioutil"
	"os"
	"testing"
)

func TestNameCmd(t *testing.T) {
	dir := testDir + "namecmd"
	receiver := svc.Svc{
		Dir: dir,
	}
	receiver.Init()
	defer os.RemoveAll(dir)
	err := os.Chdir(dir)
	if err != nil {
		t.Fatal(err)
	}
	// go-doudou name -f /Users/wubin1989/workspace/chengdutreeyee/team3-cloud-analyse/vo/vo.go -o
	_, _, err = ExecuteCommandC(rootCmd, []string{"name", "-f", dir + "/vo/vo.go", "-o"}...)
	if err != nil {
		t.Fatal(err)
	}
	expect := `// Code generated by go generate; DO NOT EDIT.
// This file was generated by go-doudou
package vo

import (
	"encoding/json"
	"reflect"

	"github.com/unionj-cloud/go-doudou/name/strategies"
)

func (object PageFilter) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	if !reflect.ValueOf(object.Name).IsZero() {
		objectMap[strategies.LowerCaseConvert("Name")] = object.Name
	}
	if !reflect.ValueOf(object.Dept).IsZero() {
		objectMap[strategies.LowerCaseConvert("Dept")] = object.Dept
	}
	return json.Marshal(objectMap)
}

func (object Order) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	if !reflect.ValueOf(object.Col).IsZero() {
		objectMap[strategies.LowerCaseConvert("Col")] = object.Col
	}
	if !reflect.ValueOf(object.Sort).IsZero() {
		objectMap[strategies.LowerCaseConvert("Sort")] = object.Sort
	}
	return json.Marshal(objectMap)
}

func (object Page) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	if !reflect.ValueOf(object.Orders).IsZero() {
		objectMap[strategies.LowerCaseConvert("Orders")] = object.Orders
	}
	if !reflect.ValueOf(object.PageNo).IsZero() {
		objectMap[strategies.LowerCaseConvert("PageNo")] = object.PageNo
	}
	if !reflect.ValueOf(object.Size).IsZero() {
		objectMap[strategies.LowerCaseConvert("Size")] = object.Size
	}
	return json.Marshal(objectMap)
}

func (object PageQuery) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	if !reflect.ValueOf(object.Filter).IsZero() {
		objectMap[strategies.LowerCaseConvert("Filter")] = object.Filter
	}
	if !reflect.ValueOf(object.Page).IsZero() {
		objectMap[strategies.LowerCaseConvert("Page")] = object.Page
	}
	return json.Marshal(objectMap)
}

func (object PageRet) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	if !reflect.ValueOf(object.Items).IsZero() {
		objectMap[strategies.LowerCaseConvert("Items")] = object.Items
	}
	if !reflect.ValueOf(object.PageNo).IsZero() {
		objectMap[strategies.LowerCaseConvert("PageNo")] = object.PageNo
	}
	if !reflect.ValueOf(object.PageSize).IsZero() {
		objectMap[strategies.LowerCaseConvert("PageSize")] = object.PageSize
	}
	if !reflect.ValueOf(object.Total).IsZero() {
		objectMap[strategies.LowerCaseConvert("Total")] = object.Total
	}
	if !reflect.ValueOf(object.HasNext).IsZero() {
		objectMap[strategies.LowerCaseConvert("HasNext")] = object.HasNext
	}
	return json.Marshal(objectMap)
}

func (object UserVo) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	if !reflect.ValueOf(object.Id).IsZero() {
		objectMap[strategies.LowerCaseConvert("Id")] = object.Id
	}
	if !reflect.ValueOf(object.Name).IsZero() {
		objectMap[strategies.LowerCaseConvert("Name")] = object.Name
	}
	if !reflect.ValueOf(object.Phone).IsZero() {
		objectMap[strategies.LowerCaseConvert("Phone")] = object.Phone
	}
	if !reflect.ValueOf(object.Dept).IsZero() {
		objectMap[strategies.LowerCaseConvert("Dept")] = object.Dept
	}
	return json.Marshal(objectMap)
}
`
	marshallerfile := dir + "/vo/vo_marshaller.go"
	f, err := os.Open(marshallerfile)
	if err != nil {
		t.Fatal(err)
	}
	content, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != expect {
		t.Errorf("want %s, got %s\n", expect, string(content))
	}
}

func TestGetImportPath(t *testing.T) {
	dir := testDir + "importpath"
	receiver := svc.Svc{
		Dir: dir,
	}
	receiver.Init()
	defer os.RemoveAll(dir)
	err := os.Chdir(dir)
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		file string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "1",
			args: args{
				file: dir + "/domain",
			},
			want: "testfilesimportpath/domain",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := astutils.GetImportPath(tt.args.file); got != tt.want {
				t.Errorf("GetImportPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
