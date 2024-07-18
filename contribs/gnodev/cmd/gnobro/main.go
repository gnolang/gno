package main

import (
	"context"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	charmlog "github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"

	"github.com/gnolang/gno/contribs/gnodev/pkg/browser"
	"github.com/gnolang/gno/contribs/gnodev/pkg/events"
	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

//go:embed assets/banner_land_1.txt
var banner string

const gnoPrefix = "gno.land"

type broCfg struct {
	readonly       bool
	remote         string
	devEndpoint    string
	chainID        string
	defaultAccount string
	defaultRealm   string
	sshListener    string
	sshHostKeyPath string
}

var defaultBroOptions = broCfg{
	remote:         "127.0.0.1:26657",
	devEndpoint:    "",
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
			ShortUsage: "gnobro [flags]",
			ShortHelp:  "runs a cli browser.",
			LongHelp:   `run a cli browser`,
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
		"target",
		defaultBroOptions.remote,
		"target gnoland address",
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
		"default realm to display when gnobro start and no argument are provided",
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

	fs.StringVar(
		&c.devEndpoint,
		"dev",
		defaultBroOptions.devEndpoint,
		"dev endpoint, if empty will be default to `ws://<target>:8888`",
	)

	fs.BoolVar(
		&c.readonly,
		"readonly",
		defaultBroOptions.readonly,
		"readonly mode, no command allowed",
	)
}

func execBrowser(cfg *broCfg, args []string, io commands.IO) error {
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

	signer, err := getSignerForAccount(io, address, kb, cfg)
	if err != nil {
		return fmt.Errorf("unable to get signer for account %q: %w", address, err)
	}

	target := resolveUnixOrTCPAddr(cfg.remote)
	cl, err := client.NewHTTPClient(target)
	if err != nil {
		return fmt.Errorf("unable to create http client for %q: %w", target, err)
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
		path = strings.TrimLeft(cfg.defaultRealm, gnoPrefix)
	}

	bcfg := browser.DefaultConfig()
	bcfg.Readonly = cfg.readonly
	bcfg.Renderer = lipgloss.DefaultRenderer()
	bcfg.GnoClient = gnocl
	bcfg.URLDefaultValue = path
	bcfg.URLPrefix = gnoPrefix

	if cfg.sshListener == "" {
		return runLocal(ctx, cfg, bcfg, io)
	}

	return runServer(ctx, cfg, bcfg, io)
}

func runLocal(ctx context.Context, cfg *broCfg, bcfg browser.Config, io commands.IO) error {
	var err error

	model := browser.New(bcfg)
	p := tea.NewProgram(model,
		tea.WithAltScreen(), // use the full size of the terminal in its "alternate screen buffer"

		tea.WithMouseCellMotion(), // turn on mouse support so we can track the mouse wheel
	)

	devpoint, err := getDevEndpoint(cfg)
	if err != nil {
		return fmt.Errorf("unable to parse dev endpoint: %w", err)
	}

	var wg sync.WaitGroup

	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	// setup trap signal
	osm.TrapSignal(func() {
		cancel(nil)
		p.Quit()
	})

	if devpoint != "" {
		var devcl browser.DevClient
		devcl.Handler = func(typ events.Type, data any) error {
			switch typ {
			case events.EvtReload, events.EvtReset, events.EvtTxResult:
				p.Send(browser.RefreshRealm())
			default:
			}

			return nil
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := devcl.Run(ctx, devpoint, nil); err != nil {
				cancel(fmt.Errorf("dev connection failed: %w", err))
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := p.Run()
		cancel(err)
	}()

	wg.Wait()

	io.Println("Bye!")
	return context.Cause(ctx)
}

func runServer(ctx context.Context, cfg *broCfg, bcfg browser.Config, io commands.IO) error {
	// setup logger
	charmlogger := charmlog.New(io.Out())
	charmlogger.SetLevel(charmlog.DebugLevel)
	logger := slog.New(charmlogger)

	teaHandler := func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
		shortid := fmt.Sprintf("%.10s", s.Context().SessionID())

		bcfgCopy := bcfg // copy config

		bcfgCopy.Logger = logger.WithGroup(shortid)
		bcfgCopy.Renderer = bubbletea.MakeRenderer(s)

		switch len(s.Command()) {
		case 0:
			bcfgCopy.Banner = fmt.Sprintf(banner, s.User())
		case 1:
			// use command argument as path
			path := filepath.Clean(s.Command()[0])
			bcfgCopy.URLDefaultValue = path
		default:
		}

		bcfgCopy.Logger.Info("session started",
			"time", time.Now(),
			"path", bcfgCopy.URLDefaultValue,
			"sid", s.Context().SessionID(),
			"user", s.User())
		model := browser.New(bcfgCopy)
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
			logging.StructuredMiddlewareWithLogger(
				charmlogger, charmlog.DebugLevel,
			),
			activeterm.Middleware(), // Bubble Tea apps usually require a PTY.
			CommandLimiterMiddleware(),
		),
	)

	ctx, cancelCause := context.WithCancelCause(ctx)
	defer cancelCause(nil)

	// setup trap signal
	osm.TrapSignal(func() {
		logger.Info("stopping SSH server")
		if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			cancelCause(fmt.Errorf("could not stop server: %w", err))
		} else {
			cancelCause(nil)
		}
	})

	go func() {
		logger.Info("starting SSH server", "addr", sshaddr.String())
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			cancelCause(err)
		} else {
			cancelCause(nil)
		}
	}()

	<-ctx.Done()

	io.Println("Bye!")

	return context.Cause(ctx)
}

