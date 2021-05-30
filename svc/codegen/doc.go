package codegen

import (
	"encoding/json"
	"fmt"
	"github.com/unionj-cloud/go-doudou/constants"
	"github.com/unionj-cloud/go-doudou/sliceutils"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"
	"unicode"

	"github.com/iancoleman/strcase"
	"github.com/sirupsen/logrus"
	"github.com/unionj-cloud/go-doudou/astutils"
	v3 "github.com/unionj-cloud/go-doudou/openapi/v3"
	"github.com/unionj-cloud/go-doudou/stringutils"
)

var schemas map[string]v3.Schema

/**
bool

string

int  int8  int16  int32  int64
uint uint8 uint16 uint32 uint64 uintptr

byte // alias for uint8

rune // alias for int32
     // represents a Unicode code point

float32 float64

complex64 complex128

TODO 支持匿名结构体
*/
func schemaOf(field astutils.FieldMeta) *v3.Schema {
	ft := strings.TrimPrefix(field.Type, "*")
	switch ft {
	case "int", "int8", "int16", "int32", "uint", "uint8", "uint16", "uint32", "byte", "rune":
		return v3.Int
	case "int64", "uint64", "uintptr":
		return v3.Int64
	case "bool":
		return v3.Bool
	case "string", "error":
		return v3.String
	case "float32":
		return v3.Float32
	case "float64":
		return v3.Float64
	case "complex64", "complex128":
		return v3.Any
	case "multipart.FileHeader":
		return v3.File
	case "":
		return v3.Any
	default:
		if strings.Contains(ft, "map[") {
			elem := ft[strings.Index(ft, "]")+1:]
			elem = strings.TrimPrefix(elem, "*")
			return &v3.Schema{
				Type: v3.ObjectT,
				AdditionalProperties: schemaOf(astutils.FieldMeta{
					Type: elem,
				}),
			}
		}
		if strings.Contains(ft, "[") {
			elem := ft[strings.Index(ft, "]")+1:]
			elem = strings.TrimPrefix(elem, "*")
			return &v3.Schema{
				Type: v3.ArrayT,
				Items: schemaOf(astutils.FieldMeta{
					Type: elem,
				}),
			}
		}
		ft = strings.TrimPrefix(ft, "embed:")
		if !strings.Contains(ft, ".") {
			title := ft
			if unicode.IsUpper(rune(title[0])) {
				return &v3.Schema{
					Ref: "#/components/schemas/" + title,
				}
			}
		}
		if strings.HasPrefix(ft, "vo.") {
			title := strings.TrimPrefix(ft, "vo.")
			return &v3.Schema{
				Ref: "#/components/schemas/" + title,
			}
		}
		return v3.Any
	}
}

func schemasOf(vofile string) []v3.Schema {
	fset := token.NewFileSet()
	root, err := parser.ParseFile(fset, vofile, nil, 0)
	if err != nil {
		panic(err)
	}
	var sc astutils.StructCollector
	ast.Walk(&sc, root)
	var ret []v3.Schema
	for _, item := range sc.Structs {
		if unicode.IsLower(rune(item.Name[0])) {
			continue
		}
		properties := make(map[string]*v3.Schema)
		for _, field := range item.Fields {
			properties[strcase.ToLowerCamel(field.Name)] = schemaOf(field)
		}
		ret = append(ret, v3.Schema{
			Title:      item.Name,
			Type:       v3.ObjectT,
			Properties: properties,
		})
	}
	return ret
}

func vosOf(ic astutils.InterfaceCollector) []string {
	if len(ic.Interfaces) <= 0 {
		return nil
	}
	vomap := make(map[string]int)
	var vos []string
	inter := ic.Interfaces[0]
	for _, method := range inter.Methods {
		for _, field := range method.Params {
			if strings.Contains(field.Type, "*vo.") || strings.Contains(field.Type, "vo.") {
				title := strings.TrimPrefix(strings.TrimPrefix(field.Type, "*"), "vo.")
				if _, ok := vomap[title]; !ok {
					vomap[title] = 1
					vos = append(vos, title)
				}
			}
		}
		for _, field := range method.Results {
			if strings.Contains(field.Type, "*vo.") || strings.Contains(field.Type, "vo.") {
				title := strings.TrimPrefix(strings.TrimPrefix(field.Type, "*"), "vo.")
				if _, ok := vomap[title]; !ok {
					vomap[title] = 1
					vos = append(vos, title)
				}
			}
		}
	}
	return vos
}

const (
	get    = "GET"
	post   = "POST"
	put    = "PUT"
	delete = "DELETE"
)

func IsSimple(field astutils.FieldMeta) bool {
	simples := []interface{}{v3.Int, v3.Int64, v3.Bool, v3.String, v3.Float32, v3.Float64}
	pschema := schemaOf(field)
	return sliceutils.Contains(simples, pschema) || (pschema.Type == v3.ArrayT && sliceutils.Contains(simples, pschema.Items))
}

