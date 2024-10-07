package web

import (
	"std"
	"strconv"
	"strings"
	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

// Example Usage: 
// _router := NewRouter()
// _router.SetFallback(RouteAction)
// _router.AddRoute(string, RouteAction)
// ...
// _req, _action := _router.Resolve(path_)
// return _action(_req).Render()


// ROUTE

type Route struct{
	PatternStr string
	Parts []string
	PartsLength int	
	Action RouteAction		
}

// ROUTER

type Router struct{
	routes []*Route
	fallback RouteAction
}

func (r *Router) AddRoute(pattern_ string, action_ RouteAction) bool {
	_route := NewRoute(pattern_, action_)
	r.routes = append(r.routes, _route)
	return true
}

func (r *Router) SetFallback(action_ RouteAction) bool {
	r.fallback = action_
	return true
}

func (r *Router) Resolve(path_ string) (*Request, RouteAction) {
	path_ = attify(path_)
	// Loop through first looking for exact string pattern matches
	for _, _route := range r.routes {
		if path_ == _route.PatternStr {
			_req := NewRequestExact(path_)
			return _req, _route.Action
		}		
	}	
	
    _reqParts := trimSplit(path_)
	_reqLength := len(_reqParts)

	// If no exact match, loop through processing pattern keys
	for _, _route := range r.routes {
		if _reqLength != _route.PartsLength {				
			continue			
		}		
		
		_matches := 0
		_keys := []string{}
		_values := avl.Tree{}
		for _i, _rtPart := range _route.Parts {
			if _rtPart == _reqParts[_i] { 
				_matches++ 
				_keys = append(_keys, _rtPart)
				_values.Set(_rtPart,_rtPart)
				continue
			}

			_k := partKey(_rtPart)
			if _k != "" {
				_matches++ 
				_keys = append(_keys, _k)
				_values.Set(_k,_reqParts[_i])
			} else {
				//see `subPathResolve` for start of subpath support. will likely cull
				//example url of idea: `a:{ID1}/{ID2}:b` matching `a:noun/person:b`
			}			
		}

		if _matches == _route.PartsLength {
			_req := NewRequest(path_, _keys, _values)
			return _req, _route.Action
		}		
	}	

	if r.fallback != nil {
		_req := NewRequestExact(path_)
		return _req, r.fallback
	}
	
	panic("Router: No matching route")	
}

// REQUEST

type Request struct{
	Path string
	Keys []string
	Values avl.Tree
	Data avl.Tree
}

func (r *Request) Get(key_ string) (interface{}, bool) {
	return r.Data.Get(key_)	
}

func (r *Request) Set(key_ string, value_ interface{}) {
	r.Data.Set(key_, value_)	
}

func (r *Request) Value(key_ string) string {
	_value, _exists := r.Values.Get(key_)
	if !_exists {
		return ""
	}

	return string(_value)
}

func (r *Request) Respond(data_ interface{}, renderer_ Renderer) *Response {
	_body := renderer_(data_)
	return NewResponse(r, _body)
}

// RESPONSE

type Response struct{
	req *Request
	Body string
	Data avl.Tree
}

//TODO: determine if Response cloning Req data makes sense/is useful, otherwise cull?
func (r *Response) Render() string{
	return r.Body
}

func (r *Response) Get(key_ string) (interface{}, bool) {
	return r.Data.Get(key_)
}

func (r *Response) Set(key_ string, value_ interface{}) {
	r.Data.Set(key_, value_)
}

// KV 

type KV struct {
	keys []string
	values avl.Tree
}

func (k *KV) Load(keys_ []string, values_ avl.Tree) bool {
	k.keys = keys_
	k.values = values_
	return true
}

func (k *KV) Add(key_ string, value_ interface{}) bool {
	if k.values.Has(key_) {
		return false
	}
	k.keys = append(k.keys, key_)
	k.values.Set(key_, value_)
	return true
}

func (k *KV) Keys() []string {	
	return k.keys
}

func (k *KV) Values() avl.Tree {	
	return k.values
}

// KR (untested, use at own risk)
//TODO out of bounds/existance checking

type KR struct {
	keys []string
	records []avl.Tree
}

func (k *KR) Load(keys_ []string, records_ []avl.Tree) bool {
	k.keys = keys_
	k.records = records_
	return true
}

func (k *KR) Count() int {
	return len(k.records)
}

func (k *KR) AddKey(key_ string) bool {
	k.keys = append(k.keys, key_)
	return true
}

func (k *KR) AddRecord(record_ avl.Tree) bool {
	k.records = append(k.records, record_)
	return true
}

func (k *KR) RecordAt(idx_ int) avl.Tree {	
	if idx_ < 0 || idx_ > (k.Count()-1) {
		panic(ufmt.Sprintf("Invalid record index requested: %d",idx_))
	}
	return k.records[idx_]
}

func (k *KR) ValueAt(idx_ int, key_ string) interface{} {	
	_v, _e := k.RecordAt(idx_).Get(key_)
	if !_e {
		panic(ufmt.Sprintf("Invalid value requested: %d,%s",idx_,key_))
	}
	return _v
}

func (k *KR) Filter(selKeys_ []string, whereKeys_ []string, whereValues_ avl.Tree) (*KR, int) {	
	_kr := KRLoad(selKeys_, []avl.Tree{})
	_whereCount := len(whereKeys_)
	_matchedRecordCount := 0
	for _c, _record := range k.records {
		_whereMatches := 0		
		for _wk, _whereKey := range whereKeys_ {			
			_v, _e := _record.Get(_whereKey)
			_qv, _qe := whereValues_.Get(_whereKey)
			//key doesn't exist at record or in the where (would be weird, but possible)
			if !_e || !_qe {
				break
			}
			//value at record doesn't match where value (TODO: add operand support?)			
			if Stringer(_qv) != Stringer(_v) {
				break
			}			
			_whereMatches++
		}		
		if _whereMatches == _whereCount {
			_sel := avl.Tree{}
			if len(selKeys_) == 0 {
				//NOTE: this will have a weird behaviour, returning a KR with no keys.
				//spose the object being filtered would still have its keys, but worth noting
				_sel = _record
			}else {
				for _sk, _selKey := range selKeys_ {
					_sv, _se := _record.Get(_selKey)
					_sel.Set(_selKey,_sv)
				}
			}			
			_kr.AddRecord(_sel)	
			_matchedRecordCount++		
		}
	}
	return _kr, _matchedRecordCount
}


func (k *KR) Keys() []string {	
	return k.keys
}

func (k *KR) Records() []avl.Tree {	
	return k.records
}
