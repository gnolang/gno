package mux

import "testing"

func TestRouter_Render(t *testing.T) {
	// Define handlers and route configuration
	router := NewRouter()
	router.HandleFunc("hello/{name}", func(res *ResponseWriter, req *Request) {
		name := req.GetVar("name")
		if name != "" {
			res.Write("Hello, " + name + "!")
		} else {
			res.Write("Hello, world!")
		}
	})
	router.HandleFunc("hi", func(res *ResponseWriter, req *Request) {
		res.Write("Hi, earth!")
	})

	// handler func for query case
	router.HandleFunc("hello/{user}/?", func(res *ResponseWriter, req *Request) {
		name := req.GetVar("user")
		queryPath := req.GetQueryPath()
		urlQ := req.Query()

		if len(urlQ) != 0 && queryPath != "" {
			queryNum := 0
			dataToWrite := ""
			for key, values := range urlQ {
				queryNum++
				valuesResult := ""
				for _, vl := range values {
					valuesResult += vl + ","
				}
				dataToWrite += " key: " + key + " and value: " + valuesResult
			}
			res.Write("Hello, " + name + " with path " + queryPath + " ===" + dataToWrite)
		} else {
			res.Write("Hello, world!")
		}
	})

	cases := []struct {
		path           string
		expectedOutput string
	}{
		{"hello/Alice", "Hello, Alice!"},
		{"hi", "Hi, earth!"},
		{"hello/Bob", "Hello, Bob!"},
		{"hello/Bob/action?testaction=punch", "Hello, Bob with path action === key: testaction and value: punch,"},
		{"hello/Alice/finding?docs=gno&book=gnobook", "Hello, Alice with path finding === key: docs and value: gno, key: book and value: gnobook,"},
		{"hello/Celena/finding?color=red&color=blue", "Hello, Celena with path finding === key: color and value: red,blue,"},
		// TODO: {"hello", "Hello, world!"},
		// TODO: hello/, /hello, hello//Alice, hello/Alice/, hello/Alice/Bob, etc
	}
	for _, tt := range cases {
		t.Run(tt.path, func(t *testing.T) {
			output := router.Render(tt.path)
			if output != tt.expectedOutput {
				t.Errorf("Expected output %q, but got %q", tt.expectedOutput, output)
			}
		})
	}
}
