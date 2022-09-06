package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const gnolandContainerName = "int_gnoland"

func TestDockerIntegration(t *testing.T) {
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

	cmd := createCommand(t, []string{
		"docker", "exec", gnolandContainerName,
		"gnokey", "query", "auth/accounts/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", // test1
	})
	output, err := cmd.CombinedOutput()
	require.NoError(t, err)
	// FIXME: this will break frequently. we need a reliable test.
	require.Contains(t, string(output), "9999980000000ugnot")

	// FIXME: perform TXs.
}

func checkDocker(t *testing.T) {
	t.Helper()
	// FIXME: check if `docker` is installed and compatible.
}

func buildDockerImage(t *testing.T) {
	t.Helper()

	cmd := createCommand(t, []string{
		"docker",
		"build",
		"-t", "gno:integration",
		filepath.Join("..", ".."),
	})
	output, err := cmd.Output()
	require.NoError(t, err)
	// FIXME: is this check reliable?
	require.Contains(t, string(output), "Successfully built")
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
		"gno:integration",
		"gnoland",
	})
	output, err := cmd.CombinedOutput()
	require.NoError(t, err)
	require.NotEmpty(t, string(output)) // should be the hash of the container.

	//t.Cleanup(func() { cleanupGnoland(t) })
}

func waitGnoland(t *testing.T) {
	t.Helper()

	t.Log("waiting...")
	// FIXME: tail logs and wait for blockchain to be ready.
	time.Sleep(5000 * time.Millisecond)
}

func cleanupGnoland(t *testing.T) {
	t.Helper()

	// FIXME: detect if container exists before killing it.

	cmd := createCommand(t, []string{"docker", "kill", gnolandContainerName})
	_, _ = cmd.Output()

	cmd = createCommand(t, []string{"docker", "rm", "-f", gnolandContainerName})
	_, _ = cmd.Output()
}