func getDevEndpoint(cfg *broCfg) (string, error) {
	var err error

	// use remote address as default
	host, port1 := cfg.remote, "8888"
	if cfg.devEndpoint != "" {
		// if any dev endpoint as been set, fallback on this
		host, port1, err = net.SplitHostPort(cfg.devEndpoint)
		if err != nil {
			return "", fmt.Errorf("unable to parse dev endpoint: %w", err)
		}
	}

	// ensure having a (any) protocol scheme
	if strings.Index(host, "://") < 0 {
		host = "http://" + host
	}

	// parse full host including port
	devpoint, err := url.Parse(host)
	if err != nil {
		return "", fmt.Errorf("unable to construct devaddr: %w", err)
	}

	host, _, _ = net.SplitHostPort(devpoint.Host)
	if port1 != "" {
		devpoint.Host = host + ":" + port1
	} else {
		devpoint.Host = host
	}

	switch devpoint.Scheme {
	case "ws", "wss":
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
	signer.ChainID = cfg.chainID // XXX: override this

	if ok, err := kb.HasByNameOrAddress(address); !ok || err != nil {
		if err != nil {
			return nil, fmt.Errorf("invalid name: %w", err)
		}

		return nil, fmt.Errorf("unknown name/address: %q", address)
	}

	// try empty password first
	if _, err := kb.ExportPrivKeyUnsafe(address, ""); err != nil {
		prompt := fmt.Sprintf("[%.10s] Enter password:", address)
		signer.Password, err = io.GetPassword(prompt, true)
		if err != nil {
			return nil, fmt.Errorf("error while reading password: %w", err)
		}

		if _, err := kb.ExportPrivKeyUnsafe(address, string(signer.Password)); err != nil {
			return nil, fmt.Errorf("invalid password: %w", err)
		}
	}

	return signer, nil
}

func resolveUnixOrTCPAddr(in string) (out string) {
	var err error
	var addr net.Addr

	if strings.HasPrefix(in, "unix://") {
		in = strings.TrimPrefix(in, "unix://")
		if addr, err := net.ResolveUnixAddr("unix", in); err == nil {
			return fmt.Sprintf("%s://%s", addr.Network(), addr.String())
		}

		err = fmt.Errorf("unable to resolve unix address `unix://%s`: %w", in, err)
	} else { // don't bother to checking prefix
		in = strings.TrimPrefix(in, "tcp://")
		if addr, err = net.ResolveTCPAddr("tcp", in); err == nil {
			return fmt.Sprintf("%s://%s", addr.Network(), addr.String())
		}

		err = fmt.Errorf("unable to resolve tcp address `tcp://%s`: %w", in, err)
	}

	panic(err)
}

func CommandLimiterMiddleware() wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {
			if len(s.Command()) > 1 {
				s.Exit(1)
			} else {
				next(s)
			}
		}
	}
}
