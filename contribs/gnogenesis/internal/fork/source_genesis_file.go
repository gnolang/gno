package fork

import (
	"context"
	"fmt"
	"os"

	"github.com/gnolang/gno/tm2/pkg/amino"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
)

// fileGenesisSource reads the source chain's genesis from a local .json file.
type fileGenesisSource struct {
	path string
}

func newFileGenesisSource(path string) (*fileGenesisSource, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("genesis file %q: %w", path, err)
	}
	return &fileGenesisSource{path: path}, nil
}

func (s *fileGenesisSource) Description() string { return "genesis file" }
func (s *fileGenesisSource) Close() error        { return nil }

func (s *fileGenesisSource) FetchGenesis(_ context.Context) (*bft.GenesisDoc, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", s.path, err)
	}
	var genDoc bft.GenesisDoc
	if err := amino.UnmarshalJSON(data, &genDoc); err != nil {
		return nil, fmt.Errorf("parsing genesis %s: %w", s.path, err)
	}
	return &genDoc, nil
}
