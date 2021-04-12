package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/tendermint/classic/abci/example/errors"
	"github.com/tendermint/classic/abci/tests/testcli"
)

var abciType string

func init() {
	abciType = os.Getenv("ABCI")
	if abciType == "" {
		abciType = "socket"
	}
}

func main() {
	testCounter()
}

const (
	maxABCIConnectTries = 10
)

func ensureABCIIsUp(typ string, n int) error {
	var err error
	cmdString := "abci-cli echo hello"
	if typ != "socket" {
		panic(fmt.Sprintf("abci server type %v not supported", typ))
	}

	for i := 0; i < n; i++ {
		cmd := exec.Command("bash", "-c", cmdString) // nolint: gas
		_, err = cmd.CombinedOutput()
		if err == nil {
			break
		}
		<-time.After(500 * time.Millisecond)
	}
	return err
}

func testCounter() {
	abciApp := os.Getenv("ABCI_APP")
	if abciApp == "" {
		panic("No ABCI_APP specified")
	}

	fmt.Printf("Running %s test with abci=%s\n", abciApp, abciType)
	cmd := exec.Command("bash", "-c", fmt.Sprintf("abci-cli %s", abciApp)) // nolint: gas
	cmd.Stdout = os.Stdout
	if err := cmd.Start(); err != nil {
		log.Fatalf("starting %q err: %v", abciApp, err)
	}
	defer cmd.Wait()
	defer cmd.Process.Kill()

	if err := ensureABCIIsUp(abciType, maxABCIConnectTries); err != nil {
		log.Fatalf("echo failed: %v", err)
	}

	client := testcli.StartSocketClient()
	defer client.Stop()

	err := compose(
		[]func() error{
			func() error { return testcli.InitChain(client) },
			func() error { return testcli.SetOption(client, "serial", "on") },
			func() error { return testcli.Commit(client, nil) },
			func() error { return testcli.DeliverTx(client, []byte("abc"), errors.BadNonce{}, nil) },
			func() error { return testcli.Commit(client, nil) },
			func() error { return testcli.DeliverTx(client, []byte{0x00}, nil, nil) },
			func() error { return testcli.Commit(client, []byte{0, 0, 0, 0, 0, 0, 0, 1}) },
			func() error { return testcli.DeliverTx(client, []byte{0x00}, errors.BadNonce{}, nil) },
			func() error { return testcli.DeliverTx(client, []byte{0x01}, nil, nil) },
			func() error { return testcli.DeliverTx(client, []byte{0x00, 0x02}, nil, nil) },
			func() error { return testcli.DeliverTx(client, []byte{0x00, 0x03}, nil, nil) },
			func() error { return testcli.DeliverTx(client, []byte{0x00, 0x00, 0x04}, nil, nil) },
			func() error {
				return testcli.DeliverTx(client, []byte{0x00, 0x00, 0x06}, errors.BadNonce{}, nil)
			},
			func() error { return testcli.Commit(client, []byte{0, 0, 0, 0, 0, 0, 0, 5}) },
		})
	if err != nil {
		log.Fatalf("test failed: %v", err)
	}
}

func compose(fs []func() error) error {
	if len(fs) == 0 {
		return nil
	} else {
		err := fs[0]()
		if err == nil {
			return compose(fs[1:])
		} else {
			return err
		}
	}
}
