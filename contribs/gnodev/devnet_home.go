package main

import (
	"cmp"
	_ "embed"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"text/template"

	"github.com/gnolang/gno/contribs/gnodev/pkg/emitter"
	"github.com/gnolang/gno/contribs/gnodev/pkg/events"
	"github.com/gnolang/gno/contribs/gnodev/pkg/packages"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
)

//go:embed devnet_home.md.tmpl
var devnetHomeSource string

var devnetHomeTemplate = template.Must(template.New("devnet_home").Parse(devnetHomeSource))

const (
	homeMaxPackages = 5
	homeMaxTxs      = 5
)

var (
	// homePathChars are the characters a path or flag value may show with:
	// anything else could break out of a markdown link or a code block.
	homePathChars = regexp.MustCompile(`^[a-zA-Z0-9_./-]+$`)
	homeFlagChars = regexp.MustCompile(`^[a-zA-Z0-9_.:/-]+$`)
	homeDirChars  = regexp.MustCompile(`^[a-zA-Z0-9_./ -]+$`)
)

// devnetHome is the page gnoweb shows at / in local mode: the developer's
// packages, their latest transactions and how to use this chain. It sits
// between the node and the event server to see each transaction.
type devnetHome struct {
	next    emitter.Emitter
	chainID string
	rpc     string
	domain  string
	cwd     string
	roots   []string            // dirs holding the developer's packages
	primary map[string]struct{} // paths gnodev was started on

	mu       sync.Mutex
	listPkgs func() []*packages.Package
	pkgs     []homePackage // as of the last reset or reload
	morePkgs int
	gen      int // the last package refresh started, and applied
	applied  int
	txs      []homeTx // newest first
}

type homePackage struct {
	Name    string
	URL     string // empty for a path gnoweb does not serve
	Realm   bool
	path    string
	dir     string
	primary bool
}

type homeTx struct {
	Action string
	Target string
	OK     bool
	Gas    string
}

func newDevnetHome(next emitter.Emitter, cfg *AppConfig, roots, primary []string) *devnetHome {
	h := &devnetHome{
		next:    next,
		chainID: flagValue(cfg.chainId, "<chain-id>"),
		rpc:     flagValue(cfg.nodeRPCListenerAddr, "<rpc>"),
		domain:  DefaultDomain,
		primary: make(map[string]struct{}, len(primary)),
	}
	if homePathChars.MatchString(cfg.chainDomain) {
		h.domain = cfg.chainDomain
	}
	h.cwd, _ = os.Getwd()
	for _, root := range roots {
		if abs, err := filepath.Abs(root); err == nil {
			h.roots = append(h.roots, abs)
		}
	}
	for _, p := range primary {
		h.primary[p] = struct{}{}
	}
	return h
}

// watch reads the node's packages now, and again after each reset or reload.
func (h *devnetHome) watch(listPkgs func() []*packages.Package) {
	h.mu.Lock()
	h.listPkgs = listPkgs
	h.mu.Unlock()
	h.refreshPackages()
}

// Emit records each transaction, and passes every event on. A reload replays
// the transactions, so the list stays; a reset drops them, so it clears.
func (h *devnetHome) Emit(evt events.Event) {
	switch e := evt.(type) {
	case events.TxResult:
		h.record(e)
	case events.Reset, *events.Reset:
		h.mu.Lock()
		h.txs = nil
		h.mu.Unlock()
		go h.refreshPackages()
	case events.Reload, *events.Reload:
		go h.refreshPackages()
	}
	h.next.Emit(evt)
}

// refreshPackages runs off the event path: the node emits while it holds the
// lock that ListPkgs takes.
func (h *devnetHome) refreshPackages() {
	h.mu.Lock()
	listPkgs := h.listPkgs
	h.gen++
	gen := h.gen
	h.mu.Unlock()
	if listPkgs == nil {
		return
	}

	pkgs, more := h.ownPackages(listPkgs())

	h.mu.Lock()
	defer h.mu.Unlock()
	if gen > h.applied {
		h.pkgs, h.morePkgs, h.applied = pkgs, more, gen
	}
}

func (h *devnetHome) record(res events.TxResult) {
	if len(res.Tx.Msgs) == 0 {
		return
	}
	tx := homeTx{OK: res.Response.IsOK(), Gas: formatGas(res.Response.GasUsed)}
	switch msg := res.Tx.Msgs[0].(type) {
	case vm.MsgCall:
		tx.Action, tx.Target = msg.Func, packageName(msg.PkgPath)
	case vm.MsgAddPackage:
		tx.Action = "addpkg"
		if msg.Package != nil {
			tx.Target = packageName(msg.Package.Path)
		}
	case vm.MsgRun:
		tx.Action = "run"
	case bank.MsgSend:
		tx.Action = "send"
	default:
		tx.Action = msg.Type()
	}
	tx.Action = identifier(tx.Action)

	h.mu.Lock()
	defer h.mu.Unlock()
	h.txs = slices.Insert(h.txs, 0, tx)
	if len(h.txs) > homeMaxTxs {
		h.txs = h.txs[:homeMaxTxs]
	}
}

