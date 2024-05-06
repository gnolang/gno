package json

import (
	"bytes"
	"encoding/json"
	"testing"
)

type testNode struct {
	name  string
	input []byte
	_type ValueType
	value []byte
}

func simpleValid(test *testNode, t *testing.T) {
	root, err := Unmarshal(test.input)
	if err != nil {
		t.Errorf("Error on Unmarshal(%s): %s", test.input, err.Error())
	} else if root == nil {
		t.Errorf("Error on Unmarshal(%s): root is nil", test.name)
	} else if root.nodeType != test._type {
		t.Errorf("Error on Unmarshal(%s): wrong type", test.name)
	} else if !bytes.Equal(root.source(), test.value) {
		t.Errorf("Error on Unmarshal(%s): %s != %s", test.name, root.source(), test.value)
	}
}

func simpleInvalid(test *testNode, t *testing.T) {
	root, err := Unmarshal(test.input)
	if err == nil {
		t.Errorf("Error on Unmarshal(%s): error expected, got '%s'", test.name, root.source())
	} else if root != nil {
		t.Errorf("Error on Unmarshal(%s): root is not nil", test.name)
	}
}

func simpleCorrupted(name string) *testNode {
	return &testNode{name: name, input: []byte(name)}
}

func TestUnmarshal_StringSimpleSuccess(t *testing.T) {
	tests := []*testNode{
		{name: "blank", input: []byte("\"\""), _type: String, value: []byte("\"\"")},
		{name: "char", input: []byte("\"c\""), _type: String, value: []byte("\"c\"")},
		{name: "word", input: []byte("\"cat\""), _type: String, value: []byte("\"cat\"")},
		{name: "spaces", input: []byte("  \"good cat or dog\"\r\n "), _type: String, value: []byte("\"good cat or dog\"")},
		{name: "backslash", input: []byte("\"good \\\"cat\\\"\""), _type: String, value: []byte("\"good \\\"cat\\\"\"")},
		{name: "backslash 2", input: []byte("\"good \\\\\\\"cat\\\"\""), _type: String, value: []byte("\"good \\\\\\\"cat\\\"\"")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			simpleValid(test, t)
		})
	}
}

func TestUnmarshal_NumericSimpleSuccess(t *testing.T) {
	tests := []*testNode{
		{name: "1", input: []byte("1"), _type: Number, value: []byte("1")},
		{name: "-1", input: []byte("-1"), _type: Number, value: []byte("-1")},

		{name: "1234567890", input: []byte("1234567890"), _type: Number, value: []byte("1234567890")},
		{name: "-123", input: []byte("-123"), _type: Number, value: []byte("-123")},

		{name: "123.456", input: []byte("123.456"), _type: Number, value: []byte("123.456")},
		{name: "-123.456", input: []byte("-123.456"), _type: Number, value: []byte("-123.456")},

		{name: "1e3", input: []byte("1e3"), _type: Number, value: []byte("1e3")},
		{name: "1e+3", input: []byte("1e+3"), _type: Number, value: []byte("1e+3")},
		{name: "1e-3", input: []byte("1e-3"), _type: Number, value: []byte("1e-3")},
		{name: "-1e3", input: []byte("-1e3"), _type: Number, value: []byte("-1e3")},
		{name: "-1e-3", input: []byte("-1e-3"), _type: Number, value: []byte("-1e-3")},

		{name: "1.123e3456", input: []byte("1.123e3456"), _type: Number, value: []byte("1.123e3456")},
		{name: "1.123e-3456", input: []byte("1.123e-3456"), _type: Number, value: []byte("1.123e-3456")},
		{name: "-1.123e3456", input: []byte("-1.123e3456"), _type: Number, value: []byte("-1.123e3456")},
		{name: "-1.123e-3456", input: []byte("-1.123e-3456"), _type: Number, value: []byte("-1.123e-3456")},

		{name: "1E3", input: []byte("1E3"), _type: Number, value: []byte("1E3")},
		{name: "1E-3", input: []byte("1E-3"), _type: Number, value: []byte("1E-3")},
		{name: "-1E3", input: []byte("-1E3"), _type: Number, value: []byte("-1E3")},
		{name: "-1E-3", input: []byte("-1E-3"), _type: Number, value: []byte("-1E-3")},

		{name: "1.123E3456", input: []byte("1.123E3456"), _type: Number, value: []byte("1.123E3456")},
		{name: "1.123E-3456", input: []byte("1.123E-3456"), _type: Number, value: []byte("1.123E-3456")},
		{name: "-1.123E3456", input: []byte("-1.123E3456"), _type: Number, value: []byte("-1.123E3456")},
		{name: "-1.123E-3456", input: []byte("-1.123E-3456"), _type: Number, value: []byte("-1.123E-3456")},

		{name: "-1.123E-3456 with spaces", input: []byte(" \r -1.123E-3456 \t\n"), _type: Number, value: []byte("-1.123E-3456")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root, err := Unmarshal(test.input)
			if err != nil {
				t.Errorf("Error on Unmarshal(%s): %s", test.name, err.Error())
			} else if root == nil {
				t.Errorf("Error on Unmarshal(%s): root is nil", test.name)
			} else if root.nodeType != test._type {
				t.Errorf("Error on Unmarshal(%s): wrong type", test.name)
			} else if !bytes.Equal(root.source(), test.value) {
				t.Errorf("Error on Unmarshal(%s): %s != %s", test.name, root.source(), test.value)
			}
		})
	}
}

