package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	mock "github.com/gnolang/gno/contribs/gnodev/internal/mock/emitter"
	"github.com/gnolang/gno/contribs/gnodev/pkg/events"
	"github.com/gnolang/gno/contribs/gnodev/pkg/packages"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testDevnetHome builds a home whose own root, and working directory, is a
// temp dir; relative package dirs are below it.
func testDevnetHome(t *testing.T, pkgs map[string]string, primary ...string) (*devnetHome, *mock.ServerEmitter) {
	t.Helper()
	root := t.TempDir()
	list := make([]*packages.Package, 0, len(pkgs))
	for path, dir := range pkgs {
		if dir != "" && !filepath.IsAbs(dir) {
			dir = filepath.Join(root, dir)
		}
		list = append(list, &packages.Package{ImportPath: path, Dir: dir, Kind: packages.KindFS})
	}
	next := &mock.ServerEmitter{}
	cfg := defaultLocalAppConfig
	h := newDevnetHome(next, &cfg, []string{root}, primary)
	h.cwd = root
	h.watch(func() []*packages.Package { return list })
	return h, next
}

func txResult(msg std.Msg, gas int64, err error) events.TxResult {
	res := events.TxResult{Tx: std.Tx{Msgs: []std.Msg{msg}}}
	res.Response.GasUsed = gas
	if err != nil {
		res.Response.Error = abci.StringError(err.Error())
	}
	return res
}

func TestDevnetHomePackages(t *testing.T) {
	t.Parallel()

	h, _ := testDevnetHome(t, map[string]string{
		"gno.land/p/me/lib":       "lib",
		"gno.land/r/me/a":         "a",
		"gno.land/r/me/b":         "nested/b",
		"example.com/r/x":         "x",
		"gno.land/r/me/a)b":       "odd",
		"gno.land/r/demo/boards":  "/elsewhere/examples/boards",
		"gno.land/p/demo/flagged": "/elsewhere/examples/flagged",
		"gno.land/r/me/remote":    "",
		"gno.land/p/me/undecided": "../outside",
	}, "gno.land/r/me/b", "gno.land/p/demo/flagged")

	page := h.Render()
	assert.Contains(t, page, "- [r/me/b](/r/me/b)\\\n  [Actions](/r/me/b$help) · [State](/r/me/b$state)\n"+
		"- [p/demo/flagged](/p/demo/flagged)\n"+
		"- [r/me/a](/r/me/a)\\\n  [Actions](/r/me/a$help) · [State](/r/me/a$state)\n"+
		"- example.com/r/x\n"+
		"- [p/me/lib](/p/me/lib)\n",
		"the packages gnodev was started on come first, wherever they are, then realms")
	for _, gone := range []string{"boards", "me/remote", "undecided", "a)b", "more\n"} {
		assert.NotContains(t, page, gone)
	}

	// The commands name the first realm and its folder.
	assert.Contains(t, page, `-pkgpath "gno.land/r/me/b"`)
	assert.Contains(t, page, `-pkgpath "gno.land/r/<your-g1-addr>/b"`)
	assert.Contains(t, page, `-pkgdir "./nested/b"`)
	assert.Contains(t, page, "-chainid dev -remote 127.0.0.1:26657")
	assert.Contains(t, page, "it signs as `g1jg8m…`")
}

func TestDevnetHomeMorePackages(t *testing.T) {
	t.Parallel()

	pkgs := map[string]string{}
	for i := range homeMaxPackages + 2 {
		pkgs[fmt.Sprintf("gno.land/p/me/p%d", i)] = fmt.Sprintf("p%d", i)
	}
	h, _ := testDevnetHome(t, pkgs)

	page := h.Render()
	assert.Contains(t, page, "- [p/me/p4](/p/me/p4)\n- and 2 more\n")
	assert.NotContains(t, page, "p/me/p5")
	// No realm: the commands keep their placeholders.
	assert.Contains(t, page, `-pkgpath "gno.land/r/<your-realm>"`)
	assert.Contains(t, page, `-pkgdir "."`)
}