func operationOf(method astutils.MethodMeta, httpMethod string) v3.Operation {
	var ret v3.Operation
	var params []v3.Parameter

	// If http method is "POST" and each parameters' type is one of v3.Int, v3.Int64, v3.Bool, v3.String, v3.Float32, v3.Float64,
	// then we use application/x-www-form-urlencoded as Content-type and we make one ref schema from them as request body.
	var pschemas []*v3.Schema
	for _, item := range method.Params {
		if IsSimple(item) {
			pschemas = append(pschemas, schemaOf(item))
		}
	}
	if httpMethod == post && len(pschemas) == len(method.Params) {
		title := method.Name + "Req"
		reqSchema := v3.Schema{
			Type:       v3.ObjectT,
			Title:      title,
			Properties: make(map[string]*v3.Schema),
		}
		for _, item := range method.Params {
			key := item.Name
			reqSchema.Properties[strcase.ToLowerCamel(key)] = schemaOf(item)
		}
		schemas[title] = reqSchema
		mt := &v3.MediaType{
			Schema: &v3.Schema{
				Ref: "#/components/schemas/" + title,
			},
		}
		var content v3.Content
		reflect.ValueOf(&content).Elem().FieldByName("FormUrl").Set(reflect.ValueOf(mt))
		ret.RequestBody = &v3.RequestBody{
			Content:  &content,
			Required: true,
		}
	} else {
		// Simple parameters such as v3.Int, v3.Int64, v3.Bool, v3.String, v3.Float32, v3.Float64 and corresponding Array type
		// will be put into query parameter as url search params no matter what http method is.
		// Complex parameters such as structs in vo package, map and corresponding slice/array type
		// will be put into request body as json content type.
		// File and file array parameter will be put into request body as multipart/form-data content type.
		for _, item := range method.Params {
			if item.Type == "context.Context" {
				continue
			}
			pschema := schemaOf(item)
			if reflect.DeepEqual(pschema, v3.FileArray) || pschema == v3.File {
				var content v3.Content
				mt := &v3.MediaType{
					Schema: pschema,
				}
				reflect.ValueOf(&content).Elem().FieldByName("FormData").Set(reflect.ValueOf(mt))
				ret.RequestBody = &v3.RequestBody{
					Content:  &content,
					Required: true,
				}
			} else if IsSimple(item) {
				params = append(params, v3.Parameter{
					Name:   strcase.ToLowerCamel(item.Name),
					In:     v3.InQuery,
					Schema: pschema,
				})
			} else {
				var content v3.Content
				mt := &v3.MediaType{
					Schema: pschema,
				}
				reflect.ValueOf(&content).Elem().FieldByName("Json").Set(reflect.ValueOf(mt))
				ret.RequestBody = &v3.RequestBody{
					Content:  &content,
					Required: true,
				}
			}
		}
	}

	ret.Parameters = params
	var respContent v3.Content
	var hasFile bool
	for _, item := range method.Results {
		if item.Type == "*os.File" {
			hasFile = true
			break
		}
	}
	if hasFile {
		respContent.Stream = &v3.MediaType{
			Schema: v3.File,
		}
	} else {
		title := method.Name + "Resp"
		respSchema := v3.Schema{
			Type:       v3.ObjectT,
			Title:      title,
			Properties: make(map[string]*v3.Schema),
		}
		for _, item := range method.Results {
			key := item.Name
			if stringutils.IsEmpty(key) {
				key = item.Type[strings.LastIndex(item.Type, ".")+1:]
			}
			respSchema.Properties[strcase.ToLowerCamel(key)] = schemaOf(item)
		}
		schemas[title] = respSchema
		respContent.Json = &v3.MediaType{
			Schema: &v3.Schema{
				Ref: "#/components/schemas/" + title,
			},
		}
	}
	ret.Responses = &v3.Responses{
		Resp200: &v3.Response{
			Content: &respContent,
		},
	}
	return ret
}

func pathOf(method astutils.MethodMeta) v3.Path {
	var ret v3.Path
	hm := httpMethod(method.Name)
	op := operationOf(method, hm)
	reflect.ValueOf(&ret).Elem().FieldByName(strings.Title(strings.ToLower(hm))).Set(reflect.ValueOf(&op))
	return ret
}

func pathsOf(ic astutils.InterfaceCollector) map[string]v3.Path {
	if len(ic.Interfaces) <= 0 {
		return nil
	}
	pathmap := make(map[string]v3.Path)
	inter := ic.Interfaces[0]
	for _, method := range inter.Methods {
		v3path := pathOf(method)
		endpoint := fmt.Sprintf("/%s/%s", strings.ToLower(inter.Name), pattern(method.Name))
		pathmap[endpoint] = v3path
	}
	return pathmap
}

// Currently not suport alias type in vo file. TODO
func GenDoc(dir string, ic astutils.InterfaceCollector) {
	var (
		err     error
		svcname string
		docfile string
		vofile  string
		fi      os.FileInfo
		api     v3.Api
		data    []byte
		vos     []v3.Schema
		paths   map[string]v3.Path
	)
	schemas = make(map[string]v3.Schema)
	svcname = ic.Interfaces[0].Name
	docfile = filepath.Join(dir, strings.ToLower(svcname)+"_openapi3.json")
	fi, err = os.Stat(docfile)
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}
	if fi != nil {
		logrus.Warningln("file " + docfile + " will be overwrited")
	}
	vofile = filepath.Join(dir, "vo/vo.go")
	vos = schemasOf(vofile)
	for _, item := range vos {
		schemas[item.Title] = item
	}
	paths = pathsOf(ic)
	api = v3.Api{
		Openapi: "3.0.2",
		Info: &v3.Info{
			Title:          svcname,
			Description:    "",
			TermsOfService: "",
			Contact:        nil,
			License:        nil,
			Version:        fmt.Sprintf("v%s", time.Now().Local().Format(constants.FORMAT10)),
		},
		Paths: paths,
		Components: &v3.Components{
			Schemas: schemas,
		},
	}
	data, err = json.Marshal(api)
	err = ioutil.WriteFile(docfile, data, os.ModePerm)
	if err != nil {
		panic(err)
	}
}