func TestUnmarshal_StringSimpleCorrupted(t *testing.T) {
	tests := []*testNode{
		{name: "white NL", input: []byte("\"foo\nbar\"")},
		{name: "white R", input: []byte("\"foo\rbar\"")},
		{name: "white Tab", input: []byte("\"foo\tbar\"")},
		{name: "wrong quotes", input: []byte("'cat'")},
		{name: "double string", input: []byte("\"Hello\" \"World\"")},
		{name: "quotes in quotes", input: []byte("\"good \"cat\"\"")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			simpleInvalid(test, t)
		})
	}
}

func TestUnmarshal_ObjectSimpleSuccess(t *testing.T) {
	tests := []*testNode{
		{name: "{}", input: []byte("{}"), _type: Object, value: []byte("{}")},
		{name: `{ \r\n }`, input: []byte("{ \r\n }"), _type: Object, value: []byte("{ \r\n }")},
		{name: `{"key":1}`, input: []byte(`{"key":1}`), _type: Object, value: []byte(`{"key":1}`)},
		{name: `{"key":true}`, input: []byte(`{"key":true}`), _type: Object, value: []byte(`{"key":true}`)},
		{name: `{"key":"value"}`, input: []byte(`{"key":"value"}`), _type: Object, value: []byte(`{"key":"value"}`)},
		{name: `{"foo":"bar","baz":"foo"}`, input: []byte(`{"foo":"bar", "baz":"foo"}`), _type: Object, value: []byte(`{"foo":"bar", "baz":"foo"}`)},
		{name: "spaces", input: []byte(`  {  "foo"  :  "bar"  , "baz"   :   "foo"   }    `), _type: Object, value: []byte(`{  "foo"  :  "bar"  , "baz"   :   "foo"   }`)},
		{name: "nested", input: []byte(`{"foo":{"bar":{"baz":{}}}}`), _type: Object, value: []byte(`{"foo":{"bar":{"baz":{}}}}`)},
		{name: "array", input: []byte(`{"array":[{},{},{"foo":[{"bar":["baz"]}]}]}`), _type: Object, value: []byte(`{"array":[{},{},{"foo":[{"bar":["baz"]}]}]}`)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			simpleValid(test, t)
		})
	}
}

func TestUnmarshal_ObjectSimpleCorrupted(t *testing.T) {
	tests := []*testNode{
		simpleCorrupted("{{{\"key\": \"foo\"{{{{"),
		simpleCorrupted("}"),
		simpleCorrupted("{ }}}}}}}"),
		simpleCorrupted(" }"),
		simpleCorrupted("{,}"),
		simpleCorrupted("{:}"),
		simpleCorrupted("{100000}"),
		simpleCorrupted("{1:1}"),
		simpleCorrupted("{'1:2,3:4'}"),
		simpleCorrupted(`{"d"}`),
		simpleCorrupted(`{"foo"}`),
		simpleCorrupted(`{"foo":}`),
		simpleCorrupted(`{:"foo"}`),
		simpleCorrupted(`{"foo":bar}`),
		simpleCorrupted(`{"foo":"bar",}`),
		simpleCorrupted(`{}{}`),
		simpleCorrupted(`{},{}`),
		simpleCorrupted(`{[},{]}`),
		simpleCorrupted(`{[,]}`),
		simpleCorrupted(`{[]}`),
		simpleCorrupted(`{}1`),
		simpleCorrupted(`1{}`),
		simpleCorrupted(`{"x"::1}`),
		simpleCorrupted(`{null:null}`),
		simpleCorrupted(`{"foo:"bar"}`),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			simpleInvalid(test, t)
		})
	}
}

