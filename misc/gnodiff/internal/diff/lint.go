package diff

import (
	"fmt"
	"io"
	"sync"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// RunLint type-checks src with the same go/types pass gno lint uses.
// A program gc rejects and this pass also rejects, while RunGno still
// executes it, is classified as "run accepts what lint rejects".
func RunLint(src string, opt Options) (res Result) {
	opt, err := opt.resolve()
	if err != nil {
		return Result{Kind: KindInternal, Detail: err.Error()}
	}
	defer func() {
		if r := recover(); r != nil {
			res = resultFromRecover(r)
			if res.Kind == KindPanic {
				res = Result{Kind: KindInternal, Detail: clip(fmt.Sprint(r), detailLimit)}
			}
		}
	}()

	st, err := lintGetter(opt.RootDir)
	if err != nil {
		return Result{Kind: KindInternal, Detail: err.Error()}
	}

	const pkgPath = "gno.land/p/demo/gnodiff/main"
	mpkg := &std.MemPackage{
		Type: gnolang.MPUserProd,
		Name: "main",
		Path: pkgPath,
		Files: []*std.MemFile{
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
			{Name: "main.gno", Body: src},
		},
	}

	lintMu.Lock()
	defer lintMu.Unlock()
	_, tcErr := gnolang.TypeCheckMemPackage(mpkg, gnolang.TypeCheckOptions{
		Getter:     st,
		TestGetter: st,
		Mode:       gnolang.TCLatestRelaxed,
	})
	if tcErr != nil {
		return Result{Kind: KindReject, Detail: clip(tcErr.Error(), detailLimit)}
	}
	return Result{Kind: KindOK}
}

var (
	lintMu    sync.Mutex
	lintOnce  sync.Once
	lintStore gnolang.Store
	lintErr   error
)

func lintGetter(root string) (gnolang.Store, error) {
	lintOnce.Do(func() {
		defer func() {
			if r := recover(); r != nil {
				lintErr = fmt.Errorf("lint store: %v", r)
			}
		}()
		_, lintStore = test.ProdStore(root, io.Discard, nil)
	})
	if lintErr != nil {
		return nil, lintErr
	}
	if lintStore == nil {
		return nil, fmt.Errorf("lint store was not created")
	}
	return lintStore, nil
}
