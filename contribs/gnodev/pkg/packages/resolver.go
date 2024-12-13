package packages

import (
	"errors"
	"fmt"
	"go/token"
	"strings"
)

var (
	ErrResolverPackageNotFound = errors.New("package not found")
	ErrResolverPackageSkip     = errors.New("package has been skip")
)

type Resolver interface {
	Name() string
	Resolve(fset *token.FileSet, path string) (*Package, error)
}

// Chain Resolver

type ChainedResolver []Resolver

func ChainResolvers(rs ...Resolver) Resolver {
	return ChainedResolver(rs)
}

func (cr ChainedResolver) Name() string {
	var name strings.Builder

	for i, r := range cr {
		rname := r.Name()
		if rname == "" {
			continue
		}

		if i > 0 {
			name.WriteRune('/')
		}
		name.WriteString(rname)
	}

	return name.String()
}

func (cr ChainedResolver) Resolve(fset *token.FileSet, path string) (*Package, error) {
	for _, resolver := range cr {
		pkg, err := resolver.Resolve(fset, path)
		if err == nil {
			return pkg, nil
		} else if errors.Is(err, ErrResolverPackageNotFound) {
			continue
		}

		return nil, fmt.Errorf("unable to resolve %q: %w", path, err)
	}

	return nil, ErrResolverPackageNotFound
}

// // Cache Resolver

// type inMemoryCacheResolver struct {
// 	subr     Resolver
// 	cacheMap map[string] /* path */ *Package
// }

// func Cache(r Resolver) Resolver {
// 	return &inMemoryCacheResolver{
// 		subr:     r,
// 		cacheMap: map[string]*Package{},
// 	}
// }

// func (r *inMemoryCacheResolver) Name() string {
// 	return "cache_" + r.subr.Name()
// }

// func (r *inMemoryCacheResolver) Resolve(fset *token.FileSet, path string) (*Package, error) {
// 	if p, ok := r.cacheMap[path]; ok {
// 		return p, nil
// 	}

// 	p, err := r.subr.Resolve(fset, path)
// 	if err != nil {
// 		return nil, err
// 	}

// 	r.cacheMap[path] = p
// 	return p, nil
// }

// // Filter Resolver

// func NoopFilterFunc(path string) bool { return false }
// func FilterAllFunc(path string) bool  { return true }

// type FilterHandler func(path string) bool

// type filterResolver struct {
// 	Resolver
// 	FilterHandler
// }

// func FilterResolver(handler FilterHandler, r Resolver) Resolver {
// 	return &filterResolver{Resolver: r, FilterHandler: handler}
// }

// func (filterResolver) Name() string {
// 	return "filter"
// }

// func (r *filterResolver) Resolve(fset *token.FileSet, path string) (*Package, error) {
// 	if r.FilterHandler(path) {
// 		return nil, ErrResolverPackageSkip
// 	}

// 	return r.Resolver.Resolve(fset, path)
// }

// Utility Resolver

// // Log Resolver

// type logResolver struct {
// 	logger *slog.Logger
// 	Resolver
// }

// func LogResolver(l *slog.Logger, r Resolver) Resolver {
// 	return &logResolver{l, r}
// }

// func (l logResolver) Resolve(fset *token.FileSet, path string) (*Package, error) {
// 	start := time.Now()
// 	pkg, err := l.Resolver.Resolve(fset, path)
// 	if err == nil {
// 		l.logger.Debug("path resolved",
// 			"resolver", l.Resolver.Name(),
// 			"path", path,
// 			"name", pkg.Name,
// 			"took", time.Since(start).String(),
// 			"location", pkg.Location)
// 	} else if errors.Is(err, ErrResolverPackageNotFound) {
// 		l.logger.Warn("path not found",
// 			"resolver", l.Resolver.Name(),
// 			"path", path,
// 			"took", time.Since(start).String())
// 	} else {
// 		l.logger.Error("unable to resolve path",
// 			"resolver", l.Resolver.Name(),
// 			"path", path,
// 			"took", time.Since(start).String(),
// 			"err", err)
// 	}

// 	return pkg, err
// }
