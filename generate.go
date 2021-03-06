package main

import (
	"fmt"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"net/http"
	"regexp"
	"strings"
)

var ginPkg = protogen.GoImportPath("github.com/gin-gonic/gin")

func generateFile(gen *protogen.Plugin, f *protogen.File) {
	if len(f.Services) == 0 {
		return
	}
	fileName := fmt.Sprintf("%s.ginsev.go", f.GeneratedFilenamePrefix)
	g := gen.NewGeneratedFile(fileName, f.GoImportPath)
	g.P("// Code generated by protoc-gen-ginsev. DO NOT EDIT.")
	g.P()
	g.P("package ", f.GoPackageName)
	g.P()
	g.P("import (")
	g.P("\tgin ", ginPkg.String())
	g.P(")")

	for i := range f.Services {
		generateService(gen, g, f, f.Services[i])
	}
}

func generateService(gen *protogen.Plugin, g *protogen.GeneratedFile, f *protogen.File, sev *protogen.Service) {
	sevComments := getComments(sev.Comments.Leading, sev.Desc.Options().(*descriptorpb.ServiceOptions).GetDeprecated())

	sevName := strings.TrimSuffix(sev.GoName, "Service") + "Service"
	g.P(fmt.Sprintf("%stype %s interface {", sevComments, sevName))
	for _, m := range sev.Methods {
		comments := getComments(m.Comments.Leading, m.Desc.Options().(*descriptorpb.MethodOptions).GetDeprecated())
		g.P(fmt.Sprintf("%s%s (c *gin.Context)", comments, m.GoName))
	}
	g.P("}")
	g.P()
	generateRegister(gen, g, f, sev)
}

func generateRegister(gen *protogen.Plugin, g *protogen.GeneratedFile, f *protogen.File, sev *protogen.Service) {
	sevName := strings.TrimSuffix(sev.GoName, "Service") + "Service"
	registerFuncName := fmt.Sprintf("Register%s", sevName)
	g.P(fmt.Sprintf("// %s register router", registerFuncName))
	g.P(fmt.Sprintf("func %s (router gin.IRouter, service %s) {", registerFuncName, sevName))
	for _, m := range sev.Methods {
		httpRule, _ := proto.GetExtension(m.Desc.Options(), annotations.E_Http).(*annotations.HttpRule)
		method, path := getMethod(sev, m, httpRule)
		g.P(fmt.Sprintf("router.Handle(\"%s\", \"%s\", service.%s)", method, path, m.GoName))
	}
	g.P("}")
}

func getComments(prefix protogen.Comments, deprecated bool) (comments protogen.Comments) {
	if !deprecated {
		return prefix
	}
	if prefix != "" {
		prefix += "\n"
	}
	return prefix + " Deprecated: Do not use.\n"
}

func getMethod(sev *protogen.Service, m *protogen.Method, rule *annotations.HttpRule) (method, path string) {
	if rule == nil {
		return defaultMethodParam(sev, m)
	}

	switch pattern := rule.Pattern.(type) {
	case *annotations.HttpRule_Get:
		path = pattern.Get
		method = http.MethodGet
	case *annotations.HttpRule_Post:
		path = pattern.Post
		method = http.MethodPost
	case *annotations.HttpRule_Put:
		path = pattern.Put
		method = http.MethodPut
	case *annotations.HttpRule_Delete:
		path = pattern.Delete
		method = http.MethodDelete
	case *annotations.HttpRule_Patch:
		path = pattern.Patch
		method = http.MethodPatch
	case *annotations.HttpRule_Custom:
		method = pattern.Custom.GetKind()
		path = pattern.Custom.GetPath()
	default:
		panic("unsupported request type")
	}

	// turn {key} to :key
	paths := strings.Split(path, "/")
	for i, p := range paths {
		if strings.HasPrefix(p, "{") && strings.HasSuffix(p, "}") {
			paths[i] = ":" + p[1:len(p)-1]
		}
	}
	path = strings.Join(paths, "/")

	return
}

func defaultMethodParam(sev *protogen.Service, m *protogen.Method) (method, path string) {
	names := strings.Split(toSnakeCase(m.GoName), "_")

	switch strings.ToUpper(names[0]) {
	case http.MethodGet, "FIND", "QUERY", "LIST", "SEARCH":
		method = http.MethodGet
	case http.MethodPost, "CREATE":
		method = http.MethodPost
	case http.MethodPut, "UPDATE":
		method = http.MethodPut
	case http.MethodPatch:
		method = http.MethodPatch
	case http.MethodDelete:
		method = http.MethodDelete
	default:
		method = "POST"
	}

	// to camelCase
	names[0] = strings.ToLower(names[0])
	for i := 1; i < len(names); i++ {
		names[i] = strings.ToUpper(names[i][:1]) + strings.ToLower(names[i][1:])
	}

	path = fmt.Sprintf("%s/%s", strings.ToLower(strings.TrimRight(sev.GoName, "Service")), strings.Join(names, ""))
	return
}

var matchFirstCap = regexp.MustCompile("([A-Z])([A-Z][a-z])")
var matchAllCap = regexp.MustCompile("([a-z\\d])([A-Z])")

func toSnakeCase(input string) string {
	output := matchFirstCap.ReplaceAllString(input, "${1}_${2}")
	output = matchAllCap.ReplaceAllString(output, "${1}_${2}")
	output = strings.ReplaceAll(output, "-", "_")
	return strings.ToLower(output)
}
