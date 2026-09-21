package main

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoteArr_Parse(t *testing.T) {
	for _, tc := range []struct {
		name    string
		args    []string
		want    map[string]string
		wantErr string
	}{
		{
			name: "single",
			args: []string{"-remote", "gno.land=https://rpc.gno.land"},
			want: map[string]string{"gno.land": "https://rpc.gno.land"},
		},
		{
			name: "multiple",
			args: []string{
				"-remote", "gno.land=https://rpc.gno.land",
				"-remote", "test.gno=http://localhost:26657",
			},
			want: map[string]string{
				"gno.land": "https://rpc.gno.land",
				"test.gno": "http://localhost:26657",
			},
		},
		{
			name:    "missing equals",
			args:    []string{"-remote", "gno.land"},
			wantErr: "expected domain=rpc",
		},
		{
			name:    "empty domain",
			args:    []string{"-remote", "=https://rpc"},
			wantErr: "empty domain",
		},
		{
			name:    "empty rpc",
			args:    []string{"-remote", "gno.land="},
			wantErr: "empty rpc",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			m := map[string]string{}
			fs.Var((*remoteArr)(&m), "remote", "")

			err := fs.Parse(tc.args)
			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, m)
		})
	}
}

func TestRemoteArr_StringDeterministic(t *testing.T) {
	m := remoteArr{
		"gno.land":    "https://rpc.gno.land",
		"test.gno":    "http://localhost:26657",
		"staging.gno": "https://rpc.staging.gno.land",
		"alpha.gno":   "https://rpc.alpha.gno.land",
	}
	first := (&m).String()
	for range 50 {
		assert.Equal(t, first, (&m).String())
	}
	assert.Equal(t,
		"alpha.gno=https://rpc.alpha.gno.land,gno.land=https://rpc.gno.land,staging.gno=https://rpc.staging.gno.land,test.gno=http://localhost:26657",
		first,
	)
}

func TestRemoteArr_StringEmpty(t *testing.T) {
	var nilArr *remoteArr
	assert.Equal(t, "", nilArr.String())

	empty := remoteArr{}
	assert.Equal(t, "", (&empty).String())
}