func TestDevnetHomeEmpty(t *testing.T) {
	t.Parallel()

	cfg := defaultLocalAppConfig
	h := newDevnetHome(&mock.ServerEmitter{}, &cfg, nil, nil)

	page := h.Render()
	assert.Contains(t, page, "# Your Gno.land devnet\n")
	assert.Contains(t, page, "None yet: run `gnodev` in a folder with a `gnomod.toml`.\n")
	assert.Contains(t, page, "None yet: run a command from a realm's Actions page.\n")
	assert.Contains(t, page, `-pkgpath "gno.land/r/<your-g1-addr>/<your-realm>"`)
	assert.NotContains(t, page, "{{")
	assert.NotContains(t, page, "failed to build")
}

func TestDevnetHomeRefreshesPackages(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	var list []*packages.Package
	root := t.TempDir()
	cfg := defaultLocalAppConfig
	h := newDevnetHome(&mock.ServerEmitter{}, &cfg, []string{root}, nil)
	h.watch(func() []*packages.Package {
		mu.Lock()
		defer mu.Unlock()
		return list
	})
	require.Contains(t, h.Render(), "None yet: run `gnodev`")

	for i, evt := range []events.Event{&events.Reload{}, events.Reset{}} {
		name := fmt.Sprintf("r/me/r%d", i)
		mu.Lock()
		list = append(list, &packages.Package{ImportPath: "gno.land/" + name, Dir: filepath.Join(root, name)})
		mu.Unlock()

		h.Emit(evt)
		assert.Eventually(t, func() bool {
			return strings.Contains(h.Render(), "- ["+name+"]")
		}, 5*time.Second, 10*time.Millisecond, "a %T reads the packages again", evt)
	}
}

func TestDevnetHomeTransactions(t *testing.T) {
	t.Parallel()

	h, next := testDevnetHome(t, nil)
	evts := []events.Event{
		txResult(vm.MsgCall{PkgPath: "gno.land/r/me/a", Func: "Inc"}, 1_234_567, nil),
		txResult(vm.MsgAddPackage{Package: &std.MemPackage{Path: "gno.land/p/me/lib"}}, 850_000, fmt.Errorf("boom")),
		&events.Reload{},
		txResult(vm.MsgRun{}, 999, nil),
	}
	for _, evt := range evts {
		h.Emit(evt)
	}
	for _, want := range evts {
		assert.Equal(t, want, next.NextEvent(), "every event goes on to the event server")
	}

	page := h.Render()
	assert.Contains(t, page, "- `run` · 999 gas\n"+
		"- `addpkg` · lib · failed · 850k gas\n"+
		"- `Inc` · a · 1.23M gas\n",
		"newest first, kept across a reload")

	for i := range homeMaxTxs {
		h.Emit(txResult(vm.MsgCall{PkgPath: "gno.land/r/me/a", Func: fmt.Sprintf("F%d", i)}, 1, nil))
	}
	page = h.Render()
	assert.Contains(t, page, "- `F4`")
	assert.NotContains(t, page, "`run`", "only the latest are kept")

	h.Emit(&events.Reset{})
	assert.Contains(t, h.Render(), "None yet: run a command from a realm's Actions page.\n")
}

func TestDevnetHomeHostileInput(t *testing.T) {
	t.Parallel()

	h, _ := testDevnetHome(t, nil)
	h.Emit(txResult(vm.MsgCall{PkgPath: "gno.land/r/x)](javascript:alert(1)", Func: "A`](javascript:alert(1))<b>"}, 1, fmt.Errorf("x")))

	page := h.Render()
	assert.Contains(t, page, "- `Ajavascriptalert1b` · failed · 1 gas\n")
	assert.NotContains(t, page, "javascript:")

	cfg := defaultLocalAppConfig
	cfg.chainId = "dev\n\n<a href=\"javascript:alert(1)\">x</a>"
	cfg.nodeRPCListenerAddr = "tcp://127.0.0.1:26657"
	cfg.chainDomain = "gno.land\n<b>"
	page = newDevnetHome(&mock.ServerEmitter{}, &cfg, nil, nil).Render()
	assert.Contains(t, page, "-chainid <chain-id> -remote tcp://127.0.0.1:26657")
	assert.Contains(t, page, `-pkgpath "gno.land/r/<your-realm>"`)
	assert.NotContains(t, page, "javascript:")
	assert.NotContains(t, page, "<b>")
}

