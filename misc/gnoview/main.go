package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime/debug"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/encoding"
	"github.com/gnolang/gno/logos"
	gno "github.com/gnolang/gno/pkgs/gnolang"
)

var (
	row      = 0
	tc_style = tcell.StyleDefault
)

func bootGnoland() (*gno.PackageValue, *bytes.Buffer) {
	// Create a new machine.
	rr := makeRealmer()
	pn := gno.NewPackageNode("main", "gno.land/r/main", &gno.FileSet{})
	pv := pn.NewPackage(rr)
	rlm := pv.GetRealm()
	out := new(bytes.Buffer)
	m := gno.NewMachineWithOptions(gno.MachineOptions{
		Package:  pv,
		Output:   out,
		Importer: makeImporter(out),
	})

	// Run the file from machine.
	path := "./data/gnoland/main.go"
	bz, err := os.ReadFile(path)
	if err != nil {
		panic("could not read file")
	}
	var rec interface{}
	func() {
		defer func() {
			if rec = recover(); rec != nil {
				fmt.Println("====================")
				fmt.Printf("panic: %v\n", rec)
				debug.PrintStack()
				fmt.Println("====================")
			}
		}()
		n := gno.MustParseFile(path, string(bz))
		fmt.Println("running files")
		m.RunFiles(n)
		fmt.Println("running main")
		m.RunMain()
		fmt.Println("done running main")
	}()
	return pv, out
}

func main() {
	// bootGnoland
	pv, out := bootGnoland()

	encoding.Register()

	// construct screen
	s, e := tcell.NewScreen()
	if e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
	// initialize screen
	if e = s.Init(); e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
	s.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorWhite))
	s.Clear()
	sw, sh := s.Size()
	size := logos.Size{Width: sw, Height: sh}

	// make a buffered stack.
	stack := logos.NewStack(size)
	stack.PushLayer(makeMainPage(pv))
	bstack := logos.NewBufferedElemView(stack, size)
	bstack.Render()
	// fmt.Println(bstack.Sprint())
	// return
	bstack.DrawToScreen(s)

	// recover any panics.
	var rec interface{}
	var recStack []byte

	// show the screen
	quit := make(chan struct{})
	s.Show()
	go func() {
		// capture panics to print error better.
		defer func() {
			if rec = recover(); rec != nil {
				recStack = debug.Stack()
				close(quit)
				return
			}
		}()
		// handle event
		for {
			ev := s.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyCtrlQ:
					close(quit)
					return
				case tcell.KeyCtrlR:
					// TODO somehow make it clearer that it happened.
					bstack.DrawToScreen(s)
					s.Sync()
				default:
					bstack.ProcessEventKey(ev)
					if bstack.Render() {
						bstack.DrawToScreen(s)
						s.Sync()
					}
				}
			case *tcell.EventResize:
				s.Sync()
			}
		}
	}()

	// wait to quit
	<-quit
	s.Fini()
	fmt.Println("goodbye!")

	fmt.Println("====================")
	fmt.Println("out:", out.String())
	if rec != nil {
		fmt.Println("panic:", rec)
		fmt.Println("stacktrace:\n", string(recStack))
	}
	fmt.Println("====================")
	fmt.Println(bstack.StringIndented("  "))
	fmt.Println(bstack.Base.(*logos.Stack).
		Elems[0].(*logos.BufferedElemView).
		Base.(*logos.Page).
		Elems[0].(*logos.TextElem).
		Buffer.Sprint())

	fmt.Println(bstack.Base.(*logos.Stack).
		Elems[0].(*logos.BufferedElemView).
		Buffer.Sprint())
}

func makeRealmer() gno.Realmer {
	rlm := gno.NewRealm("gno.land/r/main") // the root.
	rlm.SetLogRealmOps(true)               // for debug.
	return gno.Realmer(func(pkgPath string) *gno.Realm {
		if pkgPath == "gno.land/r/main" {
			return rlm
		} else {
			panic("should not happen")
		}
	})
}

func makeImporter(out io.Writer) gno.Importer {
	return func(pkgPath string) *gno.PackageValue {
		switch pkgPath {
		case "fmt":
			pkg := gno.NewPackageNode("fmt", "fmt", nil)
			pkg.DefineGoNativeType(reflect.TypeOf((*fmt.Stringer)(nil)).Elem())
			pkg.DefineGoNativeValue("Println", func(a ...interface{}) (n int, err error) {
				res := fmt.Sprintln(a...)
				return out.Write([]byte(res))
			})
			pkg.DefineGoNativeValue("Printf", func(format string, a ...interface{}) (n int, err error) {
				res := fmt.Sprintf(format, a...)
				return out.Write([]byte(res))
			})
			pkg.DefineGoNativeValue("Sprintf", fmt.Sprintf)
			return pkg.NewPackage(nil)
		case "strings":
			pkg := gno.NewPackageNode("strings", "strings", nil)
			pkg.DefineGoNativeValue("SplitN", strings.SplitN)
			pkg.DefineGoNativeValue("HasPrefix", strings.HasPrefix)
			return pkg.NewPackage(nil)
		default:
			panic("unknown package path " + pkgPath)
		}
	}
}

func makeMainPage(pv *gno.PackageValue) logos.Elem {
	elem := MakeValueElem(pv, bstyle)
	return elem
}

func makeMainPage2() *logos.BufferedElemView {
	// make a buffered page.
	ts := `testing
this is a test string.
testing.`
	tc_style := &logos.Style{
		Padding: logos.Padding{2, 2, 2, 2},
		Border:  logos.DefaultBorder(),
	}
	// TODO width shouldn't matter.
	page := logos.NewPage(ts, 84, true, tc_style)
	bpv := logos.NewBufferedElemView(page, logos.Size{})
	return bpv
}
