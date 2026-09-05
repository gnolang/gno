package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// verifyOneCmdName is the subcommand the daemon re-invokes itself with to
// verify a single package.
//
// The indirection buys the one thing a goroutine cannot give: a budget that is
// ENFORCED rather than merely observed. Go cannot kill a goroutine, so an
// in-process deadline can only abandon the work — which leaves it running,
// still consuming the CPU the budget was meant to bound, and still mutating
// stores the next attempt reads. A process can be killed.
//
// Three further properties come from the boundary rather than from code:
// per-attempt isolation (each child builds its own stores, so nothing survives
// to be raced), crash containment (the typechecker and preprocessor report
// errors by panicking, and a native-code crash or OOM would otherwise take the
// daemon with it), and responsive shutdown.
//
// The cost is a process spawn plus a store build per candidate. For a daemon
// whose entire job is deciding whether a package compiles quickly, that is the
// right trade.
const verifyOneCmdName = "verify-one"

// verifyOneConfig is the child's configuration: only what verification needs.
// No signer, no keystore — a child cannot approve anything, so the approver key
// never enters the process that compiles untrusted code.
type verifyOneConfig struct {
	gnoRoot string
	remote  string
}

func (c *verifyOneConfig) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.gnoRoot, "gno-root", "", "gno repository root")
	fs.StringVar(&c.remote, "remote", "", "RPC address, for resolving on-chain-only imports")
}

func newVerifyOneCmd(io commands.IO) *commands.Command {
	cfg := &verifyOneConfig{}
	return commands.NewCommand(
		commands.Metadata{
			Name:       verifyOneCmdName,
			ShortUsage: verifyOneCmdName + " [flags]",
			ShortHelp:  "verify one package read from stdin (used internally)",
			// Inherit nothing. AddSubCommands copies the parent's flags onto a
			// subcommand by default, which both collides with the two declared
			// below and would drag in the signing flags — and a verifier must
			// not have those: it handles untrusted input and cannot approve
			// anything, so the approver key has no business in its process.
			NoParentFlags: true,
			LongHelp: "Reads an amino-JSON MemPackage from stdin, type-checks and preprocesses " +
				"it, and exits zero if it passes or non-zero with the reason on stderr. gpao " +
				"invokes this on itself so that the verification budget can be enforced by " +
				"killing the process; it is not meant to be run by hand.",
		},
		cfg,
		func(ctx context.Context, _ []string) error {
			return execVerifyOne(ctx, cfg, io)
		},
	)
}

func execVerifyOne(_ context.Context, cfg *verifyOneConfig, cio commands.IO) error {
	raw, err := io.ReadAll(cio.In())
	if err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}
	var mpkg std.MemPackage
	if err := amino.UnmarshalJSON(raw, &mpkg); err != nil {
		return fmt.Errorf("decode mempackage: %w", err)
	}
	// Restate Type, which arrives nil — and load-bearing, not defensive.
	//
	// Not a codec problem: amino JSON does carry Type, as a typed Any. It is
	// nil because NewMsgAddPackage never sets it, so it was never in the tx;
	// the keeper restates it server-side for the same reason. And it IS read —
	// GoParseMemPackage asserts on it unchecked, so nil panics rather than
	// defaulting.
	//
	// MPUserAll, matching what AddPackage stamps on the stored package and
	// therefore what EnablePackage reads back. Not MPUserProd: a submitted
	// package legitimately contains _test.gno files, and MPUserProd rejects
	// them outright in AddMemPackage's validation — which would make this
	// oracle refuse nearly every real package and report it to the operator as
	// bad code. Restricting to production files is the type-check's job below
	// (ProdOnly), not the package type's.
	mpkg.Type = gno.MPUserAll

	v, err := newVerifier(cfg.gnoRoot, cfg.remote, cio.Err())
	if err != nil {
		return err
	}
	if err := v.prepare(&mpkg); err != nil {
		// The network under the resolver failed before any verdict was
		// possible; same channel as below.
		fmt.Fprintln(cio.Err(), err)
		os.Exit(exitResolverUnavailable)
	}
	// Everything the compile needs is local now. The parent starts the budget
	// on this line.
	fmt.Fprintln(cio.Out(), childReadyMarker)
	if err := v.verifyPackage(&mpkg); err != nil {
		if errors.Is(err, errResolverUnavailable) {
			// Not a verdict; say so with the exit status, which is the only
			// channel the parent classifies on. The reason still goes to
			// stderr, which the parent tees and reports.
			fmt.Fprintln(cio.Err(), err)
			os.Exit(exitResolverUnavailable)
		}
		return err
	}
	return nil
}

// exitResolverUnavailable is the child's exit status when verification could
// not obtain evidence -- the network under the import resolver failed -- as
// opposed to exiting 1 with a verdict. 2 belongs to the Go runtime (panic).
//
// Exited directly rather than through commands.ExitCodeError: the test
// harness drives the command through ParseAndRun, which returns that error
// instead of translating it into a status, and the parent classifies on the
// status alone.
const exitResolverUnavailable = 3

