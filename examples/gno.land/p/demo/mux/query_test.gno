package mux

import (
	"fmt"
	"testing"
)

func TestQuery_Get(t *testing.T) {
	cases := []struct {
		name           string
		reqPath        string
		key            string
		expectedOutput string
	}{
		// get by key
		{"Get_query_key", "api/v1/?name=testname&age=12", "name", "testname"},
		{"Get_query_repeat", "api/v1/user?name=name1&name=name2&name=name3", "name", "name1"},
		{"Get_query_with_pre_path", "api/v1/unprocesspath?user=thinhnx&age=12", "user", "thinhnx"},
		{"Get_query_empty", "api/v1/?name=testname&age=12", "loc", ""},
		{"Get_query_with_slash", "api/v1/?name=testname&endpoint=testdomain.com/v1/with/slash/", "endpoint", "testdomain.com/v1/with/slash/"},
		{"Get_query_with_extra_slash", "api/v1/unprocesspath?user=thinhnx&age=12&addr=12/34/NewYork//", "addr", "12/34/NewYork//"},
	}

	for _, tt := range cases {
		name := fmt.Sprintf("%s-%s", tt.name, tt.reqPath)
		t.Run(name, func(t *testing.T) {
			req := &Request{
				HandlerPath: "",
				Path:        tt.reqPath,
			}
			output := req.Query().Get(tt.key)
			if output != tt.expectedOutput {
				t.Errorf("Expected %q, but got %q", tt.expectedOutput, output)
			}
		})
	}
}

func TestQuery_QueryList(t *testing.T) {
	cases := []struct {
		name           string
		reqPath        string
		key            string
		expectedOutput []string
	}{
		// get all queries
		{"Get_name_list", "api/v1/user?name=name0&age=12&name=name1&age=30", "name", []string{"name0", "name1"}},
		{"Get_age_list", "api/v1/user?name=name0&age=12&name=name1&age=30", "age", []string{"12", "30"}},
		{"Get_loc_list", "api/v1/user?name=name0&age=12&name=name1&age=30&loc=HN", "loc", []string{"HN"}},
		{"Get_empty_list", "api/v1/user?name=name0&age=12&name=name1&age=30&loc=HN", "addr", []string{}},
	}

	for _, tt := range cases {
		name := fmt.Sprintf("%s-%s", tt.name, tt.reqPath)
		t.Run(name, func(t *testing.T) {
			req := &Request{
				HandlerPath: "",
				Path:        tt.reqPath,
			}
			// simple check for length of expected and output list
			output := req.Query()[tt.key]
			if len(output) != len(tt.expectedOutput) {
				t.Errorf("Expected %q, but got %q", tt.expectedOutput, output)
			}
			// check the corresponding elements
			listOutput := req.Query()[tt.key]
			for i, ttExpt := range tt.expectedOutput {
				if ttExpt != listOutput[i] {
					t.Errorf("Expected %q, but got %q", ttExpt, listOutput[i])
				}
			}
		})
	}
}

func TestQuery_GetPath(t *testing.T) {
	cases := []struct {
		name           string
		reqPath        string
		expectedOutput string
	}{
		// get all queries
		{"Get_user", "api/v1/user?name=name1&name=name2&name=name3", "user"},
		{"Get_empty_path", "api/v1/?user=thinhnx&age=12&addr=12/34/NewYork", ""},
	}

	for _, tt := range cases {
		name := fmt.Sprintf("%s-%s", tt.name, tt.reqPath)
		t.Run(name, func(t *testing.T) {
			req := &Request{
				HandlerPath: "",
				Path:        tt.reqPath,
			}
			// check the corresponding elements
			outputGP := req.GetQueryPath()
			if tt.expectedOutput != outputGP {
				t.Errorf("Expected %q, but got %q", tt.expectedOutput, outputGP)
			}
		})
	}
}