func TestRealmName(t *testing.T) {
	t.Parallel()

	for url, want := range map[string]string{
		"/r/dev/counter": "counter",
		"/r/a/x/v2":      "x/v2",
		"/r/solo":        "solo",
	} {
		assert.Equal(t, want, realmName(url), url)
	}
}

func TestDevnetHomeShownDir(t *testing.T) {
	t.Parallel()

	cwd := t.TempDir()
	h := &devnetHome{cwd: cwd}
	elsewhere := filepath.Join(t.TempDir(), "realm")
	for dir, want := range map[string]string{
		"":                                 ".",
		cwd:                                ".",
		filepath.Join(cwd, "a", "b"):       "./a/b",
		elsewhere:                          elsewhere,
		filepath.Join(cwd, "it's", "here"): "<your-realm-folder>",
	} {
		assert.Equal(t, want, h.shownDir(dir), dir)
	}
}

func TestFormatGas(t *testing.T) {
	t.Parallel()

	for gas, want := range map[int64]string{
		0:          "0",
		999:        "999",
		1_000:      "1k",
		850_400:    "850k",
		1_000_000:  "1M",
		1_500_000:  "1.5M",
		1_234_567:  "1.23M",
		1_960_735:  "1.96M",
		20_000_000: "20M",
	} {
		assert.Equal(t, want, formatGas(gas), "gas %d", gas)
	}
}

func TestGnodev_DevnetHome_ServedAtRoot(t *testing.T) {
	workspace := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(workspace, "gnowork.toml"), []byte(""), 0o644))
	writeWorkspacePkg(t, filepath.Join(workspace, "hello"), "gno.land/r/me/hello",
		"package hello\n\nfunc Render(_ string) string { return \"hi\" }\n")
	t.Chdir(workspace)

	for _, tc := range []struct {
		name     string
		cfg      func(*AppConfig)
		home     bool
		status   int
		location string
	}{
		{name: "default", cfg: func(*AppConfig) {}, home: true, status: http.StatusOK},
		{name: "path", cfg: func(c *AppConfig) { c.webHome = "/r/me/hello" }, status: http.StatusFound, location: "/r/me/hello"},
		{name: "none", cfg: func(c *AppConfig) { c.webHome = ":none:" }},
		{name: "staging", cfg: func(c *AppConfig) { c.staging, c.noExamples = true, true }},
		{name: "no-web", cfg: func(c *AppConfig) { c.noWeb = true }, status: http.StatusNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := defaultLocalAppConfig
			cfg.home = filepath.Join(t.TempDir(), "nokeybase")
			cfg.nodeRPCListenerAddr = "127.0.0.1:0"
			cfg.noWatch = true
			tc.cfg(&cfg)

			app := NewApp(discardLogger(), &cfg, commands.NewTestIO())
			require.NoError(t, app.Setup(context.Background(), workspace))
			t.Cleanup(app.Close)
			require.Equal(t, tc.home, app.home != nil)

			handler, err := app.setupHandlers(context.Background())
			require.NoError(t, err)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

			if tc.status != 0 {
				assert.Equal(t, tc.status, rr.Code)
			}
			if tc.location != "" {
				assert.Equal(t, tc.location, rr.Header().Get("Location"))
			}
			if tc.home {
				assert.Contains(t, rr.Body.String(), "Your Gno.land devnet")
				assert.Contains(t, rr.Body.String(), `href="/r/me/hello$state"`)
			} else {
				assert.NotContains(t, rr.Body.String(), "Your Gno.land devnet")
			}
		})
	}
}
