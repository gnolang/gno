package merkle

import (
	"crypto/sha256"
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKeyPath(t *testing.T) {
	t.Parallel()

	var path KeyPath
	keys := make([][]byte, 10)
	alphanum := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	cc8 := rand.NewChaCha8(sha256.Sum256([]byte("abc123")))
	rng := rand.New(cc8)

	for range 1_000 {
		path = nil

		for i := range keys {
			enc := keyEncoding(rng.IntN(int(KeyEncodingMax)))
			keys[i] = make([]byte, rand.Uint32()%20)
			switch enc {
			case KeyEncodingURL:
				for j := range keys[i] {
					keys[i][j] = alphanum[rng.IntN(len(alphanum))]
				}
			case KeyEncodingHex:
				cc8.Read(keys[i])
			default:
				panic("Unexpected encoding")
			}
			path = path.AppendKey(keys[i], enc)
		}

		res, err := KeyPathToKeys(path.String())
		require.Nil(t, err)

		for i, key := range keys {
			require.Equal(t, key, res[i])
		}
	}
}
