package vm

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/doc"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
)

func Test_parseQueryEvalData(t *testing.T) {
	t.Parallel()
	tt := []struct {
		input   string
		pkgpath string
		expr    string
	}{
		{
			"gno.land/r/realm.Expression()",
			"gno.land/r/realm",
			"Expression()",
		},
		{
			"a.b/c/d.e",
			"a.b/c/d",
			"e",
		},
		{
			"a.b.c.d.e/c/d.e",
			"a.b.c.d.e/c/d",
			"e",
		},
		{
			"abcde/c/d.e",
			"abcde/c/d",
			"e",
		},
	}
	for _, tc := range tt {
		path, expr := parseQueryEvalData(tc.input)
		assert.Equal(t, tc.pkgpath, path)
		assert.Equal(t, tc.expr, expr)
	}
}

func Test_parseQueryEval_panic(t *testing.T) {
	t.Parallel()

	assert.PanicsWithValue(t, panicInvalidQueryEvalData, func() {
		parseQueryEvalData("gno.land/r/sys/users")
	})
}

func TestVmHandlerQuery_Eval(t *testing.T) {
	tt := []struct {
		input               []byte
		expectedResult      string
		expectedResultMatch string
		expectedErrorMatch  string
		expectedPanicMatch  string
		// XXX: expectedEvents
	}{
		// valid queries
		{input: []byte(`gno.land/r/hello.Echo("hello")`), expectedResult: `("echo:hello" string)`},
		{input: []byte(`gno.land/r/hello.caller()`), expectedResult: `("" .uverse.address)`}, // FIXME?
		{input: []byte(`gno.land/r/hello.GetHeight()`), expectedResult: `(42 int64)`},
		// {input: []byte(`gno.land/r/hello.time.RFC3339`), expectedResult: `test`}, // not working, but should we care?
		{input: []byte(`gno.land/r/hello.PubString`), expectedResult: `("public string" string)`},
		{input: []byte(`gno.land/r/hello.ConstString`), expectedResult: `("const string" string)`},
		{input: []byte(`gno.land/r/hello.pvString`), expectedResult: `("private string" string)`},
		{input: []byte(`gno.land/r/hello.counter`), expectedResult: `(42 int)`},
		{input: []byte(`gno.land/r/hello.GetCounter()`), expectedResult: `(42 int)`},
		{input: []byte(`gno.land/r/hello.Inc()`), expectedResult: `(43 int)`},
		{input: []byte(`gno.land/r/hello.pvEcho("hello")`), expectedResult: `("pvecho:hello" string)`},
		{input: []byte(`gno.land/r/hello.1337`), expectedResult: `(1337 int)`},
		{input: []byte(`gno.land/r/hello.13.37`), expectedResult: `(13.37 float64)`},
		{input: []byte(`gno.land/r/hello.float64(1337)`), expectedResult: `(1337 float64)`},
		{input: []byte(`gno.land/r/hello.myStructInst`), expectedResult: `(struct{(1000 int)} gno.land/r/hello.myStruct)`},
		{input: []byte(`gno.land/r/hello.myStructInst.Foo()`), expectedResult: `("myStruct.Foo" string)`},
		{input: []byte(`gno.land/r/hello.myStruct`), expectedResultMatch: `\(typeval{gno.land/r/hello.myStruct} type{}\)`},
		{input: []byte(`gno.land/r/hello.Inc`), expectedResult: `(Inc func() int)`},
		{input: []byte(`gno.land/r/hello.fn()("hi")`), expectedResult: `("echo:hi" string)`},
		{input: []byte(`gno.land/r/hello.sl`), expectedResultMatch: `(slice[ref(.*)] []int)`},    // XXX: should return the actual value
		{input: []byte(`gno.land/r/hello.sl[1]`), expectedResultMatch: `(slice[ref(.*)] []int)`}, // XXX: should return the actual value
		{input: []byte(`gno.land/r/hello.println(1234)`), expectedResultMatch: `^$`},             // XXX: compare stdout?
		{
			input:          []byte(`gno.land/r/hello.(func() string { return "hello123" + pvString })()`),
			expectedResult: `("hello123private string" string)`,
		},

		// panics
		{input: []byte(`gno.land/r/hello`), expectedPanicMatch: `expected <pkgpath>.<expression> syntax in query input data`},

		// errors
		{input: []byte(`gno.land/r/hello.doesnotexist`), expectedErrorMatch: `^:0:0: name doesnotexist not declared:`}, // multiline error
		{input: []byte(`gno.land/r/doesnotexist.Foo`), expectedErrorMatch: `^invalid package path$`},
		{input: []byte(`gno.land/r/hello.Panic()`), expectedErrorMatch: `^foo$`},
		{input: []byte(`gno.land/r/hello.sl[6]`), expectedErrorMatch: `^slice index out of bounds: 6 \(len=5\)$`},
		{input: []byte(`gno.land/r/hello.func(){ for {} }()`), expectedErrorMatch: `out of gas in location: CPUCycles`},
	}

	for _, tc := range tt {
		name := string(tc.input)
		t.Run(name, func(t *testing.T) {
			env := setupTestEnv()
			ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
			vmHandler := env.vmh

			// Give "addr1" some gnots.
			addr := crypto.AddressFromPreimage([]byte("addr1"))
			acc := env.acck.NewAccountWithAddress(ctx, addr)
			env.acck.SetAccount(ctx, acc)
			env.bankk.SetCoins(ctx, addr, std.MustParseCoins("10000000ugnot"))
			assert.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(std.MustParseCoins("10000000ugnot")))
			const pkgpath = "gno.land/r/hello"
			// Create test package.
			files := []*std.MemFile{
				{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgpath)},
				{Name: "hello.gno", Body: `
package hello

import (
	"chain/runtime"
	"time"
)

var _ = time.RFC3339
func caller() address { return runtime.OriginCaller() }
var GetHeight = runtime.ChainHeight
var sl = []int{1,2,3,4,5}
func fn() func(string) string { return Echo }
type myStruct struct{a int}
var myStructInst = myStruct{a: 1000}
func (ms myStruct) Foo() string { return "myStruct.Foo" }
func Panic() { panic("foo") }
var counter int = 42
var pvString = "private string"
var PubString = "public string"
const ConstString = "const string"
func Echo(msg string) string { return "echo:"+msg }
func GetCounter() int { return counter }
func Inc() int { counter += 1; return counter }
func pvEcho(msg string) string { return "pvecho:"+msg }
`},
			}
			pkgPath := "gno.land/r/hello"
			msg1 := NewMsgAddPackage(addr, pkgPath, files)
			err := env.vmk.AddPackage(ctx, msg1)
			assert.NoError(t, err)
			env.vmk.CommitGnoTransactionStore(ctx)

			defer func() {
				if r := recover(); r != nil {
					output := fmt.Sprintf("%v", r)
					assert.Regexp(t, tc.expectedPanicMatch, output)
				} else {
					assert.Equal(t, tc.expectedPanicMatch, "", "should not panic")
				}
			}()

			req := abci.RequestQuery{
				Path: "vm/qeval",
				Data: tc.input,
			}
			res := vmHandler.Query(env.ctx, req)

			if tc.expectedPanicMatch == "" {
				if tc.expectedErrorMatch == "" {
					assert.True(t, res.IsOK(), "should not have error")
					if tc.expectedResult != "" {
						assert.Equal(t, tc.expectedResult, string(res.Data))
					}
					if tc.expectedResultMatch != "" {
						assert.Regexp(t, tc.expectedResultMatch, string(res.Data))
					}
				} else {
					assert.False(t, res.IsOK(), "should have an error")
					errmsg := res.Error.Error()
					assert.Regexp(t, tc.expectedErrorMatch, errmsg)
				}
			}
		})
	}
}

