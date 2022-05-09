// main.go

package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/bft/rpc/client"
	"github.com/gnolang/gno/pkgs/std"
	"github.com/gorilla/mux"
	"github.com/gotuna/gotuna"

	"github.com/gnolang/gno/gnoland/website/static" // for static files
	"github.com/gnolang/gno/pkgs/sdk/vm"            // for error types
)

var flags struct {
	bindAddr string
}

func init() {
	flag.StringVar(&flags.bindAddr, "bind", "127.0.0.1:8888", "server listening address")
}

func main() {
	flag.Parse()

	app := gotuna.App{
		ViewFiles: os.DirFS("./views/"),
		Router:    gotuna.NewMuxRouter(),
		Static:    static.EmbeddedStatic,
		// StaticPrefix: "static/",
	}

	app.Router.Handle("/", handlerHome(app))
	app.Router.Handle("/faucet", handlerFaucet(app))
	app.Router.Handle("/r/boards:gnolang/6", handlerRedirect(app))
	// NOTE: see rePathPart.
	app.Router.Handle("/r/{rlmname:[a-z][a-z0-9_]*}", handlerRealmMain(app))
	app.Router.Handle("/r/{rlmname:[a-z][a-z0-9_]*}:{querystr:.*}", handlerRealmRender(app))
	app.Router.Handle("/r/{rlmname:[a-z][a-z0-9_]*}/{filename:.*}", handlerRealmFile(app))
	app.Router.Handle("/p/{filepath:.*}", handlerPackageFile(app))
	app.Router.Handle("/static/{path:.+}", handlerStaticFile(app))
	app.Router.Handle("/favicon.ico", handlerFavicon(app))

	fmt.Printf("Running on http://%s\n", flags.bindAddr)
	err := http.ListenAndServe(flags.bindAddr, app.Router)
	if err != nil {
		fmt.Fprintf(os.Stderr, "HTTP server stopped with error: %+v\n", err)
	}
}

func handlerHome(app gotuna.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.NewTemplatingEngine().
			Render(w, r, "home.html", "header.html")
	})
}

func handlerFaucet(app gotuna.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.NewTemplatingEngine().
			Render(w, r, "faucet.html", "header.html")
	})
}

// XXX temporary.
func handlerRedirect(app gotuna.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/r/boards:gnolang/3", http.StatusFound)
		app.NewTemplatingEngine().
			Render(w, r, "home.html", "header.html")
	})
}

func handlerRealmMain(app gotuna.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		rlmname := vars["rlmname"]
		rlmpath := "gno.land/r/" + rlmname
		query := r.URL.Query()
		if query.Has("help") {
			// Render function helper.
			funcName := query.Get("__func")
			qpath := "vm/qfuncs"
			data := []byte(rlmpath)
			res, err := makeRequest(qpath, data)
			if err != nil {
				writeError(w, err)
				return
			}
			var fsigs vm.FunctionSignatures
			amino.MustUnmarshalJSON(res, &fsigs)
			// Fill fsigs with query parameters.
			for i := range fsigs {
				fsig := &(fsigs[i])
				for j := range fsig.Params {
					param := &(fsig.Params[j])
					value := query.Get(param.Name)
					param.Value = value
				}
			}
			// Render template.
			tmpl := app.NewTemplatingEngine()
			tmpl.Set("FuncName", funcName)
			tmpl.Set("RealmPath", rlmpath)
			tmpl.Set("DirPath", pathOf(rlmpath))
			tmpl.Set("FunctionSignatures", fsigs)
			tmpl.Render(w, r, "realm_help.html", "header.html")
		} else {
			// Ensure realm exists. TODO optimize.
			qpath := "vm/qfile"
			data := []byte(rlmpath)
			_, err := makeRequest(qpath, data)
			if err != nil {
				writeError(w, errors.New("error querying realm package"))
				return
			}
			// Render blank query path, /r/REALM:.
			handleRealmRender(app, w, r)
		}
	})
}

type pathLink struct {
	URL  string
	Text string
}

func handlerRealmRender(app gotuna.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleRealmRender(app, w, r)
	})
}

func handleRealmRender(app gotuna.App, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rlmname := vars["rlmname"]
	rlmpath := "gno.land/r/" + rlmname
	querystr := vars["querystr"]
	if r.URL.Path == "/r/"+rlmname+":" {
		// Redirect to /r/REALM if querypath is empty.
		http.Redirect(w, r, "/r/"+rlmname, http.StatusFound)
		return
	}
	qpath := "vm/qrender"
	data := []byte(fmt.Sprintf("%s\n%s", rlmpath, querystr))
	res, err := makeRequest(qpath, data)
	if err != nil {
		// XXX hack
		if strings.Contains(err.Error(), "Render not declared") {
			res = []byte("realm package has no Render() function")
		} else {
			writeError(w, err)
			return
		}
	}
	// linkify querystr.
	queryParts := strings.Split(string(querystr), "/")
	pathLinks := []pathLink{}
	for i, part := range queryParts {
		pathLinks = append(pathLinks, pathLink{
			URL:  "/r/" + rlmname + ":" + strings.Join(queryParts[:i+1], "/"),
			Text: part,
		})
	}
	// Render template.
	tmpl := app.NewTemplatingEngine()

	tmpl.Set("RealmName", rlmname)
	tmpl.Set("RealmPath", rlmpath)
	tmpl.Set("Query", string(querystr))
	tmpl.Set("PathLinks", pathLinks)
	tmpl.Set("Contents", string(res))
	tmpl.Render(w, r, "realm_render.html", "header.html")
}

