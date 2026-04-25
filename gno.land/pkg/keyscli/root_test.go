package keyscli

import "testing"

func TestGnowebURLForPkg(t *testing.T) {
	tests := []struct {
		name    string
		envURL  string
		chainID string
		pkgPath string
		want    string
	}{
		{
			name:    "empty pkgPath",
			chainID: "gnoland1",
			pkgPath: "",
			want:    "",
		},
		{
			name:    "registry hit",
			chainID: "gnoland1",
			pkgPath: "gno.land/r/demo/counter",
			want:    "https://gno.land/r/demo/counter",
		},
		{
			name:    "registry hit testnet",
			chainID: "test11",
			pkgPath: "gno.land/r/demo/counter",
			want:    "https://test11.testnets.gno.land/r/demo/counter",
		},
		{
			name:    "dev chainID",
			chainID: "dev",
			pkgPath: "gno.land/r/demo/counter",
			want:    "http://127.0.0.1:8888/r/demo/counter",
		},
		{
			name:    "unknown chain",
			chainID: "mydev",
			pkgPath: "gno.land/r/demo/counter",
			want:    "",
		},
		{
			name:    "env var overrides registry",
			envURL:  "https://my.private.net",
			chainID: "gnoland1",
			pkgPath: "gno.land/r/demo/counter",
			want:    "https://my.private.net/r/demo/counter",
		},
		{
			name:    "env var resolves unknown chain",
			envURL:  "https://my.private.net",
			chainID: "private42",
			pkgPath: "gno.land/r/demo/counter",
			want:    "https://my.private.net/r/demo/counter",
		},
		{
			name:    "env var with trailing slash",
			envURL:  "https://my.private.net/",
			chainID: "private42",
			pkgPath: "gno.land/r/demo/counter",
			want:    "https://my.private.net/r/demo/counter",
		},
		{
			name:    "env var works with empty chainID",
			envURL:  "https://my.private.net",
			chainID: "",
			pkgPath: "gno.land/r/demo/counter",
			want:    "https://my.private.net/r/demo/counter",
		},
		{
			name:    "pkgPath without gno.land prefix",
			chainID: "gnoland1",
			pkgPath: "/r/demo/counter",
			want:    "https://gno.land/r/demo/counter",
		},
		{
			name:    "pkgPath equal to gno.land",
			chainID: "gnoland1",
			pkgPath: "gno.land",
			want:    "https://gno.land/",
		},
		{
			name:    "pkgPath gno.landfoo not stripped",
			chainID: "gnoland1",
			pkgPath: "gno.landfoo/r/demo",
			want:    "https://gno.land/gno.landfoo/r/demo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(gnowebURLEnv, tt.envURL)
			got := gnowebURLForPkg(tt.chainID, tt.pkgPath)
			if got != tt.want {
				t.Errorf("gnowebURLForPkg(%q, %q) [env=%q] = %q, want %q", tt.chainID, tt.pkgPath, tt.envURL, got, tt.want)
			}
		})
	}
}
