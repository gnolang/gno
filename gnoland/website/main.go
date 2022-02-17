// main.go

package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/pkgs/bft/rpc/client"
	"github.com/gorilla/mux"
	"github.com/gotuna/gotuna"

	"github.com/gnolang/gno/gnoland/website/static" // for static files
	"github.com/gnolang/gno/pkgs/sdk/vm"            // for error types
)

func init() {
	fmt.Println(vm.Package)
}

func main() {
	app := gotuna.App{
		ViewFiles: os.DirFS("./views/"),
		Router:    gotuna.NewMuxRouter(),
		Static:    static.EmbeddedStatic,
		// StaticPrefix: "static/",
	}

	app.Router.Handle("/", handlerHome(app))
	app.Router.Handle("/r/{rlmpath:[a-z][a-z0-9_]*}/{path:.*}", handlerRealmRender(app))
	app.Router.Handle("/files/{filepath:.+}", handlerPackageFilePath(app))
	app.Router.Handle("/static/{path:.+}", handlerStaticFile(app))

	fmt.Println("Running on http://localhost:8888")
	http.ListenAndServe("127.0.0.1:8888", app.Router)
}

func handlerHome(app gotuna.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.NewTemplatingEngine().
			Render(w, r, "app.html")
	})
}

func handlerRealmRender(app gotuna.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		rlmPath := "gno.land/r/" + vars["rlmpath"]
		path := vars["path"]

		qpath := "vm/qrender"
		data := []byte(fmt.Sprintf("%s\n%s", rlmPath, path))
		writeRequestResponse(app, w, r, qpath, data)
	})
}

func handlerPackageFilePath(app gotuna.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		filepath := vars["filepath"]
		if strings.HasPrefix(filepath, "r/") || strings.HasPrefix(filepath, "p/") {
			filepath = "gno.land/" + filepath
		} else if strings.HasPrefix(filepath, "gno.land") {
			panic("should not happen")
		} else {
			// e.g. stdlibs.
		}
		qpath := "vm/qfile"
		data := []byte(filepath)
		writeRequestResponse(app, w, r, qpath, data)
	})
}

func writeRequestResponse(app gotuna.App, w http.ResponseWriter, r *http.Request, qpath string, data []byte) {
	opts2 := client.ABCIQueryOptions{
		// Height: height, XXX
		// Prove: false, XXX
	}
	remote := "127.0.0.1:26657"
	cli := client.NewHTTP(remote, "/websocket")
	qres, err := cli.ABCIQueryWithOptions(
		qpath, data, opts2)
	if err != nil {
		writeError(w, err)
		return
	}
	if qres.Response.Error != nil {
		fmt.Printf("Log: %s\n",
			qres.Response.Log)
		writeError(w, qres.Response.Error)
		return
	}
	resdata := qres.Response.Data
	resstr := string(resdata)
	tmpl := app.NewTemplatingEngine()
	tmpl.Set("Contents", template.HTML(resstr))
	tmpl.Render(w, r, "app.html")
	return
}

func handlerStaticFile(app gotuna.App) http.Handler {

	fs := http.FS(app.Static)
	fileapp := http.StripPrefix("/static", http.FileServer(fs))
	notFound := handlerNotFound(app)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		fpath := filepath.Clean(vars["path"])
		f, err := fs.Open(fpath)
		if os.IsNotExist(err) {
			notFound.ServeHTTP(w, r)
			return
		}
		stat, err := f.Stat()
		if err != nil || stat.IsDir() {
			notFound.ServeHTTP(w, r)
			return
		}

		// TODO: ModTime doesn't work for embed?
		//w.Header().Set("ETag", fmt.Sprintf("%x", stat.ModTime().UnixNano()))
		//w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%s", "31536000"))
		fileapp.ServeHTTP(w, r)
	})
}

func handlerNotFound(app gotuna.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		app.NewTemplatingEngine().
			Set("title", app.Locale.T(app.Session.GetLocale(r), "Not found")).
			SetError("title", app.Locale.T(app.Session.GetLocale(r), "Not found")).
			Render(w, r, "404.html")
	})
}

func writeError(w http.ResponseWriter, err error) {
	w.WriteHeader(500)
	w.Write([]byte(err.Error()))
}
