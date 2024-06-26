package web

import (
	"std"
	"strings"
	"regexp"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/avl"
)

// BASE RENDERER FUNCTIONS

func Stringer(data_ interface{}) string {
	return string(data_)
}

func Selfer(data_ interface{}) string {
	return Renderable(data_).Render()
}

func Linker(data_ interface{}) string {
	_link := data_.(*Link)
	return "[" + _link.Text + "](" + _link.Link + ")"
}

func Lister(data_ interface{}) string {
	_items := data_.([]string)
	_list := ""
	for _k, _item := range _items {
		_list += "* " + _item + "\n"
	}
	return _list
}

func LinkLister(data_ interface{}) string {
	_items := []string{}
	_links := data_.([]*Link)
	for _i, _link := range _links {
		_linked := Linker(_link)
		_items = append(_items, _linked)
	}

	return Lister(_items) 
}

//ie: likely the most used renderer, mapping KV=>placeholders
func Templater(data_ interface{}) string {
	_template := data_.(*Template)
	_tmp := _template.TemplateRaw 
	for _k, _key := range _template.Keys {
		_value, _ := _template.Values.Get(_key)
		_key = string(_key)
		_keyed := keyify(_key)	
		_r, _exists := _template.Renderers.Get(_key)
		if _exists {			
			_renderer := _r.(func(interface{})(string))
			_tmp = strings.ReplaceAll(_tmp, _keyed, _renderer(_value))
		} else {
			_tmp = strings.ReplaceAll(_tmp, _keyed, Stringer(_value))
		}		
	}
	return _tmp
}

func MarkdownRaw(data_ interface{}) string {
	return Stringer(data_) // think this would actually just be same.
}

func MarkdownTemplater(data_ interface{}) string {
	return Templater(data_) // ditto for this
}

//// Markdown Helpers/Renderers 
//// (adapted from ui package to comply with Renderer 
//// Note re: HR: keys with an HR key renderer will need an arbitrary value for the key 
//// or it will be skipped

func H1(data_ interface{}) string     	{ return "# " + Stringer(data_) + "\n" }
func H2(data_ interface{}) string     	{ return "## " + Stringer(data_) + "\n" }
func H3(data_ interface{}) string     	{ return "### " + Stringer(data_) + "\n" }
func H4(data_ interface{}) string     	{ return "#### " + Stringer(data_) + "\n" }
func H5(data_ interface{}) string     	{ return "##### " + Stringer(data_) + "\n" }
func H6(data_ interface{}) string     	{ return "###### " + Stringer(data_) + "\n" }
func Bold(data_ interface{}) string   	{ return "**" + Stringer(data_) + "**"} 
func Italic(data_ interface{}) string 	{ return "_" + Stringer(data_) + "_"} 
func Code(data_ interface{}) string   	{ return "`" + Stringer(data_) + "`"} 
func HR(data_ interface{}) string	  	{ return "\n---\n"}

// RENDERER SPECIFIC TYPES

type Link struct {
	Text string 
	Link string
}

type Template struct {
	TemplateRaw string 
	Keys []string 
	Values avl.Tree
	Renderers avl.Tree
}

// TODO: noodle on "loaders" to simplify loading template files?

func TemplateLoader() string {
	panic("Not yet implemented!")
	return ""
}

func MarkdownLoader() string {
	panic("Not yet implemented!")
	return ""
}