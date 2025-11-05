package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/url"
	"os"
	"os/signal"
	gopath "path"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	charmlog "github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"golang.org/x/sync/errgroup"

	"github.com/gnolang/gno/contribs/gnobro/pkg/browser"
	"github.com/gnolang/gno/contribs/gnodev/pkg/events"
	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
)

const gnoPrefix = "gno.land"

type broCfg struct {
	readonly       bool
	remote         string
	dev            bool
	devRemote      string
	chainID        string
	defaultAccount string
	defaultRealm   string
	sshListener    string
	sshHostKeyPath string
	banner         bool
	jsonlog        bool
}

var defaultBroOptions = broCfg{
	remote:         "127.0.0.1:26657",
	dev:            true,
	devRemote:      "",
	sshListener:    "",
	defaultRealm:   "gno.land/r/gnoland/home",
	chainID:        "dev",
	sshHostKeyPath: ".ssh/id_ed25519",
}

func main() {
	cfg := &broCfg{}

	stdio := commands.NewDefaultIO()
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "gnobro",
			ShortUsage: "gnobro [flags] [pkg_path]",
			ShortHelp:  "Gno Browser, a realm explorer",
			LongHelp: `Gnobro is a terminal user interface (TUI) that allows you to browse realms within your
terminal. It automatically connects to Gnodev for real-time development. In
addition to hot reload, it also has the ability to execute commands and interact
with your realm.
`,
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execBrowser(cfg, args, stdio)
		})

	cmd.Execute(context.Background(), os.Args[1:])
}

func (c *broCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.remote,
		"remote",
		defaultBroOptions.remote,
		"remote gno.land URL",
	)

	fs.StringVar(
		&c.chainID,
		"chainid",
		defaultBroOptions.chainID,
		"chainid",
	)

	fs.StringVar(
		&c.defaultAccount,
		"account",
		defaultBroOptions.defaultAccount,
		"default local account to use",
	)

	fs.StringVar(
		&c.defaultRealm,
		"default-realm",
		defaultBroOptions.defaultRealm,
		"default realm to display when gnobro starts and no argument is provided",
	)

	fs.StringVar(
		&c.sshListener,
		"ssh",
		defaultBroOptions.sshListener,
		"ssh server listener address",
	)

	fs.StringVar(
		&c.sshHostKeyPath,
		"ssh-key",
		defaultBroOptions.sshHostKeyPath,
		"ssh host key path",
	)

	fs.BoolVar(
		&c.dev,
		"dev",
		defaultBroOptions.dev,
		"enable dev mode and connect to gnodev for realtime update",
	)

	fs.StringVar(
		&c.devRemote,
		"dev-remote",
		defaultBroOptions.devRemote,
		"dev endpoint, if empty will default to `ws://<target>:8888`",
	)

	fs.BoolVar(
		&c.banner,
		"banner",
		defaultBroOptions.banner,
		"if enabled, display a banner",
	)

	fs.BoolVar(
		&c.readonly,
		"readonly",
		defaultBroOptions.readonly,
		"readonly mode, no commands allowed",
	)

	fs.BoolVar(
		&c.jsonlog,
		"jsonlog",
		defaultBroOptions.jsonlog,
		"display server log as json format",
	)
}

func execBrowser(cfg *broCfg, args []string, cio commands.IO) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	home := gnoenv.HomeDir()

	var address string
	var kb keys.Keybase
	if cfg.defaultAccount != "" {
		address = cfg.defaultAccount

		var err error
		kb, err = keys.NewKeyBaseFromDir(home)
		if err != nil {
			return fmt.Errorf("unable to load keybase: %w", err)
		}
	} else {
		// create a inmemory keybase
		kb = keys.NewInMemory()
		kb.CreateAccount(integration.DefaultAccount_Name, integration.DefaultAccount_Seed, "", "", 0, 0)
		address = integration.DefaultAccount_Name
	}

	signer, err := getSignerForAccount(cio, address, kb, cfg)
	if err != nil {
		return fmt.Errorf("unable to get signer for account %q: %w", address, err)
	}

	cl, err := client.NewHTTPClient(cfg.remote)
	if err != nil {
		return fmt.Errorf("unable to create http client for %q: %w", cfg.remote, err)
	}

	gnocl := &gnoclient.Client{
		RPCClient: cl,
		Signer:    signer,
	}

	var path string
	switch {
	case len(args) > 0:
		path = strings.TrimSpace(args[0])
		path = strings.TrimPrefix(path, gnoPrefix)
	case cfg.defaultRealm != "":
		path = strings.TrimPrefix(cfg.defaultRealm, gnoPrefix)
	}

	bcfg := browser.DefaultConfig()
	bcfg.DevMode = cfg.dev
	bcfg.Readonly = cfg.readonly
	bcfg.Renderer = lipgloss.DefaultRenderer()
	bcfg.URLDefaultValue = path
	bcfg.URLPrefix = gnoPrefix

	if cfg.sshListener == "" {
		if cfg.banner {
			bcfg.Banner = NewGnoLandBanner()
		}

		return runLocal(ctx, gnocl, cfg, bcfg, cio)
	}

	return runServer(ctx, gnocl, cfg, bcfg, cio)
}

