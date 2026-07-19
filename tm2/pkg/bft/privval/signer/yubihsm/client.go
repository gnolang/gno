package yubihsm

import (
	"fmt"

	hsm "github.com/certusone/yubihsm-go"
	"github.com/certusone/yubihsm-go/commands"
	"github.com/certusone/yubihsm-go/connector"
)

// hsmAPI is the subset of a YubiHSM2 session used by the signer. It is
// defined as an interface so tests can substitute a mock implementation
// without requiring a physical device or a running yubihsm-connector.
type hsmAPI interface {
	SendEncryptedCommand(c *commands.CommandMessage) (commands.Response, error)
	Destroy()
}

// hsmAPI is implemented by *hsm.SessionManager.
var _ hsmAPI = (*hsm.SessionManager)(nil)

// newSession opens an authenticated session with the YubiHSM2 device behind
// cfg's yubihsm-connector.
func newSession(cfg *Config) (hsmAPI, error) {
	conn := connector.NewHTTPConnector(cfg.ConnectorURL)

	sm, err := hsm.NewSessionManager(conn, cfg.AuthKeyID, cfg.Password)
	if err != nil {
		return nil, fmt.Errorf("unable to open YubiHSM2 session: %w", err)
	}

	return sm, nil
}
