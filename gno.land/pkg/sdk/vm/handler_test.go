package vm

import (
	"fmt"
	"testing"

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
		parseQueryEvalData("gno.land/r/demo/users")
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
		{input: []byte(`gno.land/r/hello.myStruct`), expectedResultMatch: `\(typeval{gno.land/r/hello.myStruct \(0x.*\)} type{}\)`},
		{input: []byte(`gno.land/r/hello.Inc`), expectedResult: `(Inc func()( int))`},
		{input: []byte(`gno.land/r/hello.fn()("hi")`), expectedResult: `("echo:hi" string)`},
		{input: []byte(`gno.land/r/hello.sl`), expectedResultMatch: `(slice[ref(.*)] []int)`},    // XXX: should return the actual value
		{input: []byte(`gno.land/r/hello.sl[1]`), expectedResultMatch: `(slice[ref(.*)] []int)`}, // XXX: should return the actual value
		{input: []byte(`gno.land/r/hello.println(1234)`), expectedResultMatch: `^$`},             // XXX: compare stdout?

		// panics
		{input: []byte(`gno.land/r/hello`), expectedPanicMatch: `expected <pkgpath>.<expression> syntax in query input data`},

		// errors
		{input: []byte(`gno.land/r/hello.doesnotexikst`), expectedErrorMatch: `^/0:0: name doesnotexist not declared:`}, // multiline error
		{input: []byte(`gno.land/r/doesnotexist.Foo`), expectedErrorMatch: `^invalid package path$`},
		{input: []byte(`gno.land/r/hello.Panic()`), expectedErrorMatch: `^foo$`},
		{input: []byte(`gno.land/r/hello.panic("bar")`), expectedErrorMatch: `^bar$`},
		{input: []byte(`gno.land/r/hello.sl[6]`), expectedErrorMatch: `^slice index out of bounds: 6 \(len=5\)$`},
	}

	for _, tc := range tt {
		name := string(tc.input)
		t.Run(name, func(t *testing.T) {
			env := setupTestEnv()
			ctx := env.ctx
			vmHandler := env.vmh

			// Give "addr1" some gnots.
			addr := crypto.AddressFromPreimage([]byte("addr1"))
			acc := env.acck.NewAccountWithAddress(ctx, addr)
			env.acck.SetAccount(ctx, acc)
			env.bank.SetCoins(ctx, addr, std.MustParseCoins("10000000ugnot"))
			assert.True(t, env.bank.GetCoins(ctx, addr).IsEqual(std.MustParseCoins("10000000ugnot")))

			// Create test package.
			files := []*std.MemFile{
				{"hello.gno", `
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
				Path: "vm/qeval",
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
			if tc.expectedErrorMatch == "" {
				assert.True(t, res.IsOK(), "should not have error")
				if tc.expectedResult != "" {
					assert.Equal(t, string(res.Data), tc.expectedResult)
				}
				if tc.expectedResultMatch != "" {
					assert.Regexp(t, tc.expectedResultMatch, string(res.Data))
				}
			} else {
				assert.False(t, res.IsOK(), "should have an error")
				errmsg := res.Error.Error()
				assert.Regexp(t, tc.expectedErrorMatch, errmsg)
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
		{input: []byte(`gno.land/r/hello`), expectedResult: `[{"FuncName":"Panic","Params":null,"Results":null},{"FuncName":"Echo","Params":[{"Name":"msg","Type":"string","Value":""}],"Results":[{"Name":"_","Type":"string","Value":""}]},{"FuncName":"GetCounter","Params":null,"Results":[{"Name":"_","Type":"int","Value":""}]},{"FuncName":"Inc","Params":null,"Results":[{"Name":"_","Type":"int","Value":""}]}]`},
		{input: []byte(`gno.land/r/doesnotexist`), expectedErrorMatch: `invalid package path`},
		{input: []byte(`std`), expectedErrorMatch: `invalid package path`},
		{input: []byte(`strings`), expectedErrorMatch: `invalid package path`},
	}

	for _, tc := range tt {
		name := string(tc.input)
		t.Run(name, func(t *testing.T) {
			env := setupTestEnv()
			ctx := env.ctx
			vmHandler := env.vmh

			// Give "addr1" some gnots.
			addr := crypto.AddressFromPreimage([]byte("addr1"))
			acc := env.acck.NewAccountWithAddress(ctx, addr)
			env.acck.SetAccount(ctx, acc)
			env.bank.SetCoins(ctx, addr, std.MustParseCoins("10000000ugnot"))
			assert.True(t, env.bank.GetCoins(ctx, addr).IsEqual(std.MustParseCoins("10000000ugnot")))

			// Create test package.
			files := []*std.MemFile{
				{"hello.gno", `
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
		expectedErrorMatch  string
		expectedPanicMatch  string
		// XXX: expectedEvents
	}{
		// valid queries
		{input: []byte(`gno.land/r/hello/hello.gno`), expectedResult: "package hello\nfunc Hello() string { return \"hello\" }"},
		{input: []byte(`gno.land/r/hello/README.md`), expectedResult: "# Hello"},
		{input: []byte(`gno.land/r/hello/doesnotexist.gno`), expectedErrorMatch: `file "gno.land/r/hello/doesnotexist.gno" is not available`},
		{input: []byte(`gno.land/r/hello`), expectedResult: "README.md\nhello.gno"},
		{input: []byte(`gno.land/r/doesnotexist`), expectedErrorMatch: `package "gno.land/r/doesnotexist" is not available`},
		{input: []byte(`gno.land/r/doesnotexist/hello.gno`), expectedErrorMatch: `file "gno.land/r/doesnotexist/hello.gno" is not available`},
	}

	for _, tc := range tt {
		name := string(tc.input)
		t.Run(name, func(t *testing.T) {
			env := setupTestEnv()
			ctx := env.ctx
			vmHandler := env.vmh

			// Give "addr1" some gnots.
			addr := crypto.AddressFromPreimage([]byte("addr1"))
			acc := env.acck.NewAccountWithAddress(ctx, addr)
			env.acck.SetAccount(ctx, acc)
			env.bank.SetCoins(ctx, addr, std.MustParseCoins("10000000ugnot"))
			assert.True(t, env.bank.GetCoins(ctx, addr).IsEqual(std.MustParseCoins("10000000ugnot")))

			// Create test package.
			files := []*std.MemFile{
				{"README.md", "# Hello"},
				{"hello.gno", "package hello\nfunc Hello() string { return \"hello\" }"},
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
			if tc.expectedErrorMatch == "" {
				assert.True(t, res.IsOK(), "should not have error")
				if tc.expectedResult != "" {
					assert.Equal(t, string(res.Data), tc.expectedResult)
				}
				if tc.expectedResultMatch != "" {
					assert.Regexp(t, tc.expectedResultMatch, string(res.Data))
				}
			} else {
				assert.False(t, res.IsOK(), "should have an error")
				errmsg := res.Error.Error()
				assert.Regexp(t, tc.expectedErrorMatch, errmsg)
			}
		})
	}
}
