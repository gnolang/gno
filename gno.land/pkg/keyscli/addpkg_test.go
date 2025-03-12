package keyscli

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRejectUnformattedFiles(t *testing.T) {
	tests := []struct {
		name    string
		dir     string
		wantErr string
	}{
		{
			"unformatted",
			"unfmted",
			"your file's size increased after formatting beyond the tolerable value",
		},
		{
			"formatted",
			"fmted",
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			absDirPath := filepath.Join("testdata", tt.dir)
			gotErr := validateThatAllFilesAreFormatted(absDirPath, "pkg")
			if tt.wantErr != "" {
				require.Error(t, gotErr, "Expecting a non-nil error")
				require.Contains(t, gotErr.Error(), tt.wantErr, "Missing substring")
				return
			}

			require.NoError(t, gotErr, "Unexpected error")
		})
	}
}