func handlerRealmFile(app gotuna.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		diruri := "gno.land/r/" + vars["rlmname"]
		filename := vars["filename"]
		renderPackageFile(app, w, r, diruri, filename)
	})
}

func handlerPackageFile(app gotuna.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		pkgpath := "gno.land/p/" + vars["filepath"]
		diruri, filename := std.SplitFilepath(pkgpath)
		if filename == "" && diruri == pkgpath {
			// redirect to diruri + "/"
			http.Redirect(w, r, "/p/"+vars["filepath"]+"/", http.StatusFound)
			return
		}
		renderPackageFile(app, w, r, diruri, filename)
	})
}

func renderPackageFile(app gotuna.App, w http.ResponseWriter, r *http.Request, diruri string, filename string) {
	if filename == "" {
		// Request is for a folder.
		qpath := "vm/qfile"
		data := []byte(diruri)
		res, err := makeRequest(qpath, data)
		if err != nil {
			writeError(w, err)
			return
		}
		files := strings.Split(string(res), "\n")
		// Render template.
		tmpl := app.NewTemplatingEngine()
		tmpl.Set("DirURI", diruri)
		tmpl.Set("DirPath", pathOf(diruri))
		tmpl.Set("Files", files)
		tmpl.Render(w, r, "package_dir.html", "header.html")
	} else {
		// Request is for a file.
		filepath := diruri + "/" + filename
		qpath := "vm/qfile"
		data := []byte(filepath)
		res, err := makeRequest(qpath, data)
		if err != nil {
			writeError(w, err)
			return
		}
		// Render template.
		tmpl := app.NewTemplatingEngine()
		tmpl.Set("DirURI", diruri)
		tmpl.Set("DirPath", pathOf(diruri))
		tmpl.Set("FileName", filename)
		tmpl.Set("FileContents", string(res))
		tmpl.Render(w, r, "package_file.html", "header.html")
	}
}

func makeRequest(qpath string, data []byte) (res []byte, err error) {
	opts2 := client.ABCIQueryOptions{
		// Height: height, XXX
		// Prove: false, XXX
	}
	remote := "127.0.0.1:26657"
	cli := client.NewHTTP(remote, "/websocket")
	qres, err := cli.ABCIQueryWithOptions(
		qpath, data, opts2)
	if err != nil {
		return nil, err
	}
	if qres.Response.Error != nil {
		fmt.Printf("Log: %s\n",
			qres.Response.Log)
		return nil, qres.Response.Error
	}
	return qres.Response.Data, nil
}

func handlerStaticFile(app gotuna.App) http.Handler {
	fs := http.FS(app.Static)
	fileapp := http.StripPrefix("/static", http.FileServer(fs))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		fpath := filepath.Clean(vars["path"])
		f, err := fs.Open(fpath)
		if os.IsNotExist(err) {
			handleNotFound(app, fpath, w, r)
			return
		}
		stat, err := f.Stat()
		if err != nil || stat.IsDir() {
			handleNotFound(app, fpath, w, r)
			return
		}

		// TODO: ModTime doesn't work for embed?
		// w.Header().Set("ETag", fmt.Sprintf("%x", stat.ModTime().UnixNano()))
		// w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%s", "31536000"))
		fileapp.ServeHTTP(w, r)
	})
}

func handlerFavicon(app gotuna.App) http.Handler {
	fs := http.FS(app.Static)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fpath := "img/favicon.ico"
		f, err := fs.Open(fpath)
		if os.IsNotExist(err) {
			handleNotFound(app, fpath, w, r)
			return
		}
		w.Header().Set("Content-Type", "image/x-icon")
		w.Header().Set("Cache-Control", "public, max-age=604800") // 7d
		io.Copy(w, f)
	})
}

func handleNotFound(app gotuna.App, path string, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	app.NewTemplatingEngine().
		Set("title", "Not found").
		Set("path", path).
		Render(w, r, "404.html", "header.html")
}

func writeError(w http.ResponseWriter, err error) {
	w.WriteHeader(500)
	w.Write([]byte(err.Error()))
}

func pathOf(diruri string) string {
	parts := strings.Split(diruri, "/")
	if parts[0] == "gno.land" {
		return "/" + strings.Join(parts[1:], "/")
	} else {
		panic(fmt.Sprintf("invalid dir-URI %q", diruri))
	}
}
