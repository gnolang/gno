package web

import (
	"std"
	"strings"
	"regexp"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/avl"
)

func NewRouter() *Router{
	return &Router{}
}

func NewRoute(pattern_ string, action_ RouteAction) *Route {
	pattern_ = attify(pattern_)	
	_pathParts := trimSplit(pattern_)
	// _parts := []string{}
	// for _, _part := range _pathParts {
	// 	_parts = append(_parts, string(_part))
	// }
	return &Route{
		PatternStr: pattern_,
		Parts: _pathParts,
		PartsLength: len(_pathParts),
		Action: action_, 
	}	
}

func NewRequestExact(path_ string) *Request {
	return &Request{
		Path: path_,
		Keys: []string{},
		Values: avl.Tree{}, 
		Data: avl.Tree{},
	}
}

func NewRequest(path_ string, keys_ []string, values_ avl.Tree) *Request {
	// var _keys []string = []string{}
	// var _values avl.Tree = avl.Tree{}

	// for _k, _part := range _pathParts {
		
	// }
	return &Request{
		Path: path_,
		Keys: keys_,
		Values: values_,
		Data: avl.Tree{},
	}
}

func NewResponse(req_ *Request, body_ string) *Response {
	return &Response{
		req: req_,
		Body: body_,
		Data: req_.Data,
	}
}



func NewKV() *KV {
	return &KV{}
}

func KVLoad(keys_ []string, values_ avl.Tree) *KV {
	_kv := NewKV()
	_kv.Load(keys_,values_)
	return _kv
}

//KR is untested. use at own risk
func NewKR() *KR {
	return &KR{}
}

func KRLoad(keys_ []string, records_ []avl.Tree) *KR {
	_kr := NewKR()
	_kr.Load(keys_,records_)
	return _kr
}

func NewLink(text_ string, link_ string) *Link {
	return &Link{
		Text: text_,
		Link: link_,
	}
}

func NewTemplate(template_ string, keys_ []string, values_ avl.Tree, renderers_ avl.Tree) *Template {
	return &Template{
		TemplateRaw: template_,
		Keys: keys_,
		Values: values_,
		Renderers: renderers_,
	}
}