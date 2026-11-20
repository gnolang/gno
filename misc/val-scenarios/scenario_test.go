package valscenarios

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	p2ptypes "github.com/gnolang/gno/tm2/pkg/p2p/types"
)

func bashCommand(t *testing.T, script string, env ...string) *exec.Cmd {
	t.Helper()
	if out, err := exec.Command("bash", "-c", "declare -A check").CombinedOutput(); err != nil {
		t.Skipf("Bash 4+ required: %s", out)
	}
	lib, err := filepath.Abs("lib/scenario.sh")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	cmd := exec.CommandContext(ctx, "bash", "-c", "source \"$SCENARIO_TEST_LIB\"\n"+script)
	cmd.Env = append(os.Environ(), append([]string{"SCENARIO_TEST_LIB=" + lib}, env...)...)
	return cmd
}

func TestWaitForRPC(t *testing.T) {
	status := func(id p2ptypes.ID) string {
		result := rpctypes.ResultStatus{NodeInfo: p2ptypes.NodeInfo{
			NetAddress: &p2ptypes.NetAddress{ID: id, IP: net.ParseIP("127.0.0.1"), Port: 26800},
		}}
		// Use the RPC codec: NetAddress has a custom string representation.
		return `{"jsonrpc":"2.0","id":"","result":` + string(amino.MustMarshalJSON(result)) + `}`
	}
	for _, tc := range []struct {
		name, response, pid string
		wantOK              bool
	}{
		{"own node", status("expected"), "$$", true},
		{"other node", status("other"), "$$", false},
		{"non RPC server", `{}`, "$$", false},
		{"dead process", status("expected"), "$dead_pid", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				fmt.Fprint(w, tc.response)
			}))
			defer server.Close()
			_, port, err := net.SplitHostPort(server.Listener.Addr().String())
			if err != nil {
				t.Fatal(err)
			}
			cmd := bashCommand(t, fmt.Sprintf(`
scenario_init rpc-test
sleep 30 &
dead_pid=$!
kill "$dead_pid"
wait "$dead_pid" 2>/dev/null || true
NODE_ID[val1]=expected
NODE_RPC_PORT[val1]=%s
NODE_PID[val1]=%s
wait_for_rpc val1 1
`, port, tc.pid), "RUNTIME=local")
			out, err := cmd.CombinedOutput()
			if (err == nil) != tc.wantOK {
				t.Fatalf("success=%v, want %v; output: %s", err == nil, tc.wantOK, out)
			}
		})
	}
}

