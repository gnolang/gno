package main

import (
	"context"
	"path/filepath"
	"testing"

	signer "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/local"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecrets_Init_KeyType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		flag     []string
		wantType string
		wantErr  string
	}{
		{
			name:     "default is ed25519",
			flag:     nil,
			wantType: "ed25519",
		},
		{
			name:     "explicit ed25519",
			flag:     []string{"--key-type", "ed25519"},
			wantType: "ed25519",
		},
		{
			name:     "explicit secp256k1",
			flag:     []string{"--key-type", "secp256k1"},
			wantType: "secp256k1",
		},
		{
			name:    "unknown scheme rejected",
			flag:    []string{"--key-type", "p256"},
			wantErr: "unsupported validator key type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			cmd := newRootCmd(commands.NewTestIO())
			args := []string{"secrets", "init", "--data-dir", tempDir}
			args = append(args, tt.flag...)

			err := cmd.ParseAndRun(context.Background(), args)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)

			fk, err := signer.LoadFileKey(filepath.Join(tempDir, defaultValidatorKeyName))
			require.NoError(t, err)

			switch tt.wantType {
			case "ed25519":
				_, ok := fk.PrivKey.(ed25519.PrivKeyEd25519)
				assert.True(t, ok, "expected ed25519 key, got %T", fk.PrivKey)
			case "secp256k1":
				_, ok := fk.PrivKey.(secp256k1.PrivKeySecp256k1)
				assert.True(t, ok, "expected secp256k1 key, got %T", fk.PrivKey)
			}
		})
	}
}

func TestSecrets_Init_KeyType_SingleValidator(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	cmd := newRootCmd(commands.NewTestIO())
	args := []string{
		"secrets",
		"init",
		"--data-dir",
		tempDir,
		"--key-type",
		"secp256k1",
		validatorPrivateKeyKey,
	}

	require.NoError(t, cmd.ParseAndRun(context.Background(), args))

	fk, err := signer.LoadFileKey(filepath.Join(tempDir, defaultValidatorKeyName))
	require.NoError(t, err)

	_, ok := fk.PrivKey.(secp256k1.PrivKeySecp256k1)
	assert.True(t, ok, "single-key init should also honour --key-type")
}
