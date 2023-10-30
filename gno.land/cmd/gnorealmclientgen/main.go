package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"

	"mvdan.cc/gofumpt/format"
)

var supportedContractFuncArgTypes = map[string]struct{}{
	"bool":       {},
	"byte":       {},
	"rune":       {},
	"int":        {},
	"int8":       {},
	"int16":      {},
	"int32":      {},
	"int64":      {},
	"float32":    {},
	"float64":    {},
	"complex64":  {},
	"complex128": {},
	"string":     {},
}

const renderFuncName string = "Render"

type ContractFunction struct {
	FuncName string
	Params   []NameType
	Results  []NameType
}

type NameType struct {
	Name string
	Type string
}

var (
	remote  = flag.String("remote", "http://localhost:26657", "remote node to connect to")
	pkgPath = flag.String("pkgpath", "", "package to generate client for")
)

func main() {

	flag.Parse()

	client := gnoclient.Client{
		RPCClient: rpcclient.NewHTTP(*remote, "/websocket"),
	}

	res, err := client.Query(
		gnoclient.QueryCfg{
			Path: "vm/qfuncs",
			Data: []byte(*pkgPath),
		},
	)
	if err != nil {
		fmt.Println("unable to generate: ", err)
		os.Exit(1)
	}

	var contractFunctions []ContractFunction
	if err := json.Unmarshal(res.Response.Data, &contractFunctions); err != nil {
		fmt.Println("unable to generate: ", err)
		os.Exit(1)
	}

	sb := &strings.Builder{}
	sb.WriteString("package main\n\n")
	sb.WriteString("import (\n")
	sb.WriteString("\"flag\"\n")
	sb.WriteString(")\n\n")
	sb.WriteString("const (\n")
	sb.WriteString("defaultPkgPath string = \"")
	sb.WriteString(*pkgPath)
	sb.WriteString("\"\n")
	sb.WriteString("defaultRemote string = \"")
	sb.WriteString(*remote)
	sb.WriteString("\"\n")
	sb.WriteString(")\n\n")

	for _, contractFunction := range contractFunctions {
		if contractFunction.FuncName == renderFuncName {
			continue
		}

		if err := genFuncOptsAndFlagset(sb, contractFunction); err != nil {
			log.Fatalf("unable to generate: %v", err)
		}

		sb.WriteString("\n\n")
	}

	formattedCode, err := format.Source([]byte(sb.String()), format.Options{LangVersion: "1.21.2"})
	if err != nil {
		log.Fatalf("unable to format code: %v", err)
	}

	fmt.Println(string(formattedCode))
}

func genFuncOptsAndFlagset(sb *strings.Builder, contractFunction ContractFunction) error {

	sb.WriteString("type ")
	sb.WriteString(contractFunction.FuncName)
	sb.WriteString("Opts struct {\n")

	for _, param := range contractFunction.Params {
		if _, ok := supportedContractFuncArgTypes[param.Type]; !ok {
			return fmt.Errorf("unsupported type %s", param.Type)
		}

		sb.WriteString(capitalize(param.Name))
		sb.WriteString(" ")
		sb.WriteString(param.Type)
		sb.WriteString("\n")
	}

	sb.WriteString("}\n")

	sb.WriteString("func (opts *")
	sb.WriteString(contractFunction.FuncName)
	sb.WriteString("Opts) flagSet() *flag.FlagSet {\n")

	sb.WriteString("fs := flag.NewFlagSet(\"")
	sb.WriteString(contractFunction.FuncName)
	sb.WriteString("\", flag.ExitOnError)\n")

	for _, param := range contractFunction.Params {
		sb.WriteString("fs.")
		sb.WriteString(capitalize(param.Type))
		sb.WriteString("Var(&opts.")
		sb.WriteString(capitalize(param.Name))
		sb.WriteString(", \"")
		sb.WriteString(param.Name)
		sb.WriteString("\", ")
		sb.WriteString(defaultValue(param.Type))
		sb.WriteString(", \"\")\n")
	}

	sb.WriteString("return fs\n")
	sb.WriteString("}\n")

	return nil
}

func capitalize(s string) string {
	return strings.ToUpper(s[:1]) + s[1:]
}

func defaultValue(typ string) string {

	switch typ {
	case "bool":
		return "false"
	case "byte":
		return "0"
	case "rune":
		return "0"
	case "int":
		return "0"
	case "int8":
		return "0"
	case "int16":
		return "0"
	case "int32":
		return "0"
	case "int64":
		return "0"
	case "float32":
		return "0"
	case "float64":
		return "0"
	case "complex64":
		return "0"
	case "complex128":
		return "0"
	case "string":
		return "\"\""
	default:
		return "nil"
	}
}
