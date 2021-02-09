package main

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/encoding"
	"github.com/gnolang/gno/logos"
)

var row = 0
var style = tcell.StyleDefault

func main() {

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
	// plain := tcell.StyleDefault
	// bold := style.Bold(true)
	s.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorWhite))
	s.Clear()

	// construct a page
	ts := makeTestString()
	page := logos.NewPage(ts, 20, true, logos.Style{}) // TODO width shouldn't matter.
	sw, sh := s.Size()
	size := logos.Size{Width: sw, Height: sh}
	bpv := logos.NewBufferedPageView(page, size)
	bpv.Render()
	bpv.DrawToScreen(s)

	// show the screen
	quit := make(chan struct{})
	s.Show()
	go func() {
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
					bpv.DrawToScreen(s)
					s.Sync()
				default:
					bpv.ProcessEventKey(ev)
					if bpv.Render() {
						bpv.DrawToScreen(s)
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
	fmt.Println("charset:", s.CharacterSet())
	fmt.Println("goodbye!")
}

func makeTestString() string {
	s := ""
	putln := func(l string) {
		s += "\n" + l
	}
	// putln("Character set: " + s.CharacterSet())
	putln("Press Ctrl-Q to Exit")
	putln("English:   October")
	putln("Icelandic: október")
	putln("Arabic:    أكتوبر")
	putln("Russian:   октября")
	putln("Greek:     Οκτωβρίου")
	putln("Chinese:   十月 (note, two double wide characters)")
	putln("Combining: A\u030a (should look like Angstrom)")
	putln("Emoticon:  \U0001f618 (blowing a kiss)")
	putln("Airplane:  \u2708 (fly away)")
	putln("Command:   \u2318 (mac clover key)")
	putln("Enclose:   !\u20e3 (should be enclosed exclamation)")
	putln("ZWJ:       \U0001f9db\u200d\u2640 (female vampire)")
	putln("ZWJ:       \U0001f9db\u200d\u2642 (male vampire)")
	putln("Family:    \U0001f469\u200d\U0001f467\u200d\U0001f467 (woman girl girl)\n")
	putln("Region:    \U0001f1fa\U0001f1f8 (USA! USA!)\n")
	putln("")
	putln("Box:")
	putln(string([]rune{
		tcell.RuneULCorner,
		tcell.RuneHLine,
		tcell.RuneTTee,
		tcell.RuneHLine,
		tcell.RuneURCorner,
	}))
	putln(string([]rune{
		tcell.RuneVLine,
		tcell.RuneBullet,
		tcell.RuneVLine,
		tcell.RuneLantern,
		tcell.RuneVLine,
	}) + "  (bullet, lantern/section)")
	putln(string([]rune{
		tcell.RuneLTee,
		tcell.RuneHLine,
		tcell.RunePlus,
		tcell.RuneHLine,
		tcell.RuneRTee,
	}))
	putln(string([]rune{
		tcell.RuneVLine,
		tcell.RuneDiamond,
		tcell.RuneVLine,
		tcell.RuneUArrow,
		tcell.RuneVLine,
	}) + "  (diamond, up arrow)")
	putln(string([]rune{
		tcell.RuneLLCorner,
		tcell.RuneHLine,
		tcell.RuneBTee,
		tcell.RuneHLine,
		tcell.RuneLRCorner,
	}))
	return s
}
