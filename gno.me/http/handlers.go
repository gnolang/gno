package http

import (
	"encoding/json"
	"fmt"
	gohttp "net/http"
	"strings"

	gno "github.com/gnolang/gno/gno.me/gno"
)

type app struct {
	Code      string `json:"code"`
	IsPackage bool   `json:"is_package"`
}

type appCall struct {
	Name      string `json:"name"`
	IsPackage bool   `json:"is_package"`
	Func      string `json:"func"`
	Args      string `json:"args"`
}

func createApp(resp gohttp.ResponseWriter, req *gohttp.Request) {
	enableCors(&resp)

	var gnoApp app
	dec := json.NewDecoder(req.Body)
	defer req.Body.Close()
	if err := dec.Decode(&gnoApp); err != nil {
		gohttp.Error(resp, err.Error(), gohttp.StatusBadRequest)
		return
	}

	if err := vm.Create(req.Context(), gnoApp.Code, gnoApp.IsPackage); err != nil {
		fmt.Println(err, "error adding package")
		gohttp.Error(resp, err.Error(), gohttp.StatusInternalServerError)
		return
	}

	resp.WriteHeader(gohttp.StatusCreated)
}

func callApp(resp gohttp.ResponseWriter, req *gohttp.Request) {
	enableCors(&resp)

	var call appCall
	dec := json.NewDecoder(req.Body)
	defer req.Body.Close()
	if err := dec.Decode(&call); err != nil {
		gohttp.Error(resp, err.Error(), gohttp.StatusBadRequest)
		return
	}

	var args []string
	if len(call.Args) > 0 {
		args = strings.Split(call.Args, ",")
	}

	res, _, err := vm.Call(req.Context(), call.Name, call.IsPackage, call.Func, args...)
	if err != nil {
		gohttp.Error(resp, err.Error(), gohttp.StatusInternalServerError)
		return
	}

	resp.WriteHeader(gohttp.StatusOK)
	resp.Write([]byte(res))
}

func run(resp gohttp.ResponseWriter, req *gohttp.Request) {
	enableCors(&resp)

	var gnoApp app
	dec := json.NewDecoder(req.Body)
	defer req.Body.Close()
	if err := dec.Decode(&gnoApp); err != nil {
		gohttp.Error(resp, err.Error(), gohttp.StatusBadRequest)
		return
	}

	res, err := vm.Run(req.Context(), gnoApp.Code)
	if err != nil {
		gohttp.Error(resp, err.Error(), gohttp.StatusInternalServerError)
		return
	}

	resp.WriteHeader(gohttp.StatusOK)
	resp.Write([]byte(res))
}

// TODO: this should call qrender so it doesn't change state accidentally,
// especially when interacting with remote apps.
func renderApp(resp gohttp.ResponseWriter, req *gohttp.Request) {
	enableCors(&resp)

	path := strings.Trim(req.URL.Path, "/")
	if path == "favicon.ico" {
		return
	}

	isPackage := strings.HasPrefix(path, gno.PkgPrefix)
	renderPathIdx := strings.LastIndex(path, ":")
	var renderPath string
	if renderPathIdx != -1 {
		if renderPathIdx != len(path)-1 {
			renderPath = path[renderPathIdx+1:]
		}
		path = path[:renderPathIdx]
	}

	var appName string
	parts := strings.Split(path, "/")
	appName = parts[len(parts)-1]
	res, _, err := vm.Call(req.Context(), appName, isPackage, "Render", renderPath)
	if err != nil {
		gohttp.Error(resp, err.Error(), gohttp.StatusInternalServerError)
		return
	}

	// Strip out gno type information and characters that are not nice when
	// rendered in a browser.
	res = res[2 : strings.LastIndex(res, "string")-2]
	res = strings.ReplaceAll(res, "\\n", "")
	res = strings.ReplaceAll(res, "\\t", "")
	res = strings.ReplaceAll(res, "\\\"", "\"")

	resp.Write([]byte(res))
}

func enableCors(w *gohttp.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
}
