package valset

import (
	"errors"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/std"
)

const initFile = "init.gno"

var errMissingInit = errors.New("missing init.gno")

// ModifyPoCDeployment modifies the initialization Realm
// for Proof of Contribution to be aligned with the given set
func ModifyPoCDeployment(
	file *std.MemPackage,
	set []types.GenesisValidator,
) error {
	return modifyValsetDeployment(file, set, pocTemplate)
}

// modifyValsetDeployment modifies the generic initialization Realm
// for a valset protocol to be aligned with the given set. Utilizes
// the given template to append the initial validator set
func modifyValsetDeployment(
	file *std.MemPackage,
	set []types.GenesisValidator,
	setTemplate string,
) error {
	// Find the `init.gno` file
	for _, f := range file.Files {
		if f.Name != initFile {
			continue
		}

		// Generate the Realm `init.gno` body
		body, err := generateInitBody(set, setTemplate)
		if err != nil {
			return fmt.Errorf("unable to prepare valset init body, %w", err)
		}

		f.Body = body

		return nil
	}

	return errMissingInit
}

const pocTemplate = `
package vals

import (
	"std"	

    "gno.land/p/sys/vals/poc"
    "gno.land/p/sys/vals/types"
)

func init() {
    set := []*types.Validator{
		{{ range . }}
        {
            Address:     std.Address("{{ .Address }}"),
            PubKey:      "{{ .PubKey }}",
            VotingPower: {{ .VotingPower }},
        },
		{{ end }}
    }

    v = &vals{
        p:       poc.NewPoC(WithInitialSet(set)),
        changes: make([]types.Validator, 0),
    }
}
`