func TestVmHandlerQuery_Funcs(t *testing.T) {
	tt := []struct {
		input              []byte
		expectedResult     string
		expectedErrorMatch string
	}{
		// valid queries
		{input: []byte(`gno.land/r/hello`), expectedResult: `[{"FuncName":"Panic","Params":null,"Results":null},{"FuncName":"Echo","Params":[{"Name":"msg","Type":"string","Value":""}],"Results":[{"Name":".res.0","Type":"string","Value":""}]},{"FuncName":"GetCounter","Params":null,"Results":[{"Name":".res.0","Type":"int","Value":""}]},{"FuncName":"Inc","Params":null,"Results":[{"Name":".res.0","Type":"int","Value":""}]}]`},
		{input: []byte(`gno.land/r/doesnotexist`), expectedErrorMatch: `invalid package path`},
		{input: []byte(`std`), expectedErrorMatch: `invalid package path`},
		{input: []byte(`strings`), expectedErrorMatch: `invalid package path`},
	}

	for _, tc := range tt {
		name := string(tc.input)
		t.Run(name, func(t *testing.T) {
			env := setupTestEnv()
			ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
			vmHandler := env.vmh

			// Give "addr1" some gnots.
			addr := crypto.AddressFromPreimage([]byte("addr1"))
			acc := env.acck.NewAccountWithAddress(ctx, addr)
			env.acck.SetAccount(ctx, acc)
			env.bankk.SetCoins(ctx, addr, std.MustParseCoins("10000000ugnot"))
			assert.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(std.MustParseCoins("10000000ugnot")))

			const pkgpath = "gno.land/r/hello"
			// Create test package.
			files := []*std.MemFile{
				{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgpath)},
				{Name: "hello.gno", Body: `
package hello

var sl = []int{1,2,3,4,5}
func fn() func(string) string { return Echo }
type myStruct struct{a int}
var myStructInst = myStruct{a: 1000}
func (ms myStruct) Foo() string { return "myStruct.Foo" }
func Panic() { panic("foo") }
var counter int = 42
var pvString = "private string"
var PubString = "public string"
const ConstString = "const string"
func Echo(msg string) string { return "echo:"+msg }
func GetCounter() int { return counter }
func Inc() int { counter += 1; return counter }
func pvEcho(msg string) string { return "pvecho:"+msg }
`},
			}
			pkgPath := "gno.land/r/hello"
			msg1 := NewMsgAddPackage(addr, pkgPath, files)
			err := env.vmk.AddPackage(ctx, msg1)
			assert.NoError(t, err)

			req := abci.RequestQuery{
				Path: "vm/qfuncs",
				Data: tc.input,
			}

			res := vmHandler.Query(env.ctx, req)
			if tc.expectedErrorMatch == "" {
				assert.True(t, res.IsOK(), "should not have error")
				if tc.expectedResult != "" {
					assert.Equal(t, string(res.Data), tc.expectedResult)
				}
			} else {
				assert.False(t, res.IsOK(), "should have an error")
				errmsg := res.Error.Error()
				assert.Regexp(t, tc.expectedErrorMatch, errmsg)
			}
		})
	}
}