// Render builds the page, on each request.
func (h *devnetHome) Render() string {
	h.mu.Lock()
	pkgs, more, txs := h.pkgs, h.morePkgs, slices.Clone(h.txs)
	h.mu.Unlock()

	data := struct {
		Packages     []homePackage
		MorePackages int
		Txs          []homeTx
		ChainID, RPC string
		Agent        string
		Realm        string
		RealmName    string
		RealmDir     string
	}{
		Packages:     pkgs,
		MorePackages: more,
		Txs:          txs,
		ChainID:      h.chainID,
		RPC:          h.rpc,
		Agent:        defaultDeployerAddress.String()[:6] + "…",
		Realm:        h.domain + "/r/<your-realm>",
		RealmName:    "<your-realm>",
		RealmDir:     ".",
	}
	if i := slices.IndexFunc(pkgs, func(p homePackage) bool { return p.Realm }); i >= 0 {
		data.Realm = pkgs[i].path
		data.RealmName = realmName(pkgs[i].URL)
		data.RealmDir = h.shownDir(pkgs[i].dir)
	}

	var page strings.Builder
	if err := devnetHomeTemplate.Execute(&page, data); err != nil {
		return "# Your Gno.land devnet\n\nThis page failed to build: " + err.Error() + "\n"
	}
	return page.String()
}

// ownPackages lists the developer's packages: the ones gnodev was started
// on, then realms, then the rest. It returns how many it left out.
func (h *devnetHome) ownPackages(all []*packages.Package) ([]homePackage, int) {
	var own []homePackage
	for _, pkg := range all {
		_, primary := h.primary[pkg.ImportPath]
		if (!primary && !h.isOwn(pkg.Dir)) || !homePathChars.MatchString(pkg.ImportPath) {
			continue
		}
		hp := homePackage{Name: pkg.ImportPath, path: pkg.ImportPath, dir: pkg.Dir, primary: primary}
		if rest, ok := strings.CutPrefix(pkg.ImportPath, h.domain+"/"); ok {
			hp.Name, hp.URL = rest, "/"+rest
			hp.Realm = strings.HasPrefix(rest, "r/")
		}
		own = append(own, hp)
	}
	slices.SortFunc(own, func(a, b homePackage) int {
		if a.primary != b.primary {
			return boolOrder(a.primary)
		}
		if a.Realm != b.Realm {
			return boolOrder(a.Realm)
		}
		return cmp.Compare(a.Name, b.Name)
	})
	if len(own) <= homeMaxPackages {
		return own, 0
	}
	return own[:homeMaxPackages], len(own) - homeMaxPackages
}

func (h *devnetHome) isOwn(dir string) bool {
	if dir == "" {
		return false
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		return false
	}
	for _, root := range h.roots {
		if rel, err := filepath.Rel(root, dir); err == nil && filepath.IsLocal(rel) {
			return true
		}
	}
	return false
}

// shownDir is a package's folder as the deploy command takes it: relative to
// where gnodev runs when it is below it.
func (h *devnetHome) shownDir(dir string) string {
	if dir == "" {
		return "."
	}
	if rel, err := filepath.Rel(h.cwd, dir); err == nil && filepath.IsLocal(rel) {
		dir = "./" + filepath.ToSlash(rel)
		if rel == "." {
			dir = "."
		}
	}
	if !homeDirChars.MatchString(dir) {
		return "<your-realm-folder>"
	}
	return dir
}

// realmName is a realm's path below its namespace, as a deploy under another
// namespace keeps it: /r/dev/counter gives counter, /r/a/x/v2 gives x/v2.
func realmName(url string) string {
	parts := strings.Split(strings.TrimPrefix(url, "/r/"), "/")
	if len(parts) < 2 {
		return parts[0]
	}
	return strings.Join(parts[1:], "/")
}

// packageName is a package's last element, or nothing for a path that could
// break the line it is written on.
func packageName(pkgPath string) string {
	if !homePathChars.MatchString(pkgPath) {
		return ""
	}
	return path.Base(pkgPath)
}

// flagValue keeps a flag value that can sit in a code span and a shell
// command as it is, and shows placeholder for any other.
func flagValue(v, placeholder string) string {
	if homeFlagChars.MatchString(v) {
		return v
	}
	return placeholder
}

// boolOrder sorts true first.
func boolOrder(b bool) int {
	if b {
		return -1
	}
	return 1
}

// identifier keeps what a function name may hold, as a transaction can carry
// any string there.
func identifier(s string) string {
	s = strings.Map(func(r rune) rune {
		if r == '_' || ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') || ('0' <= r && r <= '9') {
			return r
		}
		return -1
	}, s)
	if len(s) > 40 {
		s = s[:40]
	}
	return cmp.Or(s, "?")
}

func formatGas(gas int64) string {
	switch {
	case gas >= 1_000_000:
		m := strconv.FormatFloat(float64(gas)/1e6, 'f', 2, 64)
		return strings.TrimSuffix(strings.TrimRight(m, "0"), ".") + "M"
	case gas >= 1_000:
		return strconv.FormatInt(gas/1_000, 10) + "k"
	default:
		return strconv.FormatInt(gas, 10)
	}
}
