package astutils

import (
	"fmt"
	"github.com/iancoleman/strcase"
	"github.com/unionj-cloud/go-doudou/stringutils"
	"go/ast"
	"go/token"
	"strings"
)

type MethodMeta struct {
	Name     string
	Params   []FieldMeta
	Results  []FieldMeta
	Comments []string
}

type InterfaceMeta struct {
	Name     string
	Methods  []MethodMeta
	Comments []string
}

type InterfaceCollector struct {
	Interfaces []InterfaceMeta
	Package    PackageMeta
}

func (ic *InterfaceCollector) Visit(n ast.Node) ast.Visitor {
	return ic.Collect(n)
}

func (sc *InterfaceCollector) Collect(n ast.Node) ast.Visitor {
	switch spec := n.(type) {
	case *ast.Package:
		return sc
	case *ast.File: // actually it is package name
		sc.Package = PackageMeta{
			Name: spec.Name.Name,
		}
		return sc
	case *ast.GenDecl:
		if spec.Tok == token.TYPE {
			var comments []string
			if spec.Doc != nil {
				for _, comment := range spec.Doc.List {
					comments = append(comments, comment.Text)
				}
			}
			for _, item := range spec.Specs {
				typeSpec := item.(*ast.TypeSpec)
				typeName := typeSpec.Name.Name
				switch specType := typeSpec.Type.(type) {
				case *ast.InterfaceType:
					var methods []MethodMeta
					for _, method := range specType.Methods.List {
						if len(method.Names) == 0 {
							panic("no method name")
						}
						mn := method.Names[0].Name

						var mComments []string
						if method.Comment != nil {
							for _, comment := range method.Comment.List {
								mComments = append(mComments, comment.Text)
							}
						}

						var ft *ast.FuncType
						var ok bool
						if ft, ok = method.Type.(*ast.FuncType); !ok {
							panic("not funcType")
						}
						var params, results []FieldMeta
						pkeymap := make(map[string]int)
						for _, param := range ft.Params.List {
							var pn string
							if len(param.Names) > 0 {
								pn = param.Names[0].Name
							}
							pt := exprString(param.Type)
							if stringutils.IsEmpty(pn) {
								elemt := strings.TrimPrefix(pt, "*")
								if stringutils.IsNotEmpty(elemt) {
									if strings.Contains(elemt, "[") {
										elemt = elemt[strings.Index(elemt, "]")+1:]
										elemt = strings.TrimPrefix(elemt, "*")
									}
									splits := strings.Split(elemt, ".")
									_key := "p" + strcase.ToLowerCamel(splits[len(splits)-1][0:1])
									if _, exists := pkeymap[_key]; exists {
										pkeymap[_key]++
										pn = _key + fmt.Sprintf("%d", pkeymap[_key])
									} else {
										pkeymap[_key]++
										pn = _key
									}
								}
							}
							var pComments []string
							if param.Comment != nil {
								for _, comment := range param.Comment.List {
									pComments = append(pComments, comment.Text)
								}
							}
							params = append(params, FieldMeta{
								Name:     pn,
								Type:     pt,
								Tag:      "",
								Comments: pComments,
							})
						}
						if ft.Results != nil {
							rkeymap := make(map[string]int)
							for _, result := range ft.Results.List {
								var rn string
								if len(result.Names) > 0 {
									rn = result.Names[0].Name
								}
								rt := exprString(result.Type)
								if stringutils.IsEmpty(rn) {
									elemt := strings.TrimPrefix(rt, "*")
									if stringutils.IsNotEmpty(elemt) {
										if strings.Contains(elemt, "[") {
											elemt = elemt[strings.Index(elemt, "]")+1:]
											elemt = strings.TrimPrefix(elemt, "*")
										}
										splits := strings.Split(elemt, ".")
										_key := "r" + strcase.ToLowerCamel(splits[len(splits)-1][0:1])
										if _, exists := rkeymap[_key]; exists {
											rkeymap[_key]++
											rn = _key + fmt.Sprintf("%d", rkeymap[_key])
										} else {
											rkeymap[_key]++
											rn = _key
										}
									}
								}
								var rComments []string
								if result.Comment != nil {
									for _, comment := range result.Comment.List {
										rComments = append(rComments, comment.Text)
									}
								}
								results = append(results, FieldMeta{
									Name:     rn,
									Type:     rt,
									Tag:      "",
									Comments: rComments,
								})
							}
						}
						methods = append(methods, MethodMeta{
							Name:     mn,
							Params:   params,
							Results:  results,
							Comments: mComments,
						})
					}

					sc.Interfaces = append(sc.Interfaces, InterfaceMeta{
						Name:     typeName,
						Methods:  methods,
						Comments: comments,
					})
				}
			}
		}
	}
	return nil
}
