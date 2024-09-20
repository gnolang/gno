//go:build docker

package integration

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

const (
	gnolandContainerName = "int_gnoland"

	test1Addr         = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
	test1Seed         = "source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast"
	dockerWaitTimeout = 30
)

func TestDockerIntegration(t *testing.T) {
	t.Parallel()

	tmpdir, err := os.MkdirTemp(os.TempDir(), "*-gnoland-integration")
	require.NoError(t, err)

	checkDocker(t)
	cleanupGnoland(t)
	buildDockerImage(t)
	startGnoland(t)
	waitGnoland(t)

	runSuite(t, tmpdir)
}

func runSuite(t *testing.T, tempdir string) {
	t.Helper()

	// add test1 account to docker container keys with "pass" password
	dockerExec(t, fmt.Sprintf(
		`echo "pass\npass\n%s\n" | gnokey add -recover -insecure-password-stdin test1`,
		test1Seed,
	))
	// assert test1 account exists
	var acc gnoland.GnoAccount
	dockerExec_gnokeyQuery(t, "auth/accounts/"+test1Addr, &acc)
	require.Equal(t, test1Addr, acc.Address.String(), "test1 account not found")

	// This value is chosen arbitrarily and may not be optimal.
	// Feel free to update it to a more suitable amount.
	minCoins := std.MustParseCoins(ugnot.ValueString(9990000000000))
	require.True(t, acc.Coins.IsAllGTE(minCoins),
		"test1 account coins expected at least %s, got %s", minCoins, acc.Coins)

	// add gno.land/r/demo/tests package as tests_copy
	dockerExec(t,
		`echo 'pass' | gnokey maketx addpkg -insecure-password-stdin \
			-gas-fee 1000000ugnot -gas-wanted 2000000 \
			-broadcast -chainid dev \
			-pkgdir /opt/gno/src/examples/gno.land/r/demo/tests/ \
			-pkgpath gno.land/r/demo/tests_copy \
			-deposit 100000000ugnot \
			test1`,
	)
	// assert gno.land/r/demo/tests_copy has been added
	var qfuncs vm.FunctionSignatures
	dockerExec_gnokeyQuery(t, `-data "gno.land/r/demo/tests_copy" vm/qfuncs`, &qfuncs)
	require.True(t, len(qfuncs) > 0, "gno.land/r/demo/tests_copy not added")

	// broadcast a package TX
	dockerExec(t,
		`echo 'pass' | gnokey maketx call -insecure-password-stdin \
			-gas-fee 1000000ugnot -gas-wanted 2000000 \
			-broadcast -chainid dev \
			-pkgpath "gno.land/r/demo/tests_copy" -func "InitTestNodes" \
			test1`,
	)
}

func checkDocker(t *testing.T) {
	t.Helper()
	output, err := createCommand(t, []string{"docker", "info"}).CombinedOutput()
	require.NoError(t, err, "docker daemon not running: %s", string(output))
}

func buildDockerImage(t *testing.T) {
	t.Helper()

	cmd := createCommand(t, []string{
		"docker",
		"build",
		"-t", "gno:integration",
		filepath.Join("..", ".."),
	})
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))
}

// dockerExec runs docker exec with cmd as argument
func dockerExec(t *testing.T, cmd string) []byte {
	t.Helper()

	cmds := append(
		[]string{"docker", "exec", gnolandContainerName, "sh", "-c"},
		cmd,
	)
	bz, err := createCommand(t, cmds).CombinedOutput()
	require.NoError(t, err, string(bz))
	return bz
}

// dockerExec_gnokeyQuery runs dockerExec with gnokey query prefix and parses
// the command output to out.
func dockerExec_gnokeyQuery(t *testing.T, cmd string, out any) {
	t.Helper()

	output := dockerExec(t, "gnokey query "+cmd)
	// parses the output of gnokey query:
	// height: h
	// data: { JSON }
	var resp struct {
		Height int64 `yaml:"height"`
		Data   any   `yaml:"data"`
	}
	err := yaml.Unmarshal(output, &resp)
	require.NoError(t, err)
	bz, err := json.Marshal(resp.Data)
	require.NoError(t, err)
	err = amino.UnmarshalJSON(bz, out)
	require.NoError(t, err)
}

func createCommand(t *testing.T, args []string) *exec.Cmd {
	t.Helper()
	msg := strings.Join(args, " ")
	t.Log(msg)
	return exec.Command(args[0], args[1:]...)
}

func startGnoland(t *testing.T) {
	t.Helper()

	cmd := createCommand(t, []string{
		"docker", "run",
		"-d",
		"--name", gnolandContainerName,
		"-w", "/opt/gno/src/gno.land",
		"gno:integration",
		"gnoland",
		"start",
	})
	output, err := cmd.CombinedOutput()
	require.NoError(t, err)
	require.NotEmpty(t, string(output)) // should be the hash of the container.

	// t.Cleanup(func() { cleanupGnoland(t) })
}

func waitGnoland(t *testing.T) {
	t.Helper()
	t.Log("waiting...")
	for i := 0; i < dockerWaitTimeout; i++ {
		output, _ := createCommand(t,
			[]string{"docker", "logs", gnolandContainerName},
		).CombinedOutput()
		if strings.Contains(string(output), "Committed state") {
			// ok blockchain is ready
			t.Log("gnoland ready")
			return
		}
		time.Sleep(time.Second)
	}
	// cleanupGnoland(t)
	panic("gnoland start timeout")
}

func cleanupGnoland(t *testing.T) {
	t.Helper()
	createCommand(t, []string{"docker", "rm", "-f", gnolandContainerName}).Run()
}
