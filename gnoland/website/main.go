// main.go

package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gnolang/gno/pkgs/bft/rpc/client"
	"github.com/gorilla/mux"
	"github.com/gotuna/gotuna"

	"github.com/gnolang/gno/pkgs/sdk/vm" // for error types
)

func init() {
	fmt.Println(vm.Package)
}

func main() {
	app := gotuna.App{
		ViewFiles: os.DirFS("."),
		Router:    gotuna.NewMuxRouter(),
	}

	app.Router.Handle("/", handlerHome(app))
	app.Router.Handle("/p/{pkgpath:.*}", handlerPackage(app))
	app.Router.Handle("/r/{rlmpath}", handlerRealm(app))
	app.Router.Handle("/r/{rlmpath}/{expr}", handlerRealmExpr(app))
	//app.Router.Handle("/login", handlerLogin(app)).Methods(http.MethodGet, http.MethodPost)

	fmt.Println("Running on http://localhost:8888")
	http.ListenAndServe("127.0.0.1:8888", app.Router)
}

func handlerHome(app gotuna.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.NewTemplatingEngine().
			Render(w, r, "app.html")
	})
}

func handlerPackage(app gotuna.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		pkgPath := "gno.land/p/" + vars["pkgpath"]
		fmt.Println("pkgPath:", pkgPath)
		// TODO implement query handler for fetching package files.
		return
	})
}

func handlerRealm(app gotuna.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// vars := mux.Vars(r)
		// rlmPath := "gno.land/r/" + vars["rlmpath"]
		// TODO implement query handler for fetching package files.
		// TODO unless there is a Home(), render that.
		// TODO also implement query handler for state, and show that state.
		return
	})
}

func handlerRealmExpr(app gotuna.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		rlmPath := "gno.land/r/" + vars["rlmpath"]
		expr := vars["expr"]

		qpath := "vm/qeval"
		data := []byte(fmt.Sprintf("%s\n%s", rlmPath, expr))
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
		// NOTE: HACKY.
		if strings.HasSuffix(resstr, " string)") {
			resstr2 := resstr[1 : len(resstr)-len(" string)")]
			resstr3, err := strconv.Unquote(resstr2)
			if err != nil {
				w.WriteHeader(500)
				w.Write([]byte(
					fmt.Sprintf("error unquoting result: %q", resstr2)))
				return
			}
			tmpl := app.NewTemplatingEngine()
			tmpl.Set("Contents", template.HTML(resstr3))
			tmpl.Render(w, r, "app.html")
			return
		} else {
			w.WriteHeader(200)
			w.Write([]byte(resstr))
			return
		}
	})
}

/*
func handlerLogin(app gotuna.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Login form...")
	})
}
*/

func writeError(w http.ResponseWriter, err error) {
	w.WriteHeader(500)
	w.Write([]byte(err.Error()))
}