func TestLocalPortSelectionWithoutHelper(t *testing.T) {
	binDir := t.TempDir()
	for _, binary := range []string{"gnoland", "gnokey", "gnogenesis"} {
		if err := os.WriteFile(filepath.Join(binDir, binary), []byte("#!/bin/sh\nexit 0\n"), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	port := listener.Addr().(*net.TCPAddr).Port
	cmd := bashCommand(t, fmt.Sprintf(`
ensure_runtime_ready
_pick_free_port %d
first=$_PICKED_PORT
[ "$first" -gt %d ]
_pick_free_port "$first"
[ "$_PICKED_PORT" -gt "$first" ]
`, port, port), "RUNTIME=local", "BIN_DIR="+binDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("local port selection requires a helper or reused a port: %v: %s", err, out)
	}
}

func TestComposeRetry(t *testing.T) {
	for _, tc := range []struct {
		name, message string
		failures      int
		wantAttempts  string
		wantOK        bool
	}{
		{"bind", "bind: address already in use", 1, "2", true},
		{"allocated", "port is already allocated", 1, "2", true},
		{"permanent bind", "bind: address already in use", 9, "3", false},
		{"other error", "image not found", 1, "1", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := filepath.Join(t.TempDir(), "attempts")
			if err := os.WriteFile(state, []byte("0"), 0o600); err != nil {
				t.Fatal(err)
			}
			cmd := bashCommand(t, `
sleep() { :; }
compose() {
  [ "$1" = up ] || return 0
  local attempt
  attempt=$(cat "$STATE")
  attempt=$((attempt + 1))
  echo "$attempt" > "$STATE"
  if [ "$attempt" -le "$FAILURES" ]; then echo "$MESSAGE" >&2; return 1; fi
}
compose_up_one val1
`, "RUNTIME=docker", "STATE="+state, "FAILURES="+fmt.Sprint(tc.failures), "MESSAGE="+tc.message)
			out, err := cmd.CombinedOutput()
			if (err == nil) != tc.wantOK {
				t.Errorf("success=%v, want %v: %s", err == nil, tc.wantOK, out)
			}
			attempts, err := os.ReadFile(state)
			if err != nil {
				t.Fatal(err)
			}
			if strings.TrimSpace(string(attempts)) != tc.wantAttempts {
				t.Errorf("attempts=%s, want %s", attempts, tc.wantAttempts)
			}
		})
	}
}

func TestDockerStartupRetryPreservesOtherServices(t *testing.T) {
	dir := t.TempDir()
	cmd := bashCommand(t, `
scenario_init retry-test
SCENARIO_DIR="$TEST_DIR"
COMPOSE_FILE="$TEST_DIR/compose.yml"
gen_sentry sentry1
gen_validator val1 --sentry sentry1 --controllable-signer
NODE_DATA_DIR[sentry1]="$TEST_DIR"
NODE_DATA_DIR[val1]="$TEST_DIR"
NODE_ID[val1]=id
NODE_ADDRESS[val1]=address
NODE_PUBKEY[val1]=pubkey
echo 0 > "$TEST_DIR/attempts"
compose() {
  case "$1" in
    up)
      if [ "${@: -1}" = val1 ]; then
        local attempt
        attempt=$(cat "$TEST_DIR/attempts")
        attempt=$((attempt + 1))
        echo "$attempt" > "$TEST_DIR/attempts"
        if [ "$attempt" -eq 1 ]; then
          echo 'bind: address already in use'
          return 1
        fi
      fi ;;
    port) echo '0.0.0.0:49123' ;;
    down) touch "$TEST_DIR/torn-down" ;;
  esac
}
sleep() { :; }
_capture_node_logs() { :; }
wait_for_rpc() { :; }
wait_for_control() { :; }
start_all_nodes
[ "$(cat "$TEST_DIR/attempts")" -eq 2 ]
[ ! -f "$TEST_DIR/torn-down" ]
`, "RUNTIME=docker", "TEST_DIR="+dir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("startup retry tore down sentry/signer or failed: %v: %s", err, out)
	}
}

func TestDockerMappingsRemainEphemeral(t *testing.T) {
	dir := t.TempDir()
	cmd := bashCommand(t, `
scenario_init mapping-test
SCENARIO_DIR="$TEST_DIR"
COMPOSE_FILE="$TEST_DIR/compose.yml"
gen_validator val1 --controllable-signer
NODE_DATA_DIR[val1]="$TEST_DIR"
NODE_ID[val1]=id
NODE_ADDRESS[val1]=address
NODE_PUBKEY[val1]=pubkey
write_compose_file
# Replace Docker and its log stream, retaining real configuration and resolution.
compose() {
  if [ "$1" = port ]; then echo '0.0.0.0:49123'; fi
}
_capture_node_logs() { :; }
wait_for_rpc() { :; }
wait_for_control() { :; }
start_all_nodes
cat "$COMPOSE_FILE"
`, "RUNTIME=docker", "TEST_DIR="+dir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("start_all_nodes: %v: %s", err, out)
	}
	for _, mapping := range []string{`"::26657"`, `"::8080"`} {
		if !strings.Contains(string(out), mapping) {
			t.Errorf("missing ephemeral mapping %s after startup: %s", mapping, out)
		}
	}
}

func TestDockerRestartRefreshesInventory(t *testing.T) {
	cmd := bashCommand(t, `
scenario_init inventory-test
SCENARIO_DIR="$TEST_DIR"
gen_validator val1
NODE_ID[val1]=id
NODE_ADDRESS[val1]=address
NODE_PUBKEY[val1]=pubkey
NODE_RPC_PORT[val1]=49123
write_inventory
compose() { if [ "$1" = port ]; then echo '0.0.0.0:49124'; fi; }
_capture_node_logs() { :; }
wait_for_rpc() { :; }
start_node val1
jq -e '.validators[0].rpc_url == "http://127.0.0.1:49124"' "$TEST_DIR/inventory.json"
`, "RUNTIME=docker", "TEST_DIR="+t.TempDir())
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("inventory retains the old RPC port after restart: %v: %s", err, out)
	}
}

func TestBuildBinariesRelativeDir(t *testing.T) {
	dir := t.TempDir()
	makefile, err := os.ReadFile("Makefile")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "Makefile"), makefile, 0o600); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"tools", "core", "gen/contribs/gnogenesis", "signer"} {
		if err := os.MkdirAll(filepath.Join(dir, name), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	// Only replace the compiler: emulate its output path semantics after each cd.
	fakeGo := `#!/usr/bin/env bash
set -eu
shift
[ "$1" = -o ]
output=$2
shift 2
if [[ "$output" = */ ]]; then
  mkdir -p "$output"
  for pkg in "$@"; do touch "$output/$(basename "$pkg")"; done
else
  touch "$output"
fi
`
	if err := os.WriteFile(filepath.Join(dir, "tools/go"), []byte(fakeGo), 0o700); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("make", "build-binaries", "RUNTIME=local", "BIN_DIR=custom-bin",
		"LOCAL_GNO_ROOT="+dir, "CORE_GNO_ROOT="+filepath.Join(dir, "core"),
		"GNOGENESIS_GNO_ROOT="+filepath.Join(dir, "gen"), "VALSIGNER_GNO_ROOT="+filepath.Join(dir, "signer"))
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "PATH="+filepath.Join(dir, "tools")+":"+os.Getenv("PATH"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("make: %v: %s", err, output)
	}
	for _, binary := range []string{"gnoland", "gnokey", "gnogenesis", "valsignerd"} {
		if _, err := os.Stat(filepath.Join(dir, "custom-bin", binary)); err != nil {
			t.Errorf("binary %s missing from BIN_DIR: %v", binary, err)
		}
	}
}
