package paginate

import (
	"strconv"
	"testing"

	"gno.land/p/demo/avl"
)

func paginTest(page int, size int, reverse bool) string {
	if reverse {
		return "://here is the url/" + strconv.Itoa(page)
	}
	return "://here is the url/" + strconv.Itoa(size-page+1)
}

type testItems struct {
	content string
}

func (t *testItems) Render() string {
	return "CONTENT OF TESTITEM: { " + t.content + " }"
}

var mock avl.Tree

func init() {
	mock.Set("1", &testItems{content: "content 1"})
	mock.Set("2", &testItems{content: "content 2"})
	mock.Set("3", &testItems{content: "content 3"})
	mock.Set("4", &testItems{content: "content 4"})
	mock.Set("5", &testItems{content: "content 5"})
	mock.Set("6", &testItems{content: "content 6"})
	mock.Set("7", &testItems{content: "content 7"})
	mock.Set("8", &testItems{content: "content 8"})
}

func TestPaginate(t *testing.T) {
	const testString = `[1](://here%20is%20the%20url/1) *2* [3](://here%20is%20the%20url/3) ... [6](://here%20is%20the%20url/6) [7](://here%20is%20the%20url/7) [8](://here%20is%20the%20url/8) 

CONTENT OF TESTITEM: { content 2 }

[1](://here%20is%20the%20url/1) *2* [3](://here%20is%20the%20url/3) ... [6](://here%20is%20the%20url/6) [7](://here%20is%20the%20url/7) [8](://here%20is%20the%20url/8) 
`
	res := RenderTreeByPage(mock, paginTest, 2, 1, true)
	if res != testString {
		t.Errorf("Expected %s, got %s", testString, res)
	}
}

func TestNavigation(t *testing.T) {
	tests := []struct {
		current  int
		numPages int
		expected string
	}{
		{
			current:  3,
			numPages: 8,
			expected: "[1](x) [2](x) *3* [4](x) [5](x) [6](x) [7](x) [8](x) ",
		},
		{
			current:  2,
			numPages: 8,
			expected: "[1](x) *2* [3](x) ... [6](x) [7](x) [8](x) ",
		},
		{
			current:  5,
			numPages: 10,
			expected: "[1](x) [2](x) [3](x) [4](x) *5* [6](x) [7](x) [8](x) [9](x) [10](x) ",
		},
		{
			current:  5,
			numPages: 12,
			expected: "[1](x) [2](x) [3](x) [4](x) *5* [6](x) ... [10](x) [11](x) [12](x) ",
		},
		{
			current:  18,
			numPages: 12,
			expected: "[1](x) [2](x) [3](x) ... [10](x) [11](x) [12](x) ",
		},
	}

	for _, te := range tests {
		res := navigation(te.numPages, te.current, true, func(page int, size int, reverse bool) string {
			return "x"
		})
		if res != te.expected {
			t.Errorf("Expected %s, got %s", te.expected, res)
		}
	}
}
