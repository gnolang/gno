package doc

import (
	"strings"
)

type Package struct {
	ImportPath string
	Name       string
	Doc        string
	Filenames  []string
	Funcs      []*Func
	Methods    []*Func
	Vars       []*Value
	Consts     []*Value
	Types      []*Type
	Examples   []*Example
}

func (p *Package) populateType() {
	p.populateTypeWithMethods()
	p.populateTypeWithFuncs()
	p.populateTypeWithValue()
}

func (p *Package) populateTypeWithMethods() {
	for _, t := range p.Types {
		matchedFuncs := make([]*Func, 0)
		remainingFuncs := make([]*Func, 0)
		for _, fn := range p.Funcs {
			if fn.Recv == nil {
				remainingFuncs = append(remainingFuncs, fn)
				continue
			}
			for _, n := range fn.Recv {
				if n == t.Name || "*"+n == t.Name {
					matchedFuncs = append(matchedFuncs, fn)
				} else {
					remainingFuncs = append(remainingFuncs, fn)
				}
			}
		}
		t.Methods = matchedFuncs
		p.Funcs = remainingFuncs
	}
}

func (p *Package) populateTypeWithFuncs() {
	for _, t := range p.Types {
		var matchedFuncs []*Func
		var remainingFuncs []*Func

		for _, fn := range p.Funcs {
			foundMatch := false

			for _, r := range fn.Returns {
				if r.Type == t.Name || strings.HasPrefix(r.Type, "*") && r.Type[1:] == t.Name {
					matchedFuncs = append(matchedFuncs, fn)
					foundMatch = true
					break
				}
			}

			if !foundMatch {
				remainingFuncs = append(remainingFuncs, fn)
			}
		}

		t.Funcs = matchedFuncs
		p.Funcs = remainingFuncs
	}
}

func (p *Package) populateTypeWithValue() {
	for _, t := range p.Types {
		var matchedVars []*Value
		var matchedConsts []*Value

		var remainingVars []*Value
		for _, v := range p.Vars {
			var matched bool
			for _, item := range v.Items {
				if item.Type == t.Name {
					matchedVars = append(matchedVars, v)
					matched = true
					break
				}
			}
			if !matched {
				remainingVars = append(remainingVars, v)
			}
		}
		p.Vars = remainingVars

		var remainingConsts []*Value
		for _, c := range p.Consts {
			var matched bool
			for _, item := range c.Items {
				if item.Type == t.Name {
					matchedConsts = append(matchedConsts, c)
					matched = true
					break
				}
			}
			if !matched {
				remainingConsts = append(remainingConsts, c)
			}
		}
		p.Consts = remainingConsts

		t.Vars = matchedVars
		t.Consts = matchedConsts
	}
}

type Example struct {
	Name   string
	Doc    string
	Code   string
	Output []string
	Play   bool
}

type Value struct {
	Doc       string
	Names     []string
	Items     []*ValueItem
	Signature string
}

type ValueItem struct {
	Doc   string
	Type  string
	Name  string
	Value string
}

type FuncParam struct {
	Type  string
	Names []string
}

type FuncReturn struct {
	Type  string
	Names []string
}

type Func struct {
	Doc       string
	Name      string
	Params    []*FuncParam
	Returns   []*FuncReturn
	Recv      []string
	Signature string
}

type Type struct {
	Doc        string
	Name       string
	Definition string
	Consts     []*Value
	Vars       []*Value
	Funcs      []*Func
	Methods    []*Func
}
