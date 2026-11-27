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
			res = Result{Kind: KindInternal, Detail: clip(fmt.Sprint(r), detailLimit)}
		}
	}()

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
	env, err := lintEnvLocked(opt.RootDir)
	if err != nil {
		return Result{Kind: KindInternal, Detail: err.Error()}
	}
	_, tcErr := gnolang.TypeCheckMemPackage(mpkg, gnolang.TypeCheckOptions{
		Getter:   env.store,
		Mode:     gnolang.TCLatestRelaxed,
		Cache:    env.cache,
		ProdOnly: true,
	})
	if tcErr != nil {
		return Result{Kind: KindReject, Detail: clip(tcErr.Error(), detailLimit)}
	}
	return Result{Kind: KindOK}
}

type lintEnv struct {
	store gnolang.Store
	cache gnolang.TypeCheckCache
}

var (
	lintMu   sync.Mutex
	lintEnvs = map[string]*lintEnv{}
)

func lintEnvLocked(root string) (env *lintEnv, err error) {
	if existing, ok := lintEnvs[root]; ok {
		return existing, nil
	}
	defer func() {
		if r := recover(); r != nil {
			env = nil
			err = fmt.Errorf("lint store: %v", r)
		}
	}()
	_, s := test.ProdStore(root, io.Discard, nil)
	if s == nil {
		return nil, fmt.Errorf("lint store was not created")
	}
	env = &lintEnv{
		store: s,
		cache: gnolang.TypeCheckCache{},
	}
	lintEnvs[root] = env
	return env, nil
}