// childReadyMarker is the one line the child writes to stdout, once every
// source the compile needs is local. The parent starts the budget when it
// arrives, so what the budget measures is the compile and nothing else.
const childReadyMarker = "gpao: ready"

// verify runs one verification in a child process, killed if it outlasts the
// budget.
//
// This is the oracle's actual job. The chain re-runs both the type-check and
// the preprocess at MsgEnablePackage time and cannot bound how long they take —
// wall-clock is not a consensus quantity — so a correctness-only oracle
// contributes nothing the chain does not already do for itself. "This finishes
// quickly" is the claim only an off-chain actor can make, and it is what gates
// approval here.
//
// The child runs in two phases and the parent times them apart. Until the
// child reports ready it is fetching what the compile needs, from disk and
// from the node; that is bounded by the prepare budget and its expiry is
// unavailability. From ready onward it is compiling, which is what the
// validator will pay for; that is bounded by the verify budget and its expiry
// is an overrun. Timing the two together made a slow node read as a slow
// package.
//
// Exit status is the verdict: a clean exit passes, and a non-zero exit from a
// child that ran to completion is a rejection carrying the child's stderr as
// the reason. Two exits are not verdicts at all: exitResolverUnavailable, which
// says verification could not obtain evidence and becomes errVerifyUnavailable,
// and a deadline, which becomes errVerifyBudget. Upstream treats both as "no
// verdict yet" rather than as a rejection, and that distinction matters — a
// rejected package is settled, a slow one may just have lost a race with
// whatever else the machine was doing.
func (o *oracle) verify(ctx context.Context, mpkg *std.MemPackage) error {
	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot locate own binary to spawn a verifier: %w", err)
	}
	payload, err := amino.MarshalJSON(mpkg)
	if err != nil {
		return fmt.Errorf("encode mempackage: %w", err)
	}

	// Both deadlines cancel the CHILD's context, so expiry kills the process
	// rather than merely returning from the wait and leaving it running.
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	clock := newChildClock(cancel, o.cfg.prepareBudget, o.cfg.verifyBudget)
	defer clock.stop()

	args := []string{verifyOneCmdName, "-gno-root", o.cfg.gnoRoot}
	if o.cfg.remote != "" {
		args = append(args, "-remote", o.cfg.remote)
	}
	cmd := exec.CommandContext(runCtx, self, args...)
	cmd.Stdin = bytes.NewReader(payload)
	// Tee, don't capture. The buffer supplies the rejection reason on a
	// non-zero exit, but the child's stores also write diagnostics on the
	// SUCCESS path, and capturing alone silently dropped every one of those.
	// Bounded. Volume here is attacker-influenced: go/types is given an error
	// handler with no cap, so a package crafted to emit many errors would
	// otherwise be buffered in full and mirrored verbatim into the operator's
	// log. The reason for a rejection is in the first few KB.
	stderr := &boundedBuffer{limit: maxChildStderr}
	cmd.Stderr = io.Writer(stderr)
	if w := o.io.Err(); w != nil {
		// Guarded: commands.NewTestIO leaves Err nil, and MultiWriter over a
		// nil writer segfaults inside exec's copier goroutine rather than
		// anywhere near here.
		cmd.Stderr = io.MultiWriter(stderr, w)
	}
	// Stdout carries exactly one line, the ready marker, and starting the
	// budget on it is worth the pipe and the copying goroutine it costs.
	cmd.Stdout = &readyWatch{onReady: clock.ready}

	// Explicit env, so the invariant this file claims is actually true: the
	// process that compiles untrusted code does not hold the approver's key
	// material. Inheriting the parent's environment handed it $GPAO_MNEMONIC and
	// $GPAO_PASSWORD. Nothing in the child reads them, so this is
	// defence-in-depth -- but the comment above promised it and the code did not
	// deliver it. GNOROOT and HOME are kept because the type-checker resolves
	// stdlib paths through them.
	cmd.Env = childEnv()

	// If the parent's stderr is a blocked pipe -- `gpao | head`, a full
	// supervisor log -- Run waits on the copier goroutine as well as the
	// process, so the call could outlast the budget even though the child was
	// already killed. WaitDelay bounds that tail.
	cmd.WaitDelay = 5 * time.Second

	runErr := cmd.Run()

	// A clean exit is unambiguous, so it is checked first. Checking the
	// deadlines ahead of it would report a pass that landed just as a timer
	// fired as an overrun, and count it against the per-path budget allowance.
	if runErr == nil {
		return nil
	}

	// Only a child that RAN and chose to exit non-zero has judged the package.
	// Everything else is our own infrastructure failing -- fork hitting EAGAIN or
	// ENOMEM, the binary having been replaced under us, the OOM killer, a
	// SIGSEGV inside the type checker -- and treating those as a verdict marks a
	// perfectly good package permanently rejected, with no path that re-offers
	// it. Under a brief memory squeeze that condemns a whole queue.
	// ExitCode() is -1 when the process did not exit normally (signal death),
	// which is portable where ProcessState.Sys() is not.
	var ee *exec.ExitError
	if errors.As(runErr, &ee) && ee.ExitCode() >= 0 {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = runErr.Error()
		}
		if ee.ExitCode() == exitResolverUnavailable {
			return fmt.Errorf("%w: %s", errVerifyUnavailable, msg)
		}
		return errors.New(msg)
	}

	// A killed child: one of the two deadlines, or the parent shutting down.
	// Which deadline decides what it means, and reading either as a rejection
	// would permanently settle a package that was only slow, or only unlucky
	// in its node.
	if expired := clock.stop(); expired != nil {
		return expired
	}
	if ctx.Err() != nil {
		// Parent shutdown, not a verdict about this package.
		return fmt.Errorf("%w: shutting down", errVerifyBudget)
	}
	return fmt.Errorf("%w: the verifier could not complete, this is not a "+
		"verdict about the package: %v", errVerifyUnavailable, runErr)
}

