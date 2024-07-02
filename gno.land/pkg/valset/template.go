package valset

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
)

// validator defines the Go version of Gno's:
// gno.land/p/sys/vals/types Validator
type validator struct {
	Address     string // bech32 address
	PubKey      string // bech32 representation of the public key
	VotingPower uint64
}

// generateInitBody generates the templated initialization
// body for the gno.land/r/sys/vals Realm
func generateInitBody(
	set []types.GenesisValidator,
	setTemplate string,
) (string, error) {
	// Parse the template
	tmpl, err := template.New("init-realm").Parse(setTemplate)
	if err != nil {
		return "", fmt.Errorf("unable to parse template, %w", err)
	}

	vals := make([]validator, 0, len(set))

	for _, v := range set {
		vals = append(vals, validator{
			Address:     v.Address.String(),
			PubKey:      v.PubKey.String(),
			VotingPower: uint64(v.Power),
		})
	}

	// Apply the valset to the template
	var builder strings.Builder

	if err = tmpl.Execute(&builder, vals); err != nil {
		return "", fmt.Errorf("unable to execute template, %w", err)
	}

	return builder.String(), nil
}
