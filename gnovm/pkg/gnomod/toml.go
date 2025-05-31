package gnomod

import (
	"fmt"

	"github.com/pelletier/go-toml"
)

// ParseTomlBytes parses the gnomod.toml file from bytes.
func parseTomlBytes(fname string, data []byte) (*File, error) {
	var f File
	if err := toml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("error parsing gnomod.toml file at %q: %w", fname, err)
	}
	return &f, nil
}

// WriteTomlString writes the gnomod.toml file to a string.
func (f *File) WriteString() string {
	buf, err := toml.Marshal(f)
	if err != nil {
		panic(err)
	}
	return string(buf)
}
