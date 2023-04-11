package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
)

const zeroVersion = "v0.0.0"

var gnoModules []string = []string{
	"gno.land",
}

var goModules []string = []string{
	"github.com/gnolang/gno",
}

type moduleType int

const (
	invalid moduleType = iota
	gnomod
	gomod
)

func main() {
	cli := NewTMClient("127.0.0.1:26657")

	router := mux.NewRouter()
	router.Use(loggingMiddleware)
	router.HandleFunc("/{module:.+}/@v/list", list).Methods(http.MethodGet)
	router.Handle("/{module:.+}/@v/{version}.info", version(cli, false)).Methods(http.MethodGet)
	router.Handle("/{module:.+}/@latest", version(cli, true)).Methods(http.MethodGet)
	router.HandleFunc("/{module:.+}/@v/{version}.mod", mod).Methods(http.MethodGet)
	router.Handle("/{module:.+}/@v/{version}.zip", archive(cli)).Methods(http.MethodGet)
	http.ListenAndServe(":9999", router)
}

func goModuleToGno(module string) string {
	if c, ok := strings.CutPrefix(module, gnolang.GnoRealmPkgsPrefixAfter); ok {
		return path.Join(gnolang.GnoRealmPkgsPrefixBefore, c)
	}

	if c, ok := strings.CutPrefix(module, gnolang.GnoPackagePrefixAfter); ok {
		return path.Join(gnolang.GnoPackagePrefixBefore, c)
	}

	return module
}

func getModuleAndVersion(r *http.Request) (string, string, string) {
	return mux.Vars(r)["module"], goModuleToGno(mux.Vars(r)["module"]), mux.Vars(r)["version"]
}

func getModuleType(r *http.Request) moduleType {
	m := mux.Vars(r)["module"]

	for _, gom := range goModules {
		if strings.HasPrefix(m, gom) {
			return gomod
		}
	}

	for _, gnom := range gnoModules {
		if strings.HasPrefix(m, gnom) {
			return gnomod
		}
	}

	return invalid
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		logRespWriter := NewLogResponseWriter(w)
		next.ServeHTTP(logRespWriter, r)

		log.Printf(
			"duration=%s status=%d body=%s",
			time.Since(startTime).String(),
			logRespWriter.statusCode,
			logRespWriter.buf.String())
	})
}

func version(cli *TMClient, isLatest bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if getModuleType(r) == invalid {
			http.NotFound(w, r)
			return
		}

		_, m, v := getModuleAndVersion(r)

		if v != zeroVersion && !isLatest {
			http.Error(w, fmt.Sprintf("only version %q is supported", zeroVersion), http.StatusInternalServerError)
			return
		}

		// TODO improve this obtaining block hash directly
		h := sha256.New()
		err := cli.GetGnoZip(m, zeroVersion, h)
		if errors.Is(err, ErrPackageNotFound) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// TODO obtain the deployment date
		now := time.Now()

		hh := hex.EncodeToString(h.Sum(nil))

		// all versions are the latest version right now
		finalVersion := fmt.Sprintf(
			"v0.0.0-%s-%s", // if we have tagged versions in the future, fill this
			now.Format("20060102150405"),
			hh[0:12],
		)

		fmt.Printf("obtaining version %q for module %q\n", finalVersion, m)

		err = json.NewEncoder(w).Encode(map[string]string{
			"Version": finalVersion,
			"Time":    now.Format(time.RFC3339),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func mod(w http.ResponseWriter, r *http.Request) {
	// TODO implement when gno.mod is implemented
	m := mux.Vars(r)["module"]
	w.Write([]byte(fmt.Sprintf("module %s", m)))
	return
}

func list(w http.ResponseWriter, r *http.Request) {
	// TODO implement when gno.mod is implemented
	w.Write([]byte(zeroVersion))
	w.Write([]byte("\n"))
}

func archive(cli *TMClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mt := getModuleType(r)
		m, mtranslated, v := getModuleAndVersion(r)

		// TODO precompile if needed
		var err error
		switch mt {
		case gnomod:
			err = cli.GetGnoZip(m, v, w)
		case gomod:
			err = cli.GetGoZip(mtranslated, m, v, w)
		case invalid:
			http.NotFound(w, r)
			return
		default:
			http.NotFound(w, r)
			return
		}

		if errors.Is(err, ErrPackageNotFound) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}