func TestVmHandlerQuery_File(t *testing.T) {
	tt := []struct {
		input               []byte
		expectedResult      string
		expectedResultMatch string
		expectedError       error
		expectedPanicMatch  string
		expectedLogMatch    string
		// XXX: expectedEvents
	}{
		// valid queries
		{input: []byte(`gno.land/r/hello/hello.gno`), expectedResult: "package hello\n\nfunc Hello() string { return \"hello\" }\n"},
		{input: []byte(`gno.land/r/hello/README.md`), expectedResult: "# Hello"},
		{
			input:            []byte(`gno.land/r/hello/doesnotexist.gno`),
			expectedError:    &InvalidFileError{},
			expectedLogMatch: `file "gno.land/r/hello/doesnotexist.gno" is not available`,
		},
		{input: []byte(`gno.land/r/hello`), expectedResult: "README.md\ngnomod.toml\nhello.gno"},
		{
			input:            []byte(`gno.land/r/doesnotexist`),
			expectedError:    &InvalidPackageError{},
			expectedLogMatch: `package "gno.land/r/doesnotexist" is not available`,
		},
		{
			input:            []byte(`gno.land/r/doesnotexist/hello.gno`),
			expectedError:    &InvalidFileError{},
			expectedLogMatch: `file "gno.land/r/doesnotexist/hello.gno" is not available`,
		},
	}

	for _, tc := range tt {
		name := string(tc.input)
		t.Run(name, func(t *testing.T) {
			env := setupTestEnv()
			ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
			vmHandler := env.vmh

			// Give "addr1" some gnots.
			addr := crypto.AddressFromPreimage([]byte("addr1"))
			acc := env.acck.NewAccountWithAddress(ctx, addr)
			env.acck.SetAccount(ctx, acc)
			env.bankk.SetCoins(ctx, addr, std.MustParseCoins("10000000ugnot"))
			assert.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(std.MustParseCoins("10000000ugnot")))

			const pkgpath = "gno.land/r/hello"
			// Create test package.
			files := []*std.MemFile{
				{Name: "README.md", Body: "# Hello"},
				{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgpath)},
				{Name: "hello.gno", Body: "package hello\n\nfunc Hello() string { return \"hello\" }\n"},
			}
			pkgPath := "gno.land/r/hello"
			msg1 := NewMsgAddPackage(addr, pkgPath, files)
			err := env.vmk.AddPackage(ctx, msg1)
			assert.NoError(t, err)

			req := abci.RequestQuery{
				Path: "vm/qfile",
				Data: tc.input,
			}

			defer func() {
				if r := recover(); r != nil {
					output := fmt.Sprintf("%v", r)
					assert.Regexp(t, tc.expectedPanicMatch, output)
				} else {
					assert.Equal(t, "", tc.expectedPanicMatch, "should not panic")
				}
			}()
			res := vmHandler.Query(env.ctx, req)

			if tc.expectedError == nil {
				assert.True(t, res.IsOK(), "should not have error")
				if tc.expectedResult != "" {
					assert.Equal(t, string(res.Data), tc.expectedResult)
				}
				if tc.expectedResultMatch != "" {
					assert.Regexp(t, tc.expectedResultMatch, string(res.Data))
				}
			} else {
				assert.False(t, res.IsOK(), "should have an error")
				assert.ErrorIs(t, res.Error, tc.expectedError)
			}

			if tc.expectedLogMatch != "" {
				assert.Regexp(t, tc.expectedLogMatch, res.Log)
			}
		})
	}
}