func TestUnmarshal_NullSimpleCorrupted(t *testing.T) {
	tests := []*testNode{
		{name: "nul", input: []byte("nul")},
		{name: "nil", input: []byte("nil")},
		{name: "nill", input: []byte("nill")},
		{name: "NILL", input: []byte("NILL")},
		{name: "Null", input: []byte("Null")},
		{name: "NULL", input: []byte("NULL")},
		{name: "spaces", input: []byte("Nu ll")},
		{name: "null1", input: []byte("null1")},
		{name: "double", input: []byte("null null")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			simpleInvalid(test, t)
		})
	}
}

func TestUnmarshal_BoolSimpleSuccess(t *testing.T) {
	tests := []*testNode{
		{name: "lower true", input: []byte("true"), _type: Boolean, value: []byte("true")},
		{name: "lower false", input: []byte("false"), _type: Boolean, value: []byte("false")},
		{name: "spaces true", input: []byte("  true\r\n "), _type: Boolean, value: []byte("true")},
		{name: "spaces false", input: []byte("  false\r\n "), _type: Boolean, value: []byte("false")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			simpleValid(test, t)
		})
	}
}

func TestUnmarshal_BoolSimpleCorrupted(t *testing.T) {
	tests := []*testNode{
		simpleCorrupted("tru"),
		simpleCorrupted("fals"),
		simpleCorrupted("tre"),
		simpleCorrupted("fal se"),
		simpleCorrupted("true false"),
		simpleCorrupted("True"),
		simpleCorrupted("TRUE"),
		simpleCorrupted("False"),
		simpleCorrupted("FALSE"),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			simpleInvalid(test, t)
		})
	}
}

func TestUnmarshal_ArraySimpleSuccess(t *testing.T) {
	tests := []*testNode{
		{name: "[]", input: []byte("[]"), _type: Array, value: []byte("[]")},
		{name: "[1]", input: []byte("[1]"), _type: Array, value: []byte("[1]")},
		{name: "[1,2,3]", input: []byte("[1,2,3]"), _type: Array, value: []byte("[1,2,3]")},
		{name: "[1, 2, 3]", input: []byte("[1, 2, 3]"), _type: Array, value: []byte("[1, 2, 3]")},
		{name: "[1,[2],3]", input: []byte("[1,[2],3]"), _type: Array, value: []byte("[1,[2],3]")},
		{name: "[[],[],[]]", input: []byte("[[],[],[]]"), _type: Array, value: []byte("[[],[],[]]")},
		{name: "[[[[[]]]]]", input: []byte("[[[[[]]]]]"), _type: Array, value: []byte("[[[[[]]]]]")},
		{name: "[true,null,1,\"foo\",[]]", input: []byte("[true,null,1,\"foo\",[]]"), _type: Array, value: []byte("[true,null,1,\"foo\",[]]")},
		{name: "spaces", input: []byte("\n\r [\n1\n ]\r\n"), _type: Array, value: []byte("[\n1\n ]")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			simpleValid(test, t)
		})
	}
}

func TestUnmarshal_ArraySimpleCorrupted(t *testing.T) {
	tests := []*testNode{
		simpleCorrupted("[,]"),
		simpleCorrupted("[]\\"),
		simpleCorrupted("[1,]"),
		simpleCorrupted("[[]"),
		simpleCorrupted("[]]"),
		simpleCorrupted("1[]"),
		simpleCorrupted("[]1"),
		simpleCorrupted("[[]1]"),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			simpleInvalid(test, t)
		})
	}
}

