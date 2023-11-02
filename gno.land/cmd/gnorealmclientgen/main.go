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

func main3() {

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
		fmt.Println("unable to query contract: ", err)
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
	sb.WriteString("\"flag\"\n\n")
	sb.WriteString("\"github.com/gnolang/gno/gno.land/pkg/gnoclient\"\n")
	sb.WriteString("rpcclient \"github.com/gnolang/gno/tm2/pkg/bft/rpc/client\"\n")
	sb.WriteString("keysclient \"github.com/gnolang/gno/tm2/pkg/crypto/keys/client\"\n\n")
	sb.WriteString("\"github.com/peterbourgon/ff/v3/ffcli\"\n")
	sb.WriteString(")\n\n")
	sb.WriteString("const (\n")
	sb.WriteString("defaultPkgPath string = \"")
	sb.WriteString(*pkgPath)
	sb.WriteString("\"\n")
	sb.WriteString("defaultRemote string = \"")
	sb.WriteString(*remote)
	sb.WriteString("\"\n")
	sb.WriteString(")\n\n")

	genRequiredOptsAndFlagset(sb)
	sb.WriteString("\n")

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
	sb.WriteString("requiredOptions\n")

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

	sb.WriteString("fs := opts.requiredOptions.flagSet(\"")
	sb.WriteString(contractFunction.FuncName)
	sb.WriteString("\")\n")

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

func genRequiredOptsAndFlagset(sb *strings.Builder) {
	sb.WriteString(`type requiredOptions struct {` + "\n")
	sb.WriteString(`keysclient.BaseOptions` + "\n")
	sb.WriteString(`GasWanted       int64` + "\n")
	sb.WriteString(`GasFee          string` + "\n")
	sb.WriteString(`ChainID         string` + "\n")
	sb.WriteString(`KeyNameOrBech32 string` + "\n")
	sb.WriteString(`PkgPath         string` + "\n")
	sb.WriteString(`Debug           bool` + "\n")
	sb.WriteString(`Command         string` + "\n")
	sb.WriteString(`CallAction      bool` + "\n")
	sb.WriteString(`QueryAction     bool` + "\n")
	sb.WriteString(`}` + "\n")
	sb.WriteString(`` + "\n")
	sb.WriteString(`func (opts *requiredOptions) flagSet(name string) *flag.FlagSet {` + "\n")
	sb.WriteString(`` + "\n")
	sb.WriteString(`fs := flag.NewFlagSet(name, flag.ExitOnError)` + "\n")
	sb.WriteString(`defaultHome := keysclient.DefaultBaseOptions.Home` + "\n")
	sb.WriteString(`` + "\n")
	sb.WriteString(`fs.BoolVar(&opts.Debug, "debug", false, "verbose output")` + "\n")
	sb.WriteString(`fs.Int64Var(&opts.GasWanted, "gas-wanted", 2000000, "gas requested for tx")` + "\n")
	sb.WriteString(`fs.StringVar(&opts.GasFee, "gas-fee", "1000000ugnot", "gas payment fee")` + "\n")
	sb.WriteString(`fs.StringVar(&opts.ChainID, "chainid", "dev", "")` + "\n")
	sb.WriteString(`fs.StringVar(&opts.PkgPath, "pkgpath", defaultPkgPath, "blog realm path")` + "\n")
	sb.WriteString(`fs.StringVar(&opts.KeyNameOrBech32, "key", "", "key name or bech32 address")` + "\n")
	sb.WriteString(`` + "\n")
	sb.WriteString(`fs.BoolVar(&opts.CallAction, "call", false, "call function")` + "\n")
	sb.WriteString(`fs.BoolVar(&opts.QueryAction, "query", false, "query function")` + "\n")
	sb.WriteString(`` + "\n")
	sb.WriteString(`// keysclient.BaseOptions` + "\n")
	sb.WriteString(`fs.StringVar(&opts.Home, "home", defaultHome, "home directory")` + "\n")
	sb.WriteString(`fs.StringVar(&opts.Remote, "remote", defaultRemote, "remote node URL")` + "\n")
	sb.WriteString(`fs.BoolVar(&opts.Quiet, "quiet", false, "for parsing output")` + "\n")
	sb.WriteString(`fs.BoolVar(&opts.InsecurePasswordStdin, "insecure-password-stdin", false, "WARNING! take password from stdin")` + "\n")
	sb.WriteString(`` + "\n")
	sb.WriteString(`return fs` + "\n")
	sb.WriteString(`}` + "\n")
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
