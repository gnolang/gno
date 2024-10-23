package gnoland

import (
	"fmt"
	"strings"
)

type Param struct {
	key   string
	value string // pre-parsing representation
}

func (p Param) Verify() error {
	// XXX: validate
	return nil
}

func (p *Param) Parse(entry string) error {
	parts := strings.Split(strings.TrimSpace(entry), "=") // <key>.<kind>=<value>
	if len(parts) != 2 {
		return fmt.Errorf("malformed entry: %q", entry)
	}

	p.key = parts[0]
	p.value = parts[1]

	return p.Verify()
}

func (p Param) String() string {
	return fmt.Sprintf("%s=%s", p.key, p.value)
}

/*
func (p *Param) UnmarshalAmino(rep string) error {
	return p.Parse(rep)
}

func (p Param) MarshalAmino() (string, error) {
	return p.String(), nil
}
*/