// Examples from https://json.org/example.html
func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{
			name: "glossary",
			value: `{
				"glossary": {
					"title": "example glossary",
					"GlossDiv": {
						"title": "S",
						"GlossList": {
							"GlossEntry": {
								"ID": "SGML",
								"SortAs": "SGML",
								"GlossTerm": "Standard Generalized Markup Language",
								"Acronym": "SGML",
								"Abbrev": "ISO 8879:1986",
								"GlossDef": {
									"para": "A meta-markup language, used to create markup languages such as DocBook.",
									"GlossSeeAlso": ["GML", "XML"]
								},
								"GlossSee": "markup"
							}
						}
					}
				}
			}`,
		},
		{
			name: "menu",
			value: `{"menu": {
				"id": "file",
				"value": "File",
				"popup": {
				  "menuitem": [
					{"value": "New", "onclick": "CreateNewDoc()"},
					{"value": "Open", "onclick": "OpenDoc()"},
					{"value": "Close", "onclick": "CloseDoc()"}
				  ]
				}
			}}`,
		},
		{
			name: "widget",
			value: `{"widget": {
				"debug": "on",
				"window": {
					"title": "Sample Konfabulator Widget",
					"name": "main_window",
					"width": 500,
					"height": 500
				},
				"image": { 
					"src": "Images/Sun.png",
					"name": "sun1",
					"hOffset": 250,
					"vOffset": 250,
					"alignment": "center"
				},
				"text": {
					"data": "Click Here",
					"size": 36,
					"style": "bold",
					"name": "text1",
					"hOffset": 250,
					"vOffset": 100,
					"alignment": "center",
					"onMouseUp": "sun1.opacity = (sun1.opacity / 100) * 90;"
				}
			}}    `,
		},
		{
			name: "web-app",
			value: `{"web-app": {
				"servlet": [   
				  {
					"servlet-name": "cofaxCDS",
					"servlet-class": "org.cofax.cds.CDSServlet",
					"init-param": {
					  "configGlossary:installationAt": "Philadelphia, PA",
					  "configGlossary:adminEmail": "ksm@pobox.com",
					  "configGlossary:poweredBy": "Cofax",
					  "configGlossary:poweredByIcon": "/images/cofax.gif",
					  "configGlossary:staticPath": "/content/static",
					  "templateProcessorClass": "org.cofax.WysiwygTemplate",
					  "templateLoaderClass": "org.cofax.FilesTemplateLoader",
					  "templatePath": "templates",
					  "templateOverridePath": "",
					  "defaultListTemplate": "listTemplate.htm",
					  "defaultFileTemplate": "articleTemplate.htm",
					  "useJSP": false,
					  "jspListTemplate": "listTemplate.jsp",
					  "jspFileTemplate": "articleTemplate.jsp",
					  "cachePackageTagsTrack": 200,
					  "cachePackageTagsStore": 200,
					  "cachePackageTagsRefresh": 60,
					  "cacheTemplatesTrack": 100,
					  "cacheTemplatesStore": 50,
					  "cacheTemplatesRefresh": 15,
					  "cachePagesTrack": 200,
					  "cachePagesStore": 100,
					  "cachePagesRefresh": 10,
					  "cachePagesDirtyRead": 10,
					  "searchEngineListTemplate": "forSearchEnginesList.htm",
					  "searchEngineFileTemplate": "forSearchEngines.htm",
					  "searchEngineRobotsDb": "WEB-INF/robots.db",
					  "useDataStore": true,
					  "dataStoreClass": "org.cofax.SqlDataStore",
					  "redirectionClass": "org.cofax.SqlRedirection",
					  "dataStoreName": "cofax",
					  "dataStoreDriver": "com.microsoft.jdbc.sqlserver.SQLServerDriver",
					  "dataStoreUrl": "jdbc:microsoft:sqlserver://LOCALHOST:1433;DatabaseName=goon",
					  "dataStoreUser": "sa",
					  "dataStorePassword": "dataStoreTestQuery",
					  "dataStoreTestQuery": "SET NOCOUNT ON;select test='test';",
					  "dataStoreLogFile": "/usr/local/tomcat/logs/datastore.log",
					  "dataStoreInitConns": 10,
					  "dataStoreMaxConns": 100,
					  "dataStoreConnUsageLimit": 100,
					  "dataStoreLogLevel": "debug",
					  "maxUrlLength": 500}},
				  {
					"servlet-name": "cofaxEmail",
					"servlet-class": "org.cofax.cds.EmailServlet",
					"init-param": {
					"mailHost": "mail1",
					"mailHostOverride": "mail2"}},
				  {
					"servlet-name": "cofaxAdmin",
					"servlet-class": "org.cofax.cds.AdminServlet"},
			   
				  {
					"servlet-name": "fileServlet",
					"servlet-class": "org.cofax.cds.FileServlet"},
				  {
					"servlet-name": "cofaxTools",
					"servlet-class": "org.cofax.cms.CofaxToolsServlet",
					"init-param": {
					  "templatePath": "toolstemplates/",
					  "log": 1,
					  "logLocation": "/usr/local/tomcat/logs/CofaxTools.log",
					  "logMaxSize": "",
					  "dataLog": 1,
					  "dataLogLocation": "/usr/local/tomcat/logs/dataLog.log",
					  "dataLogMaxSize": "",
					  "removePageCache": "/content/admin/remove?cache=pages&id=",
					  "removeTemplateCache": "/content/admin/remove?cache=templates&id=",
					  "fileTransferFolder": "/usr/local/tomcat/webapps/content/fileTransferFolder",
					  "lookInContext": 1,
					  "adminGroupID": 4,
					  "betaServer": true}}],
				"servlet-mapping": {
				  "cofaxCDS": "/",
				  "cofaxEmail": "/cofaxutil/aemail/*",
				  "cofaxAdmin": "/admin/*",
				  "fileServlet": "/static/*",
				  "cofaxTools": "/tools/*"},
			   
				"taglib": {
				  "taglib-uri": "cofax.tld",
				  "taglib-location": "/WEB-INF/tlds/cofax.tld"}}}`,
		},
		{
			name: "SVG Viewer",
			value: `{"menu": {
				"header": "SVG Viewer",
				"items": [
					{"id": "Open"},
					{"id": "OpenNew", "label": "Open New"},
					null,
					{"id": "ZoomIn", "label": "Zoom In"},
					{"id": "ZoomOut", "label": "Zoom Out"},
					{"id": "OriginalView", "label": "Original View"},
					null,
					{"id": "Quality"},
					{"id": "Pause"},
					{"id": "Mute"},
					null,
					{"id": "Find", "label": "Find..."},
					{"id": "FindAgain", "label": "Find Again"},
					{"id": "Copy"},
					{"id": "CopyAgain", "label": "Copy Again"},
					{"id": "CopySVG", "label": "Copy SVG"},
					{"id": "ViewSVG", "label": "View SVG"},
					{"id": "ViewSource", "label": "View Source"},
					{"id": "SaveAs", "label": "Save As"},
					null,
					{"id": "Help"},
					{"id": "About", "label": "About Adobe CVG Viewer..."}
				]
			}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Unmarshal([]byte(test.value))
			if err != nil {
				t.Errorf("Error on Unmarshal: %s", err.Error())
			}
		})
	}
}

func TestUnmarshalSafe(t *testing.T) {
	json := []byte(`{ "store": {
		"book": [ 
		  { "category": "reference",
			"author": "Nigel Rees",
			"title": "Sayings of the Century",
			"price": 8.95
		  },
		  { "category": "fiction",
			"author": "Evelyn Waugh",
			"title": "Sword of Honour",
			"price": 12.99
		  },
		  { "category": "fiction",
			"author": "Herman Melville",
			"title": "Moby Dick",
			"isbn": "0-553-21311-3",
			"price": 8.99
		  },
		  { "category": "fiction",
			"author": "J. R. R. Tolkien",
			"title": "The Lord of the Rings",
			"isbn": "0-395-19395-8",
			"price": 22.99
		  }
		],
		"bicycle": {
		  "color": "red",
		  "price": 19.95
		}
	  }
	}`)
	safe, err := UnmarshalSafe(json)
	if err != nil {
		t.Errorf("Error on Unmarshal: %s", err.Error())
	} else if safe == nil {
		t.Errorf("Error on Unmarshal: safe is nil")
	} else {
		root, err := Unmarshal(json)
		if err != nil {
			t.Errorf("Error on Unmarshal: %s", err.Error())
		} else if root == nil {
			t.Errorf("Error on Unmarshal: root is nil")
		} else if !bytes.Equal(root.source(), safe.source()) {
			t.Errorf("Error on UnmarshalSafe: values not same")
		}
	}
}

// BenchmarkGoStdUnmarshal-8   	   61698	     19350 ns/op	     288 B/op	       6 allocs/op
// BenchmarkUnmarshal-8        	   45620	     26165 ns/op	   21889 B/op	     367 allocs/op
//
// type bench struct {
// 	Name  string `json:"name"`
// 	Value int    `json:"value"`
// }

// func BenchmarkGoStdUnmarshal(b *testing.B) {
// 	data := []byte(webApp)
// 	for i := 0; i < b.N; i++ {
// 		err := json.Unmarshal(data, &bench{})
// 		if err != nil {
// 			b.Fatal(err)
// 		}
// 	}
// }

// func BenchmarkUnmarshal(b *testing.B) {
// 	data := []byte(webApp)
// 	for i := 0; i < b.N; i++ {
// 		_, err := Unmarshal(data)
// 		if err != nil {
// 			b.Fatal(err)
// 		}
// 	}
// }