func runLocal(ctx context.Context, gnocl *gnoclient.Client, cfg *broCfg, bcfg browser.Config, io commands.IO) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	model := browser.New(bcfg, gnocl)
	p := tea.NewProgram(model,
		tea.WithContext(ctx),
		tea.WithAltScreen(),       // use the full size of the terminal in its "alternate screen buffer"
		tea.WithMouseCellMotion(), // turn on mouse support so we can track the mouse wheel
	)

	var errgs errgroup.Group
	if cfg.dev {
		devpoint, err := getDevEndpoint(cfg)
		if err != nil {
			return fmt.Errorf("unable to parse dev endpoint: %w", err)
		}

		var devcl browser.DevClient
		devcl.Handler = func(typ events.Type, data any) error {
			switch typ {
			case events.EvtReload, events.EvtReset, events.EvtTxResult:
				p.Send(browser.RefreshRealmMsg())
			default:
			}

			return nil
		}

		devcl.HandlerStatus = func(s browser.ClientStatus, remote string) {
			p.Send(browser.DevClientStatusUpdateMsg(s, remote))
		}

		errgs.Go(func() error {
			defer cancel()

			if err := devcl.Run(ctx, devpoint, nil); err != nil {
				return fmt.Errorf("dev connection failed: %w", err)
			}

			return nil
		})
	}

	errgs.Go(func() error {
		defer cancel()

		_, err := p.Run()
		return err
	})

	if err := errgs.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	io.Println("Bye!")
	return nil
}

func runServer(ctx context.Context, gnocl *gnoclient.Client, cfg *broCfg, bcfg browser.Config, io commands.IO) error {
	// setup logger
	logger := newLogger(io.Out(), cfg.jsonlog)

	teaHandler := func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
		shortid := fmt.Sprintf("%.10s", s.Context().SessionID())

		bcfgCopy := bcfg // copy config

		bcfgCopy.Logger = logger.WithGroup(shortid)
		bcfgCopy.Renderer = bubbletea.MakeRenderer(s)

		if cfg.banner {
			bcfgCopy.Banner = NewGnoLandBanner()
		}

		pval := s.Context().Value("path")
		if path, ok := pval.(string); ok && len(path) > 0 {
			// Erase banner on specifc command
			bcfgCopy.Banner = browser.ModelBanner{}
			// Set up url
			bcfgCopy.URLDefaultValue = path
		}

		bcfgCopy.Logger.Info("session started",
			"time", time.Now(),
			"path", bcfgCopy.URLDefaultValue,
			"sid", s.Context().SessionID(),
			"user", s.User())
		model := browser.New(bcfgCopy, gnocl)

		return model, []tea.ProgramOption{
			tea.WithAltScreen(),       // use the full size of the terminal in its "alternate screen buffer"
			tea.WithMouseCellMotion(), // turn on mouse support so we can track the mouse wheel
		}
	}

	sshaddr, err := net.ResolveTCPAddr("", cfg.sshListener)
	if err != nil {
		return fmt.Errorf("unable to resolve address: %w", err)
	}

	s, err := wish.NewServer(
		wish.WithAddress(sshaddr.String()),
		wish.WithHostKeyPath(cfg.sshHostKeyPath),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			activeterm.Middleware(), // ensure PTY
			ValidatePathCommandMiddleware(bcfg.URLPrefix),
			StructuredMiddlewareWithLogger(
				ctx, logger, slog.LevelInfo,
			),
			// XXX: add ip throttler
		),
	)
	if err != nil {
		return fmt.Errorf("unable to create SSH server: %w", err)
	}

	var errgs errgroup.Group

	errgs.Go(func() error {
		logger.Info("starting SSH server", "addr", sshaddr.String())
		return s.ListenAndServe()
	})

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	errgs.Go(func() error {
		<-ctx.Done()

		logger.Info("stopping SSH server... (5s timeout)")

		sctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		return s.Shutdown(sctx)
	})

	if err := errgs.Wait(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		return err
	}

	if !cfg.jsonlog {
		io.Println("Bye!")
	}
	return nil
}

