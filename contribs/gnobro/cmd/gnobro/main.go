package main

import (
	"container/list"
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

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	charmlog "github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"

	"github.com/gnolang/gno/contribs/gnodev/pkg/events"
	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	tmlog "github.com/gnolang/gno/tm2/pkg/log"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	zone "github.com/lrstanley/bubblezone"
)

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

//go:embed banner2.txt
var banner string

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

	logger := tmlog.NewNoopLogger()
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

	// remoteAddr := resolveUnixOrTCPAddr(cfg.target)
	cl, err := client.NewHTTPClient(cfg.remote)
	if err != nil {
		return fmt.Errorf("unable to create http client for %q: %w", remoteAddr, err)
	}

	gnocl := &gnoclient.Client{
		RPCClient: cl,
		Signer:    signer,
	}
	base := gnoclient.BaseTxCfg{
		GasFee:    "1000000ugnot",
		GasWanted: 2000000,
	}

	broclient := NewBroClient(logger, base, gnocl)

	renderer := lipgloss.DefaultRenderer()
	input := initURLInput(renderer)

	var targetRealm string
	if len(args) > 0 {
		targetRealm = args[0]
	}
	switch {
	case targetRealm != "":
		path := strings.TrimLeft(targetRealm, gnoPrefix)
		input.SetValue(path)
	case cfg.defaultRealm != "":
		path := strings.TrimLeft(cfg.defaultRealm, gnoPrefix)
		input.SetValue(path)
	}

	cmd := initCommandInput(renderer)
	mod := model{
		render:       renderer,
		readonly:     cfg.readonly,
		client:       broclient,
		listFuncs:    newFuncList(),
		urlInput:     input,
		commandInput: cmd,
		zone:         zone.New(),
		pageurls:     map[string]string{},
		history:      list.New(),
	}

	if cfg.sshListener == "" {
		p := tea.NewProgram(mod,
			tea.WithAltScreen(), // use the full size of the terminal in its "alternate screen buffer"

			tea.WithMouseCellMotion(), // turn on mouse support so we can track the mouse wheel
		)

		host1, port1 := cfg.remote, "8888"
		if cfg.devEndpoint != "" {
			host1, port1, err = net.SplitHostPort(cfg.devEndpoint)
			if err != nil {
				return fmt.Errorf("unable to parse dev endpoint: %w", err)
			}
		}

		if !strings.HasPrefix(host1, "http://") && !strings.HasPrefix(host1, "https://") {
			host1 = "http://" + host1
		}

		devpoint, err := url.Parse(host1)
		if err != nil {
			return fmt.Errorf("unable to construct devaddr: %w", err)
		}

		host, port2, _ := net.SplitHostPort(devpoint.Host)
		devpoint.Host = host
		devpoint.Scheme = "ws"
		devpoint.Path = "_events"
		switch {
		case port1 != "":
			devpoint.Host += ":" + port1
		case port2 != "":
			devpoint.Host += ":" + port2
		}

		// var wg sync.WaitGroup
		var devcl DevClient
		// devcl.Logger = log.ZapLoggerToSlog(log.NewZapConsoleLogger(io.Out(), zapcore.DebugLevel))
		devcl.Handler = func(typ events.Type, data any) error {
			switch typ {
			case events.EvtReload, events.EvtReset, events.EvtTxResult:
				p.Send(UpdateRenderMsg{})
			default:
			}

			return nil
		}

		var wg sync.WaitGroup

		ctx, cancel := context.WithCancelCause(ctx)
		defer cancel(nil)

		// Setup trap signal
		osm.TrapSignal(func() {
			cancel(nil)
			p.Quit()
		})

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := devcl.Run(ctx, devpoint.String(), nil); err != nil {
				logger.Error("dev connection failed", "err", err)
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := p.Run()
			cancel(err)
			io.Println("Bye!")
		}()

		wg.Wait()
		return nil
	}

	teaHandler := func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
		model := mod // copy model
		model.readonly = cfg.readonly
		model.banner = fmt.Sprintf(banner, s.User())
		model.render = bubbletea.MakeRenderer(s)
		model.history = list.New() // set a new history

		if len(s.Command()) == 1 {
			path := filepath.Clean(s.Command()[0])
			model.urlInput.SetValue(path)
			model.urlInput.CursorEnd()
		}

		return model, []tea.ProgramOption{
			tea.WithAltScreen(),       // use the full size of the terminal in its "alternate screen buffer"
			tea.WithMouseCellMotion(), // turn on mouse support so we can track the mouse wheel
		}
	}

	sshaddr, err := net.ResolveTCPAddr("", cfg.sshListener)
	if err != nil {
		return fmt.Errorf("unable to resolve address: %w", err)
	}

	logger = slog.New(charmlog.New(io.Out()))
	s, err := wish.NewServer(
		wish.WithAddress(sshaddr.String()),
		wish.WithHostKeyPath(cfg.sshHostKeyPath),
		wish.WithMiddleware(
			NoCommandMiddleware(),
			activeterm.Middleware(), // Bubble Tea apps usually require a PTY.
			bubbletea.Middleware(teaHandler),
			logging.Middleware(),
		),
	)

	ctx, cancelCause := context.WithCancelCause(ctx)
	defer cancelCause(nil)

	// Setup trap signal
	osm.TrapSignal(func() {
		logger.Info("Stopping SSH server")
		if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			cancelCause(fmt.Errorf("Could not stop server: %w", err))
		} else {
			cancel()
		}
	})

	go func() {
		logger.Info("Starting SSH server", "addr", sshaddr.String())
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			cancelCause(err)
		} else {
			cancel()
		}
	}()

	<-ctx.Done()
	return context.Cause(ctx)
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

func NoCommandMiddleware() wish.Middleware {
	return func(sh ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {
			if len(s.Command()) > 0 {
				s.Exit(1)
			} else {
				sh(s)
			}
		}
	}
}