// childClock times the child's two phases apart. From spawn until the child
// reports ready it runs against the prepare budget; from ready it runs against
// the verify budget. Whichever expires cancels the child's context, which kills
// it, and is remembered so the kill can be classified afterwards.
type childClock struct {
	mu      sync.Mutex
	kill    context.CancelFunc
	verify  time.Duration
	timer   *time.Timer
	expired error
}

func newChildClock(kill context.CancelFunc, prepare, verify time.Duration) *childClock {
	c := &childClock{kill: kill, verify: verify}
	c.timer = time.AfterFunc(prepare, func() {
		c.expire(fmt.Errorf("%w: the verifier did not have its sources within the "+
			"-prepare-budget of %s", errVerifyUnavailable, prepare))
	})
	return c
}

// ready switches from the prepare budget to the verify budget.
func (c *childClock) ready() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.expired != nil {
		return
	}
	c.timer.Stop()
	c.timer = time.AfterFunc(c.verify, func() {
		c.expire(fmt.Errorf("%w: exceeded %s and was killed", errVerifyBudget, c.verify))
	})
}

func (c *childClock) expire(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.expired == nil {
		c.expired = err
		c.kill()
	}
}

// stop disarms the running timer and reports which deadline expired, if any.
func (c *childClock) stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.timer.Stop()
	return c.expired
}

// readyWatch is the child's stdout: it calls onReady once, on the first line
// that is the ready marker, wherever that line lands. Nothing else is expected
// on stdout, and whatever else arrives is dropped rather than read into.
type readyWatch struct {
	onReady func()
	line    bytes.Buffer
	ready   bool
}

// maxReadyLine bounds what is buffered of a line still waiting for its newline;
// a longer one cannot be the marker and is dropped as it streams.
const maxReadyLine = 256

func (w *readyWatch) Write(p []byte) (int, error) {
	// Report the full length whatever is consumed: a short write would make
	// exec's copier report an error, which would be misread as the child
	// failing.
	n := len(p)
	for !w.ready && len(p) > 0 {
		i := bytes.IndexByte(p, '\n')
		if i < 0 {
			if w.line.Len()+len(p) <= maxReadyLine {
				w.line.Write(p)
			} else {
				w.line.Reset()
				w.line.WriteByte(0) // poison: this line is already too long
			}
			break
		}
		w.line.Write(p[:i])
		if strings.TrimSpace(w.line.String()) == childReadyMarker {
			w.ready = true
			w.onReady()
		}
		w.line.Reset()
		p = p[i+1:]
	}
	return n, nil
}

// maxChildStderr bounds what we retain and mirror from a child.
const maxChildStderr = 16 << 10

// boundedBuffer keeps at most limit bytes and records that it truncated.
type boundedBuffer struct {
	buf       bytes.Buffer
	limit     int
	truncated bool
}

func (b *boundedBuffer) Write(p []byte) (int, error) {
	if room := b.limit - b.buf.Len(); room > 0 {
		if len(p) <= room {
			b.buf.Write(p)
		} else {
			b.buf.Write(p[:room])
			b.truncated = true
		}
	} else if len(p) > 0 {
		b.truncated = true
	}
	// Report the full length: a short write would make exec's copier report an
	// error, which would then be misread as the child failing.
	return len(p), nil
}

func (b *boundedBuffer) String() string {
	if b.truncated {
		return b.buf.String() + "\n[truncated]"
	}
	return b.buf.String()
}

// childEnv is the environment the verifier runs with: enough to resolve stdlib
// paths, and none of the approver's credentials.
func childEnv() []string {
	var out []string
	for _, k := range childEnvAllowed {
		if v := os.Getenv(k); v != "" {
			out = append(out, k+"="+v)
		}
	}
	return out
}

// childEnvAllowed is everything the verifier is given. Notably absent:
// GPAO_MNEMONIC and GPAO_PASSWORD.
//
// GPAO_TEST_SPIN_HEARTBEAT is a test hook -- it puts a child into an endless
// heartbeat loop so the budget test can prove the process is actually killed
// rather than merely abandoned. It is listed here rather than left to
// inheritance so that the allow-list is the single, readable statement of what
// crosses into the process that compiles untrusted code.
var childEnvAllowed = []string{
	"HOME", "GNOROOT", "PATH", "TMPDIR",
	"GPAO_TEST_SPIN_HEARTBEAT",
}
