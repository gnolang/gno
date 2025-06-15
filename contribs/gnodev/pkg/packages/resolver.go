package packages

import (
	"errors"
	"fmt"
	"go/parser"
	"go/scanner"
	"go/token"
	"log/slog"
	"strings"
	"time"
)

var (
	ErrResolverPackageNotFound = errors.New("package not found")
	ErrResolverPackageSkip     = errors.New("package has been skip")
)

type Resolver interface {
	Name() string
	Resolve(fset *token.FileSet, path string) (*Package, error)
}

type NoopResolver struct{}

func (NoopResolver) Name() string { return "" }
func (NoopResolver) Resolve(fset *token.FileSet, path string) (*Package, error) {
	return nil, ErrResolverPackageNotFound
}

// Chain Resolver

type ChainedResolver []Resolver

func ChainResolvers(rs ...Resolver) Resolver {
	switch len(rs) {
	case 0:
		return &NoopResolver{}
	case 1:
		return rs[0]
	default:
		return ChainedResolver(rs)
	}
}

func (cr ChainedResolver) Name() string {
	names := make([]string, 0, len(cr))
	for _, r := range cr {
		rname := r.Name()
		if rname == "" {
			continue
		}

		names = append(names, rname)
	}

	return strings.Join(names, "/")
}

func (cr ChainedResolver) Resolve(fset *token.FileSet, path string) (*Package, error) {
	for _, resolver := range cr {
		pkg, err := resolver.Resolve(fset, path)
		if err == nil {
			return pkg, nil
		} else if errors.Is(err, ErrResolverPackageNotFound) {
			continue
		}

		return nil, fmt.Errorf("resolver %q error: %w", resolver.Name(), err)
	}

	return nil, ErrResolverPackageNotFound
}

type MiddlewareHandler func(fset *token.FileSet, path string, next Resolver) (*Package, error)

type middlewareResolver struct {
	Handler MiddlewareHandler
	Next    Resolver
}

func MiddlewareResolver(r Resolver, handlers ...MiddlewareHandler) Resolver {
	// Start with the final resolver
	start := r

	// Wrap each handler around the previous one
	for _, handler := range handlers {
		start = &middlewareResolver{
			Next:    start,
			Handler: handler,
		}
	}

	return start
}

func (r middlewareResolver) Name() string {
	return r.Next.Name()
}

func (r *middlewareResolver) Resolve(fset *token.FileSet, path string) (*Package, error) {
	if r.Handler != nil {
		return r.Handler(fset, path, r.Next)
	}

	return r.Next.Resolve(fset, path)
}

// LogMiddleware creates a logging middleware handler.
func LogMiddleware(logger *slog.Logger) MiddlewareHandler {
	return func(fset *token.FileSet, path string, next Resolver) (*Package, error) {
		start := time.Now()
		pkg, err := next.Resolve(fset, path)
		switch {
		case err == nil:
			logger.Debug("path resolved",
				"resolver", next.Name(),
				"path", path,
				"name", pkg.Name,
				"took", time.Since(start).String(),
				"location", pkg.Location,
			)
		case errors.Is(err, ErrResolverPackageSkip):
			logger.Debug(err.Error(),
				"resolver", next.Name(),
				"path", path,
				"took", time.Since(start).String(),
			)

		case errors.Is(err, ErrResolverPackageNotFound):
			logger.Warn(err.Error(),
				"resolver", next.Name(),
				"path", path,
				"took", time.Since(start).String())

		default:
			logger.Error(err.Error(),
				"resolver", next.Name(),
				"path", path,
				"took", time.Since(start).String())
		}

		return pkg, err
	}
}

type ShouldCacheFunc func(pkg *Package) bool

func CacheAll(_ *Package) bool { return true }

// CacheMiddleware creates a caching middleware handler.
func CacheMiddleware(shouldCache ShouldCacheFunc) MiddlewareHandler {
	cacheMap := make(map[string]*Package)
	return func(fset *token.FileSet, path string, next Resolver) (*Package, error) {
		if pkg, ok := cacheMap[path]; ok {
			return pkg, nil
		}

		pkg, err := next.Resolve(fset, path)
		if pkg != nil && shouldCache(pkg) {
			cacheMap[path] = pkg
		}

		return pkg, err
	}
}

// FilterPathHandler defines the function signature for filter handlers.
type FilterPathHandler func(path string) bool

func FilterPathMiddleware(name string, filter FilterPathHandler) MiddlewareHandler {
	return func(fset *token.FileSet, path string, next Resolver) (*Package, error) {
		if filter(path) {
			return nil, fmt.Errorf("filter %q: %w", name, ErrResolverPackageSkip)
		}

		return next.Resolve(fset, path)
	}
}

var FilterStdlibs = FilterPathMiddleware("stdlibs", isStdPath)

func isStdPath(path string) bool {
	if i := strings.IndexRune(path, '/'); i > 0 {
		if j := strings.IndexRune(path[:i], '.'); j >= 0 {
			return false
		}
	}

	return true
}

// PackageCheckerMiddleware creates a middleware handler for post-processing syntax.
func PackageCheckerMiddleware(logger *slog.Logger) MiddlewareHandler {
	return func(fset *token.FileSet, path string, next Resolver) (*Package, error) {
		// First, resolve the package using the next resolver in the chain.
		pkg, err := next.Resolve(fset, path)
		if err != nil {
			return nil, err
		}

		// Post-process each file in the package.
		for _, file := range pkg.Files {
			fname := file.Name
			if !isGnoFile(fname) {
				continue
			}

			logger.Debug("checking syntax", "path", path, "filename", fname)
			_, err := parser.ParseFile(fset, file.Name, file.Body, parser.AllErrors)
			if err == nil {
				continue
			}

			if el, ok := err.(scanner.ErrorList); ok {
				for _, e := range el {
					logger.Error("syntax error",
						"path", path,
						"filename", fname,
						"err", e.Error(),
					)
				}
			}

			return nil, fmt.Errorf("file %q have error(s)", file.Name)
		}

		return pkg, nil
	}
}