func TestVmHandlerQuery_Doc(t *testing.T) {
	expected := &doc.JSONDocumentation{
		PackagePath: "gno.land/r/hello",
		PackageLine: "package hello // import \"hello\"",
		PackageDoc:  "hello is a package for testing\n",
		Values: []*doc.JSONValueDecl{
			{
				Signature: "const prefix = \"Hello\"",
				Const:     true,
				Doc:       "The prefix for the hello message\n",
				Values: []*doc.JSONValue{
					{
						Name: "prefix",
						Doc:  "",
						Type: "",
					},
				},
			},
		},
		Funcs: []*doc.JSONFunc{
			{
				Type:      "",
				Name:      "Hello",
				Signature: "func Hello(msg string) (res string)",
				Doc:       "",
				Params: []*doc.JSONField{
					{Name: "msg", Type: "string"},
				},
				Results: []*doc.JSONField{
					{Name: "res", Type: "string"},
				},
			},
			{
				Type:      "myStruct",
				Name:      "Foo",
				Signature: "func (ms myStruct) Foo() string",
				Doc:       "",
				Params:    []*doc.JSONField{},
				Results: []*doc.JSONField{
					{Name: "", Type: "string"},
				},
			},
		},
		Types: []*doc.JSONType{
			{
				Name: "myStruct",
				Type: "struct{ a int }",
				Doc:  "myStruct is a struct for testing\n",
				Kind: "struct",
				Fields: []*doc.JSONField{
					{Name: "a", Type: "int", Doc: ""},
				},
			},
		},
	}

	tt := []struct {
		input              []byte
		expectedResult     string
		expectedErrorMatch string
	}{
		// valid queries
		{input: []byte(`gno.land/r/hello`), expectedResult: expected.JSON()},
		{input: []byte(`gno.land/r/doesnotexist`), expectedErrorMatch: `invalid package path`},
	}

	for _, tc := range tt {
		name := string(tc.input)
		t.Run(name, func(t *testing.T) {
			env := setupTestEnv()
			ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
			vmHandler := env.vmh

			// Give "addr1" some gnots.
			addr := crypto.AddressFromPreimage([]byte("addr1"))
			acc := env.acck.NewAccountWithAddress(ctx, addr)
			env.acck.SetAccount(ctx, acc)
			env.bankk.SetCoins(ctx, addr, std.MustParseCoins("10000000ugnot"))
			assert.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(std.MustParseCoins("10000000ugnot")))

			// Create test package.
			const pkgpath = "gno.land/r/hello"
			// Create test package.
			files := []*std.MemFile{
				{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgpath)},
				{Name: "hello.gno", Body: `
// hello is a package for testing
package hello

// myStruct is a struct for testing
type myStruct struct{a int}
func (ms myStruct) Foo() string { return "myStruct.Foo" }
// The prefix for the hello message
const prefix = "Hello"
func Hello(msg string) (res string) { res = prefix+" "+msg; return }
`},
			}
			pkgPath := "gno.land/r/hello"
			msg1 := NewMsgAddPackage(addr, pkgPath, files)
			err := env.vmk.AddPackage(ctx, msg1)
			assert.NoError(t, err)

			req := abci.RequestQuery{
				Path: "vm/qdoc",
				Data: tc.input,
			}

			res := vmHandler.Query(env.ctx, req)
			if tc.expectedErrorMatch == "" {
				assert.True(t, res.IsOK(), "should not have error")
				if tc.expectedResult != "" {
					assert.Equal(t, tc.expectedResult, string(res.Data))
				}
			} else {
				assert.False(t, res.IsOK(), "should have an error")
				errmsg := res.Error.Error()
				assert.Regexp(t, tc.expectedErrorMatch, errmsg)
			}
		})
	}
}
