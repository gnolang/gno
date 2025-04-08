package txindexer

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	txIndexer = "tx-indexer"
	start     = "start"
)

// process manages the tx-indexer process
type process struct {
	cmd       *exec.Cmd
	logger    *slog.Logger
	cmdGetter func() *exec.Cmd
}

func newProcess(logger *slog.Logger, cfg Config) *process {
	return &process{
		logger:    logger,
		cmdGetter: newStartCmdGetter(cfg),
	}
}

func (p *process) start(ctx context.Context) error {
	p.newCmd()

	if p.cmd == nil {
		return fmt.Errorf("tx-indexer command is nil")
	}

	if err := p.syncCmdLogs(ctx); err != nil {
		return fmt.Errorf("failed to sync logs with stdout/stderr: %w", err)
	}

	err := p.cmd.Start()
	switch {
	case err == nil, strings.Contains(err.Error(), "already started"):
		return nil
	default:
		return fmt.Errorf("failed to start tx-indexer: %w", err)
	}
}

func (p *process) stop(ctx context.Context) error {
	if p.cmd == nil || p.cmd.Process == nil {
		return nil
	}

	defer p.reset()

	err := p.cmd.Process.Signal(os.Interrupt)
	switch {
	case err == nil:
		p.wait(ctx)
	case errors.Is(err, os.ErrProcessDone):
	default:
		return fmt.Errorf("failed to send interrupt signal: %w", err)
	}

	return nil
}

func (p *process) wait(ctx context.Context) {
	const graceful = time.Second * 3

	done := make(chan error)
	go func() {
		if p.cmd == nil || p.cmd.Process == nil {
			done <- nil
			return
		}

		done <- p.cmd.Wait()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case err := <-done:
			if err != nil && !errors.Is(err, os.ErrProcessDone) {
				p.logger.Error("tx-indexer process exited with error", "error", err)
			}
			return
		case <-time.After(graceful):
			_ = p.cmd.Process.Kill()
			p.logger.Error("tx-indexer process killed after timeout", "timeout", graceful)
			return
		}
	}
}

func (p *process) syncCmdLogs(ctx context.Context) error {
	stdout, err := p.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := p.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	log := func(r io.Reader) {
		s := bufio.NewScanner(r)
		for s.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
				p.logger.Info(s.Text())
			}
		}
	}

	go func() { log(stdout) }()
	go func() { log(stderr) }()

	return nil
}

func (p *process) newCmd() {
	if p.cmd != nil {
		return
	}
	p.cmd = p.cmdGetter()
}

func (p *process) reset() {
	p.cmd = nil
}

func startArgs(cfg Config) []string {
	args := []string{start}
	if cfg.DBPath != "" {
		args = append(args, "-db-path="+cfg.DBPath)
	}
	if cfg.HTTPRateLimit != nil {
		args = append(args, "-http-rate-limit="+fmt.Sprint(*cfg.HTTPRateLimit))
	}
	if cfg.ListenAddress != "" {
		args = append(args, "-listen-address="+cfg.ListenAddress)
	}
	if cfg.LogLevel != nil {
		args = append(args, "-log-level="+*cfg.LogLevel)
	}
	if cfg.MaxChunkSize != nil {
		args = append(args, "-max-chunk-size="+fmt.Sprint(*cfg.MaxChunkSize))
	}
	if cfg.MaxSlots != nil {
		args = append(args, "-max-slots="+fmt.Sprint(*cfg.MaxSlots))
	}
	if cfg.Remote != nil {
		args = append(args, "-remote="+*cfg.Remote)
	}

	return args
}

func newStartCmdGetter(cfg Config) func() *exec.Cmd {
	return func() *exec.Cmd {
		return exec.Command(txIndexer, startArgs(cfg)...)
	}
}
