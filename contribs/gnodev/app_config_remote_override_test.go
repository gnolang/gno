package main

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoteOverrideArr_Parse(t *testing.T) {
	for _, tc := range []struct {
		name    string
		args    []string
		want    map[string]string
		wantErr string
	}{
		{
			name: "single",
			args: []string{"-remote-override", "gno.land=https://rpc.gno.land"},
			want: map[string]string{"gno.land": "https://rpc.gno.land"},
		},
		{
			name: "multiple",
			args: []string{
				"-remote-override", "gno.land=https://rpc.gno.land",
				"-remote-override", "test.gno=http://localhost:26657",
			},
			want: map[string]string{
				"gno.land": "https://rpc.gno.land",
				"test.gno": "http://localhost:26657",
			},
		},
		{
			name:    "missing equals",
			args:    []string{"-remote-override", "gno.land"},
			wantErr: "expected domain=rpc",
		},
		{
			name:    "empty domain",
			args:    []string{"-remote-override", "=https://rpc"},
			wantErr: "empty domain",
		},
		{
			name:    "empty rpc",
			args:    []string{"-remote-override", "gno.land="},
			wantErr: "empty rpc",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			m := map[string]string{}
			fs.Var((*remoteOverrideArr)(&m), "remote-override", "")

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
