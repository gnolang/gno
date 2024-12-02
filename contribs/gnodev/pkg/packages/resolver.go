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

var ErrResolverPackageNotFound = errors.New("package not found")

type Resolver interface {
	Name() string
	Resolve(fset *token.FileSet, path string) (*Package, error)
}

type logResolver struct {
	logger *slog.Logger
	Resolver
}

func LogResolver(l *slog.Logger, r Resolver) Resolver {
	return &logResolver{l, r}
}

func (l logResolver) Resolve(fset *token.FileSet, path string) (*Package, error) {
	start := time.Now()
	pkg, err := l.Resolver.Resolve(fset, path)
	if err == nil {
		l.logger.Info("path resolved",
			"resolver", l.Resolver.Name(),
			"took", time.Since(start).String(),
			"path", path,
			"name", pkg.Name,
			"location", pkg.Location)
	} else if errors.Is(err, ErrResolverPackageNotFound) {
		l.logger.Debug("path not found",
			"resolver", l.Resolver.Name(),
			"took", time.Since(start).String(),
			"path", path)

	} else {
		l.logger.Error("unable to resolve path",
			"resolver", l.Resolver.Name(),
			"took", time.Since(start).String(),
			"path", path,
			"err", err)
	}

	return pkg, err
}

type ChainedResolver []Resolver

func ChainResolvers(rs ...Resolver) Resolver {
	return ChainedResolver(rs)
}

func ChainWithLogger(logger *slog.Logger, rs ...Resolver) Resolver {
	loggedResolvers := make([]Resolver, len(rs))
	for i, r := range rs {
		loggedResolvers[i] = LogResolver(logger, r)
	}
	return ChainedResolver(loggedResolvers)
}

func (cr ChainedResolver) Name() string {
	var name strings.Builder

	for i, r := range cr {
		if i > 0 {
			name.WriteRune('/')
		}
		name.WriteString(r.Name())
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

type inMemoryCacheResolver struct {
	subr     Resolver
	cacheMap map[string] /* path */ *Package
}

func Cache(r Resolver) Resolver {
	return &inMemoryCacheResolver{
		subr:     r,
		cacheMap: map[string]*Package{},
	}
}

func (r *inMemoryCacheResolver) Name() string {
	return "cache_" + r.subr.Name()
}

func (r *inMemoryCacheResolver) Resolve(fset *token.FileSet, path string) (*Package, error) {
	if p, ok := r.cacheMap[path]; ok {
		return p, nil
	}

	p, err := r.subr.Resolve(fset, path)
	if err != nil {
		return nil, err
	}

	r.cacheMap[path] = p
	return p, nil
}

type SyntaxErrorHandler func(path string, filename string, serr *scanner.Error)

type SyntaxCheckerResolver struct {
	SyntaxErrorHandler
	Resolver
}

func SyntaxChecker(r Resolver, handler SyntaxErrorHandler) Resolver {
	return &SyntaxCheckerResolver{
		SyntaxErrorHandler: handler,
		Resolver:           r,
	}
}

func (SyntaxCheckerResolver) Name() string {
	return "syntax_checker"
}

func (r *SyntaxCheckerResolver) Resolve(fset *token.FileSet, path string) (*Package, error) {
	p, err := r.Resolver.Resolve(fset, path)
	if err != nil {
		return nil, err
	}

	for _, file := range p.Files {
		_, err = parser.ParseFile(fset, file.Name, file.Body, parser.AllErrors)
		if err == nil {
			continue
		}

		if el, ok := err.(scanner.ErrorList); ok {
			for _, e := range el {
				r.SyntaxErrorHandler(path, file.Name, e)
			}
		}

		return nil, fmt.Errorf("unable to parse %q: %w",
			file.Name, err)
	}

	return p, err
}
