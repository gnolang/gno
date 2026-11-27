package gen

import (
	"fmt"
	"strings"
)

// Seed is one enumerated corpus program.
type Seed struct {
	Category string
	Name     string
	Source   string
}

type item struct {
	cat, name, src string
}

func mainOnly(body string) string {
	return "package main\n\nfunc main() {\n" + body + "}\n"
}

// RawCatalog is the enumerated seed corpus before gofmt.
func RawCatalog() []Seed {
	var items []item
	items = append(items, integerItems()...)
	items = append(items, convertItems()...)
	items = append(items, languageItems()...)
	out := make([]Seed, len(items))
	for i, it := range items {
		out[i] = Seed{Category: it.cat, Name: it.name, Source: it.src}
	}
	return out
}

type width struct {
	name   string
	max    string
	min    string
	signed bool
}

func widths() []width {
	return []width{
		{"int8", "127", "-128", true},
		{"int16", "32767", "-32768", true},
		{"int32", "2147483647", "-2147483648", true},
		{"int64", "9223372036854775807", "-9223372036854775808", true},
		{"int", "9223372036854775807", "-9223372036854775808", true},
		{"uint8", "255", "", false},
		{"uint16", "65535", "", false},
		{"uint32", "4294967295", "", false},
		{"uint64", "18446744073709551615", "", false},
		{"uint", "18446744073709551615", "", false},
		{"byte", "255", "", false},
		{"rune", "2147483647", "-2147483648", true},
	}
}

func integerItems() []item {
	var out []item
	for _, w := range widths() {
		out = append(out, item{"arith", "addwrap_" + w.name, mainOnly(fmt.Sprintf(
			"\tvar a %s = %s\n\tvar one %s = 1\n\tprintln(a + one)\n\tprintln(a + a)\n",
			w.name, w.max, w.name))})
		if w.signed {
			out = append(out, item{"arith", "subwrap_" + w.name, mainOnly(fmt.Sprintf(
				"\tvar a %s = %s\n\tvar one %s = 1\n\tprintln(a - one)\n",
				w.name, w.min, w.name))})
			out = append(out, item{"arith", "mulwrap_" + w.name, mainOnly(fmt.Sprintf(
				"\tvar a %s = %s\n\tvar neg %s = -1\n\tprintln(a * neg)\n",
				w.name, w.min, w.name))})
			out = append(out, item{"arith", "divmod_" + w.name, mainOnly(fmt.Sprintf(
				"\tvar a %s = -5\n\tvar b %s = 2\n\tprintln(a / b)\n\tprintln(a %% b)\n\tvar c %s = 5\n\tprintln(c / b)\n\tprintln(c %% b)\n",
				w.name, w.name, w.name))})
			out = append(out, item{"arith", "shr_" + w.name, mainOnly(fmt.Sprintf(
				"\tvar n %s = -8\n\tprintln(n >> 1)\n\tvar p %s = 8\n\tprintln(p >> 1)\n\tprintln(p << 1)\n",
				w.name, w.name))})
			out = append(out, item{"arith", "div0_" + w.name, mainOnly(fmt.Sprintf(
				"\tvar a %s = 1\n\tvar b %s = 0\n\tprintln(\"before\")\n\tprintln(a / b)\n",
				w.name, w.name))})
			out = append(out, item{"arith", "negshift_" + w.name, mainOnly(fmt.Sprintf(
				"\tvar a %s = 1\n\tvar s %s = -1\n\tprintln(\"before\")\n\tprintln(a << s)\n",
				w.name, w.name))})
		} else {
			out = append(out, item{"arith", "subwrap_" + w.name, mainOnly(fmt.Sprintf(
				"\tvar z %s = 0\n\tvar one %s = 1\n\tprintln(z - one)\n",
				w.name, w.name))})
			out = append(out, item{"arith", "mulwrap_" + w.name, mainOnly(fmt.Sprintf(
				"\tvar a %s = %s\n\tvar two %s = 2\n\tprintln(a * two)\n",
				w.name, w.max, w.name))})
			out = append(out, item{"arith", "divmod_" + w.name, mainOnly(fmt.Sprintf(
				"\tvar a %s = 5\n\tvar b %s = 2\n\tprintln(a / b)\n\tprintln(a %% b)\n",
				w.name, w.name))})
			out = append(out, item{"arith", "shr_" + w.name, mainOnly(fmt.Sprintf(
				"\tvar a %s = %s\n\tprintln(a >> 1)\n\tvar one %s = 1\n\tprintln(one << 1)\n\tprintln(one << 0)\n",
				w.name, w.max, w.name))})
			out = append(out, item{"arith", "div0_" + w.name, mainOnly(fmt.Sprintf(
				"\tvar a %s = 1\n\tvar b %s = 0\n\tprintln(\"before\")\n\tprintln(a / b)\n",
				w.name, w.name))})
		}
		out = append(out, item{"arith", "bits_" + w.name, mainOnly(fmt.Sprintf(
			"\tvar a %s = %s\n\tvar b %s = 1\n\tprintln(a | b)\n\tprintln(a & b)\n\tprintln(a ^ b)\n\tprintln(a &^ b)\n\tprintln(^a)\n",
			w.name, w.max, w.name))})
		out = append(out, item{"arith", "cmp_" + w.name, mainOnly(fmt.Sprintf(
			"\tvar a %s = 1\n\tvar b %s = 2\n\tprintln(a < b)\n\tprintln(a == a)\n\tprintln(a > b)\n\tprintln(a != b)\n\tprintln(a <= a)\n",
			w.name, w.name))})
	}
	return out
}

func convertItems() []item {
	ws := widths()
	var out []item
	for _, src := range ws {
		for _, dst := range ws {
			vals := []string{"0", "1", src.max}
			if src.signed {
				vals = append(vals, src.min, "-1")
			}
			var body strings.Builder
			for i, v := range vals {
				body.WriteString(fmt.Sprintf("\tvar a%d %s = %s\n\tprintln(%s(a%d))\n", i, src.name, v, dst.name, i))
			}
			out = append(out, item{
				"convert",
				src.name + "_to_" + dst.name,
				mainOnly(body.String()),
			})
		}
	}
	return out
}