func getDevEndpoint(cfg *broCfg) (string, error) {
	var err error

	// use remote address as default
	host, port := cfg.remote, "8888"
	if cfg.devRemote != "" {
		// if any dev endpoint as been set, fallback on this
		host, port, err = net.SplitHostPort(cfg.devRemote)
		if err != nil {
			return "", fmt.Errorf("unable to parse dev endpoint: %w", err)
		}
	}

	// ensure having a (any) protocol scheme
	if !strings.Contains(host, "://") {
		host = "http://" + host
	}

	// parse full host including port
	devpoint, err := url.Parse(host)
	if err != nil {
		return "", fmt.Errorf("unable to construct devaddr: %w", err)
	}

	host, _, _ = net.SplitHostPort(devpoint.Host)
	if port != "" {
		devpoint.Host = host + ":" + port
	} else {
		devpoint.Host = host
	}

	switch devpoint.Scheme {
	case "ws", "wss": // already good
	case "https":
		devpoint.Scheme = "wss"
	default:
		devpoint.Scheme = "ws"
	}
	devpoint.Path = "_events"

	return devpoint.String(), nil
}

func getSignerForAccount(io commands.IO, address string, kb keys.Keybase, cfg *broCfg) (gnoclient.Signer, error) {
	var signer gnoclient.SignerFromKeybase

	signer.Keybase = kb
	signer.Account = address
	signer.ChainID = cfg.chainID

	if ok, err := kb.HasByNameOrAddress(address); !ok || err != nil {
		if err != nil {
			return nil, fmt.Errorf("invalid name: %w", err)
		}

		return nil, fmt.Errorf("unknown name/address: %q", address)
	}

	// try empty password first
	if _, err := kb.ExportPrivKey(address, ""); err != nil {
		prompt := fmt.Sprintf("[%.10s] Enter password:", address)
		signer.Password, err = io.GetPassword(prompt, true)
		if err != nil {
			return nil, fmt.Errorf("error while reading password: %w", err)
		}

		if _, err := kb.ExportPrivKey(address, signer.Password); err != nil {
			return nil, fmt.Errorf("invalid password: %w", err)
		}
	}

	return signer, nil
}

func ValidatePathCommandMiddleware(pathPrefix string) wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {
			switch cmd := s.Command(); len(cmd) {
			case 0: // ok
				next(s)
				return
			case 1: // check for valid path
				path := cmd[0]
				if strings.HasPrefix(path, pathPrefix) && gopath.Clean(path) == path {
					s.Context().SetValue("path", path)
					next(s)
					return
				}

				fmt.Fprintln(s.Stderr(), "provided path is invalid")
			default:
				fmt.Fprintln(s.Stderr(), "too many arguments")
			}

			s.Exit(1)
		}
	}
}

func StructuredMiddlewareWithLogger(ctx context.Context, logger *slog.Logger, level slog.Level) wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(sess ssh.Session) {
			ct := time.Now()
			hpk := sess.PublicKey() != nil
			pty, _, _ := sess.Pty()
			logger.Log(
				ctx,
				level,
				"connect",
				"user", sess.User(),
				"remote-addr", sess.RemoteAddr().String(),
				"public-key", hpk,
				"command", sess.Command(),
				"term", pty.Term,
				"width", pty.Window.Width,
				"height", pty.Window.Height,
				"client-version", sess.Context().ClientVersion(),
			)
			next(sess)
			logger.Log(
				ctx,
				level,
				"disconnect",
				"user", sess.User(),
				"remote-addr", sess.RemoteAddr().String(),
				"duration", time.Since(ct),
			)
		}
	}
}

func newLogger(out io.Writer, json bool) *slog.Logger {
	if json {
		return slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}

	charmlogger := charmlog.New(out)
	charmlogger.SetLevel(charmlog.DebugLevel)
	return slog.New(charmlogger)
}
