package web

import (
	"std"
	"strconv"
	"strings"
	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

type RouteAction func (*Request) *Response

type Renderer func(interface{}) string

type Renderable interface {
	Render() string
}

type IRouter interface {
	AddRoute(string, RouteAction) bool 
	SetFallback(RouteAction) bool
	Resolve(string) (*Request, RouteAction)	
}

type IRequest interface {	
	Respond(interface{},Renderer) *Response	
	Value(string) string
	Get(string) (interface{}, bool)
	Set(string,interface{})
}

type IResponse interface {
	Render() string
	Get(string) (interface{}, bool)
	Set(string,interface{})
}

type IKV interface {
	Load([]string, avl.Tree) bool
	Add(string, interface{}) bool
	Keys() []string
	Values() avl.Tree
}

type IKR interface {
	Load([]string, []avl.Tree) bool
	Filter([]string, []string, avl.Tree) (*KR, int)
	AddKey(string) bool
	AddRecord(interface{}) bool
	RecordAt(int) avl.Tree
	ValueAt(int, string) interface{}
	Keys() []string
	Records() avl.Tree
	Count() int
}