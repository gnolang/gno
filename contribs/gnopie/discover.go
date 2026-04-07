package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pelletier/go-toml"
)

// Remote represents a discovered network configuration.
type Remote struct {
	Name    string `toml:"name,omitempty"`
	RPC     string `toml:"rpc"`
	Indexer string `toml:"indexer,omitempty"`
	ChainID string `toml:"chain_id"`
}

const (
	discoverTimeout  = 10 * time.Second
	cacheMaxAge      = 24 * time.Hour
	localhostWebPort = "8888" // gnodev default web port
)

var metaRe = regexp.MustCompile(`<meta\s+name="gnoconnect:(\w+)"\s+content="([^"]*)"`)

// DebugFunc is an optional debug logging function.
type DebugFunc func(format string, args ...any)

func noopDebug(string, ...any) {}

// DiscoverRemote discovers network configuration by fetching the domain's
// gnoweb homepage and reading <meta name="gnoconnect:*"> tags.
// Results are cached on disk for cacheMaxAge.
func DiscoverRemote(home, domain string, dbg DebugFunc) (*Remote, error) {
	if dbg == nil {
		dbg = noopDebug
	}

	// Check cache first
	if r, err := loadCachedRemote(home, domain); err == nil {
		dbg("cache hit for %s (rpc=%s, chainid=%s)", domain, r.RPC, r.ChainID)
		return r, nil
	}
	dbg("cache miss for %s, discovering...", domain)

	// Determine URL to fetch
	var url string
	if domain == "localhost" || strings.HasPrefix(domain, "127.0.0.1") {
		url = fmt.Sprintf("http://localhost:%s/", localhostWebPort)
	} else {
		url = fmt.Sprintf("https://%s/", domain)
	}

	dbg("fetching %s", url)
	client := &http.Client{Timeout: discoverTimeout}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("discovering %s: %w", domain, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", url, err)
	}

	// Parse gnoconnect meta tags
	remote := &Remote{Name: domain}
	matches := metaRe.FindAllSubmatch(body, -1)
	for _, match := range matches {
		key := string(match[1])
		value := string(match[2])
		dbg("found gnoconnect:%s = %s", key, value)
		switch key {
		case "rpc":
			remote.RPC = value
		case "chainid":
			remote.ChainID = value
		case "indexer":
			remote.Indexer = value
		}
	}

	if remote.RPC == "" {
		return nil, fmt.Errorf("no gnoconnect:rpc meta tag found at %s", url)
	}
	if remote.ChainID == "" {
		return nil, fmt.Errorf("no gnoconnect:chainid meta tag found at %s", url)
	}

	dbg("discovered %s: rpc=%s chainid=%s", domain, remote.RPC, remote.ChainID)

	// Cache the result
	_ = saveCachedRemote(home, domain, remote)
	dbg("cached %s", domain)

	return remote, nil
}

func cachePath(home, domain string) string {
	// Use hash to avoid path issues with dots/slashes
	h := sha256.Sum256([]byte(domain))
	return filepath.Join(home, "gnopie", "cache", fmt.Sprintf("%x.toml", h[:8]))
}

type cachedRemote struct {
	Remote
	CachedAt time.Time `toml:"cached_at"`
}

func loadCachedRemote(home, domain string) (*Remote, error) {
	path := cachePath(home, domain)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cached cachedRemote
	if err := toml.Unmarshal(data, &cached); err != nil {
		return nil, err
	}

	if time.Since(cached.CachedAt) > cacheMaxAge {
		return nil, fmt.Errorf("cache expired")
	}

	cached.Remote.Name = domain
	return &cached.Remote, nil
}

func saveCachedRemote(home, domain string, r *Remote) error {
	path := cachePath(home, domain)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	cached := cachedRemote{
		Remote:   *r,
		CachedAt: time.Now(),
	}

	data, err := toml.Marshal(cached)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
