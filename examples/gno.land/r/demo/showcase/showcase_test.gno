package showcase

func TestRender(t *testing.T) {
	cases := []struct {
		name           string
		path           string
		shouldContains string
	}{
		{"home", "", "hello"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			println(name, tc.path, tc.shouldContains)
		})
	}
}
