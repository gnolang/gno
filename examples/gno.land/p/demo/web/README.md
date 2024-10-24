# `web` Framework

This is a simple web framework made to simplify routing/rendering logic. Made for Gno, it could likely be generalized into a handy grab-n-go Go native framework as well with the addition of an HTTP handling/serving component. 

## Core Structs

### `Router`

The central component, the router can have new routes added with `AddRoute(string,RouteAction)`, and can also have a fallback/404 renderer set with `SetFallback(RouteAction)`. To resolve a path, simply call `Resolve(string)` from the realm's `Render(path)` function, which returns a `Request` pointer and the `RouteAction` function associated with the `Route`. Paths support keys in the form of `{someVar}` which are then available via the `Request` object returned on resolve. 

Example usage: 

```
_r := web.NewRouter()
_r.SetFallback(NotFound)
_r.AddRoute("",Home)
_r.AddRoute("{communeID}/proposals/{proposalID}",ProposalProfile)
_r.AddRoute("{communeID}",CommuneProfile)

[...]
// inside realm's Render(path) function 
_req, _action := _r.Resolve(path_)
return _action(_req).Render()
```

### `Route`

The `Route` struct is how the `Resolve` function is able to determine matching routes. It tracks the `PatternStr` as passed to the `AddRoute` call, `Parts` as a `[]string`, the number of parts at `PartsLength`, and the `RouteAction` conforming function is stored at `Action`. 

When matching routes, first the list of routes is iterated through looking for exact matches between the request's path and the `PatternStr`. If no exact match is found it iterates through a second time, first comparing `PartsLength` to the request's split part length to skip routes that obviously won't match, and then comparing each part first exactly, then as a key, tallying the matched number of parts for final comparison.

### `Request`

The `Request` struct tracks a string `Path` representing the path the user is requesting, along with `Keys` (`[]string`), `Values` (`avl.Tree`) and `Data` (`avl.Tree`) fields. 

URL matched keys are accessible with request struct's `Value(key)` function. The struct also has `Set` and `Get` functions mapping to the same functions on the `Data` tree to load arbitrary data onto the request object. Most importantly, the request struct has a `Respond(interface{}, Renderer)` function that takes in data and the desired `Renderer` conforming function and returns a `Response` pointer.

### `Response`

The `Response` struct tracks the `Request` pointer at  `req`, and contains `Body` (`string`) and `Data` (`avl.Tree`) fields. Like the `Request` struct, `Set` and `Get` functions mapping to the same functions on the `Data` tree also exist on `Response` to load arbitrary data onto the response object. Its `Render()` function is used to return the contents of `Body`, which is set in the request's `Respond` function using the `Renderer` and data passed.


### `KV`

The `KV` struct simplifies dealing with keyed lists, providing an iterable `[]string` field at `keys`, and an `avl.Tree` at `values`. It has an `Add(string,interface{})` function to add new k=>v state, a `Keys()` function to return the keys, and `Values()` to return the values. This simplifies some things, particularly templating. A future version will likely leverage it internally (like in `Request` or `Template` for their keys/values). You can generate a new empty KV pointer with `NewKV()`, or alternatively can load a keys string and values avl.Tree into a KV wrapped version with `KVLoad([]string, avl.Tree)`.

## Function Types

## `RouteAction`

The `RouteAction` type is what devs will implement for handling their app logic and mapping it to a template or otherwise generating the response body. It takes in a `Request` pointer, and returns a `Response` pointer by calling `Respond(interface{},Renderer)` on the passed request struct. 

## `Renderer`

A renderer function is any that takes in an `interface{}` param, and returns a `string`. This allows for both an assortment of base renderers like `Stringer`, `Selfer`, `Linker`, etc (see `renderers.gno` for full list), as well as custom render functions, which could then be used to handle view partials before final/top-level template rendering.

## `RouteAction` & `Renderer` Usage Example

Below is an example of a `RouteAction` function as implemented by `CommuneProfile`, along with how `KV` and the `web.Templater` `Renderer` might be used together.

```
router.AddRoute("{communeID}",CommuneProfile)

[...]

const viewCommune string = 
`# {communeID}
  
Admin: {admin}

Link: {link}

List: 
  {list}


Link List: 

  {linkList}`

[...]    

func CommuneProfile(req_ *web.Request) *web.Response {
	_communeStr := req_.Value("communeID")
	_communeID := identity.IDString(_communeStr)
	_exists := daoRegistry.Exists(_communeID)
	if !_exists {
		_res := ufmt.Sprintf("You are on commune: %s. Does not exist.",_communeStr)
		return req_.Respond(_res, web.Stringer)
	}

	_kv := web.NewKV()	
	_d := daoRegistry.DAO(_communeID)
	
	_list := []string{"one","three","two","apple"}
	_link := web.NewLink("google","https://google.com")
	_links := []*web.Link{_link,_link,_link}


	// return req_.Respond(_d, web.Selfer)
	_kv.Add("communeID", _communeID)
	_kv.Add("admin",string(_d.Identity.Account()))
	_kv.Add("link",web.Linker(_link))
	_kv.Add("list",web.Lister(_list))
	_kv.Add("linkList",web.LinkLister(_links))
	_template := web.NewTemplate(viewCommune, _kv.Keys(), _kv.Values(), avl.Tree{})	

	return req_.Respond(_template, web.Templater)
}
```

