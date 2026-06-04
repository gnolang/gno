package millipede

import (
	"testing"

	"gno.land/p/nt/uassert/v0"
)

func TestRender(t *testing.T) {
	cases := []struct {
		path     string
		expected string
	}{
		{
			path: "",
			expected: "```" + `
    ╚⊙ ⊙╝
  ╚═(███)═╝
 ╚═(███)═╝
╚═(███)═╝
 ╚═(███)═╝
  ╚═(███)═╝
   ╚═(███)═╝
    ╚═(███)═╝
    ╚═(███)═╝
   ╚═(███)═╝
  ╚═(███)═╝
 ╚═(███)═╝
╚═(███)═╝
 ╚═(███)═╝
  ╚═(███)═╝
   ╚═(███)═╝
    ╚═(███)═╝
    ╚═(███)═╝
   ╚═(███)═╝
  ╚═(███)═╝
 ╚═(███)═╝
` + "```\n[19](/r/demo/art/millipede:19)<  >[21](/r/demo/art/millipede:21)",
		},
		{
			path: "4",
			expected: "```" + `
    ╚⊙ ⊙╝
  ╚═(███)═╝
 ╚═(███)═╝
╚═(███)═╝
 ╚═(███)═╝
` + "```\n[3](/r/demo/art/millipede:3)<  >[5](/r/demo/art/millipede:5)",
		},
	}

	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			got := Render(tc.path)
			uassert.Equal(t, tc.expected, got)
		})
	}
}
