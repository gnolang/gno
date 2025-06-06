package merkle

import (
	// it is ok to use math/rand here: we do not need a cryptographically secure random
	// number generator here and we can run the tests a bit faster
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKeyPath(t *testing.T) {
	t.Parallel()

	var path KeyPath
	keys := make([][]byte, 10)
	alphanum := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	for range 1_000 {
		path = nil

		for i := range keys {
			enc := keyEncoding(rand.Intn(int(KeyEncodingMax)))
			keys[i] = make([]byte, rand.Uint32()%20)
			switch enc {
			case KeyEncodingURL:
				for j := range keys[i] {
					keys[i][j] = alphanum[rand.Intn(len(alphanum))]
				}
			case KeyEncodingHex:
				rand.Read(keys[i])
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