## Renderer Related Structs

### `Link`

This struct type expects a `Text` field and a `Link` field, and is used by the `Linker` and `LinkLister` renderers. Links can be generated with `web.NewLink(text_,link_)`.

### `Template`

This struct type expects a `TemplateRaw` field containing a string template, a `Keys` field that's `[]string`, a `Values` field that's an `avl.Tree` of the keyed values, and a `Renderers` field that's an `avl.Tree` tracking per-key Renderers. This is used by the `Templater` and `MarkdownTemplater` renderers (currently the same thing). Templates can be generated with `web.NewTemplate(raw_, keys_, values_, renderers_)`.

The `Templater` renderer uses `{}` wrapped keys in the `TemplateRaw`, and then for each key simply calls `ReplaceAll` on the wrapped key, replacing it with the data from `Values`, processed through a `Renderer` (defaults to `Stringer` if no key-renderer specified). The `viewCommune` const in the above snippet shows an example of what this looks like (could also be loaded from template files).

Alternate method to the above using per-key `Renderer` functions: 

```
_renderers := avl.Tree{}
_renderers.Set("link",web.Linker)
_renderers.Set("list",web.Lister)
_renderers.Set("linkList",web.LinkLister)

[...]

_kv.Add("link",_link)
_kv.Add("list",_list)
_kv.Add("linkList",_links)

_template := web.NewTemplate(viewCommune, _kv.Keys(), _kv.Values(), _renderers)	
return req_.Respond(_template, web.Templater)
```

This method allows for default/app level key renderers to be defined, enabling consistent rules like using `web.H1` for any `{title}` keys, or assigning a custom renderer to `{username}` that converts it to a link to their profile (which could leverage `web.Linker` internally once creating the user's `Link` struct).

### Markdown Renderers

Effectively a carbon copy of the functions defined in the `ui` package, the following functions have been adapted to conform to the `Renderer` function requirements, allowing them to be used like any other Renderer. This could be especially helpful for per-key renderers, allowing for template files that focus on structure rather than style while style is moved to a global template key=>Renderer tree.

```
func H1(data_ interface{}) string     	{ return "# " + Stringer(data_) + "\n" }
func H2(data_ interface{}) string     	{ return "## " + Stringer(data_) + "\n" }
func H3(data_ interface{}) string     	{ return "### " + Stringer(data_) + "\n" }
func H4(data_ interface{}) string     	{ return "#### " + Stringer(data_) + "\n" }
func H5(data_ interface{}) string     	{ return "##### " + Stringer(data_) + "\n" }
func H6(data_ interface{}) string     	{ return "###### " + Stringer(data_) + "\n" }
func Bold(data_ interface{}) string   	{ return "**" + Stringer(data_) + "**"} 
func Italic(data_ interface{}) string 	{ return "_" + Stringer(data_) + "_"} 
func Code(data_ interface{}) string   	{ return "`" + Stringer(data_) + "`"} 
func HR(data_ interface{}) string       { return "\n---\n"}
```

## Utility Functions

While other internal utility functions exist, the public facing functions are most of note and for the moment deal mostly with type conversions. String=>Int conversions under the hood just use `strconv.Atoi` before casting/returning the desired type. Supported conversion functions are:

- `Str(interface{}) string` <- just shorthand for `Stringer`
- `Int(interface{}) int`
- `UI64(int) uint64`
- `I64(int) int64`
- `UI32(int) uint32`
- `I32(int) int32`
- `UI16(int) uint16`
- `I16(int) int16`
- `UI8(int) uint8`
- `I8(int) int8`
- `StrUI64(string) uint64` 
- `StrI64(string) int64` 
- `StrUI32(string) uint32` 
- `StrI32(string) int32` 
- `StrUI16(string) uint16` 
- `StrI16(string) int16` 
- `StrUI8(string) uint8` 
- `StrI8(string) int8` 

## Experimental Structs

### `KR`

Not used/tested yet, `KR` extends the idea of the `KV` construct into records of values. It tracks `keys` as `[]string` like `KV`, and `records` as a `[]avl.Tree`. It has `AddKey(string)` and `AddRecord(avl.Tree)` functions, as well as getters for both `Keys()` and `Records()`. It also has both `RecordAt(int)` and `ValueAt(int, string)` functions for selecting records/values at known indexes. It can also return the number of records with `Count()`.

More ambitiously, it contains a `Filter(selKeys_ []string, whereKeys_ []string, whereValues_ avl.Tree)` function that returns a new `KR` pointer of the resultset, along with a count of the number of matches. Internally it compares the results of the values after passing them through `Stringer` to avoid type issues (while likely creating some more).

Again... this is not tested. Assume borked. 