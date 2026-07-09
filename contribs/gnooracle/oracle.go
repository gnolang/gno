package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	// Importing the vm package also registers its message types (MsgAddPackage,
	// MsgEnablePackage, ...) with amino so block txs decode correctly.
	vm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/gnolang/gno/tm2/pkg/amino"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

// oracle watches a chain, typechecks submitted packages off-chain, and
// broadcasts approvals for the ones that pass.
type oracle struct {
	cfg      config
	io       commands.IO
	client   gnoclient.Client
	approver crypto.Address

	// Disk-backed stores + cache used to typecheck candidate packages,
	// resolving stdlib and examples/ imports. Built once and reused.
	prodbs storetypes.CommitStore
	prodgs gno.Store
	testbs storetypes.CommitStore
	testgs gno.Store
	cache  gno.TypeCheckCache

	// seen dedupes packages already processed in this run.
	seen map[string]struct{}
}

func newOracle(cfg config, io commands.IO) (*oracle, error) {
	signer, err := gnoclient.SignerFromBip39(cfg.mnemonic, cfg.chainID, "", 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to build signer: %w", err)
	}
	if err := signer.Validate(); err != nil {
		return nil, fmt.Errorf("invalid signer: %w", err)
	}
	info, err := signer.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to read signer info: %w", err)
	}

	rpc, err := rpcclient.NewHTTPClient(cfg.remote)
	if err != nil {
		return nil, fmt.Errorf("failed to build RPC client: %w", err)
	}

	// Type-check stores mirror `gno lint`: production files against stdlibs +
	// examples, with a test-stdlib overlay. PreprocessOnly avoids executing
	// imported code — we only need the type information.
	prodbs, prodgs := test.StoreWithOptions(
		cfg.gnoRoot, io.Err(),
		test.StoreOptions{PreprocessOnly: true, WithExamples: true},
	)
	testbs, testgs := test.StoreWithOptions(
		cfg.gnoRoot, io.Err(),
		test.StoreOptions{PreprocessOnly: true, WithExamples: true, Testing: true, SourceStore: prodgs},
	)

	return &oracle{
		cfg:      cfg,
		io:       io,
		client:   gnoclient.Client{Signer: signer, RPCClient: rpc},
		approver: info.GetAddress(),
		prodbs:   prodbs,
		prodgs:   prodgs,
		testbs:   testbs,
		testgs:   testgs,
		cache:    make(gno.TypeCheckCache),
		seen:     make(map[string]struct{}),
	}, nil
}

// run polls the node for new blocks and processes each one, until ctx is done.
func (o *oracle) run(ctx context.Context) error {
	height := o.cfg.startHeight
	if height <= 0 {
		status, err := o.client.RPCClient.Status(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to query node status: %w", err)
		}
		height = status.SyncInfo.LatestBlockHeight + 1
	}

	ticker := time.NewTicker(o.cfg.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			o.io.Println("gnooracle: shutting down")
			return nil
		case <-ticker.C:
		}

		status, err := o.client.RPCClient.Status(ctx, nil)
		if err != nil {
			o.io.ErrPrintfln("gnooracle: status query failed: %v", err)
			continue
		}
		latest := status.SyncInfo.LatestBlockHeight

		for ; height <= latest; height++ {
			if err := o.processBlock(ctx, height); err != nil {
				o.io.ErrPrintfln("gnooracle: block %d processing failed: %v", height, err)
			}
		}
	}
}

// processBlock decodes a block's transactions and handles every MsgAddPackage.
func (o *oracle) processBlock(ctx context.Context, height int64) error {
	res, err := o.client.RPCClient.Block(ctx, &height)
	if err != nil {
		return err
	}
	if res.Block == nil {
		return nil
	}
	for _, raw := range res.Block.Data.Txs {
		var tx std.Tx
		if err := amino.Unmarshal(raw, &tx); err != nil {
			o.io.ErrPrintfln("gnooracle: skipping undecodable tx at height %d: %v", height, err)
			continue
		}
		for _, msg := range tx.Msgs {
			add, ok := msg.(vm.MsgAddPackage)
			if !ok || add.Package == nil {
				continue
			}
			o.handleCandidate(add.Package)
		}
	}
	return nil
}

// handleCandidate typechecks a submitted package and, if it passes, broadcasts
// a MsgEnablePackage to activate it on-chain.
func (o *oracle) handleCandidate(mpkg *std.MemPackage) {
	path := mpkg.Path
	if _, done := o.seen[path]; done {
		return
	}
	o.seen[path] = struct{}{}

	o.io.Printfln("gnooracle: typechecking %q", path)
	if err := o.typecheck(mpkg); err != nil {
		o.io.Printfln("gnooracle: %q rejected, not approving: %v", path, err)
		return
	}

	o.io.Printfln("gnooracle: %q passed typecheck, broadcasting approval", path)
	if err := o.enable(path); err != nil {
		o.io.ErrPrintfln("gnooracle: failed to approve %q: %v", path, err)
		return
	}
	o.io.Printfln("gnooracle: %q approved and enabled", path)
}

// typecheck runs the Gno typechecker on a candidate package, mirroring the
// on-chain check the validator will re-run at MsgEnablePackage time. Any panic
// from the typechecker is converted into an error so a single bad package can't
// crash the daemon.
func (o *oracle) typecheck(mpkg *std.MemPackage) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("typecheck panicked: %v", r)
		}
	}()

	// Best-effort: preload imports resolvable from disk (stdlibs + examples).
	// Missing on-chain-only imports surface as typecheck errors below.
	_ = test.LoadImports(o.testgs, mpkg, false)

	// Fresh transactions isolate each typecheck from the persistent base stores.
	newProdGnoStore := func() gno.Store {
		cw := o.prodbs.CacheWrap()
		return o.prodgs.BeginTransaction(cw, cw, nil, nil)
	}
	newTestGnoStore := func() gno.Store {
		cw := o.testbs.CacheWrap()
		return o.testgs.BeginTransaction(cw, cw, nil, nil)
	}

	_, errs := gno.TypeCheckMemPackage(mpkg, gno.TypeCheckOptions{
		Getter:     newProdGnoStore(),
		TestGetter: newTestGnoStore(),
		Mode:       gno.TCLatestStrict,
		Cache:      o.cache,
	})
	return errs
}

// enable builds, signs and broadcasts a MsgEnablePackage for pkgPath.
func (o *oracle) enable(pkgPath string) error {
	gasFee, err := std.ParseCoin(o.cfg.gasFee)
	if err != nil {
		return fmt.Errorf("invalid gas fee %q: %w", o.cfg.gasFee, err)
	}

	tx := std.Tx{
		Msgs:       []std.Msg{vm.MsgEnablePackage{Approver: o.approver, PkgPath: pkgPath}},
		Fee:        std.NewFee(o.cfg.gasWanted, gasFee),
		Signatures: nil,
	}

	// accountNumber/sequenceNumber == 0 lets SignTx auto-query the chain.
	signed, err := o.client.SignTx(tx, 0, 0)
	if err != nil {
		return fmt.Errorf("sign: %w", err)
	}
	// BroadcastTxCommit returns an error if CheckTx or DeliverTx failed.
	if _, err := o.client.BroadcastTxCommit(signed); err != nil {
		return fmt.Errorf("broadcast: %w", err)
	}
	return nil
}
