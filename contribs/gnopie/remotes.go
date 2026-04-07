package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pelletier/go-toml"
)

type Remote struct {
	Name    string `toml:"-"`
	RPC     string `toml:"rpc"`
	Indexer string `toml:"indexer,omitempty"`
	ChainID string `toml:"chain_id"`
	Default bool   `toml:"default,omitempty"`
}

type RemoteConfig struct {
	Remotes map[string]*Remote
	path    string
}

func DefaultRemotes() map[string]*Remote {
	return map[string]*Remote{
		"gno.land": {
			Name: "gno.land", RPC: "https://rpc.gno.land:443",
			Indexer: "https://indexer.gno.land/graphql", ChainID: "gnoland1", Default: true,
		},
		"localhost": {
			Name: "localhost", RPC: "http://127.0.0.1:26657", ChainID: "dev",
		},
	}
}

func remotesPath(home string) string {
	return filepath.Join(home, "gnopie", "remotes.toml")
}

func LoadRemotes(home string) (*RemoteConfig, error) {
	path := remotesPath(home)
	cfg := &RemoteConfig{path: path}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg.Remotes = DefaultRemotes()
			return cfg, nil
		}
		return nil, fmt.Errorf("reading remotes: %w", err)
	}
	remotes := make(map[string]*Remote)
	if err := toml.Unmarshal(data, &remotes); err != nil {
		return nil, fmt.Errorf("parsing remotes: %w", err)
	}
	for name, r := range remotes {
		r.Name = name
	}
	cfg.Remotes = remotes
	return cfg, nil
}

func (c *RemoteConfig) Save() error {
	if err := os.MkdirAll(filepath.Dir(c.path), 0o755); err != nil {
		return err
	}
	data, err := toml.Marshal(c.Remotes)
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, data, 0o644)
}

func (c *RemoteConfig) GetDefault() (*Remote, error) {
	for _, r := range c.Remotes {
		if r.Default {
			return r, nil
		}
	}
	return nil, fmt.Errorf("no default remote (use 'gnopie remotes default <name>')")
}

func (c *RemoteConfig) Get(name string) (*Remote, error) {
	if r, ok := c.Remotes[name]; ok {
		return r, nil
	}
	return nil, fmt.Errorf("unknown remote %q", name)
}

func (c *RemoteConfig) SetDefault(name string) error {
	if _, ok := c.Remotes[name]; !ok {
		return fmt.Errorf("unknown remote %q", name)
	}
	for n, r := range c.Remotes {
		r.Default = (n == name)
	}
	return nil
}

func (c *RemoteConfig) Resolve(networkFlag, domain string) (*Remote, error) {
	if networkFlag != "" {
		return c.Get(networkFlag)
	}
	if domain != "" {
		if r, ok := c.Remotes[domain]; ok {
			return r, nil
		}
	}
	return c.GetDefault()
}

func (c *RemoteConfig) SortedNames() []string {
	names := make([]string, 0, len(c.Remotes))
	for name := range c.Remotes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (c *RemoteConfig) Add(name, rpc, chainID, indexer string) error {
	if _, ok := c.Remotes[name]; ok {
		return fmt.Errorf("remote %q already exists", name)
	}
	c.Remotes[name] = &Remote{Name: name, RPC: rpc, ChainID: chainID, Indexer: indexer}
	return nil
}

func (c *RemoteConfig) Remove(name string) error {
	if _, ok := c.Remotes[name]; !ok {
		return fmt.Errorf("unknown remote %q", name)
	}
	delete(c.Remotes, name)
	return nil
}

func (c *RemoteConfig) Update(name, rpc, chainID, indexer string) error {
	r, ok := c.Remotes[name]
	if !ok {
		return fmt.Errorf("unknown remote %q", name)
	}
	if rpc != "" {
		r.RPC = rpc
	}
	if chainID != "" {
		r.ChainID = chainID
	}
	if indexer != "" {
		r.Indexer = indexer
	}
	return nil
}

func (c *RemoteConfig) FormatTable() string {
	var sb strings.Builder
	for _, name := range c.SortedNames() {
		r := c.Remotes[name]
		marker := "  "
		if r.Default {
			marker = "* "
		}
		fmt.Fprintf(&sb, "%s%-15s %s (chain: %s)\n", marker, name, r.RPC, r.ChainID)
	}
	return sb.String()
}
